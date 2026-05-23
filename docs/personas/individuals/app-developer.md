# App developer

Embedding cxr in a product feature; treats it as a building block.

## Use cxr when

- Feature needs to route a user prompt to one of N LLM handlers.
- Don't want to build dispatch + cassette plumbing from scratch.
- Acceptable to write one custom `Handler` for your in-house model.

## Constraints

- Single feature scope; not building a framework.
- Acceptance criteria expressed as user stories, mapped to tests.
- LLM cost watched per-PR.

## What they get from cxr

- Drop-in `Router` + `Executor`: write a `Handler`, register it,
  done.
- `xrr.Session` integration prebuilt; just provide a cassette dir.
- Tests run with CGO_ENABLED=0 (no native deps for app builds).

## Stories

- [US-0001 Resolve handler by explicit ID](../../stories/US-0001-resolve-handler-by-explicit-id.md)
- [US-0004 Default to first handler](../../stories/US-0004-default-handler.md)
- [US-0006 Cassette-replayed dispatch in CI](../../stories/US-0006-cassette-replay-ci.md)
- [US-0009 Handler error wrapping](../../stories/US-0009-handler-error-wrapping.md)
