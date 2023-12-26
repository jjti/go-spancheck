package testdata

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
)

// unused, not assigned
func _() {
	otel.Tracer("foo").Start(context.Background(), "bar") // want `span is unassigned, probable memory leak`
}

// unused, empty assignee
func _() {
	ctx, _ := otel.Tracer("foo").Start(context.Background(), "bar") // want `span is unassigned, probable memory leak`
	print(ctx.Done())
}

// no .End()
func _() {
	ctx, span := otel.Tracer("foo").Start(context.Background(), "bar") // want `span.End is not called on all paths, possible memory leak`
	print(ctx.Done(), span.IsRecording())
} // want `this return statement may be reached without calling span.End`

// no .End()
func _() {
	var ctx, span = otel.Tracer("foo").Start(context.Background(), "bar") // want `span.End is not called on all paths, possible memory leak`
	print(ctx.Done(), span.IsRecording())
} // want `this return statement may be reached without calling span.End`

// no .End(), re-assigned
func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want `span.End is not called on all paths, possible memory leak`
	_, span = otel.Tracer("foo").Start(context.Background(), "bar")
	fmt.Print(span)
	defer span.End()
} // want `this return statement may be reached without calling span.End`

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	return nil
}

func _() {
	// correct
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	_, span = otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()
}
