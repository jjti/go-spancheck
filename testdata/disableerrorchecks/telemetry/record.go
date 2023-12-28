package telemetry

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Record(span trace.Span, err error) error {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
	return err
}
