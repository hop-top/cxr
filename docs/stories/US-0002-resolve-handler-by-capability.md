---
status: shipped
personas: [multi-llm-saas, framework-author]
priority: P0
---

# US-0002: Resolve handler by capability intersection

As a tool author, I want cxr to pick a handler whose declared
capabilities overlap with the request, so users get the right model
without naming it.

## Use this when

- Request has `Capabilities: []string{"image"}` or similar.
- Multiple handlers are registered; you want cxr to pick.
- Capability set is small and stable (text, image, code, audio, …).

## Result

`Router.Route(req)` returns the first registered handler whose
`CanHandle(req)` returns `true`. Order in `NewRouter([]Handler{...})`
is significant — earlier handlers win ties.

## Steps

```go
claude := myClaudeHandler{} // declares caps: ["text", "code"]
gemini := myGeminiHandler{} // declares caps: ["image"]
router := cxr.NewRouter([]cxr.Handler{claude, gemini})

h, _ := router.Route(cxr.Request{Capabilities: []string{"image"}})
// h is gemini
```

## Verify

```bash
go test -run TestRouter_CapabilityMatch ./...
go test -run TestProcessHandler_CanHandle ./...
```

Expected: both PASS.

## How it works

`CanHandle` is the handler's own predicate — cxr does not introspect
caps. The built-in `ProcessHandler` matches when any of the request's
capabilities appears in its declared capability set.

## Tests

- [`router_test.go:TestRouter_CapabilityMatch`](../../router_test.go)
- [`process_test.go:TestProcessHandler_CanHandle`](../../process_test.go)
