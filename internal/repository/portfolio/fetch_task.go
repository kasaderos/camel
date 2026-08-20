package portfolio

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
)

func (r *Repository) FetchTasks(
	ctx context.Context,
	portfolioID string,
	statuses []model.TaskStatus,
) ([]model.Task, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio id is required")
	}

	q := r.db.WithContext(ctx).Where("portfolio_id = ?", portfolioID)

	if len(statuses) > 0 {
		statusStrings := make([]string, 0, len(statuses))
		for _, status := range statuses {
			statusStrings = append(statusStrings, string(status))
		}
		q = q.Where("status IN ?", statusStrings)
	}

	var rows []Task
	err := q.Order("created_at ASC").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("fetch tasks: %w", err)
	}

	out := make([]model.Task, 0, len(rows))
	for _, row := range rows {
		out = append(out, toModel(row))
	}

	return out, nil
}
