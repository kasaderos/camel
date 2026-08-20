package portfolio

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type Exchanger interface {
	FetchPrice(
		ctx context.Context,
		assetID string,
	) (float64, time.Time, error)
	CreateMarketOrder(
		ctx context.Context,
		symbol string,
		qty float64,
		side model.OrderSide,
	) (*model.Order, error)
	FetchOrder(ctx context.Context, orderID string) (*model.Order, error)
	CancelOrder(ctx context.Context, orderID string) error
}

type PortfolioRepository interface {
	CreatePortfolio(ctx context.Context, p model.Portfolio) error
	FetchPortfolio(ctx context.Context, id string) (model.Portfolio, error)
	UpdatePortfolio(ctx context.Context, p model.Portfolio) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task model.Task) error
	UpdateTask(ctx context.Context, task model.UpdateTask) error
	FetchTasks(
		ctx context.Context,
		portfolioID string,
		statuses []model.TaskStatus,
	) ([]model.Task, error)
}

type AnalyticsService interface {
	FetchStockScores(ctx context.Context, stockIDs []model.StockID) (map[model.StockID]float64, error)
}
