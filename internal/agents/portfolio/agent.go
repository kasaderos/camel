package portfolio

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/alpaca"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type Agent struct {
	model.PortfolioAgent

	repo     AgentRepository
	market   MarketService
	exchange Exchanger
}

func NewAgent(
	agent model.PortfolioAgent,
	repo AgentRepository,
	market MarketService,
	exchange Exchanger,
) *Agent {
	return &Agent{
		PortfolioAgent: agent,
		repo:           repo,
		market:         market,
		exchange:       exchange,
	}
}

func (a *Agent) ClosePosition(ctx context.Context) error {
	if a.AssetQty <= 1e-3 {
		return nil
	}

	_, err := a.exchange.CreateMarketOrder(
		ctx,
		a.AssetID,
		a.AssetQty,
		model.OrderSideSell,
	)
	if err != nil {
		return fmt.Errorf("create market order: %w", err)
	}

	a.AssetQty = 0.0

	err = a.repo.UpdatePortfolioAgent(ctx, a.PortfolioAgent)
	if err != nil {
		return fmt.Errorf("update portfolio agent: %w", err)
	}

	return nil
}

func (a *Agent) AdjustTargetSum(
	ctx context.Context,
	targetSum float64,
) (*model.Order, error) {
	currentPrice, err := a.exchange.FetchPrice(ctx, a.AssetID)
	if err != nil {
		return nil, fmt.Errorf("fetch current price: %w", err)
	}

	currentSum := a.AssetQty * currentPrice
	slog.Info(
		"adjusting target sum",
		"asset_id", a.AssetID,
		"current_value", currentSum,
		"target_sum", targetSum,
	)

	// Skip if already within one unit of the target in either direction.
	if currentSum > targetSum && currentSum-currentPrice < targetSum {
		slog.Info("no adjustment needed: over by less than one unit")
		return nil, nil
	}

	if math.Abs(currentSum-targetSum) < 0.01 {
		slog.Info("no adjustment needed: sum diff < 0.01$")
		return nil, nil
	}

	targetQty := targetSum / currentPrice
	deltaQty := targetQty - a.AssetQty

	var side string
	if deltaQty > 0 {
		side = model.OrderSideBuy
	} else {
		side = model.OrderSideSell
	}

	assetQty := math.Ceil(math.Abs(deltaQty))

	slog.Info(
		"create order",
		"asset_id", a.AssetID,
		"price", currentPrice,
		"qty", assetQty,
		"side", side,
	)

	order, err := a.exchange.CreateMarketOrder(
		ctx,
		a.AssetID,
		assetQty,
		side,
	)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	agentQtyChange := order.Qty
	if order.Side == model.OrderSideSell {
		agentQtyChange = -order.Qty
	}

	a.PortfolioAgent.AssetQty += agentQtyChange

	err = a.repo.UpdatePortfolioAgent(ctx, a.PortfolioAgent)
	if err != nil {
		return nil, fmt.Errorf("update portfolio agent: %w", err)
	}

	return order, nil
}

func (a *Agent) FetchScore(ctx context.Context) (float64, error) {
	// Asset agent settings
	const (
		lookback = 3
		window   = 3
	)

	now := time.Now()

	mBars, err := a.market.FetchBars(
		ctx,
		a.AssetID,
		now.Add(-24*time.Hour*365),
		now,
	)
	if err != nil {
		return 0.0, fmt.Errorf("failed to fetch market data: %w", err)
	}

	bars := lo.Map(mBars, func(b alpaca.Bar, _ int) model.Bar {
		return model.Bar{
			Timestamp: b.Timestamp,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    b.Volume,
		}
	})

	score := calcEMAChange(bars, window, lookback)

	return score, nil
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
