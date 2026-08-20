package model

import (
	"fmt"
	"io"
	"time"
)

type Portfolio struct {
	ID string

	// Free cash
	Cash float64

	// Stocks in the portfolio
	Stocks []PortfolioStock

	// Total cost of the portfolio
	Cost float64

	UpdatedAt time.Time
}

type PortfolioStock struct {
	StockID    StockID
	EntryPrice float64
	Quantity   float64
}

func (p Portfolio) Print(w io.Writer) {
	fmt.Fprintln(w, "----------------")
	fmt.Fprintf(w, "ID:       %s\n", p.ID)
	fmt.Fprintf(w, "Updated:  %s\n", p.UpdatedAt.Format(time.RFC3339))
	fmt.Fprintln(w, "----------------")
	fmt.Fprintf(w, "Cash:     %.2f\n", p.Cash)
	fmt.Fprintf(w, "Cost:     %.2f\n", p.Cost)
	fmt.Fprintln(w, "----------------")
	fmt.Fprintln(w, "Positions")
	for _, stock := range p.Stocks {
		if stock.Quantity <= 0 {
			continue
		}
		fmt.Fprintf(w, "%-12s %g\n", stock.StockID, stock.Quantity)
	}
}

func (p *Portfolio) AddStockQuantity(stockID StockID, qty float64) error {
	for i := range p.Stocks {
		if p.Stocks[i].StockID == stockID {
			p.Stocks[i].Quantity += qty
			return nil
		}
	}

	return fmt.Errorf("stock not found in portfolio: %s", stockID)
}

func (p *Portfolio) ApplyOrder(order *Order) error {
	if order == nil || order.ID == "" {
		return fmt.Errorf("order id is empty")
	}

	qtyChange := order.Qty
	if order.Side == OrderSideSell {
		qtyChange = -qtyChange
	}

	if err := p.AddStockQuantity(order.AssetID, qtyChange); err != nil {
		return err
	}

	cashChange := order.Sum()
	if order.Side == OrderSideBuy {
		p.Cash -= cashChange
	} else {
		p.Cash += cashChange
	}

	return nil
}
