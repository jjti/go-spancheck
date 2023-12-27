# go-spanlint

![Latest release](https://img.shields.io/github/v/release/jjti/go-spanlint)
[![build](https://github.com/jjti/go-spanlint/actions/workflows/build.yaml/badge.svg)](https://github.com/jjti/go-spanlint/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjti/go-spanlint)](https://goreportcard.com/report/github.com/jjti/go-spanlint)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Checks usage of [OpenTelemetry spans](https://opentelemetry.io/docs/instrumentation/go/manual/) from [go.opentelemetry.io/otel/trace](go.opentelemetry.io/otel/trace).

## Installation & Usage

```bash
go install github.com/jjti/go-spanlint/cmd/spanlint@latest
spanlint ./...
```

## Configuration

```bash

```

## Background

Tracing is a celebrated [[1](https://andydote.co.uk/2023/09/19/tracing-is-better/),[2](https://charity.wtf/2022/08/15/live-your-best-life-with-structured-events/)] and well marketed [[3](https://docs.datadoghq.com/tracing/),[4](https://www.honeycomb.io/distributed-tracing)] pillar of observability. But self-instrumented traces requires a lot of easy-to-forget boilerplate:

```go
import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)


func task(ctx context.Context) error {
    ctx, span := otel.Tracer("foo").Start(ctx, "bar")
    defer span.End() // call `.End()`

    if err := subTask(ctx); err != nil {
        span.SetStatus(codes.Error, err.Error()) // call SetStatus(codes.Error, msg) to set status:error
        span.RecordError(err) // call RecordError(err) to record an error event
        return err
    }

    return nil
}
```

1. OpenTelemetry docs: [Creating spans](https://opentelemetry.io/docs/instrumentation/go/manual/#creating-spans)
1. Uptrace tutorial: [OpenTelemetry Go Tracing API](https://uptrace.dev/opentelemetry/go-tracing.html#quickstart)

### Forgetting to call `span.End()`

Not calling `End` can cause memory leaks and prevents spans from being closed.

> Any Span that is created MUST also be ended. This is the responsibility of the user. Implementations of this API may leak memory or other resources if Spans are not ended.

[source: trace.go](https://github.com/open-telemetry/opentelemetry-go/blob/98b32a6c3a87fbee5d34c063b9096f416b250897/trace/trace.go#L523)

```go
func task(ctx context.Context) error {
    otel.Tracer("app").Start(ctx, "foo") // span is unassigned, probable memory leak
    _, span := otel.Tracer().Start(ctx, "foo") // span.End is not called on all paths, possible memory leak
    return nil // return can be reached without calling span.End
}
```

### Forgetting to call `span.SetStatus(codes.Error, "msg")`

Not calling `SetStatus` prevents the `status` attribute from being set to `error` which decreases the utility of spans because:

1. observability platforms and APMs differentiate "success" vs "failure" using [span's status codes](https://docs.datadoghq.com/tracing/metrics/).
1. telemetry collector agents, like the [Open Telemetry Collector's Tail Sampling Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/tailsamplingprocessor/README.md#:~:text=Sampling%20Processor.-,status_code,-%3A%20Sample%20based%20upon), are configurable to sample `Error` spans at a higher rate than `OK` spans.
1. observability platforms, like [DataDog, have trace retention filters that use spans' status](https://docs.datadoghq.com/tracing/trace_pipeline/trace_retention/). In other words, `status:error` spans often receive special treatment with the assumption they are more useful for debugging. And forgetting to set the status can lead to spans, with useful debugging information, being dropped.

```go
func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // span.SetStatus is not called on all paths
	defer span.End()

	if err := subTask(); err != nil {
        span.RecordError(err)
		return errors.New(err) // return can be reached without calling span.SetStatus
	}

	return nil
}
```

OpenTelemetry docs: [Set span status](https://opentelemetry.io/docs/instrumentation/go/manual/#set-span-status).

### Forgetting to call `span.RecordError(err)`

Calling `RecordError` creates a new exception-type [event (structured log message)](https://opentelemetry.io/docs/concepts/signals/traces/#span-events) on the span. This is recommended to capture the error's stack trace.

```go
func _() error {
    _, span := otel.Tracer("foo").Start(context.Background(), "bar") // span.RecordError is not called on all paths
    defer span.End()

	if err := subTask(); err != nil {
        span.SetStatus(codes.Error, err.Error())
        return errors.New(err) // return can be reached without calling span.RecordError
	}

    return nil
}
```

OpenTelemetry docs: [Record errors](https://opentelemetry.io/docs/instrumentation/go/manual/#record-errors).
