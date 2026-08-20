package portfolio

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) UpdatePortfolio(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	err := s.repo.UpdatePortfolio(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	return nil
}
