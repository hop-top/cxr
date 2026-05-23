# Multi-LLM SaaS

Product company shipping a feature that calls multiple LLM providers.

## Use cxr when

- Product offers users a choice of model/provider (or routes
  automatically).
- Engineering needs cassette-replayed tests in CI to avoid live LLM
  cost + flakiness.
- Provider mix changes faster than the product code; want a stable
  abstraction over claude/gemini/llm/openai/custom.

## Constraints

- CI budget: cannot hit real LLMs on every PR.
- Compliance: must log every dispatch with capability + tool intent.
- Latency: cxr overhead must stay <1ms; the LLM call dominates.

## What they get from cxr

- `Router.Route` resolves capability/tool intent to a handler
  deterministically — testable in isolation.
- `Executor.Exec` wraps the handler call with `xrr.Session` so the
  same test runs against cassettes in CI and live providers in dev.
- `Handler` interface: implement once per provider; new providers
  drop in without touching dispatch.

## Stories

- [US-0001 Resolve handler by explicit ID](../../stories/US-0001-resolve-handler-by-explicit-id.md)
- [US-0002 Resolve handler by capability intersection](../../stories/US-0002-resolve-handler-by-capability.md)
- [US-0004 Default to first handler](../../stories/US-0004-default-handler.md)
- [US-0006 Cassette-replayed dispatch in CI](../../stories/US-0006-cassette-replay-ci.md)
