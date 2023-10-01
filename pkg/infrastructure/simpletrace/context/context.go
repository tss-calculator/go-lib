package context

import (
	stdcontext "context"

	"github.com/tss-calculator/go-lib/pkg/common/maybe"
)

var (
	ctxTraceID = struct{}{}
	ctxDepth   = struct{}{}
)

type Trace struct {
	TraceID string
	Depth   int
}

func GetTrace(ctx stdcontext.Context) maybe.Maybe[Trace] {
	var trace Trace

	traceID, ok := ctx.Value(ctxTraceID).(string)
	if !ok {
		return maybe.None[Trace]()
	}
	trace.TraceID = traceID

	depth, ok := ctx.Value(ctxDepth).(int)
	if !ok {
		return maybe.None[Trace]()
	}
	trace.Depth = depth

	return maybe.New(trace)
}

func SetTrace(ctx stdcontext.Context, trace Trace) stdcontext.Context {
	maybeTrace := GetTrace(ctx)
	v, ok := maybe.Just(maybeTrace)
	if !ok {
		traceCtx := setTraceID(ctx, trace.TraceID)
		return setDepth(traceCtx, trace.Depth)
	}
	if v.TraceID != trace.TraceID {
		return ctx
	}
	if v.Depth > trace.Depth {
		return setDepth(ctx, trace.Depth)
	}
	return ctx
}

func setTraceID(ctx stdcontext.Context, traceID string) stdcontext.Context {
	return stdcontext.WithValue(ctx, ctxTraceID, traceID)
}

func setDepth(ctx stdcontext.Context, depth int) stdcontext.Context {
	return stdcontext.WithValue(ctx, ctxDepth, depth)
}
