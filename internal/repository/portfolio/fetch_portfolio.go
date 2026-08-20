package portfolio

import (
	"context"
	"errors"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
	"gorm.io/gorm"
)

func (r *Repository) FetchPortfolio(
	ctx context.Context,
	id string,
) (model.Portfolio, error) {
	if id == "" {
		return model.Portfolio{}, fmt.Errorf("portfolio id is required")
	}

	var row Portfolio
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Portfolio{}, fmt.Errorf("portfolio not found")
	}
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("fetch portfolio: %w", err)
	}

	var stockRows []PortfolioStock
	err = r.db.WithContext(ctx).
		Where("portfolio_id = ?", id).
		Order("stock_id").
		Find(&stockRows).Error
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("fetch portfolio stocks: %w", err)
	}

	stocks := make([]model.PortfolioStock, 0, len(stockRows))
	for _, s := range stockRows {
		stocks = append(stocks, toPortfolioStock(s))
	}

	return model.Portfolio{
		ID:        row.ID,
		Cash:      row.Cash,
		Cost:      row.Cost,
		Stocks:    stocks,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
