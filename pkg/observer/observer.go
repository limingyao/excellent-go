package observer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var defaultObserver = func(ctx context.Context, name string) (context.Context, func(error)) {
	ctx, span := otel.Tracer("").Start(ctx, name)
	return ctx, func(err error) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
	}
}

func New(ctx context.Context, name string) (context.Context, func(error)) {
	return defaultObserver(ctx, name)
}

func SetObserver(observer func(ctx context.Context, name string) (context.Context, func(error))) {
	defaultObserver = observer
}
