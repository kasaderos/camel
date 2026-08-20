package portfolio

import "github.com/kasaderos/camel/internal/model"

func (s *RepositorySuite) TestCreateTask() {
	ctx := s.T().Context()

	portfolio := model.Portfolio{
		ID:   "portfolio-1",
		Cash: 10000,
		Stocks: []model.PortfolioStock{
			{StockID: "AAPL", EntryPrice: 0, Quantity: 0},
		},
	}
	err := s.repo.CreatePortfolio(ctx, portfolio)
	s.Require().NoError(err)

	task := model.Task{
		ID:          "task-1",
		PortfolioID: portfolio.ID,
		StockID:     "AAPL",
		Side:        model.OrderSideBuy,
		Quantity:    5,
		Status:      model.TaskStatusCreated,
	}

	err = s.repo.CreateTask(ctx, task)
	s.Require().NoError(err)

	fetched, err := s.repo.FetchTasks(
		ctx,
		portfolio.ID,
		[]model.TaskStatus{model.TaskStatusCreated},
	)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal(task.ID, fetched[0].ID)
	s.Equal(task.PortfolioID, fetched[0].PortfolioID)
	s.Equal(task.StockID, fetched[0].StockID)
	s.Equal(task.Side, fetched[0].Side)
	s.Equal(task.Quantity, fetched[0].Quantity)
	s.Equal(task.Status, fetched[0].Status)
	s.Nil(fetched[0].Order)
}

func (s *RepositorySuite) TestCreateTask_GeneratesID() {
	ctx := s.T().Context()

	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:   "portfolio-1",
		Cash: 1000,
	})
	s.Require().NoError(err)

	err = s.repo.CreateTask(ctx, model.Task{
		PortfolioID: "portfolio-1",
		StockID:     "AAPL",
		Side:        model.OrderSideBuy,
		Quantity:    1,
		Status:      model.TaskStatusCreated,
	})
	s.Require().NoError(err)

	fetched, err := s.repo.FetchTasks(ctx, "portfolio-1", nil)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.NotEmpty(fetched[0].ID)
}
