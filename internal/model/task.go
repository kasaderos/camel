package model

import (
	"fmt"
	"math"
	"time"
)

type TaskStatus string

const (
	TaskStatusCreated     TaskStatus = "created"
	TaskStatusOrderFailed TaskStatus = "order_failed"
	TaskStatusOrderSent   TaskStatus = "order_sent"
	TaskStatusOrderFilled TaskStatus = "order_filled"
	TaskStatusCompleted   TaskStatus = "completed"
)

type Task struct {
	ID          int64
	PortfolioID string
	StockID     StockID
	Side        OrderSide
	Quantity    float64
	Status      TaskStatus

	Order *Order

	ErrorMessage string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateTask struct {
	ID           int64
	Quantity     *float64
	Status       *TaskStatus
	Order        *Order
	ErrorMessage *string
}

func (t *Task) ValidateOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}

	if math.Abs(order.Qty-t.Quantity) > 0.01 {
		return fmt.Errorf("order quantity mismatch: %f != %f", order.Qty, t.Quantity)
	}

	return nil
}
