// Package cxr — Capability eXecution Router.
// Domain-agnostic dispatch runtime: routes requests to handlers by
// capability/tool intersection, executes subprocesses, and wraps
// xrr record/replay transparently.
package cxr

import "context"

// Request is the unit of dispatch.
type Request struct {
	Prompt       string
	Capabilities []string
	Tools        []string
	Config       map[string]any // operation-scoped; e.g. config["prompt"]["model"]
	Operation    string         // logical hint: "prompt", "embed", "exec", …
}

// Result carries handler output.
type Result struct {
	Output map[string]any
	Raw    []byte
}

// Capabilities holds what a handler discovered at runtime via Probe.
type Capabilities struct {
	Tools  []string
	Models []string
	Flags  []string // known --flag names
}

// Handler is implemented by every backend (claude, gemini, llm, custom, …).
type Handler interface {
	// ID returns the stable identifier for this handler (e.g. "claude").
	ID() string

	// CanHandle reports whether this handler can service req.
	CanHandle(req Request) bool

	// Exec runs the request and returns a structured result.
	Exec(ctx context.Context, req Request) (Result, error)

	// Probe discovers available tools, models, and flags by running the
	// underlying CLI. Caching is the handler's responsibility — implementations
	// typically memoize the result for the handler's lifetime (e.g. via
	// sync.Once). The Router does not cache Probe output.
	Probe(ctx context.Context) (*Capabilities, error)
}
