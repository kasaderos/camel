package portfolio

import "github.com/kasaderos/camel/internal/model"

func toPortfolioStock(s PortfolioStock) model.PortfolioStock {
	return model.PortfolioStock{
		StockID:    s.StockID,
		EntryPrice: s.EntryPrice,
		Quantity:   s.Quantity,
	}
}

func toModel(t Task) model.Task {
	task := model.Task{
		ID:           t.ID,
		PortfolioID:  t.PortfolioID,
		StockID:      t.StockID,
		Side:         model.OrderSide(t.Side),
		Quantity:     t.Quantity,
		Status:       model.TaskStatus(t.Status),
		ErrorMessage: t.ErrorMessage,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}

	if t.OrderID != nil && *t.OrderID != "" {
		order := &model.Order{
			ID:           *t.OrderID,
			AssetID:      t.StockID,
			AvgFillPrice: t.AvgFillPrice,
			Qty:          t.Quantity,
			FilledQty:    t.FilledQty,
			Side:         model.OrderSide(t.Side),
		}
		if t.SubmittedAt != nil {
			order.CreatedAt = *t.SubmittedAt
		}
		if t.FilledAt != nil {
			order.FilledAt = *t.FilledAt
		}
		task.Order = order
	}

	return task
}
