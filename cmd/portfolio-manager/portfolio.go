package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kasaderos/camel/internal/agents/portfolio"
	"github.com/kasaderos/camel/internal/model"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
)

func createPortfolio(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	agent := do.MustInvoke[*portfolio.Agent](injector)

	assets, err := readAssetsCSV(c.String("csv"))
	if err != nil {
		return err
	}

	err = agent.CreatePortfolio(ctx, assets)
	if err != nil {
		return err
	}

	agent.PrintInfo(ctx, c.Writer)

	return nil
}

func portfolioInfo(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	agent := do.MustInvoke[*portfolio.Agent](injector)

	err = agent.Initialize(ctx, c.String("id"))
	if err != nil {
		return err
	}

	agent.PrintInfo(ctx, c.Writer)

	return nil
}

func rebalance(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	agent := do.MustInvoke[*portfolio.Agent](injector)

	if err := agent.Rebalance(ctx, c.String("id")); err != nil {
		return err
	}

	err = agent.Rebalance(ctx, c.String("id"))
	if err != nil {
		return err
	}

	agent.PrintInfo(ctx, c.Writer)

	fmt.Fprintln(c.Writer, "rebalance OK")
	return nil
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
