package model

import "time"

const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
)

type PendingOrder struct {
	PortfolioID int64
	Symbol      string
	Sum         float64

	CreatedAt time.Time
}

type Order struct {
	ID     string
	Symbol string

	Amount float64
	Price  float64

	CreatedAt time.Time
}
