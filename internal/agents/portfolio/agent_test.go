package portfolio

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	depmock "github.com/kasaderos/camel/gen/mocks/portfolio"
	"github.com/kasaderos/camel/internal/model"
)

type PortfolioAgentTestSuite struct {
	suite.Suite
	ctx         context.Context
	repo        *depmock.AgentRepository
	assetAgent1 *depmock.AssetAgent
	assetAgent2 *depmock.AssetAgent
	assetAgent3 *depmock.AssetAgent
	agent       *Agent
}

func (s *PortfolioAgentTestSuite) SetupTest() {
	s.ctx = s.T().Context()
	s.repo = depmock.NewAgentRepository(s.T())

	s.assetAgent1 = depmock.NewAssetAgent(s.T())
	s.assetAgent2 = depmock.NewAssetAgent(s.T())
	s.assetAgent3 = depmock.NewAssetAgent(s.T())

	portfolioAgent := model.PortfolioAgent{
		ID:          "portfolio-1",
		PortfolioID: "portfolio-1",
	}

	assetAgents := []AssetAgent{
		s.assetAgent1,
		s.assetAgent2,
		s.assetAgent3,
	}

	s.agent = NewAgent(portfolioAgent, s.repo, assetAgents)
}

func (s *PortfolioAgentTestSuite) TestCoordinate() {
	callCount := 0

	s.assetAgent1.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-1"}).
		Once()

	s.assetAgent2.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-2"}).
		Once()

	s.assetAgent3.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-3"}).
		Once()

	err := s.agent.Coordinate(s.ctx, func(ctx context.Context, agent AssetAgent) error {
		agent.FetchInfo(ctx)
		callCount++
		return nil
	})

	s.NoError(err)
	s.Equal(3, callCount)
}

func (s *PortfolioAgentTestSuite) TestPortfolio_CalculatesWeights() {
	// Setup agents with different EMA changes
	state1 := model.State{}
	state1.SetEmaChange(0.05) // 5% above threshold
	s.assetAgent1.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).
		Once()
	s.assetAgent1.EXPECT().
		FetchState(s.ctx).
		Return(state1).
		Once()

	state2 := model.State{}
	state2.SetEmaChange(0.03) // 3% above threshold
	s.assetAgent2.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).
		Once()
	s.assetAgent2.EXPECT().
		FetchState(s.ctx).
		Return(state2).
		Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold (0.02)
	s.assetAgent3.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).
		Once()
	s.assetAgent3.EXPECT().
		FetchState(s.ctx).
		Return(state3).
		Once()

	weights, err := s.agent.Portfolio(s.ctx, 0.02)

	s.NoError(err)
	s.Len(weights, 2) // Only AAPL and GOOGL should be included

	// Weights should sum to 1.0
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	s.InDelta(1.0, totalWeight, 0.0001)

	// AAPL should have higher weight than GOOGL
	s.Greater(weights["AAPL"], weights["GOOGL"])
}

func (s *PortfolioAgentTestSuite) TestPortfolio_NoValidAssets() {
	// All assets below threshold
	state1 := model.State{}
	state1.SetEmaChange(0.01)
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchState(s.ctx).Return(state1).Once()

	state2 := model.State{}
	state2.SetEmaChange(-0.02)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchState(s.ctx).Return(state2).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.005)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchState(s.ctx).Return(state3).Once()

	weights, err := s.agent.Portfolio(s.ctx, 0.02)

	s.NoError(err)
	s.Empty(weights)
}

func (s *PortfolioAgentTestSuite) TestRebalance_FullCycle() {
	// Step 1: Update states
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(nil).Once()

	// Step 2: Portfolio calculation - setup states
	state1 := model.State{}
	state1.SetEmaChange(0.06) // 60% of total score
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchState(s.ctx).Return(state1).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% of total score
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchState(s.ctx).Return(state2).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold, not included
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchState(s.ctx).Return(state3).Once()

	// Step 3: Calculate portfolio value using FetchTotalSum
	// Total: 1600 + 1050 + 200 = 2850
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(200.0, nil).Once()

	// Step 4: Sell/Withdraw pass
	// AAPL: target 60% of 2850 = 1710, current 1600, no withdrawal needed
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()

	// GOOGL: target 40% of 2850 = 1140, current 1050, no withdrawal needed
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()

	// MSFT: not in target weights, close position
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 200.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 200 from MSFT)
	// AAPL: needs 110 more (1710 - 1600), freeCash = 200, can deposit
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()
	s.assetAgent1.EXPECT().DepositWithBuy(s.ctx, 110.0).Return(model.AssetAgent{}, nil).Once()

	// GOOGL: needs 90 more (1140 - 1050), freeCash = 90 (200 - 110), can deposit
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 90.0).Return(model.AssetAgent{}, nil).Once()

	// MSFT: not in weights, but still needs FetchInfo and FetchTotalSum in buy pass
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(0.0, nil).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}

func (s *PortfolioAgentTestSuite) TestRebalance_UpdateStateError() {
	s.assetAgent1.EXPECT().
		UpdateState(s.ctx).
		Return(errors.New("state update failed")).
		Once()

	err := s.agent.Rebalance(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "update asset agent states")
}

func (s *PortfolioAgentTestSuite) TestRebalance_PriceError() {
	// Step 1: Update states
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(nil).Once()

	// Step 2: Portfolio calculation
	state1 := model.State{}
	state1.SetEmaChange(0.05)
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchState(s.ctx).Return(state1).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.03)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchState(s.ctx).Return(state2).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchState(s.ctx).Return(state3).Once()

	// Step 3: FetchTotalSum fails
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(0.0, errors.New("fetch total sum failed")).Once()

	err := s.agent.Rebalance(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "agent fetch total sum")
}

func (s *PortfolioAgentTestSuite) TestRebalance_SellPartial() {
	// Test case where an asset needs to be partially sold using WithdrawWithSell
	// Step 1: Update states
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(nil).Once()

	// Step 2: Portfolio calculation - AAPL weight decreases from 70% to 60%
	state1 := model.State{}
	state1.SetEmaChange(0.06) // 60% target
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchState(s.ctx).Return(state1).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% target
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchState(s.ctx).Return(state2).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchState(s.ctx).Return(state3).Once()

	// Step 3: Calculate portfolio value
	// Total = 1400
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1000.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(0.0, nil).Once()

	// Step 4: Withdraw/Sell pass
	// AAPL target: 60% of 1400 = 840 (currently 1000, need to withdraw 160)
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1000.0, nil).Once()
	s.assetAgent1.EXPECT().WithdrawWithSell(s.ctx, 160.0).Return(model.AssetAgent{Cash: 160.0}, nil).Once()

	// GOOGL: target 40% = 560 (currently 400, no withdrawal)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()

	// MSFT: not in weights, close position
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 0.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 160 from AAPL)
	// AAPL: no deposit needed
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(840.0, nil).Once()

	// GOOGL target: 560 (currently 400, needs 160, freeCash = 160)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 160.0).Return(model.AssetAgent{}, nil).Once()

	// MSFT: not in weights, but still needs FetchInfo and FetchTotalSum in buy pass
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(0.0, nil).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}

func TestPortfolioAgentTestSuite(t *testing.T) {
	suite.Run(t, new(PortfolioAgentTestSuite))
}
