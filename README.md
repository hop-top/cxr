# cxr — Capability eXecution Router

[![Go Reference](https://pkg.go.dev/badge/hop.top/cxr.svg)](https://pkg.go.dev/hop.top/cxr)

Domain-agnostic dispatch runtime: routes requests to handlers by
capability / tool intersection, executes them (subprocess or in-process),
and wraps execution with [`hop.top/xrr`](https://github.com/hop-top/xrr)
record / replay transparently.

> [!WARNING]
> **Alpha — API and tag history may break.** `cxr/v0.1.x-alpha.*` is the
> first published line. Pin to an exact tag, not a range. Breaking changes
> may land between alpha tags; see [releases](https://github.com/hop-top/cxr/releases).

## What it does

cxr resolves a `Request` (`{Prompt, Capabilities, Tools, Config,
Operation}`) to a `Handler` (claude, gemini, llm, a custom backend, …)
using a deterministic resolution chain:

1. **Explicit handler ID** — `req.Config["handler"]`
2. **Capability intersection** — `handler.caps ∩ req.Capabilities`
3. **Tools intersection** — `handler.tools ∩ req.Tools`
4. **Default handler** — first registered

The chosen handler's `Exec` runs the request; if an `xrr.Session` +
`xrr.Adapter` are wired in, the call is recorded / replayed as a cassette
for deterministic tests.

## Install

```bash
go get hop.top/cxr@latest
```

## Quick start

```go
import (
    "context"
    "hop.top/cxr"
)

router := cxr.NewRouter([]cxr.Handler{claudeHandler, llmHandler})

// No cassette wrapping
exec := cxr.NewExecutor(router, nil, nil)

result, err := exec.Exec(ctx, cxr.Request{
    Prompt:       "summarize this",
    Capabilities: []string{"text"},
    Operation:    "prompt",
})
```

With xrr cassette wrapping:

```go
import "hop.top/xrr"

session := xrr.NewFileSession("testdata/cassettes/foo")
adapter := myCxrXrrAdapter{} // implements xrr.Adapter

exec := cxr.NewExecutor(router, session, adapter)
result, err := exec.Exec(ctx, req)
```

## Implementing a Handler

A `Handler` is any type that satisfies:

```go
type Handler interface {
    ID() string
    CanHandle(req Request) bool
    Exec(ctx context.Context, req Request) (Result, error)
    Probe(ctx context.Context) (*Capabilities, error)
}
```

`Probe` is called by the consumer when capability discovery is needed.
Caching is the handler's responsibility — `ProcessHandler` memoizes the
result for the handler's lifetime via `sync.Once`. The `Router` itself
does not cache `Probe` output.

## Content-based routing

`ClassifyMiddleware` steers dispatch by prompt content. A `Classifier`
labels the prompt; a confident, mapped verdict sets the explicit handler
ID (chain step 1) on a copy of the request. A nil classifier, a
classifier error, an empty / unmapped label, or a confidence below
`MinConfidence` all leave the request untouched — the deterministic
chain above stays authoritative. An explicit `Config["handler"]` set by
the caller always outranks the classifier.

```go
exec.Use(cxr.ClassifyMiddleware(cxr.ClassifierRouting{
    Classifier: myClassifier, // implements cxr.Classifier
    Routes: map[string]cxr.ClassifiedRoute{
        "code": {Handler: "claude"},
    },
    MinConfidence: 0.6,
}))
```

`Executor.Use` accepts any `Middleware` (`func(ctx, Request) Request`);
middleware run in registration order before the router resolves the
request.

## Subprocess execution

`process.go` provides helpers for the common case where a Handler shells
out to a CLI (claude, gemini, llm, …). It captures stdout (returned as
`Result.Raw`) and surfaces stderr in error messages on non-zero exit.
Timeouts and cancellation are driven by the caller's `context.Context`
(passed straight through to `exec.CommandContext`); set a deadline on
`ctx` to bound wall-clock execution. When an `xrr.Session` + `xrr.Adapter`
are wired in via `Executor`, the Request/Result envelope is recorded as a
cassette.

## Status

- **Module:** `hop.top/cxr`
- **First published tag:** `cxr/v0.1.0-alpha.0`
- **Go version:** 1.26+
- **Dependencies:** `hop.top/xrr`

## License

See [LICENSE](LICENSE).
