# Framework author

Building a higher-level orchestration framework that delegates to cxr.

## Use cxr when

- Framework needs a swappable dispatch substrate; want to expose
  cxr's `Handler` interface as the framework's plugin contract.
- Want xrr-backed determinism so framework tests don't drift.
- Need to compose cxr `Executor`s into a larger pipeline (chain of
  responsibility, fan-out, etc.).

## Constraints

- API stability for downstream consumers — cannot ship cxr's
  internals as-is.
- Documentation overhead: must explain cxr concepts to framework
  users in framework's own vocabulary.

## What they get from cxr

- `Handler` is the only required interface — small enough to embed
  in framework's plugin docs without exposing cxr internals.
- `Capabilities` discovery contract is explicit (`Probe` returns
  Tools/Models/Flags) — frameworks can layer their own capability
  resolution on top.
- `Executor.Exec` is a pure function of `(Router, xrr.Session,
  xrr.Adapter)` — composes cleanly.

## Stories

- [US-0001 Resolve handler by explicit ID](../../stories/US-0001-resolve-handler-by-explicit-id.md)
- [US-0002 Resolve handler by capability intersection](../../stories/US-0002-resolve-handler-by-capability.md)
- [US-0008 Handler.Probe discovery contract](../../stories/US-0008-handler-probe-discovery.md)
