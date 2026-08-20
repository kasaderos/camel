package portfolio

import "github.com/kasaderos/camel/internal/model"

func (s *RepositorySuite) TestFetchTasks_FiltersByStatus() {
	ctx := s.T().Context()

	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:   "portfolio-1",
		Cash: 1000,
	})
	s.Require().NoError(err)

	err = s.repo.CreateTask(ctx, model.Task{
		ID:          "task-created",
		PortfolioID: "portfolio-1",
		StockID:     "AAPL",
		Side:        model.OrderSideBuy,
		Quantity:    1,
		Status:      model.TaskStatusCreated,
	})
	s.Require().NoError(err)

	err = s.repo.CreateTask(ctx, model.Task{
		ID:          "task-completed",
		PortfolioID: "portfolio-1",
		StockID:     "MSFT",
		Side:        model.OrderSideSell,
		Quantity:    2,
		Status:      model.TaskStatusCompleted,
	})
	s.Require().NoError(err)

	fetched, err := s.repo.FetchTasks(
		ctx,
		"portfolio-1",
		[]model.TaskStatus{model.TaskStatusCreated},
	)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal("task-created", fetched[0].ID)

	fetched, err = s.repo.FetchTasks(ctx, "portfolio-1", nil)
	s.Require().NoError(err)
	s.Len(fetched, 2)
}
