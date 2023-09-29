package http

import (
	stdcontext "context"
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/tss-calculator/go-lib/pkg/common/maybe"
	"github.com/tss-calculator/go-lib/pkg/infrastructure/simpletrace/context"

	"github.com/pkg/errors"
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

func traceContextFromHeaders(ctx stdcontext.Context, headers http.Header) (stdcontext.Context, error) {
	traceID := headers.Get(traceIDHeader)
	if traceID == "" {
		return nil, errors.Wrap(errUnexpectedTraceID, "trace id header not found")
	}

	v := headers.Get(depthHeader)
	if v == "" {
		return nil, errors.Wrap(errUnexpectedDepth, "depth header not found")
	}
	depth, err := strconv.Atoi(v)
	if err != nil {
		return nil, errors.Wrap(errUnexpectedDepth, err.Error())
	}

	return context.SetTrace(ctx, context.Trace{TraceID: traceID, Depth: depth}), nil
}

func traceContextToRequest(request *http.Request) (*http.Request, error) {
	trace, ok := maybe.Just(context.GetTrace(request.Context()))
	if !ok {
		return nil, errors.WithStack(errTraceContextNotFound)
	}
	request.Header.Add(traceIDHeader, trace.TraceID)
	request.Header.Add(depthHeader, strconv.Itoa(trace.Depth))
	return request, nil
}
