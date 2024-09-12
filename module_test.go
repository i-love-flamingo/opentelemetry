package opentelemetry_test

import (
	"testing"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/opentelemetry"
)

type (
	loggerModule struct{}
)

// Configure DI
func (m *loggerModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(flamingo.Logger)).To(new(flamingo.NullLogger))
}

func TestModule_Configure(t *testing.T) {
	t.Parallel()

	if err := config.TryModules(nil, new(loggerModule), new(opentelemetry.Module)); err != nil {
		t.Error(err)
	}
}
