package alpaca

import (
	"context"
	"fmt"
)

func (s *Service) CancelOrder(ctx context.Context, orderID string) error {
	if err := s.client.CancelOrder(orderID); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	return nil
}
