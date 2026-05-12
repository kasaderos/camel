package agents

import (
	"context"
	"testing"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetAgent_CreateAndFetch(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// First create a portfolio agent directly in DB
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-1",
		AssetID:          "AAPL",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         10.5,
		Cash:             1000.0,
		State:            model.State{},
	}

	// Create agent
	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Fetch agent
	fetched, err := repo.FetchInfo(ctx, "test-agent-1")
	require.NoError(t, err)

	assert.Equal(t, agent.ID, fetched.ID)
	assert.Equal(t, agent.AssetID, fetched.AssetID)
	assert.Equal(t, agent.AssetQty, fetched.AssetQty)
	assert.Equal(t, agent.Cash, fetched.Cash)
}

func TestAssetAgent_FetchNotFound(t *testing.T) {
	repo, _ := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.FetchInfo(ctx, "non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestAssetAgent_Deposit(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// Create portfolio agent
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-2",
		AssetID:          "GOOGL",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         0,
		Cash:             500.0,
		State:            model.State{},
	}

	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Deposit cash
	err = repo.Deposit(ctx, "test-agent-2", 250.0)
	require.NoError(t, err)

	// Verify balance
	fetched, err := repo.FetchInfo(ctx, "test-agent-2")
	require.NoError(t, err)
	assert.Equal(t, 750.0, fetched.Cash)
}

func TestAssetAgent_Withdraw_Success(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// Create portfolio agent
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-3",
		AssetID:          "MSFT",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         0,
		Cash:             1000.0,
		State:            model.State{},
	}

	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Withdraw cash
	err = repo.Withdraw(ctx, "test-agent-3", 300.0)
	require.NoError(t, err)

	// Verify balance
	fetched, err := repo.FetchInfo(ctx, "test-agent-3")
	require.NoError(t, err)
	assert.Equal(t, 700.0, fetched.Cash)
}

func TestAssetAgent_Withdraw_InsufficientFunds(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// Create portfolio agent
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-4",
		AssetID:          "TSLA",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         0,
		Cash:             100.0,
		State:            model.State{},
	}

	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Try to withdraw more than available
	err = repo.Withdraw(ctx, "test-agent-4", 200.0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	// Verify balance unchanged
	fetched, err := repo.FetchInfo(ctx, "test-agent-4")
	require.NoError(t, err)
	assert.Equal(t, 100.0, fetched.Cash)
}

func TestAssetAgent_UpdateState(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// Create portfolio agent
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-5",
		AssetID:          "NVDA",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         5.0,
		Cash:             500.0,
		State:            model.State{},
	}

	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Update state
	newState := model.State{}
	newState.SetEmaChange(0.05)
	testDate, _ := time.Parse(time.DateOnly, "2024-01-01")
	newState.SetDate(testDate)

	err = repo.UpdateState(ctx, "test-agent-5", newState)
	require.NoError(t, err)

	// Fetch and verify state
	fetched, err := repo.FetchInfo(ctx, "test-agent-5")
	require.NoError(t, err)

	emaChange, ok := fetched.State.EmaChange()
	require.True(t, ok)
	assert.Equal(t, 0.05, emaChange)

	date, ok := fetched.State.Date()
	require.True(t, ok)
	assert.Equal(t, "2024-01-01", date)
}

func TestAssetAgent_MultipleOperations(t *testing.T) {
	repo, pgContainer := setupTestDB(t)
	ctx := context.Background()

	// Create portfolio agent
	_, err := pgContainer.DB.ExecContext(ctx, `
		INSERT INTO portfolio_agents (id, portfolio_id)
		VALUES ('portfolio-1', 'portfolio-1')
	`)
	require.NoError(t, err)

	portfolioAgentID := "portfolio-1"
	agent := &model.AssetAgent{
		ID:               "test-agent-6",
		AssetID:          "AMD",
		PortfolioAgentID: &portfolioAgentID,
		AssetQty:         10.0,
		Cash:             2000.0,
		State:            model.State{},
	}

	err = repo.CreateAgent(ctx, agent)
	require.NoError(t, err)

	// Perform multiple operations
	err = repo.Deposit(ctx, "test-agent-6", 500.0)
	require.NoError(t, err)

	err = repo.Withdraw(ctx, "test-agent-6", 300.0)
	require.NoError(t, err)

	err = repo.Deposit(ctx, "test-agent-6", 100.0)
	require.NoError(t, err)

	// Verify final balance: 2000 + 500 - 300 + 100 = 2300
	fetched, err := repo.FetchInfo(ctx, "test-agent-6")
	require.NoError(t, err)
	assert.Equal(t, 2300.0, fetched.Cash)
}
