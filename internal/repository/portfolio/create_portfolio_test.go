package portfolio

import "github.com/kasaderos/camel/internal/model"

func (s *RepositorySuite) TestCreatePortfolio() {
	ctx := s.T().Context()

	original := model.Portfolio{
		ID:   "portfolio-1",
		Cash: 10000,
		Cost: 0,
		Stocks: []model.PortfolioStock{
			{
				StockID:    "AAPL",
				EntryPrice: 0,
				Quantity:   0,
			},
			{
				StockID:    "MSFT",
				EntryPrice: 0,
				Quantity:   0,
			},
		},
	}

	err := s.repo.CreatePortfolio(ctx, original)
	s.Require().NoError(err)

	fetched, err := s.repo.FetchPortfolio(ctx, original.ID)
	s.Require().NoError(err)
	s.Equal(original.ID, fetched.ID)
	s.Equal(original.Cash, fetched.Cash)
	s.Equal(original.Cost, fetched.Cost)
	s.Len(fetched.Stocks, 2)
	s.Equal("AAPL", fetched.Stocks[0].StockID)
	s.Equal("MSFT", fetched.Stocks[1].StockID)
}
