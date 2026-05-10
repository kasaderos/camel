package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kasaderos/camel/internal/model"
	portfolioservice "github.com/kasaderos/camel/internal/service/agents/portfolio"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
)

func createPortfolio(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	svc := do.MustInvoke[*portfolioservice.PortfolioAgentService](injector)

	assets, err := readAssetsCSV(c.String("csv"))
	if err != nil {
		return err
	}

	agent, err := svc.CreatePortfolio(ctx, assets)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.Writer, "portfolio_agent_id=%s portfolio_id=%s asset_agents=%d\n", agent.ID, agent.PortfolioID, len(agent.AssetAgents))

	return nil
}

func portfolioInfo(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	svc := do.MustInvoke[*portfolioservice.PortfolioAgentService](injector)

	agent, err := svc.Fetch(ctx, c.String("id"))
	if err != nil {
		return err
	}

	printPortfolio(c.Writer, agent)
	return nil
}

func rebalance(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	svc := do.MustInvoke[*portfolioservice.PortfolioAgentService](injector)

	if err := svc.Rebalance(ctx, c.String("id")); err != nil {
		return err
	}

	agent, err := svc.Fetch(ctx, c.String("id"))
	if err != nil {
		return err
	}
	printPortfolio(c.Writer, agent)

	fmt.Fprintln(c.Writer, "rebalance OK")
	return nil
}

func printPortfolio(w io.Writer, agent model.PortfolioAgent) {
	fmt.Fprintf(w, "portfolio_agent_id=%s portfolio_id=%s created_at=%s updated_at=%s\n", agent.ID, agent.PortfolioID, agent.CreatedAt.Format(time.RFC3339), agent.UpdatedAt.Format(time.RFC3339))
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "asset agents:")

	for _, a := range agent.AssetAgents {
		fmt.Fprintf(w, "- id=%s asset_id=%s asset_qty=%.4f cash=%.2f state=%v\n", a.ID, a.AssetID, a.AssetQty, a.Cash, a.State)
	}

	type summary struct {
		AssetID string
		Count   int
		Qty     float64
		Cash    float64
	}

	byAsset := map[string]*summary{}
	for _, a := range agent.AssetAgents {
		s := byAsset[a.AssetID]
		if s == nil {
			s = &summary{AssetID: a.AssetID}
			byAsset[a.AssetID] = s
		}
		s.Count++
		s.Qty += a.AssetQty
		s.Cash += a.Cash
	}

	keys := make([]string, 0, len(byAsset))
	for k := range byAsset {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintln(w, "")

	weights := agent.Portfolio(0)
	if len(weights) > 0 {
		wKeys := make([]string, 0, len(weights))
		for k := range weights {
			wKeys = append(wKeys, k)
		}
		sort.Strings(wKeys)

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "portfolio weights:")
		for _, k := range wKeys {
			fmt.Fprintf(w, "- asset_id=%s weight=%.4f\n", k, weights[k])
		}
	}
}

func readAssetsCSV(path string) ([]model.Asset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	seen := map[string]struct{}{}
	var out []model.Asset
	for {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read csv: %w", err)
		}
		if len(rec) == 0 {
			continue
		}

		id := strings.TrimSpace(rec[0])
		if id == "" {
			continue
		}
		if strings.EqualFold(id, "asset_id") || strings.EqualFold(id, "symbol") {
			// header row
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, model.Asset{ID: id})
	}

	if len(out) == 0 {
		return nil, errors.New("csv contains no assets")
	}

	return out, nil
}
