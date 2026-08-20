package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (r *Repository) UpdateTask(
	ctx context.Context,
	task model.UpdateTask,
) error {
	if task.ID == "" {
		return fmt.Errorf("task id is required")
	}

	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}

	if task.Quantity != nil {
		updates["quantity"] = *task.Quantity
	}

	if task.Status != nil {
		updates["status"] = string(*task.Status)
	}

	if task.ErrorMessage != nil {
		updates["error_message"] = *task.ErrorMessage
	}

	if task.Order != nil {
		updates["order_id"] = task.Order.ID
		updates["avg_fill_price"] = task.Order.AvgFillPrice
		updates["filled_qty"] = task.Order.FilledQty

		if !task.Order.CreatedAt.IsZero() {
			updates["submitted_at"] = task.Order.CreatedAt
		}
		if !task.Order.FilledAt.IsZero() {
			updates["filled_at"] = task.Order.FilledAt
		}
	}

	res := r.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ?", task.ID).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update task: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
