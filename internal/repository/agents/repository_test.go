package agents

import (
	"context"

	"github.com/kasaderos/camel/pkg/testutils/testsuites"
)

// RepositoryTestSuite is a common test suite for all repository tests
// It provides a shared PostgreSQL database and repository instance
type RepositoryTestSuite struct {
	testsuites.PostgresSuite
	repo *AgentRepository
}

// SetupSuite runs once before all tests in the suite
func (s *RepositoryTestSuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()
	s.RunMigrations("../../../migrations")
	s.repo = New(s.DB)
}

// TearDownTest runs after each test to clean up
func (s *RepositoryTestSuite) TearDownTest() {
	s.Truncate("asset_agents", "portfolio_agents")
}

// createPortfolioAgent is a helper to create a portfolio agent for tests
func (s *RepositoryTestSuite) createPortfolioAgent(id string) {
	_, err := s.DB.ExecContext(context.Background(), `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ($1, $1)
	`, id)
	s.Require().NoError(err)
}
