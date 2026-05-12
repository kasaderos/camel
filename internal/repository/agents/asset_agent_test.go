package agents

import (
	"context"
	"testing"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/suite"
)

type AssetAgentTestSuite struct {
	RepositoryTestSuite
	portfolioAgentID string
}

func (s *AssetAgentTestSuite) SetupTest() {
	// Create a portfolio agent for each test
	s.portfolioAgentID = "portfolio-1"
	s.createPortfolioAgent(s.portfolioAgentID)
}

func (s *AssetAgentTestSuite) TestCreateAndFetch() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-1",
		AssetID:          "AAPL",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         10.5,
		Cash:             1000.0,
		State:            model.State{},
	}

	// Create agent
	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Fetch agent
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-1")
	s.Require().NoError(err)

	s.Equal(agent.ID, fetched.ID)
	s.Equal(agent.AssetID, fetched.AssetID)
	s.Equal(agent.AssetQty, fetched.AssetQty)
	s.Equal(agent.Cash, fetched.Cash)
}

func (s *AssetAgentTestSuite) TestFetchNotFound() {
	ctx := context.Background()

	_, err := s.repo.FetchInfo(ctx, "non-existent")
	s.Require().Error(err)
	s.Contains(err.Error(), "agent not found")
}

func (s *AssetAgentTestSuite) TestDeposit() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-2",
		AssetID:          "GOOGL",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         0,
		Cash:             500.0,
		State:            model.State{},
	}

	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Deposit cash
	err = s.repo.Deposit(ctx, "test-agent-2", 250.0)
	s.Require().NoError(err)

	// Verify balance
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-2")
	s.Require().NoError(err)
	s.Equal(750.0, fetched.Cash)
}

func (s *AssetAgentTestSuite) TestWithdraw_Success() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-3",
		AssetID:          "MSFT",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         0,
		Cash:             1000.0,
		State:            model.State{},
	}

	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Withdraw cash
	err = s.repo.Withdraw(ctx, "test-agent-3", 300.0)
	s.Require().NoError(err)

	// Verify balance
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-3")
	s.Require().NoError(err)
	s.Equal(700.0, fetched.Cash)
}

func (s *AssetAgentTestSuite) TestWithdraw_InsufficientFunds() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-4",
		AssetID:          "TSLA",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         0,
		Cash:             100.0,
		State:            model.State{},
	}

	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Try to withdraw more than available
	err = s.repo.Withdraw(ctx, "test-agent-4", 200.0)
	s.Require().Error(err)
	s.Contains(err.Error(), "insufficient funds")

	// Verify balance unchanged
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-4")
	s.Require().NoError(err)
	s.Equal(100.0, fetched.Cash)
}

func (s *AssetAgentTestSuite) TestUpdateState() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-5",
		AssetID:          "NVDA",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         5.0,
		Cash:             500.0,
		State:            model.State{},
	}

	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Update state
	newState := model.State{}
	newState.SetEmaChange(0.05)

	testDate, _ := time.Parse(time.DateOnly, "2024-01-01")
	newState.SetDate(testDate)

	err = s.repo.UpdateState(ctx, "test-agent-5", newState)
	s.Require().NoError(err)

	// Fetch and verify state
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-5")
	s.Require().NoError(err)

	emaChange, ok := fetched.State.EmaChange()
	s.Require().True(ok)
	s.Equal(0.05, emaChange)

	date, ok := fetched.State.Date()
	s.Require().True(ok)
	s.Equal("2024-01-01", date)
}

func (s *AssetAgentTestSuite) TestMultipleOperations() {
	ctx := context.Background()

	agent := &model.AssetAgent{
		ID:               "test-agent-6",
		AssetID:          "AMD",
		PortfolioAgentID: &s.portfolioAgentID,
		AssetQty:         10.0,
		Cash:             2000.0,
		State:            model.State{},
	}

	err := s.repo.CreateAgent(ctx, agent)
	s.Require().NoError(err)

	// Perform multiple operations
	err = s.repo.Deposit(ctx, "test-agent-6", 500.0)
	s.Require().NoError(err)

	err = s.repo.Withdraw(ctx, "test-agent-6", 300.0)
	s.Require().NoError(err)

	err = s.repo.Deposit(ctx, "test-agent-6", 100.0)
	s.Require().NoError(err)

	// Verify final balance: 2000 + 500 - 300 + 100 = 2300
	fetched, err := s.repo.FetchInfo(ctx, "test-agent-6")
	s.Require().NoError(err)
	s.Equal(2300.0, fetched.Cash)
}

func TestAssetAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AssetAgentTestSuite))
}
