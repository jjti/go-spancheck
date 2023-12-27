package testdata

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type testErr struct{}

func (e *testErr) Error() string {
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
} // want "this return statement may be reached without calling span.End"

func _() {
	var ctx, span = otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	print(ctx.Done(), span.IsRecording())
} // want "this return statement may be reached without calling span.End"

func _() {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.End is not called on all paths, possible memory leak"
	_, span = otel.Tracer("foo").Start(context.Background(), "bar")
	fmt.Print(span)
	defer span.End()
} // want "this return statement may be reached without calling span.End"

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		err := errors.New("foo")
		return err // want "this return statement may be reached without calling span.SetStatus"
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		return errors.New("foo") // want "this return statement may be reached without calling span.SetStatus"
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		return &testErr{} // want "this return statement may be reached without calling span.SetStatus"
	}

	return nil
}

func _() (string, error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		return "", &testErr{} // want "this return statement may be reached without calling span.SetStatus"
	}

	return "", nil
}

func _() (string, error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		return "", errors.New("foo") // want "this return statement may be reached without calling span.SetStatus"
	}

	return "", nil
}

func _() {
	f := func() error {
		_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
		defer span.End()

		if true {
			return errors.New("foo") // want "this return statement may be reached without calling span.SetStatus"
		}

		return nil
	}
	fmt.Println(f)
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	{
		if true {
			return errors.New("foo") // want "this return statement may be reached without calling span.SetStatus"
		}
	}

	return nil
}

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
		return err
	}

	if true {
		span.SetStatus(codes.Error, "foo")
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

func _() {
	// TODO: https://andydote.co.uk/2023/09/19/tracing-is-better/
}
