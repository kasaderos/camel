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

	repo   AgentRepository
	market MarketService
}

func NewAgent(repo AgentRepository, market MarketService, opts ...Option) (*Agent, error) {
	a := &Agent{
		repo:   repo,
		market: market,
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (a *Agent) Initalize(ctx context.Context, id string) error {
	agent, err := a.repo.FetchInfo(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch agent: %w", err)
	}

	a.AssetAgent = agent

	return nil
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

func (a *Agent) Withdraw(
	ctx context.Context,
	agentID string,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	err := a.repo.Withdraw(ctx, agentID, amount)
	if err != nil {
		return fmt.Errorf("service failed to withdraw: %w", err)
	}

	return nil
}

func (a *Agent) Deposit(
	ctx context.Context,
	agentID string,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	err := a.repo.Deposit(ctx, agentID, amount)
	if err != nil {
		return fmt.Errorf("service failed to deposit: %w", err)
	}

	return nil
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
		now.Add(-24*time.Hour*3*window),
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

func calcEMAChange(bars []model.Bar, window, lookback int) decimal.Decimal {
	if len(bars) < window {
		return decimal.NewFromFloat(0.0)
	}

	prices := extractClosePrices(bars)

	emaValues := ema(prices, window)
	changeValue := priceChange(emaValues, lookback)

	return decimal.NewFromFloat(changeValue)
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
