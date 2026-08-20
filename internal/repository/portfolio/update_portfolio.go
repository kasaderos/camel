package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) UpdatePortfolio(
	ctx context.Context,
	p model.Portfolio,
) error {
	if p.ID == "" {
		return fmt.Errorf("portfolio id is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Portfolio{}).
			Where("id = ?", p.ID).
			Updates(map[string]any{
				"cash":       p.Cash,
				"cost":       p.Cost,
				"updated_at": time.Now().UTC(),
			})
		if res.Error != nil {
			return fmt.Errorf("update portfolio: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("portfolio not found")
		}

		for _, stock := range p.Stocks {
			row := PortfolioStock{
				PortfolioID: p.ID,
				StockID:     stock.StockID,
				EntryPrice:  stock.EntryPrice,
				Quantity:    stock.Quantity,
			}

			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "portfolio_id"}, {Name: "stock_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"entry_price", "quantity"}),
			}).Create(&row).Error
			if err != nil {
				return fmt.Errorf("update portfolio stock: %w", err)
			}
		}

		return nil
	})
}

func (r *Repository) UpdatePortfolioStock(
	ctx context.Context,
	portfolioID string,
	stock model.PortfolioStock,
) error {
	if portfolioID == "" {
		return fmt.Errorf("portfolio id is required")
	}
	if stock.StockID == "" {
		return fmt.Errorf("stock id is required")
	}

	res := r.db.WithContext(ctx).
		Model(&PortfolioStock{}).
		Where("portfolio_id = ? AND stock_id = ?", portfolioID, stock.StockID).
		Updates(map[string]any{
			"entry_price": stock.EntryPrice,
			"quantity":    stock.Quantity,
		})
	if res.Error != nil {
		return fmt.Errorf("update portfolio stock: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("portfolio stock not found")
	}

	return nil
}
