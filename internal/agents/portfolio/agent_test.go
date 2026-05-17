package portfolio

import (
	"context"
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

	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-1"}).Once()
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-2"}).Once()
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{ID: "agent-3"}).Once()

	err := s.agent.Coordinate(s.ctx, func(ctx context.Context, agent AssetAgent) error {
		agent.FetchInfo(ctx)
		callCount++
		return nil
	})

	s.NoError(err)
	s.Equal(3, callCount)
}

func (s *PortfolioAgentTestSuite) TestPortfolio_CalculatesWeights() {
	// Setup States with valid EMA changes
	st1 := model.State{}
	st1.SetEmaChange(0.05)
	st2 := model.State{}
	st2.SetEmaChange(0.03)
	st3 := model.State{}
	st3.SetEmaChange(0.01) // Below threshold (0.02)

	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "AAPL", State: st1}).Once()
	s.assetAgent1.EXPECT().FetchPrice(s.ctx).Return(150.0, nil).Once()

	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "GOOGL", State: st2}).Once()
	s.assetAgent2.EXPECT().FetchPrice(s.ctx).Return(100.0, nil).Once()

	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "MSFT", State: st3}).Once()
	// FetchPrice is NOT called for MSFT — it is filtered out before that point.

	portfolio, err := s.agent.FetchPortfolio(s.ctx)

	s.NoError(err)
	s.Len(portfolio.Assets, 2)

	// Total score = 0.05 + 0.03 = 0.08. AAPL weight = 0.05 / 0.08 = 0.625
	s.Equal(0.625, portfolio.Assets["AAPL"].Weight)
	// Assets is map[string]PortfolioAsset (value type); use NotContains to check absence.
	s.NotContains(portfolio.Assets, "MSFT", "MSFT should be filtered out by threshold")
}

func (s *PortfolioAgentTestSuite) TestRebalance_FullCycle() {
	// --- STAGE 1: Initial FetchPortfolio (Current State) ---
	appleState := model.State{}
	appleState.SetEmaChange(0.03)

	googState := model.State{}
	googState.SetEmaChange(0.04)

	msftState := model.State{}
	msftState.SetEmaChange(0.05)

	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "AAPL",
			AssetQty: 6,
			Cash:     100,
		}).Once()
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "GOOGL",
			AssetQty: 3,
			Cash:     200,
		}).Once()
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "MSFT",
			AssetQty: 10,
			State:    msftState,
		}).Once()
	s.assetAgent3.EXPECT().FetchPrice(s.ctx).Return(100.0, nil).Once()
	// Initial portfolio: MSFT price=100, qty=10, Sum=1000, weight=1.0

	// --- STAGE 2: UpdatePortfolio (Target State) ---
	s.assetAgent1.EXPECT().UpdateState(s.ctx).Return(model.AssetAgent{}, nil).Once()
	s.assetAgent2.EXPECT().UpdateState(s.ctx).Return(model.AssetAgent{}, nil).Once()
	s.assetAgent3.EXPECT().UpdateState(s.ctx).Return(model.AssetAgent{}, nil).Once()

	newAppleState := model.State{}
	newAppleState.SetEmaChange(0.03)

	newGoogState := model.State{}
	newGoogState.SetEmaChange(0.04)

	newMsftState := model.State{}
	newMsftState.SetEmaChange(0.05)

	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "AAPL",
			State:    newAppleState,
			AssetQty: 6,
		}).Once()

	s.assetAgent1.EXPECT().FetchPrice(s.ctx).Return(100.0, nil).Once()
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "GOOGL",
			State:    newGoogState,
			AssetQty: 3,
		}).Once()

	s.assetAgent2.EXPECT().FetchPrice(s.ctx).Return(100.0, nil).Once()
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(
		model.AssetAgent{
			AssetID:  "MSFT",
			State:    newMsftState,
			AssetQty: 10,
		}).Once()

	// --- STAGE 3: Rebalance Pass 1 (Sell/Withdraw) ---
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "AAPL"}).Once()
	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "GOOGL"}).Once()
	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "MSFT"}).Once()
	s.assetAgent3.EXPECT().ClosePosition(s.ctx).Return(model.AssetAgent{Cash: 1000.0}, nil).Once()

	// --- STAGE 4: Rebalance Pass 2 (Buy/Deposit) ---
	// freeCash=1000; deficitSum(AAPL)=600-0=600; deficitSum(GOOGL)=400-0=400.
	s.assetAgent1.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "AAPL"}).Once()
	s.assetAgent1.EXPECT().DepositWithBuy(s.ctx, 600.0).Return(model.AssetAgent{}, nil).Once()

	s.assetAgent2.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "GOOGL"}).Once()
	s.assetAgent2.EXPECT().DepositWithBuy(s.ctx, 400.0).Return(model.AssetAgent{}, nil).Once()

	s.assetAgent3.EXPECT().FetchInfo(s.ctx).Return(model.AssetAgent{AssetID: "MSFT"}).Once()

	err := s.agent.Rebalance(s.ctx)
	s.NoError(err)
}

func TestPortfolioAgentTestSuite(t *testing.T) {
	suite.Run(t, new(PortfolioAgentTestSuite))
}
