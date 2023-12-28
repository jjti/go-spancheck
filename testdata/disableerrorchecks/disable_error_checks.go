package disableerrorchecks

import (
	"context"
	"errors"

	"github.com/jjti/go-spanlint/testdata/disableerrorchecks/telemetry"
	"go.opentelemetry.io/otel"
)

// incorrect

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths" "span.RecordError is not called on all paths"
	defer span.End()

	if true {
		err := errors.New("foo")
		return err // want "return can be reached without calling span.SetStatus" "return can be reached without calling span.RecordError"
	}

	return nil
}

// correct

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	if true {
		err := errors.New("foo")
		return telemetry.Record(span, err)
	}

	return nil
}
