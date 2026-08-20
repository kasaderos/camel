package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"gorm.io/gorm"
)

func (r *Repository) CreatePortfolio(
	ctx context.Context,
	p model.Portfolio,
) error {
	if p.ID == "" {
		return fmt.Errorf("portfolio id is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		err := tx.Create(&Portfolio{
			ID:        p.ID,
			Cash:      p.Cash,
			Cost:      p.Cost,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
		if err != nil {
			return fmt.Errorf("create portfolio: %w", err)
		}

		if len(p.Stocks) == 0 {
			return nil
		}

		rows := make([]PortfolioStock, 0, len(p.Stocks))
		for _, stock := range p.Stocks {
			rows = append(rows, PortfolioStock{
				PortfolioID: p.ID,
				StockID:     stock.StockID,
				EntryPrice:  stock.EntryPrice,
				Quantity:    stock.Quantity,
			})
		}

		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("create portfolio stocks: %w", err)
		}

		return nil
	})
}
