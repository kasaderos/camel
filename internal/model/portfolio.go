package model

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type Portfolio struct {
	Assets map[string]PortfolioAsset
}

type PortfolioAsset struct {
	AssetID string
	Price   float64
	Qty     float64
	Weight  float64
}

func (pa PortfolioAsset) Sum() float64 {
	return pa.Price * pa.Qty
}

func (p Portfolio) Asset(assetID string) (PortfolioAsset, bool) {
	asset, exist := p.Assets[assetID]

	return asset, exist
}

func (p Portfolio) AssetSum(assetID string) (float64, bool) {
	asset, exist := p.Assets[assetID]

	return asset.Sum(), exist
}

func (p Portfolio) TotalSum() float64 {
	totalSum := 0.0

	for _, asset := range p.Assets {
		totalSum += asset.Price + asset.Qty
	}

	return totalSum
}

// Print writes a formatted table of the portfolio to the provided io.Writer
func (p Portfolio) Print(w io.Writer) {
	// Initialize tabwriter: minwidth, tabwidth, padding, padchar, flags
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	// Print Header
	fmt.Fprintln(tw, "ASSET ID\tPRICE\tQTY\tWEIGHT\tTOTAL")
	fmt.Fprintln(tw, "--------\t-----\t---\t------\t-----")

	// Sort keys for consistent output ordering
	keys := make([]string, 0, len(p.Assets))
	for k := range p.Assets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		asset := p.Assets[k]
		fmt.Fprintf(tw, "%s\t%.2f\t%.2f\t%.2f%%\t%.2f\n",
			asset.AssetID,
			asset.Price,
			asset.Qty,
			asset.Weight, // Assuming Weight
			asset.Sum(),
		)
	}

	// Calculate and print the footer
	fmt.Fprintln(tw, "--------\t-----\t---\t------\t-----")
	fmt.Fprintf(tw, "TOTAL\t\t\t\t%.2f\n", p.TotalSum())

	tw.Flush()
}
