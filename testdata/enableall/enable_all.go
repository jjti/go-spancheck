package enableall

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjti/go-spancheck/testdata/enableall/util"
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

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		err := errors.New("foo")
		span.RecordError(err)
		return err // want "return can be reached without calling span.SetStatus"
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		span.RecordError(errors.New("foo"))
		return errors.New("foo") // want "return can be reached without calling span.SetStatus"
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		span.RecordError(errors.New("foo"))
		return &testError{} // want "return can be reached without calling span.SetStatus"
	}

	return nil
}

func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.RecordError is not called on all paths"
	defer span.End()

	if true {
		span.SetStatus(codes.Error, "foo")
		return &testError{} // want "return can be reached without calling span.RecordError"
	}

	return nil
}

func _() (string, error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		span.RecordError(errors.New("foo"))
		return "", &testError{} // want "return can be reached without calling span.SetStatus"
	}

	return "", nil
}

func _() (string, error) {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		span.RecordError(errors.New("foo"))
		return "", errors.New("foo") // want "return can be reached without calling span.SetStatus"
	}

	return "", nil
}

func _() {
	f := func() error {
		_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
		defer span.End()

		if true {
			span.RecordError(errors.New("foo"))
			return errors.New("foo") // want "return can be reached without calling span.SetStatus"
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
			span.RecordError(errors.New("foo"))
			return errors.New("foo") // want "return can be reached without calling span.SetStatus"
		}
	}

	return nil
}

func _() error {
	_, span := trace.StartSpan(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
	defer span.End()

	if true {
		err := errors.New("foo")
		return err // want "return can be reached without calling span.SetStatus"
	}

	return nil
}

func _() {
	span := util.TestStartTrace() // want "span.End is not called on all paths, possible memory leak"
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

func testStartTrace() *trace.Span { // no error expected because this is in extra start types
	_, span := trace.StartSpan(context.Background(), "bar")
	return span
}

// https://github.com/jjti/go-spancheck/issues/25
func _() error {
	if true {
		_, span := otel.Tracer("foo").Start(context.Background(), "bar")
		span.End()
	}

	return errors.New("test")
}

func _() error {
	if true {
		_, span := otel.Tracer("foo").Start(context.Background(), "bar") // want "span.SetStatus is not called on all paths"
		defer span.End()

		if true {
			span.RecordError(errors.New("test"))
			return errors.New("test") // want "return can be reached without calling span.SetStatus"
		}
	}

	return errors.New("test")
}

// TODO: the check below now fails after the change to fix issue 25 above.
// func _() error {
// 	var span *trace.Span

// 	if true {
// 		_, span = trace.StartSpan(context.Background(), "foo")
// 		defer span.End()
// 	}

// 	return errors.New("test")
// }
