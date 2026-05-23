---
status: shipped
personas: [multi-llm-saas, app-developer]
priority: P1
---

# US-0004: Default to first registered handler

As a tool author, I want a no-config request to land on a sane
default, so simple call sites don't have to think about capability
sets.

## Use this when

- Request has no `Config["handler"]`, no `Capabilities`, no `Tools`.
- You want to short-circuit configuration and pick a fallback.
- Single-handler setups (one provider) should "just work".

## Result

`Router.Route(req)` returns the first handler in the slice passed
to `NewRouter`. Order is the only knob.

## Steps

```go
claude := myClaudeHandler{}
router := cxr.NewRouter([]cxr.Handler{claude}) // default = claude

h, _ := router.Route(cxr.Request{}) // h is claude
```

## Verify

```bash
go test -run TestRouter_Default ./...
```

## How it works

Resolution chain in `router.go`:

1. Explicit ID
2. Capability intersection (via `CanHandle`)
3. Tool intersection (via `CanHandle`)
4. **Default** — first registered handler

Steps 2+3 are folded into `CanHandle`. If nothing matches and there
is at least one handler registered, step 4 fires.

## Tests

- [`router_test.go:TestRouter_Default`](../../router_test.go)
