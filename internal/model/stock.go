package model

import "time"

type StockID = string

type Stock struct {
	ID    StockID
	Price float64

	UpdatedAt time.Time
}
