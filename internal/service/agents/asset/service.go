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
) error {
	if agentID == "" {
		return errors.New("agent ID is required")
	}

	agent, err := s.FetchInfo(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to fetch agent: %w", err)
	}

	state, err := s.getState(ctx, agent)
	if err != nil {
		return fmt.Errorf("failed to compute state: %w", err)
	}

	return s.repo.UpdateState(ctx, agentID, state)
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
		now.Add(-24*time.Hour*window),
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}

	// Ensure we have data for today
	lastBar := bars[len(bars)-1]
	if lastBar.Timestamp.Format(time.DateOnly) != now.Format(time.DateOnly) {
		return nil, fmt.Errorf("no market data for today: %s %s", lastBar.Timestamp, now)
	}

	return map[string]string{
		model.StateDate:            now.Format(time.RFC3339),
		model.EMA20Lookback5Change: calcEMAChange(bars, window, lookback),
	}, nil
}

func calcEMAChange(bars []model.Bar, window, lookback int) string {
	if len(bars) < window+lookback {
		return "0.0"
	}

	prices := extractClosePrices(bars)

	emaValues := ema(prices, window)
	changeValue := priceChange(emaValues, lookback)

	return fmt.Sprintf("%.3f", changeValue)
}

func extractClosePrices(bars []model.Bar) []float64 {
	prices := make([]float64, len(bars))
	for i, bar := range bars {
		prices[i] = bar.Close
	}

	return prices
}

func ema(prices []float64, window int) []float64 {
	if len(prices) == 0 || window <= 0 {
		return nil
	}

	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(window+1)

	// Initialize the first EMA value as the first price
	result[0] = prices[0]

	for i := 1; i < len(prices); i++ {
		// Formula: EMA = (Price - PrevEMA) * Multiplier + PrevEMA
		// We round to 2 digits as requested previously
		val := (prices[i]-result[i-1])*multiplier + result[i-1]
		result[i] = math.Round(val*100) / 100
	}

	return result
}

func priceChange(prices []float64, lookback int) float64 {
	n := len(prices)

	// Ensure we have enough data points and a valid lookback
	if n <= lookback || lookback <= 0 {
		return 0.0
	}

	currentPrice := prices[n-1]
	oldPrice := prices[n-1-lookback]

	if oldPrice == 0 {
		return 0.0
	}

	change := currentPrice/oldPrice - 1

	return change
}
