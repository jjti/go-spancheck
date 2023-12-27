# go-spanlint

![Latest release](https://img.shields.io/github/v/release/jjti/go-spanlint)
[![build](https://github.com/jjti/go-spanlint/actions/workflows/build.yaml/badge.svg)](https://github.com/jjti/go-spanlint/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjti/go-spanlint)](https://goreportcard.com/report/github.com/jjti/go-spanlint)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Checks usage of [OpenTelemetry spans](https://pkg.go.dev/go.opentelemetry.io/otel/trace).

## Problem Statement

Tracing is an — often celebrated [[1](https://andydote.co.uk/2023/09/19/tracing-is-better/), [2](https://charity.wtf/2022/08/15/live-your-best-life-with-structured-events/)] — pillar of observability. But it's easy to shoot yourself in the foot when creating and managing OTEL spans. For two reasons:

### Forgetting to call `span.End()`

Not calling `.End()` can cause memory leaks.

> Any Span that is created MUST also be ended. This is the responsibility of the user. Implementations of this API may leak memory or other resources if Spans are not ended.

[source: trace.go](https://github.com/open-telemetry/opentelemetry-go/blob/98b32a6c3a87fbee5d34c063b9096f416b250897/trace/trace.go#L523)

```go
func task(ctx context.Context) error {
    otel.Tracer().Start(ctx, "foo") // span is unassigned, probable memory leak
    _, span := otel.Tracer().Start(ctx, "foo") // span.End is not called on all paths, possible memory leak
    return nil // this return statement may be reached without calling span.End
}
```

### Forgetting to call `span.SetStatus(codes.Error, "msg")`

Setting spans' status to `codes.Error` matters for a couple reasons:

1. observability platforms and APMs differentiate "success" vs "failure" using [span's status codes](https://docs.datadoghq.com/tracing/metrics/).
1. telemetry collector agents, like the [Open Telemetry Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/tailsamplingprocessor/README.md#:~:text=Sampling%20Processor.-,status_code,-%3A%20Sample%20based%20upon), are configurable to sample `Error` spans at a higher rate than `OK` spans. Similarly, observability platforms like DataDog support trace retention filters based on spans' status. In other words, `Error` spans often receive special treatment with the assumption they are more useful for debugging.

```go
func _() error {
	_, span := otel.Tracer("foo").Start(context.Background(), "bar") // span.SetStatus is not called on all paths
	defer span.End()

	if true {
		return errors.New("foo") // this return statement may be reached without calling span.SetStatus
	}

	return nil
}
```
