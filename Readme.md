# OpenTelemetry

This package provides effective observability to the flamingo ecosystem using
the [OpenTelemetry](https://opentelemetry.io/) instrumentation library.
With the OpenTelemetry module, your application automatically exports telemetry data such as metrics and traces.
This makes it easy to analyze your application's behavior and collect statistics about load and performance.
It provides exporters for the most common tools, i.e. Jaeger, Prometheus and OTLP components.

The metrics endpoint is provided under the systemendpoint. Once the module is activate you can access them
via `http://localhost:13210/metrics`

## Module configuration

| Config                                                        | Default Value                        | Description                                                                                                                                                           |
|---------------------------------------------------------------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `flamingo.opentelemetry.serviceName`                          | `flamingo`                           | serviceName is automatically added to all traces as `service.name` attribute                                                                                          |
| `flamingo.opentelemetry.zipkin.enable`                        | `false`                              | enables the zipkin exporter                                                                                                                                           |
| `flamingo.opentelemetry.zipkin.endpoint`                      | `http://localhost:9411/api/v2/spans` | URL to the zipkin instance                                                                                                                                            |
| `flamingo.opentelemetry.otlp.http.enable`                     | `false`                              | enables the OTLP HTTP exporter                                                                                                                                        |
| `flamingo.opentelemetry.otlp.http.endpoint`                   | `http://localhost:4318/v1/traces`    | URL to the OTLP collector                                                                                                                                             |
| `flamingo.opentelemetry.otlp.grpc.enable`                     | `false`                              | enables the OTLP gRPC exporter                                                                                                                                        |
| `flamingo.opentelemetry.otlp.grpc.endpoint`                   | `grpc://localhost:4317/v1/traces`    | URL to the OTLP collector                                                                                                                                             |
| `flamingo.opentelemetry.tracing.sampler.allowlist`            | `[]`                                 | list of URL paths that are sampled; if empty, all paths are allowed                                                                                                   |
| `flamingo.opentelemetry.tracing.sampler.blocklist`            | `[]`                                 | list of URL paths that are never sampled                                                                                                                              |
| `flamingo.opentelemetry.tracing.sampler.ignoreParentDecision` | `true`                               | if `true`, we will ignore sampling decisions of the parent span                                                                                                       |

## Adding your own tracing information

Before you can create your own spans, you have to initialize a tracer:

```go
tracer := otel.Tracer("my-app")
```

Now you can create a span based on a `context.Context`. This will automatically attach all tracing-relevant
information (e.g. trace-ID) to the span.

```go
func doSomething(ctx context.Context) {
ctx, span := tracer.Start(ctx, "my-span")
defer span.End()

// do some work to track with my-span	
}
```

To add further attributes to the span, please refer to the
official [OpenTelemetry documentation](https://opentelemetry.io/docs/instrumentation/go/manual/#traces).

## Adding your own metrics

To collect your own metrics, you have to initialize a meter:

```go
meter := otel.Meter("my-app")
```

Now you can create a new metric, e.g. a counter:

```go
counter, _ := meter.Int64Counter("my.count",
metric.WithDescription("count of something"),
)

counter.Add(ctx, 1)
```

For more information about the kinds of metrics and how to use them, please refer to the
official [OpenTelemetry documentation](https://opentelemetry.io/docs/instrumentation/go/manual/#metrics).
