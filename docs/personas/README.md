# cxr personas

UCP (Universal Consumer Persona): **tool authors**.

cxr is a library; consumers are people building developer tools on top
of it, not end-users running cxr directly. Personas split by stack
position.

## Businesses

- [Multi-LLM SaaS](businesses/multi-llm-saas.md) — product company
  shipping a multi-LLM feature, needs deterministic dispatch + replay
  in CI.
- [DevTool startup](businesses/devtool-startup.md) — pre-PMF startup
  iterating on routing logic, needs fast handler swap-out.

## Individuals

- [Framework author](individuals/framework-author.md) — building a
  higher-level orchestration framework that delegates to cxr.
- [App developer](individuals/app-developer.md) — embedding cxr in a
  product feature, treats it as a building block.
- [Internal-tool builder](individuals/internal-tool-builder.md) —
  one engineer maintaining a company-internal CLI on top of cxr.
