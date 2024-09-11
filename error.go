package opentelemetry

import (
	"flamingo.me/flamingo/v3/framework/flamingo"
)

type (
	errorHandler struct {
		logger flamingo.Logger
	}
)

func newErrorHandler(logger flamingo.Logger) *errorHandler {
	return &errorHandler{
		logger: logger,
	}
}

func (e *errorHandler) Handle(err error) {
	e.logger.
		WithField(flamingo.LogKeyModule, "opentelemetry").
		WithField(flamingo.LogKeyCategory, "internal").
		Error(err)
}
