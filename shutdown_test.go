package opentelemetry_test

import (
	"context"
	"errors"
	"testing"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"go.opentelemetry.io/otel"
	noopMetric "go.opentelemetry.io/otel/metric/noop"
	noopTrace "go.opentelemetry.io/otel/trace/noop"

	"flamingo.me/opentelemetry"
	"flamingo.me/opentelemetry/mocks"
)

type (
	tracerProvider struct {
		noopTrace.TracerProvider
		mocks.Shutdowner
	}

	meterProvider struct {
		noopMetric.MeterProvider
		mocks.Shutdowner
	}
)

var errShutdown = errors.New("shutdown error")

func TestListener_Notify(t *testing.T) { //nolint:tparallel // no parallel subtests possible because of global state manipulation
	t.Parallel()

	type args struct {
		event flamingo.Event
	}

	tests := []struct {
		name               string
		args               args
		traceShutdownError error
		meterShutdownError error
	}{
		{
			name: "shutdown meter and tracer successfully",
			args: args{
				event: new(flamingo.ShutdownEvent),
			},
			traceShutdownError: nil,
			meterShutdownError: nil,
		},
		{
			name: "error on shutdown meter and tracer",
			args: args{
				event: new(flamingo.ShutdownEvent),
			},
			traceShutdownError: errShutdown,
			meterShutdownError: errShutdown,
		},
	}

	for _, tt := range tests { //nolint:paralleltest // no parallel test possible because of global state manipulation
		t.Run(tt.name, func(t *testing.T) {
			tp := new(tracerProvider)
			tp.Shutdowner.EXPECT().Shutdown(context.Background()).Once().Return(tt.traceShutdownError)
			otel.SetTracerProvider(tp)

			mp := new(meterProvider)
			mp.Shutdowner.EXPECT().Shutdown(context.Background()).Once().Return(tt.meterShutdownError)
			otel.SetMeterProvider(mp)

			l := new(opentelemetry.Listener).Inject(new(flamingo.NullLogger))

			l.Notify(context.Background(), tt.args.event)
		})
	}
}
