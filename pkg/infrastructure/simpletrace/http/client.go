package http

import (
	"net/http"

	applogger "github.com/tss-calculator/go-lib/pkg/application/logger"
)

func RoundTripperWithTrace(logger applogger.Logger, roundTripper http.RoundTripper) http.RoundTripper {
	return &roundTripperImpl{
		logger:       logger,
		roundTripper: roundTripper,
	}
}

type roundTripperImpl struct {
	logger       applogger.Logger
	roundTripper http.RoundTripper
}

func (r *roundTripperImpl) RoundTrip(request *http.Request) (*http.Response, error) {
	requestWithTrace, err := traceContextToRequest(request)
	if err != nil {
		r.logger.Error(err, "failed append trace context to request")
		return r.roundTripper.RoundTrip(request)
	}
	return r.roundTripper.RoundTrip(requestWithTrace)
}
