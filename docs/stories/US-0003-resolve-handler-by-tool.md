---
status: shipped
personas: [devtool-startup]
priority: P1
---

# US-0003: Resolve handler by tool intersection

As a tool author building an agent loop, I want cxr to pick a
handler whose tool set overlaps with the request's tool needs, so
the resolver does the right thing even when capability isn't the
discriminator.

## Use this when

- Request comes with `Tools: []string{"Bash", "Read"}`.
- Handlers differ by which tools they expose (e.g., one wraps a
  shell-capable CLI, another doesn't).
- Capability alone is too coarse.

## Result

`ProcessHandler.CanHandle` returns `true` when any of `req.Tools`
intersects the handler's declared tools.

## Steps

```go
claudeWithBash := cxr.NewProcessHandler(
    "claude", "claude-cli",
    nil,                            // no capabilities
    []string{"Bash", "Read", "Write"}, // tools
)
h, _ := router.Route(cxr.Request{Tools: []string{"Bash"}})
```

## Verify

```bash
go test -run TestProcessHandler_CanHandle_Tools ./...
```

Expected: PASS.

## How it works

cxr's resolution chain checks `CanHandle` for every handler in
order. `ProcessHandler.CanHandle` ORs capability and tool
intersection — either match wins.

## Tests

- [`process_test.go:TestProcessHandler_CanHandle_Tools`](../../process_test.go)
