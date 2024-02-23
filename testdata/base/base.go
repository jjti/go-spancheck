package main

import (
	"context"
	"errors"
	"fmt"

	"go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type testError struct{}

func (e *testError) Error() string {
	return "foo"
}

// incorrect

func _() {
	otel.Tracer("foo").Start(context.Background(), "bar")           // want "span is unassigned, probable memory leak"
	ctx, _ := otel.Tracer("foo").Start(context.Background(), "bar") // want "span is unassigned, probable memory leak"
	fmt.Print(ctx)
}

func _() {
	ctx, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	print(ctx.Done(), span.IsRecording())
} // want "return can be reached without calling span.End"

func _() {
	var ctx, span = otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	print(ctx.Done(), span.IsRecording())
} // want "return can be reached without calling span.End"

func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	_, span = otel.Tracer("foo").Start(context.Background(), "bar")
	fmt.Print(span)
	defer span.End()
} // want "return can be reached without calling span.End"

func _() {
	_, span := trace.StartSpan(context.Background(), "foo") // want "span.End is not called on all paths, possible memory leak"
	fmt.Print(span)
} // want "return can be reached without calling span.End"

func _() {
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "foo", trace.SpanContext{}) // want "span.End is not called on all paths, possible memory leak"
	fmt.Print(span)
} // want "return can be reached without calling span.End"

// correct

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	if true {
		return nil
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	if false {
		err := errors.New("foo")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	if true {
		span.SetStatus(codes.Error, "foo")
		span.RecordError(errors.New("foo"))
		return errors.New("bar")
	}

	return nil
}

func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	_, span = otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	if true {
		span.SetStatus(codes.Error, "foo")
		return &testError{}
	}

	return nil
}

func _() (string, error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer span.End()

	if true {
		span.RecordError(errors.New("foo"))
		return "", &testError{}
	}

	return "", nil
}

func _() {
	_, span := trace.StartSpan(context.Background(), "foo")
	defer span.End()
}

func _() {
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "foo", trace.SpanContext{})
	defer span.End()
}

// This tests that we detect when the span is closed within a deferred func.
// https://github.com/jjti/go-spancheck/issues/12
func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar")
	defer func() {
		span.End()
	}()
}

// Despite above, we do not wander more than one level deep into the defer stack.
func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	defer func() {
		defer func() {
			span.End()
		}()
	}()
} // want "return can be reached without calling span.End"
