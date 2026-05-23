---
status: shipped
personas: [multi-llm-saas, app-developer]
priority: P0
---

# US-0006: Record + replay dispatch via xrr cassettes

As a tool author, I want CI to run cxr-dispatched LLM calls against
recorded cassettes so my pipeline is deterministic and free of live
LLM cost.

## Use this when

- You run integration tests that exercise the dispatch + LLM path.
- CI cannot or should not hit real LLM endpoints.
- You want a single test body to work in dev (live) and CI
  (cassette) without conditional branches.

## Result

`Executor.Exec` wraps the handler call in `xrr.Session.Record`. When
the session is in record mode, the LLM call is captured; in replay
mode, the captured response is returned without hitting the network.

## Steps

```go
session := xrr.NewFileSession("testdata/cassettes/foo")
adapter := myCxrXrrAdapter{}                    // implements xrr.Adapter
executor := cxr.NewExecutor(router, session, adapter)

result, err := executor.Exec(ctx, cxr.Request{...})
```

Pass `nil` session + adapter to skip wrapping entirely:

```go
executor := cxr.NewExecutor(router, nil, nil) // no cassette
```

## Verify

The integration is exercised in production callers (cxr ships only
the wiring). The behaviour is enforced by:

- `executor.go:NewExecutor` accepts `(router, session, adapter)`.
- `executor.go:Exec` short-circuits to direct handler call when
  `session == nil || adapter == nil`.
- `executor.go:Exec` calls `session.Record(...)` otherwise.

## How it works

xrr is provider-agnostic — `cxr.execRequest` and `cxr.execResponse`
adapt cxr's `Request`/`Result` to xrr's `Request`/`Response`
interfaces. The cassette stores the `Result` payload; on replay,
`Result.Output` is reconstructed from `xrr.RawResponse`.

## Tests

Integration; see [`executor.go`](../../executor.go) — `NewExecutor`,
`Exec`, `execRequest`, `execResponse`.

## Related

- [hop.top/xrr documentation](https://github.com/hop-top/xrr)
