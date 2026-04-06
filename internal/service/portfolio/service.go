package portfolio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
)

type Repository interface {
	FetchPortfolio(ctx context.Context, id int64) (*model.Portfolio, error)
	SearchPortfolios(ctx context.Context, offset, limit int) ([]*model.Portfolio, error)
	CreatePortfolio(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	UpdatePortfolio(ctx context.Context, p model.Portfolio) error
	DeletePortfolio(ctx context.Context, id int64) error
	LogOrder(ctx context.Context, portfolioID int64, orderID string) error
}

type Allocator interface {
	// Allocate takes a list of symbols and a cache amount,
	// and returns a map of symbol to weight, a policy ID, and an error if any.
	// The weights should be between -1 and 1
	Allocate(
		ctx context.Context,
		symbols []string,
		cache float64,
	) (
		weights map[string]float64,
		policyID string,
		err error,
	)
}

type OrderExecuter interface {
	// CreateOrder takes a symbol and a sum, and returns an order and an error if any.
	// The sum may be negative, in which case the order should be a sell order.
	// The precision of the sum should be at most 2 decimal places,
	// and it should be rounded down if necessary.
	CreateOrder(
		ctx context.Context,
		symbol string,
		sum float64,
	) (order model.Order, err error)
}

type Service struct {
	repo      Repository
	allocator Allocator
	executer  OrderExecuter
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FetchPortfolio(ctx context.Context, id int64) (*model.Portfolio, error) {
	return s.repo.FetchPortfolio(ctx, id)
}

func (s *Service) SearchPortfolios(
	ctx context.Context,
	offset int,
	limit int,
) ([]*model.Portfolio, error) {
	return s.repo.SearchPortfolios(ctx, offset, limit)
}

func (s *Service) CreatePortfolio(
	ctx context.Context,
	name string,
	symbols []string,
	cache float64,
) (int64, error) {
	weights, policyID, err := s.allocator.Allocate(ctx, symbols, cache)
	if err != nil {
		return 0, fmt.Errorf("allocate portfolio: %w", err)
	}

	portfolio := model.Portfolio{
		Name: name,

		Weights: weights,
		Cache:   cache,

		PolicyID: policyID,
	}

	portfolio, err = s.repo.CreatePortfolio(ctx, portfolio)
	if err != nil {
		return 0, fmt.Errorf("create portfolio: %w", err)
	}

	freeCache, err := s.executeOrders(ctx, portfolio)
	if err != nil {
		return 0, fmt.Errorf("execute orders: %w", err)
	}

	err = s.updatePortfolioCache(ctx, portfolio, freeCache)
	if err != nil {
		return 0, fmt.Errorf("update portfolio cache: %w", err)
	}

	slog.Debug(
		"Created portfolio",
		"id", portfolio.ID,
		"name", name,
		"weights", weights,
		"cache", cache,
		"policy_id", policyID,
	)

	return portfolio.ID, nil
}

func (s *Service) DeletePortfolio(ctx context.Context, id int64) error {
	return s.repo.DeletePortfolio(ctx, id)
}

func (s *Service) executeOrders(ctx context.Context, p model.Portfolio) (freeCache float64, err error) {
	var order model.Order

	for symbol, weight := range p.Weights {
		order, freeCache, err = s.createOrder(ctx, p.Cache, weight, symbol)
		if err != nil {
			return 0, fmt.Errorf("create order for symbol %s: %w", symbol, err)
		}

		err = s.repo.LogOrder(
			ctx,
			p.ID,
			order.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("log order for symbol %s: %w", symbol, err)
		}

		slog.Debug(
			"Executed order",
			"portfolio_id", p.ID,
			"symbol", order.Symbol,
			"amount", order.Amount,
			"order_id", order.ID,
		)
	}

	return freeCache, nil
}

func (s *Service) createOrder(
	ctx context.Context,
	initialCache float64,
	weight float64,
	symbol string,
) (o model.Order, freeCache float64, err error) {
	// Calculate the sum of the order by multiplying the initial cache by the weight
	orderSum, exact := decimal.NewFromFloat(initialCache * weight).RoundDown(2).Float64()
	if !exact {
		slog.Warn(
			"Rounded order sum down to 2 decimal places",
			"symbol", symbol,
			"initial_cache", initialCache,
			"weight", weight,
			"order_sum", orderSum,
		)
	}

	order, err := s.executer.CreateOrder(ctx, symbol, orderSum)
	if err != nil {
		return model.Order{}, 0, fmt.Errorf("create order for symbol %s: %w", symbol, err)
	}

	sum, exact := decimal.NewFromFloat(order.Amount * order.Price).RoundDown(2).Float64()
	if !exact {
		slog.Warn(
			"Rounded order sum down to 2 decimal places",
			"symbol", symbol,
			"order_amount", order.Amount,
			"order_price", order.Price,
			"order_sum", sum,
		)
	}

	// Update the free cache by subtracting the sum of the order from the initial cache
	freeCache = initialCache - sum

	return order, freeCache, nil
}

func validateWeights(weights map[string]float64) error {
	digitsAfterPoint := int32(2)

	for symbol, weight := range weights {
		if weight < -1 || weight > 1.0 {
			return fmt.Errorf(
				"weight for symbol %s must be between -1 and 1, got %f",
				symbol,
				weight,
			)
		}

		// Check if the weight has more than 2 decimal places
		if decimal.NewFromFloat(weight).Exponent() != digitsAfterPoint {
			return fmt.Errorf(
				"weight for symbol %s must have at most 2 decimal places, got %f",
				symbol,
				weight,
			)
		}
	}

	return nil

}

func (s *Service) updatePortfolioCache(ctx context.Context, p model.Portfolio, freeCache float64) error {
	p.Cache = freeCache

	err := s.repo.UpdatePortfolio(ctx, p)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	return nil
}
