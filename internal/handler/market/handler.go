package market

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	marketv1 "github.com/kasaderos/camel/gen/proto/market/v1"
	"github.com/kasaderos/camel/gen/proto/market/v1/marketv1connect"
	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/slices"
)

type Service interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start, end time.Time,
	) ([]model.Bar, error)
}

type Handler struct {
	service Service
	marketv1connect.UnimplementedMarketServiceHandler
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) FetchBars(
	ctx context.Context,
	req *marketv1.FetchBarsRequest,
) (*marketv1.FetchBarsResponse, error) {
	result, err := h.service.FetchBars(
		ctx,
		req.Symbol,
		req.GetStartDate().AsTime(),
		req.GetEndDate().AsTime(),
	)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("fetch bars: %w", err),
		)
	}

	resp, _ := slices.Map(
		result,
		func(bar model.Bar) (*marketv1.Bar, error) {
			return mapBarToProtoBar(bar), nil
		},
	)

	return &marketv1.FetchBarsResponse{
		Bars: resp,
	}, nil
}

func mapBarToProtoBar(bar model.Bar) *marketv1.Bar {
	return &marketv1.Bar{
		Timestamp: timestamppb.New(bar.Timestamp),
		Open:      bar.Open,
		High:      bar.High,
		Low:       bar.Low,
		Close:     bar.Close,
		Volume:    int64(bar.Volume),
	}
}
