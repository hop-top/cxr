// Classifier middleware: content-based dispatch steering.
//
// The package owns only the seam — [Classifier], the [ClassifierRouting]
// table, and [ClassifyMiddleware]. Concrete classifiers (score routers,
// embedding-based intent detectors, remote scorers, ...) live with the
// caller and are wired in at construction; cxr never imports them.
//
// The deterministic resolution chain (explicit ID, capability/tool
// intersection, default — see [Router]) remains authoritative: a nil
// classifier, a classifier error, an empty or unmapped label, or a
// confidence below [ClassifierRouting.MinConfidence] all leave the
// request untouched and dispatch falls through unchanged. Only a
// confident, mapped verdict steers.
package cxr

import "context"

// Classifier classifies a prompt into a routing label with a confidence
// in [0,1]. Implementations may call remote scorers; errors are never
// fatal to dispatch — [ClassifyMiddleware] treats them as "classifier
// unavailable" and falls back to the deterministic resolution chain.
//
// The shape deliberately mirrors the classifier seam in
// hop.top/kit/go/ai/llm (PromptClassifier / PickProviderClassified), so
// binding a kit classifier is a one-liner:
//
//	cxr.ClassifierFunc(func(ctx context.Context, prompt string) (string, float64, error) {
//		c, err := kitClassifier.Classify(ctx, prompt)
//		return c.Label, c.Confidence, err
//	})
type Classifier interface {
	Classify(ctx context.Context, prompt string) (label string, confidence float64, err error)
}

// ClassifierFunc adapts a plain function to [Classifier].
type ClassifierFunc func(ctx context.Context, prompt string) (string, float64, error)

// Classify implements [Classifier].
func (f ClassifierFunc) Classify(ctx context.Context, prompt string) (string, float64, error) {
	return f(ctx, prompt)
}

// ClassifiedRoute is the dispatch override applied when its label wins.
// An empty Handler is a pure annotation: the request passes through to
// the deterministic chain unchanged.
type ClassifiedRoute struct {
	// Handler is the handler ID dispatched to when this label wins.
	Handler string
}

// ClassifierRouting binds a classifier to its label routes.
type ClassifierRouting struct {
	// Classifier produces the verdict. Nil disables the stage entirely —
	// every request takes the deterministic chain.
	Classifier Classifier
	// Routes maps verdict labels to overrides. Labels absent from the
	// map are treated as ambiguous and fall back.
	Routes map[string]ClassifiedRoute
	// MinConfidence rejects verdicts below it as ambiguous. Zero accepts
	// any successful classification.
	MinConfidence float64
}

// Middleware transforms a Request before the router resolves it.
// Middleware run in registration order; each receives the previous
// one's output. See [Executor.Use].
type Middleware func(ctx context.Context, req Request) Request

// ClassifyMiddleware returns a [Middleware] that steers dispatch by
// prompt content. A confident, mapped verdict sets the explicit handler
// ID (resolution chain step 1) on a copy of the request's Config; the
// caller's map is never mutated.
//
// Fallback semantics — the request passes through untouched when:
//   - routing.Classifier is nil,
//   - the request already carries an explicit Config["handler"]
//     (caller intent outranks the classifier),
//   - the classifier returns an error,
//   - the label is empty or absent from routing.Routes,
//   - the confidence is below routing.MinConfidence,
//   - the winning route's Handler is empty.
//
// A confident verdict routed to an unregistered handler ID surfaces as
// a routing error from [Router.Route] — that is a route
// misconfiguration, not classifier ambiguity, and silently retrying the
// deterministic chain would hide it from operators.
func ClassifyMiddleware(routing ClassifierRouting) Middleware {
	return func(ctx context.Context, req Request) Request {
		if routing.Classifier == nil {
			return req
		}
		if id, ok := req.Config["handler"].(string); ok && id != "" {
			return req
		}
		label, confidence, err := routing.Classifier.Classify(ctx, req.Prompt)
		if err != nil || label == "" || confidence < routing.MinConfidence {
			return req
		}
		route, ok := routing.Routes[label]
		if !ok || route.Handler == "" {
			return req
		}

		cfg := make(map[string]any, len(req.Config)+1)
		for k, v := range req.Config {
			cfg[k] = v
		}
		cfg["handler"] = route.Handler
		req.Config = cfg
		return req
	}
}
