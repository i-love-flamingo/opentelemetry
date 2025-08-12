package opentelemetry

import (
	"fmt"
	"strings"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"flamingo.me/flamingo/v3/framework/config"
)

type configuredURLPrefixSampler struct {
	allowlist []string
	blocklist []string
}

// alwaysSampleSpanKindClient enforces sampling of outgoing http requests (client)
type alwaysSampleSpanKindClient struct {
	base tracesdk.Sampler
}

var _ tracesdk.Sampler = (*configuredURLPrefixSampler)(nil)
var _ tracesdk.Sampler = (*alwaysSampleSpanKindClient)(nil)

// Inject dependencies
func (c *configuredURLPrefixSampler) Inject(
	cfg *struct {
		Allowlist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
		Blocklist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
	},
) *configuredURLPrefixSampler {
	if cfg != nil {
		var allowed, blocked []string

		err := cfg.Allowlist.MapInto(&allowed)
		if err != nil {
			panic(fmt.Errorf("failed to map flamingo.opentelemetry.tracing.sampler.allowlist: %w", err))
		}

		err = cfg.Blocklist.MapInto(&blocked)
		if err != nil {
			panic(fmt.Errorf("failed to map flamingo.opentelemetry.tracing.sampler.blocklist: %w", err))
		}

		c.allowlist = allowed
		c.blocklist = blocked
	}

	return c
}

func (c *configuredURLPrefixSampler) ShouldSample(params tracesdk.SamplingParameters) tracesdk.SamplingResult {
	psc := trace.SpanContextFromContext(params.ParentContext)
	path := ""
	query := ""

	for _, attr := range params.Attributes {
		if attr.Key == "url.path" {
			path = attr.Value.AsString()
		}
		if attr.Key == "url.query" {
			query = attr.Value.AsString()
		}
	}

	target := path + query

	// if this is not an incoming request, we decide by parent span
	if target == "" {
		decision := tracesdk.Drop

		if psc.IsSampled() {
			decision = tracesdk.RecordAndSample
		}

		return tracesdk.SamplingResult{
			Decision:   decision,
			Tracestate: psc.TraceState(),
		}
	}

	// empty allowed means all
	sample := len(c.allowlist) == 0
	// decide if we should sample based on allow list
	for _, p := range c.allowlist {
		if strings.HasPrefix(target, p) {
			sample = true
			break
		}
	}

	// we do not sample, unless the parent is sampled
	if !sample {
		return tracesdk.SamplingResult{
			Decision:   tracesdk.Drop,
			Tracestate: psc.TraceState(),
		}
	}

	// check sampling decision against blocked
	for _, p := range c.blocklist {
		if strings.HasPrefix(target, p) {
			return tracesdk.SamplingResult{
				Decision:   tracesdk.Drop,
				Tracestate: psc.TraceState(),
			}
		}
	}

	return tracesdk.SamplingResult{
		Decision:   tracesdk.RecordAndSample,
		Tracestate: psc.TraceState(),
	}
}

func (c *configuredURLPrefixSampler) Description() string {
	allowlist := strings.Join(c.allowlist, ",")
	blocklist := strings.Join(c.blocklist, ",")

	return fmt.Sprintf("ConfiguredURLPrefixSampler{allowlist:%s,blocklist:%s}", allowlist, blocklist)
}

func (s *alwaysSampleSpanKindClient) ShouldSample(parameters tracesdk.SamplingParameters) tracesdk.SamplingResult {
	if parameters.Kind == trace.SpanKindClient {
		return tracesdk.AlwaysSample().ShouldSample(parameters)
	}

	return s.base.ShouldSample(parameters)
}

func (s *alwaysSampleSpanKindClient) Description() string {
	return fmt.Sprintf("SpanKindBasedSampler{base:%s}", s.base.Description())
}
