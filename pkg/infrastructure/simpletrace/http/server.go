package http

import (
	"net/http"

	applogger "github.com/tss-calculator/go-lib/pkg/application/logger"
)

func HandlerWithTrace(logger applogger.Logger, handler http.Handler) http.Handler {
	return &handlerImpl{
		logger:  logger,
		handler: handler,
	}
}

type handlerImpl struct {
	logger  applogger.Logger
	handler http.Handler
}

func (h *handlerImpl) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	traceCtx, err := traceContextFromHeaders(request.Context(), request.Header)
	if err != nil {
		h.logger.Error(err, "failed fetch trace context from headers")
		h.handler.ServeHTTP(writer, request)
		return
	}
	h.handler.ServeHTTP(writer, request.WithContext(traceCtx))
}
