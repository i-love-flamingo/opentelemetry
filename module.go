package opentelemetry

//go:generate go run github.com/vektra/mockery/v2@v2.53.5

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"flamingo.me/dingo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"flamingo.me/flamingo/v3/framework/flamingo"
	flamingoHttp "flamingo.me/flamingo/v3/framework/http"
	"flamingo.me/flamingo/v3/framework/systemendpoint"
	"flamingo.me/flamingo/v3/framework/systemendpoint/domain"
)

type Module struct {
	sampler                          *configuredURLPrefixSampler
	serviceName                      string
	publicEndpoint                   bool
	zipkinEnable                     bool
	zipkinEndpoint                   string
	otlpEnableHTTP                   bool
	otlpEndpointHTTP                 string
	otlpEnableGRPC                   bool
	otlpEndpointGRPC                 string
	legacyPrometheusNamingSanitation bool
}

func (m *Module) Inject(
	sampler *configuredURLPrefixSampler,
	logger flamingo.Logger,
	cfg *struct {
		ServiceName                      string `inject:"config:flamingo.opentelemetry.serviceName"`
		PublicEndpoint                   bool   `inject:"config:flamingo.opentelemetry.publicEndpoint"`
		ZipkinEnable                     bool   `inject:"config:flamingo.opentelemetry.zipkin.enable"`
		ZipkinEndpoint                   string `inject:"config:flamingo.opentelemetry.zipkin.endpoint"`
		OTLPEnableHTTP                   bool   `inject:"config:flamingo.opentelemetry.otlp.http.enable"`
		OTLPEndpointHTTP                 string `inject:"config:flamingo.opentelemetry.otlp.http.endpoint"`
		OTLPEnableGRPC                   bool   `inject:"config:flamingo.opentelemetry.otlp.grpc.enable"`
		OTLPEndpointGRPC                 string `inject:"config:flamingo.opentelemetry.otlp.grpc.endpoint"`
		LegacyPrometheusNamingSanitation bool   `inject:"config:flamingo.opentelemetry.legacyPrometheusNamingSanitation"`
	},
) *Module {
	m.sampler = sampler

	if cfg != nil {
		m.serviceName = cfg.ServiceName
		m.publicEndpoint = cfg.PublicEndpoint
		m.zipkinEnable = cfg.ZipkinEnable
		m.zipkinEndpoint = cfg.ZipkinEndpoint
		m.otlpEnableHTTP = cfg.OTLPEnableHTTP
		m.otlpEndpointHTTP = cfg.OTLPEndpointHTTP
		m.otlpEnableGRPC = cfg.OTLPEnableGRPC
		m.otlpEndpointGRPC = cfg.OTLPEndpointGRPC
		m.legacyPrometheusNamingSanitation = cfg.LegacyPrometheusNamingSanitation
	}

	otel.SetErrorHandler(newErrorHandler(logger))

	return m
}

func (m *Module) Configure(injector *dingo.Injector) {
	http.DefaultTransport = &correlationIDInjector{
		next: otelhttp.NewTransport(http.DefaultTransport),
	}

	injector.Bind(new(flamingoHttp.HandlerWrapper)).ToProvider(func() flamingoHttp.HandlerWrapper {
		return func(handler http.Handler) http.Handler {
			const maxOptions = 2

			startOptions := make([]otelhttp.Option, 0, maxOptions)
			startOptions = append(
				startOptions,
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					return operation + ": " + r.URL.Path
				}),
			)

			if m.publicEndpoint {
				startOptions = append(startOptions, otelhttp.WithPublicEndpoint())
			}

			return otelhttp.NewHandler(
				handler,
				"incoming request",
				startOptions...,
			)
		}
	})

	flamingo.BindEventSubscriber(injector).To(new(Listener))

	m.initTraces()
	m.initMetrics(injector)
}

func (m *Module) initTraces() {
	const maxTracerProviderOptions = 5
	tracerProviderOptions := make([]tracesdk.TracerProviderOption, 0, maxTracerProviderOptions)

	tracerProviderOptions = m.initOTLP(tracerProviderOptions)
	tracerProviderOptions = m.initZipkin(tracerProviderOptions)

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
		tracesdk.WithSampler(
			&alwaysSampleSpanKindClient{
				base: m.sampler,
			},
		),
	)

	tp := tracesdk.NewTracerProvider(tracerProviderOptions...)
	otel.SetTracerProvider(tp)

	opencensus.InstallTraceBridge(opencensus.WithTracerProvider(tp))

	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/context/api-propagators.md#propagators-distribution
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

// Create the OTLP HTTP exporter
func (m *Module) initOTLP(tracerProviderOptions []tracesdk.TracerProviderOption) []tracesdk.TracerProviderOption {
	if m.otlpEnableHTTP {
		u, err := url.Parse(m.otlpEndpointHTTP)
		if err != nil {
			log.Fatalf("could not parse OTLP HTTP endpoint: %v", err)
		}

		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(u.Host),
			otlptracehttp.WithURLPath(u.Path),
		}
		if u.Scheme == "http" {
			opts = append(opts, otlptracehttp.WithInsecure())
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

	return tracerProviderOptions
}

// Create the Zipkin exporter
func (m *Module) initZipkin(tracerProviderOptions []tracesdk.TracerProviderOption) []tracesdk.TracerProviderOption {
	if m.zipkinEnable {
		exp, err := zipkin.New(
			m.zipkinEndpoint,
		)
		if err != nil {
			log.Fatalf("failed to initialize Zipkin exporter: %v", err)
		}

		tracerProviderOptions = append(tracerProviderOptions, tracesdk.WithBatcher(exp))
	}

	return tracerProviderOptions
}

func (m *Module) initMetrics(injector *dingo.Injector) {
	options := []prometheus.Option{
		prometheus.WithProducer(opencensus.NewMetricProducer()),
	}

	if m.legacyPrometheusNamingSanitation {
		options = append(options, prometheus.WithTranslationStrategy(otlptranslator.UnderscoreEscapingWithSuffixes))
	}

	exp, err := prometheus.New(
		options...,
	)
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

	resp, err := rt.next.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("correlationIDInjector next RoundTrip failed: %w", err)
	}

	return resp, nil
}

func (m *Module) CueConfig() string {
	return `
flamingo: opentelemetry: {
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
	publicEndpoint: bool | *true
	tracing: sampler: {
		allowlist: [...string]
		blocklist: [...string]
	}
	legacyPrometheusNamingSanitation: bool | *true
}
`
}
