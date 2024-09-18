package opentelemetry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"flamingo.me/flamingo/v3/framework/config"

	"flamingo.me/opentelemetry"
)

func TestConfiguredURLPrefixSampler_Inject(t *testing.T) {
	t.Parallel()

	t.Run("should panic on invalid allowlist", func(t *testing.T) {
		t.Parallel()

		sampler := new(opentelemetry.ConfiguredURLPrefixSampler)

		assert.Panics(t,
			func() {
				sampler.Inject(
					&struct {
						Allowlist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
						Blocklist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
					}{
						Allowlist: []any{"1", 2, false},
					},
				)
			})
	})

	t.Run("should panic on invalid blocklist", func(t *testing.T) {
		t.Parallel()

		sampler := new(opentelemetry.ConfiguredURLPrefixSampler)

		assert.Panics(t,
			func() {
				sampler.Inject(
					&struct {
						Allowlist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
						Blocklist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
					}{
						Blocklist: []any{"1", 2, false},
					},
				)
			})
	})
}

func TestConfiguredURLPrefixSampler_ShouldSample(t *testing.T) {
	t.Parallel()

	type fields struct {
		Allowlist config.Slice
		Blocklist config.Slice
	}

	type request struct {
		path string
		want tracesdk.SamplingDecision
	}

	tests := []struct {
		name            string
		isParentSampled bool
		fields          fields
		cases           []request
	}{
		{
			name:            "empty lists should always be sampled",
			isParentSampled: true,
			cases: []request{
				{path: "/", want: tracesdk.RecordAndSample},
				{path: "/my-path", want: tracesdk.RecordAndSample},
				{path: "/nested/path", want: tracesdk.RecordAndSample},
				{path: "/static/assets/app.css", want: tracesdk.RecordAndSample},
			},
		},
		{
			name:            "only paths on allowlist should be sampled",
			isParentSampled: true,
			fields: fields{
				Allowlist: config.Slice{"/my-path", "/nested"},
			},
			cases: []request{
				{path: "/", want: tracesdk.Drop},
				{path: "/my-path", want: tracesdk.RecordAndSample},
				{path: "/nested/path", want: tracesdk.RecordAndSample},
				{path: "/static/assets/app.css", want: tracesdk.Drop},
			},
		},
		{
			name:            "paths on blocklist should not be sampled",
			isParentSampled: true,
			fields: fields{
				Blocklist: config.Slice{"/static"},
			},
			cases: []request{
				{path: "/", want: tracesdk.RecordAndSample},
				{path: "/my-path", want: tracesdk.RecordAndSample},
				{path: "/nested/path", want: tracesdk.RecordAndSample},
				{path: "/static/assets/app.css", want: tracesdk.Drop},
			},
		},
		{
			name:            "paths on allowlist can be negated by blocklist",
			isParentSampled: true,
			fields: fields{
				Allowlist: config.Slice{"/my-path", "/nested"},
				Blocklist: config.Slice{"/my-path"},
			},
			cases: []request{
				{path: "/", want: tracesdk.Drop},
				{path: "/my-path", want: tracesdk.Drop},
				{path: "/nested/path", want: tracesdk.RecordAndSample},
				{path: "/static/assets/app.css", want: tracesdk.Drop},
			},
		},
		{
			name:            "use parent decision to sample if path is not present: sample",
			isParentSampled: true,
			fields: fields{
				Allowlist: config.Slice{"/my-path", "/nested"},
				Blocklist: config.Slice{"/my-path"},
			},
			cases: []request{
				{path: "", want: tracesdk.RecordAndSample},
			},
		},
		{
			name:            "use parent decision to sample if path is not present: drop",
			isParentSampled: false,
			fields: fields{
				Allowlist: config.Slice{"/my-path", "/nested"},
				Blocklist: config.Slice{"/my-path"},
			},
			cases: []request{
				{path: "", want: tracesdk.Drop},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sampler := new(opentelemetry.ConfiguredURLPrefixSampler).
				Inject(
					&struct {
						Allowlist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.allowlist,optional"`
						Blocklist config.Slice `inject:"config:flamingo.opentelemetry.tracing.sampler.blocklist,optional"`
					}{
						Allowlist: tt.fields.Allowlist,
						Blocklist: tt.fields.Blocklist,
					},
				)

			for _, ttc := range tt.cases {
				t.Run("checking path "+ttc.path, func(t *testing.T) {
					t.Parallel()

					traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
					spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
					pscc := trace.SpanContextConfig{
						TraceID: traceID,
						SpanID:  spanID,
					}

					if tt.isParentSampled {
						pscc.TraceFlags = trace.FlagsSampled
					}

					var values []attribute.KeyValue
					if ttc.path != "" {
						values = []attribute.KeyValue{
							attribute.String("http.target", ttc.path),
						}
					}

					got := sampler.ShouldSample(
						tracesdk.SamplingParameters{
							ParentContext: trace.ContextWithSpanContext(
								context.Background(),
								trace.NewSpanContext(pscc),
							),
							TraceID:    trace.TraceID{},
							Attributes: values,
						},
					)

					assert.Equal(t, ttc.want, got.Decision, "unexpected decision for path %q", ttc.path)
				})
			}
		})
	}
}

func TestConfiguredURLPrefixSampler_Description(t *testing.T) {
	t.Parallel()

	sampler := new(opentelemetry.ConfiguredURLPrefixSampler)
	assert.Equal(t, "ConfiguredURLPrefixSampler", sampler.Description())
}

func TestSpanKindBasedSampler_ShouldSample(t *testing.T) {
	t.Parallel()

	type fields struct {
		root   tracesdk.Sampler
		config map[trace.SpanKind]tracesdk.Sampler
	}

	type args struct {
		kind trace.SpanKind
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   tracesdk.SamplingDecision
	}{
		{
			name: "fall back to root when span kind is not configured",
			fields: fields{
				root: tracesdk.AlwaysSample(),
			},
			args: args{
				kind: trace.SpanKindServer,
			},
			want: tracesdk.RecordAndSample,
		},
		{
			name: "use configured sampler for span kind",
			fields: fields{
				root: tracesdk.AlwaysSample(),
				config: map[trace.SpanKind]tracesdk.Sampler{
					trace.SpanKindClient: tracesdk.NeverSample(),
				},
			},
			args: args{
				kind: trace.SpanKindClient,
			},
			want: tracesdk.Drop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := opentelemetry.SpanKindSampler(tt.fields.root, tt.fields.config)
			assert.Equalf(t,
				tt.want,
				s.ShouldSample(tracesdk.SamplingParameters{
					Kind: tt.args.kind,
				}).Decision,
				"ShouldSample(%v)", tt.args.kind)
		})
	}
}

func TestSpanKindBasedSampler_Description(t *testing.T) {
	t.Parallel()

	s := opentelemetry.SpanKindSampler(tracesdk.AlwaysSample(), map[trace.SpanKind]tracesdk.Sampler{
		trace.SpanKindClient: tracesdk.NeverSample(),
		trace.SpanKindServer: tracesdk.TraceIDRatioBased(.5),
	})

	expectedDescription := fmt.Sprintf("SpanKindBasedSampler{root:%s,config:{%s:%s,%s:%s}}",
		tracesdk.AlwaysSample().Description(),
		trace.SpanKindClient,
		tracesdk.NeverSample().Description(),
		trace.SpanKindServer,
		tracesdk.TraceIDRatioBased(.5).Description(),
	)

	assert.Equal(t, expectedDescription, s.Description())
}
