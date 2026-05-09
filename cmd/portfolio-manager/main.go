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
						Name:     "csv",
						Usage:    "CSV file containing asset IDs (one per line or in the first column)",
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
				Usage: "Rebalance a portfolio by agent id (updates state of all asset agents)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "Portfolio agent id",
						Required: true,
					},
				},
				Action: rebalance,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("app run", "err", err)
		os.Exit(1)
	}
}
