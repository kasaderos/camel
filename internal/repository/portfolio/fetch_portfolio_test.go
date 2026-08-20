package portfolio

func (s *RepositorySuite) TestFetchPortfolio_NotFound() {
	ctx := s.T().Context()

	_, err := s.repo.FetchPortfolio(ctx, "missing")
	s.Require().Error(err)
	s.Contains(err.Error(), "portfolio not found")
}
