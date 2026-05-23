# cxr stories

Tool-author user stories. Each story is one page, intent-driven shape:
**Use this when / Result / Steps / Verify / How it works / Tests.**

| ID  | Title | Tests |
|-----|-------|-------|
| [US-0001](US-0001-resolve-handler-by-explicit-id.md) | Resolve handler by explicit ID | `router_test.go:TestRouter_ExplicitID` |
| [US-0002](US-0002-resolve-handler-by-capability.md) | Resolve handler by capability intersection | `router_test.go:TestRouter_CapabilityMatch`, `process_test.go:TestProcessHandler_CanHandle` |
| [US-0003](US-0003-resolve-handler-by-tool.md) | Resolve handler by tool intersection | `process_test.go:TestProcessHandler_CanHandle_Tools` |
| [US-0004](US-0004-default-handler.md) | Default to first registered handler | `router_test.go:TestRouter_Default` |
| [US-0005](US-0005-no-handler-error.md) | Surface no-handler-registered as a routing error | `router_test.go:TestRouter_NoHandlers` |
| [US-0006](US-0006-cassette-replay-ci.md) | Record + replay dispatch via xrr cassettes | (integration; see executor.go xrr wiring) |
| [US-0007](US-0007-process-handler-config-flags.md) | Translate config map to subprocess flags | `process_test.go:TestBuildFlagsFromConfig_Flat`, `TestBuildFlagsFromConfig_OperationOverride`, `TestBuildFlagsFromConfig_Skip` |
| [US-0008](US-0008-handler-probe-discovery.md) | Discover handler capabilities at session start | (contract; `Handler.Probe` in `cxr.go`) |
| [US-0009](US-0009-handler-error-wrapping.md) | Wrap handler errors with route context | (executor `cxr: route: %w` wrapping) |

UCP: tool authors. See [personas/](../personas/README.md) for the five
roles these stories serve.
