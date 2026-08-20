package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (r *Repository) CreateTask(
	ctx context.Context,
	task model.Task,
) error {
	if task.PortfolioID == "" {
		return fmt.Errorf("portfolio id is required")
	}
	if task.StockID == "" {
		return fmt.Errorf("stock id is required")
	}

	now := time.Now().UTC()
	row := Task{
		PortfolioID:  task.PortfolioID,
		StockID:      task.StockID,
		Quantity:     task.Quantity,
		Side:         string(task.Side),
		Status:       string(task.Status),
		ErrorMessage: task.ErrorMessage,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if task.ID != 0 {
		row.ID = task.ID
	}

	if task.Order != nil {
		orderID := task.Order.ID
		row.OrderID = &orderID
		row.AvgFillPrice = task.Order.AvgFillPrice
		row.FilledQty = task.Order.FilledQty
		if !task.Order.CreatedAt.IsZero() {
			t := task.Order.CreatedAt
			row.SubmittedAt = &t
		}
		if !task.Order.FilledAt.IsZero() {
			t := task.Order.FilledAt
			row.FilledAt = &t
		}
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}
