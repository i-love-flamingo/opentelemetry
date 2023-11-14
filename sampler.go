package opentelemetry

import (
	"net/http"
	"strings"

	"flamingo.me/flamingo/v3/framework/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type ConfiguredURLPrefixSampler struct {
	Allowlist        config.Slice
	Blocklist        config.Slice
	AllowParentTrace bool
}

// Inject dependencies
func (c *ConfiguredURLPrefixSampler) Inject(
	cfg *struct {
		Allowlist        config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
		Blocklist        config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
		AllowParentTrace bool         `inject:"config:flamingo.opentelemetry.tracing.sampler.allowParentTrace,optional"`
	},
) *ConfiguredURLPrefixSampler {
	if cfg != nil {
		c.Allowlist = cfg.Allowlist
		c.Blocklist = cfg.Blocklist
		c.AllowParentTrace = cfg.AllowParentTrace
	}
	return c
}

func (c *ConfiguredURLPrefixSampler) GetFilterOption() otelhttp.Filter {
	var allowed, blocked []string
	_ = c.Allowlist.MapInto(&allowed)
	_ = c.Blocklist.MapInto(&blocked)

	return URLPrefixSampler(allowed, blocked, c.AllowParentTrace)
}

func URLPrefixSampler(allowed, blocked []string, allowParentTrace bool) otelhttp.Filter {
	return func(request *http.Request) bool {
		path := request.URL.Path
		isParentSampled := trace.SpanContextFromContext(request.Context()).IsSampled()
		// empty allowed means all
		sample := len(allowed) == 0
		// check allowed if len is > 0, and decide if we should sample
		for _, p := range allowed {
			if strings.HasPrefix(path, p) {
				sample = true
				break
			}
		}

		// we do not sample, unless the parent is sampled
		if !sample {
			return !allowParentTrace && isParentSampled
		}

		// check sampling decision against blocked
		for _, p := range blocked {
			if strings.HasPrefix(path, p) {
				sample = false
				break
			}
		}

		// we sample, or the parent sampled
		return (!allowParentTrace && isParentSampled) || sample
	}
}
