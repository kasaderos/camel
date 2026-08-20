package portfolio

import "time"

type Task struct {
	ID          int64   `gorm:"column:id;primaryKey;autoIncrement"`
	PortfolioID string  `gorm:"column:portfolio_id"`
	StockID     string  `gorm:"column:stock_id"`
	Quantity    float64 `gorm:"column:quantity"`
	Side        string  `gorm:"column:side"`
	Status      string  `gorm:"column:status"`

	OrderID      *string    `gorm:"column:order_id"`
	AvgFillPrice float64    `gorm:"column:avg_fill_price"`
	FilledQty    float64    `gorm:"column:filled_qty"`
	SubmittedAt  *time.Time `gorm:"column:submitted_at"`
	FilledAt     *time.Time `gorm:"column:filled_at"`
	ErrorMessage string     `gorm:"column:error_message"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Task) TableName() string {
	return "rebalance_tasks"
}
