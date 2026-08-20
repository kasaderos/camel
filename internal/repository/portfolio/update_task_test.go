package portfolio

import (
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (s *RepositorySuite) TestUpdateTask() {
	ctx := s.T().Context()

	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:   "portfolio-1",
		Cash: 10000,
		Stocks: []model.PortfolioStock{
			{StockID: "AAPL"},
		},
	})
	s.Require().NoError(err)

	task := model.Task{
		ID:          "task-1",
		PortfolioID: "portfolio-1",
		StockID:     "AAPL",
		Side:        model.OrderSideBuy,
		Quantity:    5,
		Status:      model.TaskStatusCreated,
	}
	err = s.repo.CreateTask(ctx, task)
	s.Require().NoError(err)

	createdAt := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	status := model.TaskStatusOrderSent
	err = s.repo.UpdateTask(ctx, model.UpdateTask{
		ID:     task.ID,
		Status: &status,
		Order: &model.Order{
			ID:        "order-1",
			AssetID:   "AAPL",
			Qty:       5,
			Side:      model.OrderSideBuy,
			CreatedAt: createdAt,
		},
	})
	s.Require().NoError(err)

	fetched, err := s.repo.FetchTasks(
		ctx,
		"portfolio-1",
		[]model.TaskStatus{model.TaskStatusOrderSent},
	)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal(model.TaskStatusOrderSent, fetched[0].Status)
	s.Require().NotNil(fetched[0].Order)
	s.Equal("order-1", fetched[0].Order.ID)
	s.True(fetched[0].Order.CreatedAt.Equal(createdAt))

	filledStatus := model.TaskStatusOrderFilled
	filledAt := createdAt.Add(time.Minute)
	err = s.repo.UpdateTask(ctx, model.UpdateTask{
		ID:     task.ID,
		Status: &filledStatus,
		Order: &model.Order{
			ID:           "order-1",
			AssetID:      "AAPL",
			AvgFillPrice: 150.5,
			Qty:          5,
			FilledQty:    5,
			Side:         model.OrderSideBuy,
			CreatedAt:    createdAt,
			FilledAt:     filledAt,
		},
	})
	s.Require().NoError(err)

	fetched, err = s.repo.FetchTasks(
		ctx,
		"portfolio-1",
		[]model.TaskStatus{model.TaskStatusOrderFilled},
	)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal(model.TaskStatusOrderFilled, fetched[0].Status)
	s.Require().NotNil(fetched[0].Order)
	s.Equal(150.5, fetched[0].Order.AvgFillPrice)
	s.Equal(5.0, fetched[0].Order.FilledQty)
	s.True(fetched[0].Order.FilledAt.Equal(filledAt))

	failedStatus := model.TaskStatusOrderFailed
	errMsg := "rejected"
	err = s.repo.UpdateTask(ctx, model.UpdateTask{
		ID:           task.ID,
		Status:       &failedStatus,
		ErrorMessage: &errMsg,
	})
	s.Require().NoError(err)

	fetched, err = s.repo.FetchTasks(
		ctx,
		"portfolio-1",
		[]model.TaskStatus{model.TaskStatusOrderFailed},
	)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal(errMsg, fetched[0].ErrorMessage)
}
