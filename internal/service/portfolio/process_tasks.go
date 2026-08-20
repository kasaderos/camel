package portfolio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) ProcessTasks(
	ctx context.Context,
	portfolioID string,
) error {
	tasks, err := s.taskRepo.FetchTasks(
		ctx,
		portfolioID,
		[]model.TaskStatus{
			model.TaskStatusCreated,
			model.TaskStatusOrderSent,
		},
	)
	if err != nil {
		return fmt.Errorf("fetch tasks: %w", err)
	}

	var sells, buys []model.Task
	for _, task := range tasks {
		if task.Side == model.OrderSideSell {
			sells = append(sells, task)
		} else {
			buys = append(buys, task)
		}
	}

	for _, task := range sells {
		if err := s.processTask(ctx, portfolioID, task); err != nil {
			slog.Error("process sell task",
				"err", err,
				"taskID", task.ID,
				"stockID", task.StockID,
				"status", task.Status,
			)
		}
	}

	for _, task := range buys {
		if err := s.processTask(ctx, portfolioID, task); err != nil {
			slog.Error("process buy task",
				"err", err,
				"taskID", task.ID,
				"stockID", task.StockID,
				"status", task.Status,
			)
		}
	}

	return nil
}

func (s *Service) processTask(
	ctx context.Context,
	portfolioID string,
	task model.Task,
) error {
	switch task.Status {
	case model.TaskStatusCreated:
		return s.sendOrder(ctx, &task)

	case model.TaskStatusOrderSent:
		return s.checkOrderFill(ctx, &task)

	case model.TaskStatusOrderFilled:
		return s.completeTask(ctx, portfolioID, &task)

	default:
		return fmt.Errorf("unknown task status: %s", task.Status)
	}
}

func (s *Service) sendOrder(ctx context.Context, task *model.Task) error {
	order, err := s.exchange.CreateMarketOrder(
		ctx,
		task.StockID,
		task.Quantity,
		task.Side,
	)
	if err != nil {
		return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
			ID:           task.ID,
			Quantity:     &task.Quantity,
			Status:       new(model.TaskStatusOrderFailed),
			ErrorMessage: new(err.Error()),
		})
	}

	err = task.ValidateOrder(order)
	if err != nil {
		return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
			ID:           task.ID,
			Order:        order,
			Status:       new(model.TaskStatusOrderFailed),
			ErrorMessage: new(err.Error()),
		})
	}

	slog.Info("order created",
		"orderID", order.ID,
		"stockID", order.AssetID,
		"quantity", order.Qty,
		"side", order.Side,
	)

	return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
		ID:     task.ID,
		Status: new(model.TaskStatusOrderSent),
		Order:  order,
	})
}

func (s *Service) checkOrderFill(ctx context.Context, task *model.Task) error {
	if task.Order == nil || task.Order.ID == "" {
		return fmt.Errorf("order created, but order id is empty")
	}

	order, err := s.exchange.FetchOrder(ctx, task.Order.ID)
	if err != nil {
		return fmt.Errorf("fetch order %s: %w", task.Order.ID, err)
	}

	if order.Status != model.OrderStatusFilled {
		if order.Status == model.OrderStatusCancelled {
			return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
				ID:           task.ID,
				Status:       new(model.TaskStatusOrderFailed),
				ErrorMessage: new("order cancelled by user"),
			})
		}

		slog.Info("order not yet filled",
			"orderID", task.Order.ID,
			"status", order.Status,
		)

		return nil
	}

	return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
		ID:     task.ID,
		Status: new(model.TaskStatusOrderFilled),
		Order:  order,
	})
}

func (s *Service) completeTask(
	ctx context.Context,
	portfolioID string,
	task *model.Task,
) error {
	portfolio, err := s.repo.FetchPortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("fetch portfolio: %w", err)
	}

	err = portfolio.ApplyOrder(task.Order)
	if err != nil {
		return err
	}

	err = s.repo.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("update portfolio cash: %w", err)
	}

	return s.taskRepo.UpdateTask(ctx, model.UpdateTask{
		ID:     task.ID,
		Status: new(model.TaskStatusCompleted),
		Order:  task.Order,
	})
}
