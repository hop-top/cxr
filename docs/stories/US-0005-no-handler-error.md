---
status: shipped
personas: [devtool-startup]
priority: P2
---

# US-0005: Surface no-handler-registered as a routing error

As a tool author, I want a clear error when `Router.Route` is called
on an empty router, so bootstrap bugs surface immediately rather
than appearing as nil-pointer panics.

## Use this when

- Initialising cxr in a feature flag path and the handler slice
  might be empty.
- Writing tests where you assert the no-handlers contract.

## Result

`Router.Route(req)` returns `(nil, error)` with a descriptive
message: `cxr: no handlers registered`.

## Steps

```go
router := cxr.NewRouter(nil)
_, err := router.Route(cxr.Request{})
// err != nil
```

## Verify

```bash
go test -run TestRouter_NoHandlers ./...
```

## How it works

`Router.Route` falls through the resolution chain. If no handlers
exist (`len(r.handlers) == 0`), step 4 (default) can't fire, so it
returns the error.

## Tests

- [`router_test.go:TestRouter_NoHandlers`](../../router_test.go)
