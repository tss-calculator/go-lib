package grpc

import (
	stdcontext "context"

	applogger "github.com/tss-calculator/go-lib/pkg/application/logger"

	"google.golang.org/grpc"
)

func ClientTraceInterceptor(logger applogger.Logger) grpc.UnaryClientInterceptor {
	return func(ctx stdcontext.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		traceCtx, err := traceContextToMetadata(ctx)
		if err != nil {
			logger.Error(err, "failed append trace context to metadata")
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return invoker(traceCtx, method, req, reply, cc, opts...)
	}
}
