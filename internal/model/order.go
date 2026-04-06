package model

type Order struct {
	ID     string
	Symbol string

	Amount float64
	Price  float64
}
