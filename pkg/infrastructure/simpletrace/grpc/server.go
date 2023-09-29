package grpc

import (
	stdcontext "context"

	applogger "github.com/tss-calculator/go-lib/pkg/application/logger"

	"google.golang.org/grpc"
)

func ServerTraceInterceptor(logger applogger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx stdcontext.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		traceCtx, err := traceContextFromMetadata(ctx)
		if err != nil {
			logger.Error(err, "failed fetch trace context from metadata")
			return handler(ctx, req)
		}
		return handler(traceCtx, req)
	}
}
