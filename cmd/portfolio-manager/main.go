package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "portfolio-manager",
		Usage: "Manage portfolios",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a portfolio from a CSV file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "portfolio id",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "csv",
						Usage:    "CSV file containing asset IDs",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "cash",
						Usage:    "initial balance",
						Required: true,
					},
				},
				Action: createPortfolio,
			},
			{
				Name:  "info",
				Usage: "Displaya portfolio info by agent id and print details",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "Portfolio agent id",
						Required: true,
					},
				},
				Action: portfolioInfo,
			},
			{
				Name:  "rebalance",
				Usage: "Rebalance a portfolio by id",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "Portfolio id",
						Required: true,
					},
				},
				Action: rebalance,
			},
			{
				Name:  "score",
				Usage: "Prints the current score of portfolio by id",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "Portfolio id",
						Required: true,
					},
				},
				Action: score,
			},
			{
				Name:   "migrate-up",
				Usage:  "Initialize database tables",
				Flags:  []cli.Flag{},
				Action: migrateUp,
			},
			{
				Name:   "migrate-drop",
				Usage:  "Remove everything in database",
				Flags:  []cli.Flag{},
				Action: migrateDrop,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("app run", "err", err)
		os.Exit(1)
	}
}
