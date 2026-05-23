# Internal-tool builder

One engineer maintaining a company-internal CLI on top of cxr.

## Use cxr when

- Internal users (data scientists, support, ops) need a CLI that
  picks the right LLM per task.
- Cost-tracking and request logging must work for finance.
- Cannot dedicate ongoing maintenance time to dispatch infra.

## Constraints

- Solo maintainer; long bus factor risk.
- Internal users tolerate brittleness less than external users.
- LLM provider keys rotate; auth must not be hard-coded.

## What they get from cxr

- Dispatch core is small enough to grok in an hour.
- `Handler.Probe` lets the CLI surface which models are available
  to the user at session start.
- `xrr` cassettes make local repro of production incidents trivial.

## Stories

- [US-0007 Process-handler config flags](../../stories/US-0007-process-handler-config-flags.md)
- [US-0008 Handler.Probe discovery contract](../../stories/US-0008-handler-probe-discovery.md)
- [US-0009 Handler error wrapping](../../stories/US-0009-handler-error-wrapping.md)
