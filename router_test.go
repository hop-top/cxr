package cxr_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/cxr"
)

type stubHandler struct {
	id   string
	caps []string
}

func (s *stubHandler) ID() string { return s.id }
func (s *stubHandler) CanHandle(req cxr.Request) bool {
	for _, c := range s.caps {
		for _, rc := range req.Capabilities {
			if c == rc {
				return true
			}
		}
	}
	return false
}
func (s *stubHandler) Exec(_ context.Context, _ cxr.Request) (cxr.Result, error) {
	return cxr.Result{Output: map[string]any{"handler": s.id}}, nil
}
func (s *stubHandler) Probe(_ context.Context) (*cxr.Capabilities, error) {
	return &cxr.Capabilities{}, nil
}

func TestRouter_ExplicitID(t *testing.T) {
	a := &stubHandler{id: "a", caps: []string{"text"}}
	b := &stubHandler{id: "b", caps: []string{"image"}}
	r := cxr.NewRouter([]cxr.Handler{a, b})

	h, err := r.Route(cxr.Request{
		Config: map[string]any{"handler": "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, "b", h.ID())
}

func TestRouter_ExplicitID_NotFound(t *testing.T) {
	a := &stubHandler{id: "a", caps: []string{"text"}}
	r := cxr.NewRouter([]cxr.Handler{a})

	_, err := r.Route(cxr.Request{
		Config: map[string]any{"handler": "missing"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"missing"`)
}

func TestRouter_CapabilityMatch(t *testing.T) {
	a := &stubHandler{id: "a", caps: []string{"text"}}
	b := &stubHandler{id: "b", caps: []string{"image"}}
	r := cxr.NewRouter([]cxr.Handler{a, b})

	h, err := r.Route(cxr.Request{Capabilities: []string{"image"}})
	require.NoError(t, err)
	assert.Equal(t, "b", h.ID())
}

func TestRouter_Default(t *testing.T) {
	a := &stubHandler{id: "a", caps: []string{"text"}}
	r := cxr.NewRouter([]cxr.Handler{a})

	h, err := r.Route(cxr.Request{})
	require.NoError(t, err)
	assert.Equal(t, "a", h.ID())
}

func TestRouter_NoHandlers(t *testing.T) {
	r := cxr.NewRouter(nil)
	_, err := r.Route(cxr.Request{})
	assert.Error(t, err)
}
