package model

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type Portfolio struct {
	ID      string
	Weights map[string]float64
	Cash    float64
}

type PortfolioAsset struct {
	AssetID string
	Weight  float64
}

// Print writes a formatted table of the portfolio to the provided io.Writer
func (p Portfolio) Print(w io.Writer) {
	// Initialize tabwriter: minwidth, tabwidth, padding, padchar, flags
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	// Print Header
	fmt.Fprintln(tw, "ASSET ID\tWEIGHT")
	fmt.Fprintln(tw, "--------\t-------")

	// Sort assetIDs for consistent output ordering
	assetIDs := make([]string, 0, len(p.Weights))
	for k := range p.Weights {
		assetIDs = append(assetIDs, k)
	}
	sort.Strings(assetIDs)

	for _, assetID := range assetIDs {
		weight := p.Weights[assetID]
		fmt.Fprintf(tw, "%s\t%.2f\n",
			assetID,
			weight, // Assuming Weight
		)
	}

	fmt.Fprintln(tw, "--------\t-------")

	tw.Flush()
}
