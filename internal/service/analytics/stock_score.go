package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
)

const barsPeriod = 24 * time.Hour * 365

func (s *Service) FetchStockScores(
	ctx context.Context,
	stockIDs []model.StockID,
) (map[model.StockID]float64, error) {
	scores := map[model.StockID]float64{}

	for _, stockID := range stockIDs {
		score, err := s.FetchStockScore(ctx, stockID)
		if err != nil {
			return nil, fmt.Errorf("fetch stock score: %w", err)
		}

		scores[stockID] = score
	}

	return scores, nil
}

func (s *Service) FetchStockScore(
	ctx context.Context,
	stockID model.StockID,
) (float64, error) {
	// Asset agent settings
	const (
		lookback = 3
		window   = 3
	)

	now := time.Now()

	bars, err := s.market.FetchBars(
		ctx,
		stockID,
		now.Add(-barsPeriod),
		now,
	)
	if err != nil {
		return 0.0, fmt.Errorf("failed to fetch market data: %w", err)
	}

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
