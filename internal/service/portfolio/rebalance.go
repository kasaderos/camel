package portfolio

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) PlanRebalance(
	ctx context.Context,
	portfolioID string,
) error {
	slog.Info("plan rebalance", "portfolioID", portfolioID)

	tasks, err := s.FetchRebalanceTasks(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("fetch rebalance tasks: %w", err)
	}

	if len(tasks) > 0 {
		return fmt.Errorf("some tasks are not completed: %v", len(tasks))
	}

	err = s.createRebalanceTasks(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("create rebalance tasks: %w", err)
	}

	return nil
}

func (s *Service) FetchRebalanceTasks(
	ctx context.Context,
	portfolioID string,
) ([]model.Task, error) {
	tasks, err := s.taskRepo.FetchTasks(
		ctx,
		portfolioID,
		[]model.TaskStatus{
			model.TaskStatusCreated,
			model.TaskStatusOrderSent,
			model.TaskStatusOrderFilled,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("fetch tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) createRebalanceTasks(
	ctx context.Context,
	portfolioID string,
) error {
	const threshold = 0.01
	const possibleDailyChange = 0.01

	portfolio, err := s.repo.FetchPortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("fetch portfolio: %w", err)
	}

	scores, err := s.fetchStockScores(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("fetch stock scores: %w", err)
	}

	targetWeights, err := buildTargetWeights(portfolio, scores, threshold)
	if err != nil {
		return fmt.Errorf("build target weights: %w", err)
	}

	if targetWeights == nil {
		return fmt.Errorf("no target weights")
	}

	currentPrices, err := s.fetchStockPrices(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("fetch current prices: %w", err)
	}

	portfolio.Cost = calcPortfolioCost(portfolio, currentPrices, possibleDailyChange)

	tasks, err := prepareTasks(portfolio, currentPrices, targetWeights)
	if err != nil {
		return fmt.Errorf("prepare tasks: %w", err)
	}

	slog.Info("prepare tasks", "count", len(tasks))

	for _, task := range tasks {
		err = s.taskRepo.CreateTask(ctx, *task)
		if err != nil {
			return fmt.Errorf("create task: %w", err)
		}

		slog.Info("created task", "task", task)
	}

	err = s.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	return nil
}

func (s *Service) FetchPortfolioScore(
	ctx context.Context,
	portfolioID string,
) (map[model.StockID]float64, error) {
	portfolio, err := s.repo.FetchPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("fetch portfolio: %w", err)
	}

	scores, err := s.fetchStockScores(ctx, portfolio)
	if err != nil {
		return nil, fmt.Errorf("fetch stock scores: %w", err)
	}

	return scores, nil
}

func buildTask(
	portfolioID string,
	stock model.PortfolioStock,
	targetSum float64,
	price float64,
) *model.Task {
	if targetSum == 0 && stock.Quantity > 0 {
		// sell all
		return &model.Task{
			PortfolioID: portfolioID,
			StockID:     stock.StockID,
			Side:        model.OrderSideSell,
			Quantity:    math.Floor(stock.Quantity),
			Status:      model.TaskStatusCreated,
		}
	}

	currentSum := stock.Quantity * price

	if math.Abs(targetSum-currentSum) < 0.01 {
		return nil
	}

	qty := (targetSum - currentSum) / price
	if math.Abs(qty) < 1 {
		return nil
	}

	side := model.OrderSideBuy
	if qty < 0 {
		side = model.OrderSideSell
		qty = -qty
	}

	return &model.Task{
		PortfolioID: portfolioID,
		StockID:     stock.StockID,
		Side:        side,
		Quantity:    math.Floor(qty),
		Status:      model.TaskStatusCreated,
	}
}

func buildTargetWeights(
	portfolio model.Portfolio,
	scores map[model.StockID]float64,
	threshold float64,
) (map[model.StockID]float64, error) {
	longScores := map[string]float64{}
	totalScore := 0.0

	for _, stock := range portfolio.Stocks {
		score := scores[stock.StockID]

		// If the score is below the threshold
		// long only stocks with high scores
		if score > threshold {
			longScores[stock.StockID] = score
			totalScore += score
		}
	}

	if totalScore == 0 {
		return nil, nil
	}

	targetWeights := map[string]float64{}
	for _, stock := range portfolio.Stocks {
		targetWeight := longScores[stock.StockID] / totalScore

		// If the weight is close to 1, set it to 0.3
		if targetWeight > 0.999 {
			targetWeight = 0.3
		}

		targetWeights[stock.StockID] = targetWeight
	}

	return targetWeights, nil
}

func (s *Service) fetchStockScores(
	ctx context.Context,
	portfolio model.Portfolio,
) (map[model.StockID]float64, error) {
	stockIDs := make([]model.StockID, len(portfolio.Stocks))
	for i, stock := range portfolio.Stocks {
		stockIDs[i] = stock.StockID
	}

	scores, err := s.analytics.FetchStockScores(ctx, stockIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch stock scores: %w", err)
	}

	return scores, nil
}

func prepareTasks(
	portfolio model.Portfolio,
	currentPrices map[model.StockID]float64,
	targetWeights map[model.StockID]float64,
) ([]*model.Task, error) {
	tasks := []*model.Task{}

	for _, stock := range portfolio.Stocks {
		targetWeight := targetWeights[stock.StockID]

		price, ok := currentPrices[stock.StockID]
		if !ok {
			return nil, fmt.Errorf("price not found for %s", stock.StockID)
		}

		targetSum := targetWeight * portfolio.Cost

		task := buildTask(portfolio.ID, stock, targetSum, price)

		if task != nil {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func (s *Service) fetchStockPrices(
	ctx context.Context,
	portfolio model.Portfolio,
) (map[model.StockID]float64, error) {
	prices := map[model.StockID]float64{}

	for _, stock := range portfolio.Stocks {
		price, _, err := s.exchange.FetchPrice(ctx, stock.StockID)
		if err != nil {
			return nil, fmt.Errorf("fetch price: %w", err)
		}

		if price <= 0 {
			return nil, fmt.Errorf("price is not positive: %s", stock.StockID)
		}

		prices[stock.StockID] = price
	}

	return prices, nil
}

func calcPortfolioCost(
	portfolio model.Portfolio,
	currentPrices map[model.StockID]float64,
	possibleDailyChange float64,
) float64 {
	cost := portfolio.Cash

	for _, stock := range portfolio.Stocks {
		cost += stock.Quantity * currentPrices[stock.StockID]
	}

	return cost * (1 - possibleDailyChange)
}
