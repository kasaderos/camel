package alpaca

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) FetchOrder(ctx context.Context, orderID string) (*model.Order, error) {
	order, err := s.client.GetOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	return toModelOrder(order)
}
