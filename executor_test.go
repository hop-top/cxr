package cxr_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/cxr"
)

type errHandler struct {
	id  string
	err error
}

func (e *errHandler) ID() string                   { return e.id }
func (e *errHandler) CanHandle(_ cxr.Request) bool { return true }
func (e *errHandler) Probe(_ context.Context) (*cxr.Capabilities, error) {
	return &cxr.Capabilities{}, nil
}
func (e *errHandler) Exec(_ context.Context, _ cxr.Request) (cxr.Result, error) {
	return cxr.Result{}, e.err
}

// TestExecutor_RouteErrorWrapped: when the router cannot resolve a handler,
// Executor.Exec wraps the error with the "cxr: route:" prefix so callers can
// distinguish dispatch failures from handler failures.
func TestExecutor_RouteErrorWrapped(t *testing.T) {
	r := cxr.NewRouter(nil) // empty router → Route returns error
	exec := cxr.NewExecutor(r, nil, nil)

	_, err := exec.Exec(context.Background(), cxr.Request{})
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "cxr: route:"),
		"want prefix %q, got %q", "cxr: route:", err.Error())
}

// TestExecutor_HandlerErrorNotWrapped: when the handler fails, Executor.Exec
// returns the handler's error verbatim (no "cxr: route:" prefix), so callers
// can errors.Is/As against handler-defined sentinels.
func TestExecutor_HandlerErrorNotWrapped(t *testing.T) {
	sentinel := errors.New("handler exploded")
	h := &errHandler{id: "boom", err: sentinel}
	r := cxr.NewRouter([]cxr.Handler{h})
	exec := cxr.NewExecutor(r, nil, nil)

	_, err := exec.Exec(context.Background(), cxr.Request{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel),
		"handler error must propagate unwrapped; got %v", err)
	assert.False(t, strings.HasPrefix(err.Error(), "cxr: route:"),
		"handler errors must not carry route prefix; got %q", err.Error())
}

// TestExecutor_NilSessionShortCircuits: with nil session/adapter the
// Executor must call handler.Exec directly (no xrr wrapping) and return
// the handler's Result unchanged.
func TestExecutor_NilSessionShortCircuits(t *testing.T) {
	h := &stubHandler{id: "a", caps: []string{"text"}}
	r := cxr.NewRouter([]cxr.Handler{h})
	exec := cxr.NewExecutor(r, nil, nil)

	res, err := exec.Exec(context.Background(), cxr.Request{Capabilities: []string{"text"}})
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"])
}
