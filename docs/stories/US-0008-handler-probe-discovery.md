---
status: shipped
personas: [framework-author, internal-tool-builder]
priority: P1
---

# US-0008: Discover handler capabilities at session start

As a tool author, I want each handler to publish what tools, models,
and flags it actually supports at runtime — so my UI can render
available choices, and my router can match `req.Capabilities`
intelligently.

## Use this when

- Handler wraps a CLI whose capabilities depend on installed
  version, license tier, or environment.
- You want to cache capability discovery once per session, not per
  request.
- Building a `c12n doctor`-style command that surfaces what the
  user can do.

## Result

`Handler.Probe(ctx)` returns `*Capabilities{Tools, Models, Flags}`.
Caller is expected to call it once per session; cxr does not
auto-cache.

## Steps

```go
type Handler interface {
    ID() string
    CanHandle(req Request) bool
    Exec(ctx context.Context, req Request) (Result, error)
    Probe(ctx context.Context) (*Capabilities, error)
}
```

Implement `Probe`:

```go
func (h *MyHandler) Probe(ctx context.Context) (*cxr.Capabilities, error) {
    out, err := exec.CommandContext(ctx, h.bin, "--help").Output()
    if err != nil { return nil, err }
    return &cxr.Capabilities{
        Tools:  parseTools(out),
        Models: parseModels(out),
        Flags:  parseFlags(out),
    }, nil
}
```

## Verify

The contract is enforced at the type level; any handler
implementation must satisfy the `Handler` interface. cxr's stub
handler in tests (`router_test.go:stubHandler`) returns an empty
`Capabilities{}` to demonstrate the no-op case.

## How it works

`Probe` is intentionally session-scoped, not request-scoped. The
caller decides when to refresh — typically:

- Once at process start.
- After a credential rotation.
- When user invokes `<tool> doctor` or equivalent.

## Tests

Contract; see [`cxr.go`](../../cxr.go) — `Handler.Probe`.
