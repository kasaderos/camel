package portfolio

import (
	"context"
	"testing"

	"github.com/kasaderos/camel/internal/service/portfolio/mocks"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	ctx       context.Context
	exchange  *mocks.Exchanger
	repo      *mocks.PortfolioRepository
	analytics *mocks.AnalyticsService
	taskRepo  *mocks.TaskRepository
	svc       *Service
}

func (s *TestSuite) SetupTest() {
	s.ctx = s.T().Context()
	s.exchange = mocks.NewExchanger(s.T())
	s.repo = mocks.NewPortfolioRepository(s.T())
	s.analytics = mocks.NewAnalyticsService(s.T())
	s.taskRepo = mocks.NewTaskRepository(s.T())
	s.svc = &Service{
		exchange:  s.exchange,
		repo:      s.repo,
		analytics: s.analytics,
		taskRepo:  s.taskRepo,
	}
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
