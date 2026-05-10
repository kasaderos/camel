package asset

import (
	"context"

	"github.com/kasaderos/camel/internal/model"
)

type Option func(*Agent) error

func WithInitialize(ctx context.Context, id string) Option {
	return func(a *Agent) error {
		return a.Initalize(ctx, id)
	}
}

func WithCreate(ctx context.Context, agent *model.AssetAgent) Option {
	return func(a *Agent) error {
		return a.CreateAgent(ctx, agent)
	}
}
