package portfolio

import (
	"testing"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/require"
)

func TestToPortfolioStock(t *testing.T) {
	got := toPortfolioStock(PortfolioStock{
		PortfolioID: "portfolio-1",
		StockID:     "AAPL",
		EntryPrice:  150.5,
		Quantity:    10,
	})

	require.Equal(t, model.PortfolioStock{
		StockID:    "AAPL",
		EntryPrice: 150.5,
		Quantity:   10,
	}, got)
}

func TestToModel(t *testing.T) {
	createdAt := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC)
	submittedAt := time.Date(2024, 1, 2, 10, 5, 0, 0, time.UTC)
	filledAt := time.Date(2024, 1, 2, 10, 10, 0, 0, time.UTC)
	orderID := "order-1"
	emptyOrderID := ""

	base := Task{
		ID:           "task-1",
		PortfolioID:  "portfolio-1",
		StockID:      "AAPL",
		Quantity:     2.5,
		Side:         string(model.OrderSideBuy),
		Status:       string(model.TaskStatusCreated),
		ErrorMessage: "none",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	tests := []struct {
		name string
		in   Task
		want model.Task
	}{
		{
			name: "without order",
			in:   base,
			want: model.Task{
				ID:           "task-1",
				PortfolioID:  "portfolio-1",
				StockID:      "AAPL",
				Side:         model.OrderSideBuy,
				Quantity:     2.5,
				Status:       model.TaskStatusCreated,
				ErrorMessage: "none",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
		},
		{
			name: "nil order id",
			in: func() Task {
				t := base
				t.OrderID = nil
				return t
			}(),
			want: model.Task{
				ID:           "task-1",
				PortfolioID:  "portfolio-1",
				StockID:      "AAPL",
				Side:         model.OrderSideBuy,
				Quantity:     2.5,
				Status:       model.TaskStatusCreated,
				ErrorMessage: "none",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
		},
		{
			name: "empty order id",
			in: func() Task {
				t := base
				t.OrderID = &emptyOrderID
				t.AvgFillPrice = 101
				t.FilledQty = 1
				return t
			}(),
			want: model.Task{
				ID:           "task-1",
				PortfolioID:  "portfolio-1",
				StockID:      "AAPL",
				Side:         model.OrderSideBuy,
				Quantity:     2.5,
				Status:       model.TaskStatusCreated,
				ErrorMessage: "none",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
		},
		{
			name: "with order without timestamps",
			in: func() Task {
				t := base
				t.Status = string(model.TaskStatusOrderSent)
				t.OrderID = &orderID
				t.AvgFillPrice = 0
				t.FilledQty = 0
				return t
			}(),
			want: model.Task{
				ID:           "task-1",
				PortfolioID:  "portfolio-1",
				StockID:      "AAPL",
				Side:         model.OrderSideBuy,
				Quantity:     2.5,
				Status:       model.TaskStatusOrderSent,
				ErrorMessage: "none",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
				Order: &model.Order{
					ID:           "order-1",
					AssetID:      "AAPL",
					AvgFillPrice: 0,
					Qty:          2.5,
					FilledQty:    0,
					Side:         model.OrderSideBuy,
				},
			},
		},
		{
			name: "with filled order",
			in: func() Task {
				t := base
				t.Status = string(model.TaskStatusOrderFilled)
				t.OrderID = &orderID
				t.AvgFillPrice = 150.25
				t.FilledQty = 2.5
				t.SubmittedAt = &submittedAt
				t.FilledAt = &filledAt
				t.ErrorMessage = ""
				return t
			}(),
			want: model.Task{
				ID:           "task-1",
				PortfolioID:  "portfolio-1",
				StockID:      "AAPL",
				Side:         model.OrderSideBuy,
				Quantity:     2.5,
				Status:       model.TaskStatusOrderFilled,
				ErrorMessage: "",
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
				Order: &model.Order{
					ID:           "order-1",
					AssetID:      "AAPL",
					AvgFillPrice: 150.25,
					Qty:          2.5,
					FilledQty:    2.5,
					Side:         model.OrderSideBuy,
					CreatedAt:    submittedAt,
					FilledAt:     filledAt,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, toModel(tt.in))
		})
	}
}
