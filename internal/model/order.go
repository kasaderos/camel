package model

import "time"

type OrderSide string

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
	OrderSideNone OrderSide = "none"
)

type Order struct {
	ID string

	AssetID      string
	AvgFillPrice float64
	Qty          float64
	FilledQty    float64

	Side   OrderSide
	Status OrderStatus

	FilledAt  time.Time
	CreatedAt time.Time
}

func (p Order) Sum() float64 {
	return p.AvgFillPrice * p.Qty
}
