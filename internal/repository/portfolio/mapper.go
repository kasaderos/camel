package portfolio

import (
	"github.com/kasaderos/camel/internal/model"
)

func toModel(p portfolio) *model.Portfolio {
	return &model.Portfolio{
		ID:   p.ID,
		Name: p.Name,

		Weights: p.Weights,
		Cache:   p.Cache,

		PolicyID: p.PolicyID,
	}
}
