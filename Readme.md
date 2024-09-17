# OpenTelemetry

This package provides effective observability to the flamingo ecosystem using
the [OpenTelemetry](https://opentelemetry.io/) instrumentation library.
With the OpenTelemetry module, your application automatically exports telemetry data such as metrics and traces.
This makes it easy to analyze your application's behavior and collect statistics about load and performance.
It provides exporters for the most common tools, i.e. Jaeger, Prometheus and OTLP components.

The metrics endpoint is provided under the systemendpoint. Once the module is activated you can access them
via `http://localhost:13210/metrics`

## Usage with Flamingo core opencensus

This module uses https://pkg.go.dev/go.opentelemetry.io/otel/bridge/opencensus to be compatible with Flamingo core and 
other Flamingo modules which still use opencesus metrics and traces.

You should not use `opencesus.Module` and `opentelemetry.Module` at the same time. If you want to use opentelemetry on old code, 
only add the `opentelemetry.Module` and let the bridge automatically take care.

### Bridge Limitations

Please note the limitations of the bridge: https://pkg.go.dev/go.opentelemetry.io/otel/bridge/opencensus#hdr-Limitations

#### Flamingo Samplers

> Conversion of custom OpenCensus Samplers to OpenTelemetry is not implemented, and An error will be sent to the OpenTelemetry ErrorHandler.

Flamingo's `URLPrefixSampler` and config from `flamingo.opencensus.tracing.sampler.*` will not work with the bridge.

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
var tracer = otel.Tracer("my-app", trace.WithInstrumentationVersion("1.2.3"))
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
var meter = otel.Meter("my-app", metric.WithInstrumentationVersion("1.2.3"))
```

Now you can create a new metric, e.g. a counter:

```go
counter, _ := meter.Int64Counter("my.count", 
	metric.WithDescription("count of something"),
	metric.WithUnit("{something}")
)

counter.Add(ctx, 1)
```

For more information about the kinds of metrics and how to use them, please refer to the
official [OpenTelemetry documentation](https://opentelemetry.io/docs/instrumentation/go/manual/#metrics).
