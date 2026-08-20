package portfolio

import "github.com/kasaderos/camel/internal/model"

func (s *RepositorySuite) TestUpdatePortfolio() {
	ctx := s.T().Context()

	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
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
	})
	s.Require().NoError(err)

	updated := model.Portfolio{
		ID:   "portfolio-1",
		Cash: 8500,
		Cost: 1500,
		Stocks: []model.PortfolioStock{
			{
				StockID:    "AAPL",
				EntryPrice: 150,
				Quantity:   10,
			},
			{
				StockID:    "MSFT",
				EntryPrice: 300,
				Quantity:   0,
			},
		},
	}

	err = s.repo.UpdatePortfolio(ctx, updated)
	s.Require().NoError(err)

	fetched, err := s.repo.FetchPortfolio(ctx, updated.ID)
	s.Require().NoError(err)
	s.Equal(updated.Cash, fetched.Cash)
	s.Equal(updated.Cost, fetched.Cost)
	s.Equal(10.0, fetched.Stocks[0].Quantity)
	s.Equal(150.0, fetched.Stocks[0].EntryPrice)
}

func (s *RepositorySuite) TestUpdatePortfolioStock() {
	ctx := s.T().Context()

	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:   "portfolio-1",
		Cash: 10000,
		Stocks: []model.PortfolioStock{
			{StockID: "MSFT", EntryPrice: 0, Quantity: 0},
		},
	})
	s.Require().NoError(err)

	err = s.repo.UpdatePortfolioStock(ctx, "portfolio-1", model.PortfolioStock{
		StockID:    "MSFT",
		EntryPrice: 310,
		Quantity:   2,
	})
	s.Require().NoError(err)

	fetched, err := s.repo.FetchPortfolio(ctx, "portfolio-1")
	s.Require().NoError(err)
	s.Require().Len(fetched.Stocks, 1)
	s.Equal(2.0, fetched.Stocks[0].Quantity)
	s.Equal(310.0, fetched.Stocks[0].EntryPrice)
}

func (s *RepositorySuite) TestUpdatePortfolio_NotFound() {
	ctx := s.T().Context()

	err := s.repo.UpdatePortfolio(ctx, model.Portfolio{
		ID:   "missing",
		Cash: 1,
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "portfolio not found")
}
