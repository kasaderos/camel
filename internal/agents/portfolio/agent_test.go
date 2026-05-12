package portfolio

import (
	"context"
	"errors"
	"testing"

	depmock "github.com/kasaderos/camel/gen/mocks/portfolio"
	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/suite"
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
		Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", State: state1}).
		Once()

	state2 := model.State{}
	state2.SetEmaChange(0.03) // 3% above threshold
	s.assetAgent2.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", State: state2}).
		Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold (0.02)
	s.assetAgent3.EXPECT().
		FetchInfo(s.ctx).
		Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", State: state3}).
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
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(-0.02)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.005)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

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
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% of total score
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold, not included
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

	// Step 3: Calculate portfolio value using FetchTotalSum
	// Total: 1600 + 1050 + 200 = 2850
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(200.0, nil).Once()

	// Step 4: Sell/Withdraw pass
	// AAPL: target 60% of 2850 = 1710, current 1600, no withdrawal needed
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()

	// GOOGL: target 40% of 2850 = 1140, current 1050, no withdrawal needed
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 8.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()

	// MSFT: not in target weights, close position
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 200.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 200 from MSFT)
	// AAPL: needs 110 more (1710 - 1600), freeCash = 200, can deposit
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1600.0, nil).Once()
	s.assetAgent1.EXPECT().DepositWithBuy(s.ctx, 110.0).Return(model.AssetAgent{}, nil).Once()

	// GOOGL: needs 90 more (1140 - 1050), freeCash = 90 (200 - 110), can deposit
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 8.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(1050.0, nil).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 90.0).Return(model.AssetAgent{}, nil).Once()

	// MSFT: not in weights, only needs FetchInfo (early exit)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 0.0}).Once()

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
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.03)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

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
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% target
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

	// Step 3: Calculate portfolio value
	// Total = 1400
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1000.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(0.0, nil).Once()

	// Step 4: Withdraw/Sell pass
	// AAPL target: 60% of 1400 = 840 (currently 1000, need to withdraw 160)
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(1000.0, nil).Once()
	s.assetAgent1.EXPECT().WithdrawWithSell(s.ctx, 160.0).Return(model.AssetAgent{Cash: 160.0}, nil).Once()

	// GOOGL: target 40% = 560 (currently 400, no withdrawal)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()

	// MSFT: not in weights, close position
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 0.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 160 from AAPL)
	// AAPL: no deposit needed
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(840.0, nil).Once()

	// GOOGL target: 560 (currently 400, needs 160, freeCash = 160)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(400.0, nil).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 160.0).Return(model.AssetAgent{}, nil).Once()

	// MSFT: not in weights, only needs FetchInfo (early exit)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}
func (s *PortfolioAgentTestSuite) TestRebalance_BuyPartial() {
	// Test scenario: Multiple assets need to buy, but insufficient cash
	// Only the first asset gets the available cash, second asset skipped

	// Step 1: Update states
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(nil).Once()

	// Step 2: Portfolio calculation - both AAPL and GOOGL need to buy
	state1 := model.State{}
	state1.SetEmaChange(0.06) // 60% target weight
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% target weight
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold, not included
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

	// Step 3: Calculate portfolio value
	// Total portfolio = 2000
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(800.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(500.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(700.0, nil).Once()

	// Step 4: Sell/Withdraw pass
	// AAPL: target 60% of 2000 = 1200, current 800, needs to buy 400
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(800.0, nil).Once()

	// GOOGL: target 40% of 2000 = 800, current 500, needs to buy 300
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(500.0, nil).Once()

	// MSFT: not in target weights, close position (only 200 cash available)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 3.0}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 200.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 200)
	// AAPL: needs 400, but only 200 available, can't deposit (200 < 400)
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(800.0, nil).Once()

	// GOOGL: needs 300, but only 200 available, can't deposit (200 < 300)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(500.0, nil).Once()

	// MSFT: not in weights, skip
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}

func (s *PortfolioAgentTestSuite) TestRebalance_OnlyQuantitiesNoCash() {
	// Test scenario: agents hold asset quantities but have zero cash
	// Rebalancing requires selling from one asset to buy another

	// Step 1: Update states
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(nil).Once()

	// Step 2: Portfolio calculation
	state1 := model.State{}
	state1.SetEmaChange(0.06) // 60% target weight
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0, State: state1}).Once()

	state2 := model.State{}
	state2.SetEmaChange(0.04) // 40% target weight
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0, State: state2}).Once()

	state3 := model.State{}
	state3.SetEmaChange(0.01) // Below threshold, not included
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0, State: state3}).Once()

	// Step 3: Calculate portfolio value (all in asset quantities, no cash)
	// AAPL: 100 shares @ $150 = $15,000 (total value from qty only, cash = 0)
	// GOOGL: 50 shares @ $100 = $5,000 (total value from qty only, cash = 0)
	// MSFT: 20 shares @ $200 = $4,000 (total value from qty only, cash = 0)
	// Total portfolio = $24,000
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(15000.0, nil).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(5000.0, nil).Once()
	s.assetAgent3.EXPECT().FetchTotalSum(s.ctx).Return(4000.0, nil).Once()

	// Step 4: Sell/Withdraw pass
	// AAPL: target 60% of 24000 = 14400, current 15000, withdraw 600
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(15000.0, nil).Once()
	s.assetAgent1.EXPECT().WithdrawWithSell(s.ctx, 600.0).Return(model.AssetAgent{Cash: 600.0}, nil).Once()

	// GOOGL: target 40% of 24000 = 9600, current 5000, no withdrawal (needs buy)
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(5000.0, nil).Once()

	// MSFT: not in target weights, close position (sell all, get cash)
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 4000.0}, nil).Once()

	// Step 5: Deposit/Buy pass (freeCash = 600 + 4000 = 4600)
	// AAPL: already at target after selling, no action
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1", AssetID: "AAPL", AssetQty: 10.0}).Once()
	s.assetAgent1.EXPECT().FetchTotalSum(s.ctx).Return(14400.0, nil).Once()

	// GOOGL: needs 4600 more (9600 - 5000), freeCash = 4600, can deposit all
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2", AssetID: "GOOGL", AssetQty: 5.0}).Once()
	s.assetAgent2.EXPECT().FetchTotalSum(s.ctx).Return(5000.0, nil).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 4600.0).Return(model.AssetAgent{}, nil).Once()

	// MSFT: not in weights, skip
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3", AssetID: "MSFT", AssetQty: 2.0}).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}

func TestPortfolioAgentTestSuite(t *testing.T) {
	suite.Run(t, new(PortfolioAgentTestSuite))
}
