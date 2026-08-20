package market

import (
	"context"
	"testing"

	"github.com/kasaderos/camel/internal/service/market/mocks"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	ctx    context.Context
	market *mocks.MarketProvider
	repo   *mocks.Repository
	svc    *Service
}

func (s *TestSuite) SetupTest() {
	s.ctx = s.T().Context()
	s.market = mocks.NewMarketProvider(s.T())
	s.repo = mocks.NewRepository(s.T())
	s.svc = New(s.market, s.repo)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
