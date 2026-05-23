# DevTool startup

Pre-PMF startup iterating on dispatch + routing logic weekly.

## Use cxr when

- Routing strategy changes faster than handler implementations.
- Need to plug in new providers (custom internal LLM, fine-tuned
  model) without rewriting the dispatch core.
- Want xrr cassettes in CI so cross-handler comparisons are
  reproducible.

## Constraints

- Small team: one or two engineers own the LLM layer.
- Velocity matters more than abstraction perfection.
- Provider auth + rate limits change weekly.

## What they get from cxr

- Thin `Router` + `Executor` core: <500 LOC of dispatch surface to
  hold in head.
- `Handler.Probe` separates capability discovery from execution —
  cache once per session, no per-request shell-out.
- Process-handler helpers (`process.go`) for the common case of a
  CLI-backed handler with stdout/stderr capture.

## Stories

- [US-0003 Resolve handler by tool intersection](../../stories/US-0003-resolve-handler-by-tool.md)
- [US-0005 No-handler error path](../../stories/US-0005-no-handler-error.md)
- [US-0007 Process-handler config flags](../../stories/US-0007-process-handler-config-flags.md)
