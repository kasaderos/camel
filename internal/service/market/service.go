package market

type Service struct {
	market MarketProvider
	repo   Repository
}

func New(market MarketProvider, repo Repository) *Service {
	return &Service{
		market: market,
		repo:   repo,
	}
}
