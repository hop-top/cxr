package cxr

import (
	"context"
	"fmt"

	"hop.top/xrr"
)

// Executor routes requests to handlers and wraps execution with xrr
// record/replay via an xrr.Adapter.
type Executor struct {
	router     *Router
	session    xrr.Session // nil → no cassette wrapping
	adapter    xrr.Adapter // nil → no cassette wrapping
	middleware []Middleware
}

// NewExecutor returns an Executor that routes via r and records/replays
// via session+adapter. Pass nil session to skip cassette wrapping.
func NewExecutor(router *Router, session xrr.Session, adapter xrr.Adapter) *Executor {
	return &Executor{router: router, session: session, adapter: adapter}
}

// Use appends middleware to the chain. Exec applies them in
// registration order, each receiving the previous one's output, before
// the router resolves the request.
func (e *Executor) Use(mw ...Middleware) {
	e.middleware = append(e.middleware, mw...)
}

// Exec applies the middleware chain, resolves req to a handler, and
// runs it, wrapping with xrr when a session is configured.
func (e *Executor) Exec(ctx context.Context, req Request) (Result, error) {
	for _, mw := range e.middleware {
		req = mw(ctx, req)
	}

	h, err := e.router.Route(req)
	if err != nil {
		return Result{}, fmt.Errorf("cxr: route: %w", err)
	}

	if e.session == nil || e.adapter == nil {
		return h.Exec(ctx, req)
	}

	xrrReq := &execRequest{req: req}
	var result Result
	resp, err := e.session.Record(ctx, e.adapter, xrrReq, func() (xrr.Response, error) {
		r, execErr := h.Exec(ctx, req)
		if execErr != nil {
			return nil, execErr
		}
		return &execResponse{result: r}, nil
	})
	if err != nil {
		return Result{}, err
	}
	if resp != nil {
		if er, ok := resp.(*execResponse); ok {
			result = er.result
		} else if rr, ok := resp.(*xrr.RawResponse); ok {
			result = Result{Output: rr.Payload}
		}
	}
	return result, nil
}

// execRequest adapts cxr.Request to xrr.Request.
type execRequest struct{ req Request }

func (r *execRequest) AdapterID() string { return "cxr/exec" }

// execResponse adapts cxr.Result to xrr.Response.
type execResponse struct{ result Result }

func (r *execResponse) AdapterID() string { return "cxr/exec" }
