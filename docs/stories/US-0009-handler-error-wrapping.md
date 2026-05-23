---
status: shipped
personas: [app-developer, internal-tool-builder]
priority: P2
---

# US-0009: Wrap handler errors with route context

As a tool author, when `Executor.Exec` fails I want the error to
say which step of dispatch broke — routing vs handler — so logs and
incident triage point to the right place.

## Use this when

- Production error reaches your logging pipeline and you need to
  distinguish "no handler matched" from "handler ran but failed".

## Result

- Routing failures (`Router.Route` returned error) wrap with
  `cxr: route: %w`.
- Handler execution errors propagate as-is from `Handler.Exec`.
- The two are distinguishable by error string prefix.

## Steps

No special action — `Executor.Exec` does this automatically:

```go
result, err := executor.Exec(ctx, req)
if err != nil {
    // err == "cxr: route: <reason>"     → routing problem
    // err == "<provider-specific>"      → handler problem
}
```

## Verify

The wrapping is enforced in `executor.go:Exec`:

```go
h, err := e.router.Route(req)
if err != nil {
    return Result{}, fmt.Errorf("cxr: route: %w", err)
}
```

## How it works

Routing errors come from `Router.Route` (e.g. no handler with that
ID, or no handlers registered). Wrapping them with the `cxr: route:`
prefix gives downstream code a stable substring to match on, while
preserving the underlying error via `%w` for `errors.Is`/`errors.As`.

## Tests

Contract; see [`executor.go`](../../executor.go) — `Exec` error
wrapping path.
