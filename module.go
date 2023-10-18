package opentelemetry

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"flamingo.me/flamingo/v3/framework/systemendpoint"
	"flamingo.me/flamingo/v3/framework/systemendpoint/domain"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"sync"
)

var (
	createMeterOnce sync.Once
	KeyArea, _      = baggage.NewKeyProperty("area")
)

type Module struct {
	serviceName    string
	jaegerEnable   bool
	jaegerEndpoint string
	zipkinEnable   bool
	zipkinEndpoint string
}

func (m *Module) Inject(
	cfg *struct {
		ServiceName    string `inject:"config:flamingo.opentelemetry.serviceName"`
		JaegerEnable   bool   `inject:"config:flamingo.opentelemetry.jaeger.enable"`
		JaegerEndpoint string `inject:"config:flamingo.opentelemetry.jaeger.endpoint"`
		ZipkinEnable   bool   `inject:"config:flamingo.opentelemetry.zipkin.enable"`
		ZipkinEndpoint string `inject:"config:flamingo.opentelemetry.zipkin.endpoint"`
	},
) *Module {
	if cfg != nil {
		m.serviceName = cfg.ServiceName
		m.jaegerEnable = cfg.JaegerEnable
		m.jaegerEndpoint = cfg.JaegerEndpoint
		m.zipkinEnable = cfg.ZipkinEnable
		m.zipkinEndpoint = cfg.ZipkinEndpoint
	}
	return m
}

const (
	name = "flamingo.me/opentelemetry"
)

func (m *Module) Configure(injector *dingo.Injector) {
	http.DefaultTransport = &correlationIDInjector{next: otelhttp.NewTransport(http.DefaultTransport)}

	// traces
	tracerProviderOptions := make([]tracesdk.TracerProviderOption, 0, 3)

	// Create the Jaeger exporter
	if m.jaegerEnable {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(m.jaegerEndpoint)))
		if err != nil {
			log.Fatalf("failed to initialze Jeager exporter: %v", err)
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
		panic(err)
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

	// metrics
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
	meter = meterProvider.Meter(name, metric.WithInstrumentationVersion(SemVersion()))

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

type Instrumentation struct {
	Tracer trace.Tracer
	Meter  metric.Meter
}

var (
	tracer trace.Tracer
	meter  metric.Meter
)

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
	serviceName: string | *"flamingo"
	tracing: sampler: {
		whitelist: [...string]
		blacklist: [...string]
		allowParentTrace: bool | *true
	}
	publicEndpoint: bool | *true
}
`
}
