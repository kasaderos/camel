package model

import "time"

const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"

	OrderSideBuy  = "buy"
	OrderSideSell = "sell"
)

type Order struct {
	PortfolioID string

	AssetID string
	Price   float64
	Qty     float64

	Status    string
	CreatedAt time.Time
}

func (p Order) Sum() float64 {
	return p.Price * p.Qty
}
