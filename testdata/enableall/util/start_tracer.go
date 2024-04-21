package util

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestStartTrace() trace.Span {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	return span
}
