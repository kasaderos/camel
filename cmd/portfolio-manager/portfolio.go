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
	"time"

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

	stockIDs := make([]model.StockID, len(assets))
	for i, asset := range assets {
		stockIDs[i] = asset.ID
	}

	err = service.CreatePortfolio(
		ctx,
		portfolioID,
		stockIDs,
		cash,
	)
	if err != nil {
		return err
	}

	portfolio, err := service.FetchPortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("fetch portfolio: %w", err)
	}

	portfolio.Print(c.Writer)

	return nil
}

func plan(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	if err := service.PlanRebalance(ctx, c.String("id")); err != nil {
		return err
	}

	fmt.Fprintln(c.Writer, "plan OK")

	return nil
}

func rebalance(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	if err := service.ProcessTasks(ctx, c.String("id")); err != nil {
		return err
	}

	fmt.Fprintln(c.Writer, "rebalance OK")

	return nil
}

func readAssetsCSV(path string) ([]model.Stock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	seen := map[string]struct{}{}

	var out []model.Stock

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
		out = append(out, model.Stock{ID: id})
	}

	if len(out) == 0 {
		return nil, errors.New("csv contains no assets")
	}

	return out, nil
}

func info(ctx context.Context, c *cli.Command) error {
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

	portfolio.Print(c.Writer)

	return nil
}

func score(ctx context.Context, c *cli.Command) error {
	const threshold = 0.01

	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	scores, err := service.FetchPortfolioScore(ctx, c.String("id"))
	if err != nil {
		return fmt.Errorf("fetch portfolio score: %w", err)
	}

	fmt.Fprintf(c.Writer, "threshold: %f\n", threshold)

	for stockID, score := range scores {
		if score > threshold {
			fmt.Fprintf(c.Writer, "%s: %f\n", stockID, score)
		}
	}

	return nil
}

func listTasks(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	service := do.MustInvoke[*portfolio.Service](injector)

	tasks, err := service.FetchRebalanceTasks(ctx, c.String("id"))
	if err != nil {
		return fmt.Errorf("fetch tasks: %w", err)
	}

	for _, task := range tasks {
		fmt.Fprintf(
			c.Writer, "%d:\n  stock: %s\n  side: %s\n  quantity: %f\n  status: %s\n  error: %s\n  created: %s\n",
			task.ID,
			task.StockID,
			task.Side,
			task.Quantity,
			task.Status,
			task.ErrorMessage,
			task.CreatedAt.Format(time.RFC3339),
		)
	}

	return nil
}
