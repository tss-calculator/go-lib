package grpc

import (
	stdcontext "context"
	stderrors "errors"
	"strconv"

	"github.com/tss-calculator/go-lib/pkg/common/maybe"
	"github.com/tss-calculator/go-lib/pkg/infrastructure/simpletrace/context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

var (
	errUnexpectedTraceID    = stderrors.New("unexpected trace id")
	errUnexpectedDepth      = stderrors.New("unexpected depth")
	errTraceContextNotFound = stderrors.New("trace context not found")
)

const (
	traceIDHeader = "x-trace-id"
	depthHeader   = "x-trace-depth"
)

func traceContextFromMetadata(ctx stdcontext.Context) (stdcontext.Context, error) {
	v := metadata.ValueFromIncomingContext(ctx, traceIDHeader)
	if len(v) != 1 {
		return nil, errors.Wrapf(errUnexpectedTraceID, "received %v", v)
	}
	traceID := v[0]

	v = metadata.ValueFromIncomingContext(ctx, depthHeader)
	if len(v) != 1 {
		return nil, errors.Wrapf(errUnexpectedDepth, "received %v", v)
	}
	depth, err := strconv.Atoi(v[0])
	if err != nil {
		return nil, errors.Wrap(errUnexpectedDepth, err.Error())
	}

	return context.SetTrace(ctx, context.Trace{TraceID: traceID, Depth: depth}), nil
}

func traceContextToMetadata(ctx stdcontext.Context) (stdcontext.Context, error) {
	trace, ok := maybe.Just(context.GetTrace(ctx))
	if !ok {
		return nil, errors.WithStack(errTraceContextNotFound)
	}
	traceMD := metadata.New(map[string]string{
		traceIDHeader: trace.TraceID,
		depthHeader:   strconv.Itoa(trace.Depth),
	})
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		traceMD = metadata.Join(md, traceMD)
	}
	return metadata.NewOutgoingContext(ctx, traceMD), nil
}
