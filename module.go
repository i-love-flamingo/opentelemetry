package opentelemetry

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"flamingo.me/flamingo/v3/framework/systemendpoint"
	"flamingo.me/flamingo/v3/framework/systemendpoint/domain"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"

	//nolint:staticcheck
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Module struct {
	serviceName      string
	jaegerEnable     bool
	jaegerEndpoint   string
	zipkinEnable     bool
	zipkinEndpoint   string
	otlpEnableHTTP   bool
	otlpEndpointHTTP string
	otlpEnableGRPC   bool
	otlpEndpointGRPC string
}

func (m *Module) Inject(
	cfg *struct {
		ServiceName      string `inject:"config:flamingo.opentelemetry.serviceName"`
		JaegerEnable     bool   `inject:"config:flamingo.opentelemetry.jaeger.enable"`
		JaegerEndpoint   string `inject:"config:flamingo.opentelemetry.jaeger.endpoint"`
		ZipkinEnable     bool   `inject:"config:flamingo.opentelemetry.zipkin.enable"`
		ZipkinEndpoint   string `inject:"config:flamingo.opentelemetry.zipkin.endpoint"`
		OTLPEnableHTTP   bool   `inject:"config:flamingo.opentelemetry.otlp.http.enable"`
		OTLPEndpointHTTP string `inject:"config:flamingo.opentelemetry.otlp.http.endpoint"`
		OTLPEnableGRPC   bool   `inject:"config:flamingo.opentelemetry.otlp.grpc.enable"`
		OTLPEndpointGRPC string `inject:"config:flamingo.opentelemetry.otlp.grpc.endpoint"`
	},
) *Module {
	if cfg != nil {
		m.serviceName = cfg.ServiceName
		m.jaegerEnable = cfg.JaegerEnable
		m.jaegerEndpoint = cfg.JaegerEndpoint
		m.zipkinEnable = cfg.ZipkinEnable
		m.zipkinEndpoint = cfg.ZipkinEndpoint

		m.otlpEnableHTTP = cfg.OTLPEnableHTTP
		m.otlpEndpointHTTP = cfg.OTLPEndpointHTTP
		m.otlpEnableGRPC = cfg.OTLPEnableGRPC
		m.otlpEndpointGRPC = cfg.OTLPEndpointGRPC
	}
	return m
}

const (
	name = "flamingo.me/opentelemetry"
)

func (m *Module) Configure(injector *dingo.Injector) {
	http.DefaultTransport = &correlationIDInjector{next: otelhttp.NewTransport(http.DefaultTransport)}

	m.initTraces()
	m.initMetrics(injector)
}

func (m *Module) initTraces() {
	tracerProviderOptions := make([]tracesdk.TracerProviderOption, 0, 3)

	// Create the Jaeger exporter
	if m.jaegerEnable {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(m.jaegerEndpoint)))
		if err != nil {
			log.Fatalf("failed to initialze Jeager exporter: %v", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, tracesdk.WithBatcher(exp))
	}

	// Create the OTLP HTTP exporter
	if m.otlpEnableHTTP {
		u, err := url.Parse(m.otlpEndpointHTTP)
		if err != nil {
			log.Fatalf("could not parse OTLP HTTP endpoint: %v", err)
		}

		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(u.Host),
			otlptracehttp.WithURLPath(u.Path),
		}

		exp, err := otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			log.Fatalf("failed to initialze OTLP HTTP exporter: %v", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, tracesdk.WithBatcher(exp))
	}

	// Create the OTLP gRPC exporter
	if m.otlpEnableGRPC {
		exp, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpoint(m.otlpEndpointGRPC))
		if err != nil {
			log.Fatalf("failed to initialze OTLP gRPC exporter: %v", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, tracesdk.WithBatcher(exp))
	}

	// Create the Zipkin exporter
	if m.zipkinEnable {
		exp, err := zipkin.New(
			m.zipkinEndpoint,
		)
		if err != nil {
			log.Fatalf("failed to initialize Zipkin exporter: %v", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, tracesdk.WithBatcher(exp))
	}

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(m.serviceName),
			semconv.ServiceVersion(flamingo.AppVersion()),
			semconv.TelemetrySDKLanguageGo,
		))
	if err != nil {
		log.Fatalf("failed to initialize otel resource: %v", err)
	}

	tracerProviderOptions = append(tracerProviderOptions,
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	tp := tracesdk.NewTracerProvider(tracerProviderOptions...)
	otel.SetTracerProvider(tp)

	tr := tp.Tracer(name, trace.WithInstrumentationVersion(SemVersion()))
	octrace.DefaultTracer = opencensus.NewTracer(tr)

	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#propagators-distribution
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func (m *Module) initMetrics(injector *dingo.Injector) {
	bridge := opencensus.NewMetricProducer()
	exp, err := prometheus.New(prometheus.WithProducer(bridge))
	if err != nil {
		log.Fatalf("failed to initialize Prometheus exporter: %v", err)
	}

	meterProvider := sdkMetric.NewMeterProvider(sdkMetric.WithReader(exp))
	otel.SetMeterProvider(meterProvider)
	if err := runtimemetrics.Start(); err != nil {
		log.Fatal(err)
	}

	injector.BindMap((*domain.Handler)(nil), "/metrics").ToInstance(promhttp.Handler())
}

func (m *Module) Depends() []dingo.Module {
	return []dingo.Module{
		new(systemendpoint.Module),
	}
}

type correlationIDInjector struct {
	next http.RoundTripper
}

func (rt *correlationIDInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	span := trace.SpanFromContext(req.Context())
	if span.SpanContext().IsSampled() {
		req.Header.Add("X-Correlation-ID", span.SpanContext().TraceID().String())
	}
	return rt.next.RoundTrip(req)
}

func (m *Module) CueConfig() string {
	return `
flamingo: opentelemetry: {
	jaeger: {
		enable: bool | *false
		endpoint: string | *"http://localhost:14268/api/traces"
	}
	zipkin: {
		enable: bool | *false
		endpoint: string | *"http://localhost:9411/api/v2/spans"
	}
	otlp: {
		http: {
			enable: bool | *false
			endpoint: string | *"http://localhost:4318/v1/traces"
		}
		grpc: {
			enable: bool | *false
			endpoint: string | *"grpc://localhost:4317/v1/traces"
		}
	}
	serviceName: string | *"flamingo"
	tracing: sampler: {
		allowlist: [...string]
		blocklist: [...string]
		allowParentTrace: bool | *true
	}
}
`
}
