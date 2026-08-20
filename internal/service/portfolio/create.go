package portfolio

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
	"github.com/samber/lo"
)

func (s *Service) CreatePortfolio(
	ctx context.Context,
	id string,
	stockIDs []model.StockID,
	cash float64,
) error {
	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:   id,
		Cash: cash,
		Stocks: lo.Map(stockIDs, func(stockID model.StockID, _ int) model.PortfolioStock {
			return model.PortfolioStock{
				StockID:  stockID,
				Quantity: 0,
			}
		}),
		Cost: 0,
	})
	if err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}

	return nil
}
