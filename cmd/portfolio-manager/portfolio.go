package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/internal/service/portfolio"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
)

func createPortfolio(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	csvName := c.String("csv")
	portfolioID := c.String("id")
	cash := c.Float64("cash")

	assets, err := readAssetsCSV(csvName)
	if err != nil {
		return err
	}

	slog.Info("create portfolio", "portfolioID", portfolioID, "cash", cash, "csv", csvName)

	err = service.CreatePortfolio(
		ctx,
		portfolioID,
		assets,
		cash,
	)
	if err != nil {
		return err
	}

	portfolio, err := service.FetchPortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("fetch portfolio: %w", err)
	}

	err = service.Rebalance(ctx, portfolio)
	if err != nil {
		return fmt.Errorf("rebalance: %w", err)
	}

	portfolio.Print(c.Writer)

	return nil
}

func portfolioInfo(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	portfolio, err := service.FetchPortfolio(ctx, c.String("id"))
	if err != nil {
		return err
	}

	portfolio.Print(c.Writer)

	return nil
}

func rebalance(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	portfolio, err := service.FetchPortfolio(ctx, c.String("id"))
	if err != nil {
		return fmt.Errorf("fetch portfolio: %w", err)
	}

	err = service.Rebalance(
		ctx,
		portfolio,
	)
	if err != nil {
		return err
	}

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
