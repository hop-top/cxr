---
status: shipped
personas: [multi-llm-saas, framework-author, app-developer]
priority: P0
---

# US-0001: Resolve handler by explicit ID

As a tool author, I want a request to opt into a named handler via
`Config["handler"]` so users (or my UI) can pin a specific provider
when they care.

## Use this when

- Your product offers an "advanced: choose model" toggle.
- A test pins a specific handler for reproducibility.
- Routing logic upstream of cxr already decided the handler.

## Result

`Router.Route(req)` returns the handler whose `ID()` matches
`req.Config["handler"]`. Match is exact. If the named handler is not
registered, `Route` returns a clear error.

## Steps

```go
router := cxr.NewRouter([]cxr.Handler{claude, gemini})

h, err := router.Route(cxr.Request{
    Config: map[string]any{"handler": "gemini"},
})
// h.ID() == "gemini"
```

## Verify

```bash
go test -run TestRouter_ExplicitID ./...
```

Expected: `ok hop.top/cxr`.

## How it works

`Router.Route` checks `req.Config["handler"]` first (step 1 of the
resolution chain in `router.go`). Capability/tool intersection (steps
2+3) and default fallback (step 4) are skipped when an explicit ID
is set. Unknown IDs short-circuit with `cxr: no handler with id
"<name>" registered`.

## Tests

- [`router_test.go:TestRouter_ExplicitID`](../../router_test.go)
