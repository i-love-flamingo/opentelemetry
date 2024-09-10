package opentelemetry_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"flamingo.me/opentelemetry"
)

func TestURLPrefixSampler_SampleAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{
			path: "/",
			want: true,
		},
		{
			path: "/my-path",
			want: true,
		},
		{
			path: "/nested/path",
			want: true,
		},
		{
			path: "/static/assets/app.css",
			want: true,
		},
	}

	shouldSample := opentelemetry.URLPrefixSampler(nil, nil, true)

	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			t.Parallel()

			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLPrefixSampler_SampleAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{
			path: "/",
			want: false,
		},
		{
			path: "/my-path",
			want: true,
		},
		{
			path: "/nested/path",
			want: true,
		},
		{
			path: "/static/assets/app.css",
			want: false,
		},
	}

	shouldSample := opentelemetry.URLPrefixSampler([]string{"/my-path", "/nested"}, nil, true)

	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			t.Parallel()

			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLPrefixSampler_SampleBlocked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{
			path: "/",
			want: true,
		},
		{
			path: "/my-path",
			want: true,
		},
		{
			path: "/nested/path",
			want: true,
		},
		{
			path: "/static/assets/app.css",
			want: false,
		},
	}

	shouldSample := opentelemetry.URLPrefixSampler(nil, []string{"/static"}, true)

	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			t.Parallel()

			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}
