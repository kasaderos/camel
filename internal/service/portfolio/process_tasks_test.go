package portfolio

import (
	"errors"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/internal/service/portfolio/mocks"
)

func (s *TestSuite) TestSendOrder() {
	task := model.Task{
		ID:       "task-1",
		StockID:  "AAPL",
		Quantity: 2.0,
		Side:     model.OrderSideBuy,
		Status:   model.TaskStatusCreated,
	}

	tests := []struct {
		name    string
		setup   func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository)
		wantErr string
	}{
		{
			name: "marks task failed when create order fails",
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				exchange.EXPECT().
					CreateMarketOrder(s.ctx, "AAPL", 2.0, model.OrderSideBuy).
					Return(nil, errors.New("exchange down"))

				taskRepo.EXPECT().
					UpdateTask(s.ctx, model.UpdateTask{
						ID:           "task-1",
						Quantity:     new(2.0),
						Status:       new(model.TaskStatusOrderFailed),
						ErrorMessage: new("exchange down"),
					}).
					Return(nil)
			},
		},
		{
			name: "marks task failed when order validation fails",
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				order := &model.Order{
					ID:      "order-2",
					AssetID: "AAPL",
					Qty:     3.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusPending,
				}

				exchange.EXPECT().
					CreateMarketOrder(s.ctx, "AAPL", 2.0, model.OrderSideBuy).
					Return(order, nil)

				taskRepo.EXPECT().
					UpdateTask(s.ctx, model.UpdateTask{
						ID:           "task-1",
						Order:        order,
						Status:       new(model.TaskStatusOrderFailed),
						ErrorMessage: new("order quantity mismatch: 3.000000 != 2.000000"),
					}).
					Return(nil)
			},
		},
		{
			name: "marks task order sent when create succeeds",
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				order := &model.Order{
					ID:      "order-1",
					AssetID: "AAPL",
					Qty:     2.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusPending,
				}

				exchange.EXPECT().
					CreateMarketOrder(s.ctx, "AAPL", 2.0, model.OrderSideBuy).
					Return(order, nil)

				taskRepo.EXPECT().
					UpdateTask(s.ctx, model.UpdateTask{
						ID:     "task-1",
						Status: new(model.TaskStatusOrderSent),
						Order:  order,
					}).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exchange := mocks.NewExchanger(s.T())
			taskRepo := mocks.NewTaskRepository(s.T())
			svc := &Service{
				exchange: exchange,
				taskRepo: taskRepo,
			}
			tt.setup(exchange, taskRepo)

			taskCopy := task
			err := svc.sendOrder(s.ctx, &taskCopy)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *TestSuite) TestCheckOrderFill() {
	tests := []struct {
		name    string
		task    model.Task
		setup   func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository)
		wantErr string
	}{
		{
			name: "returns error when order is nil",
			task: model.Task{
				ID:     "task-1",
				Status: model.TaskStatusOrderSent,
			},
			setup:   func(*mocks.Exchanger, *mocks.TaskRepository) {},
			wantErr: "order created, but order id is empty",
		},
		{
			name: "returns error when fetch order fails",
			task: model.Task{
				ID:       "task-1",
				StockID:  "AAPL",
				Quantity: 2.0,
				Side:     model.OrderSideBuy,
				Status:   model.TaskStatusOrderSent,
				Order: &model.Order{
					ID:      "order-1",
					AssetID: "AAPL",
					Qty:     2.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusPending,
				},
			},
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				exchange.EXPECT().
					FetchOrder(s.ctx, "order-1").
					Return(nil, errors.New("unavailable"))
			},
			wantErr: "fetch order order-1: unavailable",
		},
		{
			name: "does nothing when order is not filled",
			task: model.Task{
				ID:       "task-1",
				StockID:  "AAPL",
				Quantity: 2.0,
				Side:     model.OrderSideBuy,
				Status:   model.TaskStatusOrderSent,
				Order: &model.Order{
					ID:      "order-1",
					AssetID: "AAPL",
					Qty:     2.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusPending,
				},
			},
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				exchange.EXPECT().
					FetchOrder(s.ctx, "order-1").
					Return(&model.Order{
						ID:      "order-1",
						AssetID: "AAPL",
						Qty:     2.0,
						Side:    model.OrderSideBuy,
						Status:  model.OrderStatusPending,
					}, nil)
			},
		},
		{
			name: "marks task order filled when order is filled",
			task: model.Task{
				ID:       "task-1",
				StockID:  "AAPL",
				Quantity: 2.0,
				Side:     model.OrderSideBuy,
				Status:   model.TaskStatusOrderSent,
				Order: &model.Order{
					ID:      "order-1",
					AssetID: "AAPL",
					Qty:     2.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusPending,
				},
			},
			setup: func(exchange *mocks.Exchanger, taskRepo *mocks.TaskRepository) {
				order := &model.Order{
					ID:      "order-1",
					AssetID: "AAPL",
					Qty:     2.0,
					Side:    model.OrderSideBuy,
					Status:  model.OrderStatusFilled,
				}

				exchange.EXPECT().
					FetchOrder(s.ctx, "order-1").
					Return(order, nil)

				taskRepo.EXPECT().
					UpdateTask(s.ctx, model.UpdateTask{
						ID:     "task-1",
						Status: new(model.TaskStatusOrderFilled),
						Order:  order,
					}).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exchange := mocks.NewExchanger(s.T())
			taskRepo := mocks.NewTaskRepository(s.T())
			svc := &Service{
				exchange: exchange,
				taskRepo: taskRepo,
			}
			tt.setup(exchange, taskRepo)

			taskCopy := tt.task
			err := svc.checkOrderFill(s.ctx, &taskCopy)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *TestSuite) TestCompleteTask() {
	tests := []struct {
		name    string
		task    model.Task
		setup   func(repo *mocks.PortfolioRepository, taskRepo *mocks.TaskRepository)
		wantErr string
	}{
		{
			name: "returns error when order is nil",
			task: model.Task{
				ID:     "task-1",
				Status: model.TaskStatusOrderFilled,
			},
			setup: func(repo *mocks.PortfolioRepository, taskRepo *mocks.TaskRepository) {
				repo.EXPECT().
					FetchPortfolio(s.ctx, "portfolio-1").
					Return(model.Portfolio{
						ID:   "portfolio-1",
						Cash: 1000,
						Stocks: []model.PortfolioStock{
							{StockID: "AAPL", Quantity: 1},
						},
					}, nil)
			},
			wantErr: "order id is empty",
		},
		{
			name: "returns error when stock is not in portfolio",
			task: model.Task{
				ID:     "task-1",
				Status: model.TaskStatusOrderFilled,
				Order: &model.Order{
					ID:           "order-1",
					AssetID:      "MSFT",
					Qty:          2.0,
					AvgFillPrice: 50.0,
					Side:         model.OrderSideBuy,
					Status:       model.OrderStatusFilled,
				},
			},
			setup: func(repo *mocks.PortfolioRepository, taskRepo *mocks.TaskRepository) {
				repo.EXPECT().
					FetchPortfolio(s.ctx, "portfolio-1").
					Return(model.Portfolio{
						ID:   "portfolio-1",
						Cash: 1000,
						Stocks: []model.PortfolioStock{
							{StockID: "AAPL", Quantity: 1},
						},
					}, nil)
			},
			wantErr: "stock not found in portfolio: MSFT",
		},
		{
			name: "applies buy order and marks task completed",
			task: model.Task{
				ID:     "task-1",
				Status: model.TaskStatusOrderFilled,
				Order: &model.Order{
					ID:           "order-1",
					AssetID:      "AAPL",
					Qty:          2.0,
					AvgFillPrice: 50.0,
					Side:         model.OrderSideBuy,
					Status:       model.OrderStatusFilled,
				},
			},
			setup: func(repo *mocks.PortfolioRepository, taskRepo *mocks.TaskRepository) {
				order := &model.Order{
					ID:           "order-1",
					AssetID:      "AAPL",
					Qty:          2.0,
					AvgFillPrice: 50.0,
					Side:         model.OrderSideBuy,
					Status:       model.OrderStatusFilled,
				}

				repo.EXPECT().
					FetchPortfolio(s.ctx, "portfolio-1").
					Return(model.Portfolio{
						ID:   "portfolio-1",
						Cash: 1000,
						Stocks: []model.PortfolioStock{
							{StockID: "AAPL", Quantity: 1},
						},
					}, nil)

				repo.EXPECT().
					UpdatePortfolio(s.ctx, model.Portfolio{
						ID:   "portfolio-1",
						Cash: 900,
						Stocks: []model.PortfolioStock{
							{StockID: "AAPL", Quantity: 3},
						},
					}).
					Return(nil)

				taskRepo.EXPECT().
					UpdateTask(s.ctx, model.UpdateTask{
						ID:     "task-1",
						Status: new(model.TaskStatusCompleted),
						Order:  order,
					}).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			repo := mocks.NewPortfolioRepository(s.T())
			taskRepo := mocks.NewTaskRepository(s.T())
			svc := &Service{
				repo:     repo,
				taskRepo: taskRepo,
			}
			tt.setup(repo, taskRepo)

			taskCopy := tt.task
			err := svc.completeTask(s.ctx, "portfolio-1", &taskCopy)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
		})
	}
}
