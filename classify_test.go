package cxr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/cxr"
)

// fixedClassifier returns a canned verdict (or error) for every prompt.
type fixedClassifier struct {
	label      string
	confidence float64
	err        error
}

func (c fixedClassifier) Classify(_ context.Context, _ string) (string, float64, error) {
	return c.label, c.confidence, c.err
}

// classifyRouting is the routing table shared by the middleware tests:
// label "code" steers to handler "b"; everything else falls back.
func classifyRouting(cl cxr.Classifier) cxr.ClassifierRouting {
	return cxr.ClassifierRouting{
		Classifier:    cl,
		Routes:        map[string]cxr.ClassifiedRoute{"code": {Handler: "b"}},
		MinConfidence: 0.6,
	}
}

// twoHandlerExecutor wires handlers a (caps: text) and b (caps: code) so the
// deterministic capability chain resolves "text" requests to a.
func twoHandlerExecutor(mw ...cxr.Middleware) *cxr.Executor {
	a := &stubHandler{id: "a", caps: []string{"text"}}
	b := &stubHandler{id: "b", caps: []string{"code"}}
	e := cxr.NewExecutor(cxr.NewRouter([]cxr.Handler{a, b}), nil, nil)
	e.Use(mw...)
	return e
}

func textRequest() cxr.Request {
	return cxr.Request{Prompt: "refactor this function", Capabilities: []string{"text"}}
}

func TestClassifyMiddleware_LabelSteersRouting(t *testing.T) {
	cl := fixedClassifier{label: "code", confidence: 0.9}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "b", res.Output["handler"], "confident mapped label must steer dispatch")
}

func TestClassifyMiddleware_ErrorFallsBack(t *testing.T) {
	cl := fixedClassifier{err: errors.New("scorer unavailable")}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err, "classifier errors must never fail dispatch")
	assert.Equal(t, "a", res.Output["handler"], "error must fall back to the deterministic chain")
}

func TestClassifyMiddleware_LowConfidenceFallsBack(t *testing.T) {
	cl := fixedClassifier{label: "code", confidence: 0.3}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"], "verdict below MinConfidence must fall back")
}

func TestClassifyMiddleware_NilClassifierIsPassthrough(t *testing.T) {
	mw := cxr.ClassifyMiddleware(classifyRouting(nil))

	req := textRequest()
	out := mw(context.Background(), req)
	assert.Equal(t, req, out, "nil classifier must be a no-op passthrough")

	e := twoHandlerExecutor(mw)
	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"])
}

func TestClassifyMiddleware_EmptyLabelFallsBack(t *testing.T) {
	cl := fixedClassifier{label: "", confidence: 0.9}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"], "empty label means no verdict")
}

func TestClassifyMiddleware_UnmappedLabelFallsBack(t *testing.T) {
	cl := fixedClassifier{label: "poetry", confidence: 0.9}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"], "label absent from Routes must fall back")
}

func TestClassifyMiddleware_ExplicitHandlerWins(t *testing.T) {
	cl := fixedClassifier{label: "code", confidence: 0.9}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(classifyRouting(cl)))

	req := textRequest()
	req.Config = map[string]any{"handler": "a"}
	res, err := e.Exec(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "a", res.Output["handler"], "caller's explicit handler override outranks the classifier")
}

func TestClassifyMiddleware_UnregisteredRouteSurfacesError(t *testing.T) {
	cl := fixedClassifier{label: "code", confidence: 0.9}
	routing := cxr.ClassifierRouting{
		Classifier: cl,
		Routes:     map[string]cxr.ClassifiedRoute{"code": {Handler: "missing"}},
	}
	e := twoHandlerExecutor(cxr.ClassifyMiddleware(routing))

	_, err := e.Exec(context.Background(), textRequest())
	require.Error(t, err, "a confident route to an unregistered handler is a misconfiguration, not ambiguity")
	assert.Contains(t, err.Error(), `"missing"`)
}

func TestClassifyMiddleware_DoesNotMutateCallerConfig(t *testing.T) {
	cl := fixedClassifier{label: "code", confidence: 0.9}
	mw := cxr.ClassifyMiddleware(classifyRouting(cl))

	req := textRequest()
	req.Config = map[string]any{"model": "small"}
	out := mw(context.Background(), req)

	assert.Equal(t, "b", out.Config["handler"])
	_, leaked := req.Config["handler"]
	assert.False(t, leaked, "middleware must copy Config, never mutate the caller's map")
	assert.Equal(t, "small", out.Config["model"], "existing config keys must carry over")
}

func TestClassifierFunc_Adapts(t *testing.T) {
	var cl cxr.Classifier = cxr.ClassifierFunc(
		func(_ context.Context, _ string) (string, float64, error) {
			return "code", 0.8, nil
		},
	)
	label, conf, err := cl.Classify(context.Background(), "x")
	require.NoError(t, err)
	assert.Equal(t, "code", label)
	assert.Equal(t, 0.8, conf)
}

func TestExecutor_MiddlewareOrder(t *testing.T) {
	// Later middleware sees the earlier one's output: first steers to "b",
	// second observes and redirects to "a".
	var sawHandler string
	first := func(_ context.Context, req cxr.Request) cxr.Request {
		cfg := map[string]any{"handler": "b"}
		req.Config = cfg
		return req
	}
	second := func(_ context.Context, req cxr.Request) cxr.Request {
		sawHandler, _ = req.Config["handler"].(string)
		req.Config = map[string]any{"handler": "a"}
		return req
	}
	e := twoHandlerExecutor(first, second)

	res, err := e.Exec(context.Background(), textRequest())
	require.NoError(t, err)
	assert.Equal(t, "b", sawHandler)
	assert.Equal(t, "a", res.Output["handler"])
}
