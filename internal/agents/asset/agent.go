package asset

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
)

type Agent struct {
	model.AssetAgent

	repo     AgentRepository
	market   MarketService
	exchange Exchanger
}

func NewAgent(agent model.AssetAgent, repo AgentRepository, market MarketService, exchange Exchanger) *Agent {
	return &Agent{
		AssetAgent: agent,
		repo:       repo,
		market:     market,
		exchange:   exchange,
	}
}

func (a *Agent) CreateAgent(
	ctx context.Context,
	agent *model.AssetAgent,
) error {
	return a.repo.CreateAgent(ctx, agent)
}

func (a *Agent) FetchInfo(ctx context.Context) model.AssetAgent {
	return a.AssetAgent
}

func (a *Agent) FetchState(ctx context.Context) model.State {
	return a.State
}

func (a *Agent) FetchPrice(ctx context.Context) (float64, error) {
	return a.exchange.FetchPrice(ctx, a.AssetID)
}

func (a *Agent) FetchTotalSum(ctx context.Context) (float64, error) {
	price, err := a.FetchPrice(ctx)
	if err != nil {
		return 0.0, fmt.Errorf("fetch price for %s: %w", a.AssetID, err)
	}

	agentTotalSum := a.Cash + a.AssetQty*price

	return agentTotalSum, nil
}

func (a *Agent) BuyAsset(
	ctx context.Context,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	if a.Cash < amount {
		return fmt.Errorf("insufficient cash: have %.2f, need %.2f", a.Cash, amount)
	}

	// Get current price to calculate quantity
	currentPrice, err := a.exchange.FetchPrice(ctx, a.AssetID)
	if err != nil {
		return fmt.Errorf("failed to fetch current price: %w", err)
	}

	qty := amount / currentPrice

	// Create market buy order
	order, err := a.exchange.CreateMarketOrder(ctx, a.AssetID, qty, model.OrderSideBuy)
	if err != nil {
		return fmt.Errorf("failed to create buy order: %w", err)
	}

	// Use actual order price and quantity to calculate exact cost
	actualCost := order.Sum()

	// Update agent's cash and asset quantity using actual order values
	err = a.repo.Withdraw(ctx, a.ID, actualCost)
	if err != nil {
		return fmt.Errorf("failed to withdraw cash: %w", err)
	}

	a.Cash -= actualCost
	a.AssetQty += order.Qty

	return nil
}

func (a *Agent) SellAsset(
	ctx context.Context,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	if a.AssetQty < amount {
		return fmt.Errorf("insufficient asset quantity: have %.6f, need %.6f", a.AssetQty, amount)
	}

	// Create market sell order
	order, err := a.exchange.CreateMarketOrder(ctx, a.AssetID, amount, model.OrderSideSell)
	if err != nil {
		return fmt.Errorf("failed to create sell order: %w", err)
	}

	// Use actual order price and quantity to calculate exact proceeds
	actualProceeds := order.Sum()

	// Update agent's cash and asset quantity using actual order values
	err = a.repo.Deposit(ctx, a.ID, actualProceeds)
	if err != nil {
		return fmt.Errorf("failed to deposit cash: %w", err)
	}

	a.Cash += actualProceeds
	a.AssetQty -= order.Qty

	return nil
}

func (a *Agent) ClosePosition(ctx context.Context) (model.AssetAgent, error) {
	err := a.SellAsset(ctx, a.AssetQty)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("agent sell asset: %w", err)
	}

	return a.Withdraw(ctx, a.Cash)
}

func (a *Agent) WithdrawWithSell(
	ctx context.Context,
	sum float64,
) (model.AssetAgent, error) {
	if sum < a.Cash {
		return a.Withdraw(ctx, sum)
	}

	price, err := a.FetchPrice(ctx)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("fetch price: %w", err)
	}

	sellQty := math.Floor(math.Abs(sum-a.Cash)/price) + 1.0

	err = a.SellAsset(ctx, sellQty)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("agent sell asset: %w", err)
	}

	return a.Withdraw(ctx, sum)
}

