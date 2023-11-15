package opentelemetry

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestURLPrefixSampler_SampleAll(t *testing.T) {
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

	shouldSample := URLPrefixSampler(nil, nil, true)
	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLPrefixSampler_SampleAllowed(t *testing.T) {
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

	shouldSample := URLPrefixSampler([]string{"/my-path", "/nested"}, nil, true)
	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLPrefixSampler_SampleBlocked(t *testing.T) {
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

	shouldSample := URLPrefixSampler(nil, []string{"/static"}, true)
	for _, tt := range tests {
		t.Run("checking path "+tt.path, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, tt.path, nil)
			assert.NoError(t, err)

			if got := shouldSample(request); got != tt.want {
				t.Errorf("URLPrefixSampler.shouldSample() = %v, want %v", got, tt.want)
			}
		})
	}
}
