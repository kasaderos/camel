package portfolio

import (
	"time"
)

type Portfolio struct {
	ID   string  `gorm:"column:id;primaryKey"`
	Cash float64 `gorm:"column:cash"`
	Cost float64 `gorm:"column:cost"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Portfolio) TableName() string {
	return "portfolios"
}

type PortfolioStock struct {
	PortfolioID string  `gorm:"column:portfolio_id;primaryKey"`
	StockID     string  `gorm:"column:stock_id;primaryKey"`
	EntryPrice  float64 `gorm:"column:entry_price"`
	Quantity    float64 `gorm:"column:quantity"`
}

func (PortfolioStock) TableName() string {
	return "portfolio_stocks"
}
