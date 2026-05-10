package asset

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type AssetAgentService struct {
	repo   AgentRepository
	market MarketService
}

func New(repo AgentRepository, market MarketService) *AssetAgentService {
	return &AssetAgentService{
		repo:   repo,
		market: market,
	}
}

func (s *AssetAgentService) CreateAgent(
	ctx context.Context,
	agent model.AssetAgent,
) error {
	return s.repo.CreateAgent(ctx, agent)
}

func (s *AssetAgentService) FetchInfo(
	ctx context.Context,
	agentID string,
) (model.AssetAgent, error) {
	if agentID == "" {
		return model.AssetAgent{}, errors.New("agent ID is required")
	}

	return s.repo.FetchInfo(ctx, agentID)
}

func (s *AssetAgentService) Withdraw(
	ctx context.Context,
	agentID string,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	err := s.repo.Withdraw(ctx, agentID, amount)
	if err != nil {
		return fmt.Errorf("service failed to withdraw: %w", err)
	}

	return nil
}

func (s *AssetAgentService) Deposit(
	ctx context.Context,
	agentID string,
	amount float64,
) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}

	err := s.repo.Deposit(ctx, agentID, amount)
	if err != nil {
		return fmt.Errorf("service failed to deposit: %w", err)
	}

	return nil
}

// UpdateState allows modifying the agent's state metadata
func (s *AssetAgentService) UpdateState(
	ctx context.Context,
	agentID string,
) (map[string]string, error) {
	if agentID == "" {
		return nil, errors.New("agent ID is required")
	}

	agent, err := s.FetchInfo(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent: %w", err)
	}

	state, err := s.getState(ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to compute state: %w", err)
	}

	err = s.repo.UpdateState(ctx, agentID, state)
	if err != nil {
		return nil, fmt.Errorf("update state: %w", err)
	}

	return state, nil
}

func (s *AssetAgentService) getState(
	ctx context.Context,
	agent model.AssetAgent,
) (map[string]string, error) {
	// Asset agent settings
	const (
		lookback = 5
		window   = 20
	)

	now := time.Now()

	bars, err := s.market.FetchBars(
		ctx,
		agent.AssetID,
		now.Add(-24*time.Hour*3*window),
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}

	lastBar := bars[len(bars)-1]

	return map[string]string{
		model.StateDate:            lastBar.Timestamp.Format(time.DateOnly),
		model.EMA20Lookback5Change: calcEMAChange(bars, window, lookback),
	}, nil
}

func calcEMAChange(bars []model.Bar, window, lookback int) string {
	if len(bars) < window {
		return "undefined"
	}

	prices := extractClosePrices(bars)

	emaValues := ema(prices, window)
	changeValue := priceChange(emaValues, lookback)

	return fmt.Sprintf("%.5f", changeValue)
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
