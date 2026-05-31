package extraendsignatures

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/jjti/go-spancheck/testdata/extraendsignatures/telemetry"
)

// incorrect

func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	fmt.Print(span)
} // want "return can be reached without calling span.End"

// correct

// the span is ended by a deferred call to the custom end function
func _() (err error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer func() { telemetry.End(span, err) }()

	return errors.New("foo")
}

// the custom end function is called directly on every path
func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")

	if true {
		err := errors.New("foo")
		telemetry.End(span, err)
		return err
	}

	telemetry.End(span, nil)

	return nil
}
