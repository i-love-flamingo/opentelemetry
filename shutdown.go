package opentelemetry

import (
	"context"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"go.opentelemetry.io/otel"
)

type (
	Shutdowner interface {
		Shutdown(ctx context.Context) error
	}

	Listener struct {
		logger flamingo.Logger
	}
)

// Inject dependencies
func (l *Listener) Inject(
	logger flamingo.Logger,
) *Listener {
	l.logger = logger

	return l
}

func (l *Listener) Notify(ctx context.Context, event flamingo.Event) {
	if _, ok := event.(*flamingo.ShutdownEvent); ok {
		tp := otel.GetTracerProvider()
		if s, ok := tp.(Shutdowner); ok {
			l.shutdown(ctx, s)
		}

		mp := otel.GetMeterProvider()
		if s, ok := mp.(Shutdowner); ok {
			l.shutdown(ctx, s)
		}
	}
}

func (l *Listener) shutdown(ctx context.Context, s Shutdowner) {
	l.log().Debugf("Shutdown OpenTelemetry: %T", s)

	err := s.Shutdown(ctx)
	if err != nil {
		l.log().Error("", err)
	}
}

func (l *Listener) log() flamingo.Logger {
	return l.logger.
		WithField(flamingo.LogKeyModule, "opentelemetry").
		WithField(flamingo.LogKeyCategory, "Shutdown")
}
