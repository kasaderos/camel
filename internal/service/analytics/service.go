package analytics

type Service struct {
	market MarketService
}

func NewService(
	market MarketService,
) *Service {
	return &Service{market: market}
}
