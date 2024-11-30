package disableerrorchecks

import (
	"context"
	"errors"

	"github.com/jjti/go-spancheck/testdata/disableerrorchecks/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	err := errors.New("foo")
	err = telemetry.Record(span, err)
	return err
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	err := errors.New("foo")
	recordErr(span, err)
	return err
}

func recordErr(span trace.Span, err error) {}

// https://github.com/jjti/go-spancheck/issues/24
func _() (err error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer func() {
		recordErr(span, err)

		span.End()
	}()

	return errors.New("test")
}
