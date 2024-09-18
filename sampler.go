package opentelemetry

import (
	"fmt"
	"slices"
	"strings"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"flamingo.me/flamingo/v3/framework/config"
)

type ConfiguredURLPrefixSampler struct {
	allowlist []string
	blocklist []string
}

type SpanKindBasedSampler struct {
	root   tracesdk.Sampler
	config map[trace.SpanKind]tracesdk.Sampler
}

var _ tracesdk.Sampler = (*ConfiguredURLPrefixSampler)(nil)
var _ tracesdk.Sampler = (*SpanKindBasedSampler)(nil)

// Inject dependencies
func (c *ConfiguredURLPrefixSampler) Inject(
	cfg *struct {
		Allowlist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
		Blocklist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
	},
) *ConfiguredURLPrefixSampler {
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

func (c *ConfiguredURLPrefixSampler) ShouldSample(params tracesdk.SamplingParameters) tracesdk.SamplingResult {
	psc := trace.SpanContextFromContext(params.ParentContext)
	path := ""

	for _, attr := range params.Attributes {
		if attr.Key == "http.target" {
			path = attr.Value.AsString()
		}
	}

	// if this is not an incoming request, we decide by parent span
	if path == "" {
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
	// check allowed if len is > 0, and decide if we should sample
	for _, p := range c.allowlist {
		if strings.HasPrefix(path, p) {
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
		if strings.HasPrefix(path, p) {
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

func (c *ConfiguredURLPrefixSampler) Description() string {
	return "ConfiguredURLPrefixSampler"
}

func SpanKindSampler(root tracesdk.Sampler, config map[trace.SpanKind]tracesdk.Sampler) tracesdk.Sampler {
	return &SpanKindBasedSampler{
		root:   root,
		config: config,
	}
}

func (s *SpanKindBasedSampler) ShouldSample(parameters tracesdk.SamplingParameters) tracesdk.SamplingResult {
	if sampler, ok := s.config[parameters.Kind]; ok {
		return sampler.ShouldSample(parameters)
	}

	return s.root.ShouldSample(parameters)
}

func (s *SpanKindBasedSampler) Description() string {
	cfg := make([]string, 0, len(s.config))
	for kind, sampler := range s.config {
		cfg = append(cfg, fmt.Sprintf("%s:%s", kind.String(), sampler.Description()))
	}

	slices.Sort(cfg)

	return fmt.Sprintf("SpanKindBasedSampler{root:%s,config:{%s}}", s.root.Description(), strings.Join(cfg, ","))
}