func (a *Agent) Withdraw(
	ctx context.Context,
	sum float64,
) (model.AssetAgent, error) {
	if sum <= 0 {
		return model.AssetAgent{}, model.ErrInvalidAmount
	}

	err := a.repo.Withdraw(ctx, a.ID, sum)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("service failed to withdraw: %w", err)
	}

	a.Cash -= sum

	return a.AssetAgent, nil
}

func (a *Agent) DepositWithBuy(
	ctx context.Context,
	sum float64,
) (model.AssetAgent, error) {
	if sum <= 0 {
		return model.AssetAgent{}, model.ErrInvalidAmount
	}

	// First deposit the cash
	agentInfo, err := a.Deposit(ctx, sum)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("deposit: %w", err)
	}

	// Then buy assets with all available cash
	if agentInfo.Cash > 0 {
		err = a.BuyAsset(ctx, agentInfo.Cash)
		if err != nil {
			return model.AssetAgent{}, fmt.Errorf("buy asset: %w", err)
		}
	}

	return a.AssetAgent, nil
}

func (a *Agent) Deposit(
	ctx context.Context,
	sum float64,
) (model.AssetAgent, error) {
	if sum <= 0 {
		return model.AssetAgent{}, model.ErrInvalidAmount
	}

	err := a.repo.Deposit(ctx, a.ID, sum)
	if err != nil {
		return model.AssetAgent{}, fmt.Errorf("service failed to deposit: %w", err)
	}

	a.Cash += sum

	return a.AssetAgent, nil
}

// UpdateState allows modifying the agent's state metadata
func (a *Agent) UpdateState(ctx context.Context) error {
	state, err := a.getState(ctx)
	if err != nil {
		return fmt.Errorf("failed to compute state: %w", err)
	}

	err = a.repo.UpdateState(ctx, a.ID, state)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	a.State = state

	return nil
}

func (a *Agent) getState(ctx context.Context) (model.State, error) {
	// Asset agent settings
	const (
		lookback = 5
		window   = 20
	)

	now := time.Now()

	bars, err := a.market.FetchBars(
		ctx,
		a.AssetID,
		now.Add(-24*time.Hour*365),
		now,
	)
	if err != nil {
		return model.State{}, fmt.Errorf("failed to fetch market data: %w", err)
	}

	lastBar := bars[len(bars)-1]

	st := model.State{}
	st.SetDate(lastBar.Timestamp)
	st.SetEmaChange(calcEMAChange(bars, window, lookback))

	a.State = st

	return st, nil
}

func calcEMAChange(bars []model.Bar, window, lookback int) float64 {
	if len(bars) < window {
		return 0.0
	}

	prices := extractClosePrices(bars)

	emaValues := ema(prices, window)
	changeValue := priceChange(emaValues, lookback)

	return decimal.NewFromFloat(changeValue).Round(3).InexactFloat64()
}

func extractClosePrices(bars []model.Bar) []float64 {
	prices := make([]float64, len(bars))
	for i, bar := range bars {
		prices[i] = bar.Close
	}

	return prices
}

func ema(prices []float64, n int) []float64 {
	if len(prices) == 0 || n <= 0 {
		return nil
	}

	result := make([]float64, len(prices))

	smoothing := 2.0
	multiplier := smoothing / (1.0 + float64(n))

	result[0] = prices[0]

	for i := 1; i < len(prices); i++ {
		// EMA_today = (Price_today * Multiplier) + (EMA_yesterday * (1 - Multiplier))
		todayPrice := prices[i]
		yesterdayEMA := result[i-1]

		emaValue := (todayPrice * multiplier) + (yesterdayEMA * (1.0 - multiplier))

		result[i] = math.Round(emaValue*100) / 100
	}

	return result
}

func priceChange(ema []float64, lookback int) float64 {
	n := len(ema)

	if n <= lookback || lookback <= 0 {
		return 0.0
	}

	idxStart := n - 1 - lookback
	interval := ema[idxStart:]

	if len(interval) > 2 {
		for i := 1; i < len(interval)-1; i++ {
			current := interval[i]
			prev := interval[i-1]
			next := interval[i+1]

			if (current > prev && current > next) || (current < prev && current < next) {
				return 0.0
			}
		}
	}

	startVal := interval[0]
	endVal := interval[len(interval)-1]

	if startVal == 0 {
		return 0.0
	}

	change := (endVal / startVal) - 1.0

	return change
}
