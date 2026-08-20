package portfolio

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) FetchPortfolio(
	ctx context.Context,
	id string,
) (model.Portfolio, error) {
	portfolio, err := s.repo.FetchPortfolio(ctx, id)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("fetch portfolio: %w", err)
	}

	return portfolio, nil
}
