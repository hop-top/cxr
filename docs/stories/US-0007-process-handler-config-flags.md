---
status: shipped
personas: [devtool-startup, internal-tool-builder]
priority: P1
---

# US-0007: Translate config map to subprocess flags

As a tool author wrapping a CLI-backed LLM (claude, gemini, llm),
I want a config-to-flags helper so users can express settings in
YAML/JSON and cxr renders them as `--key value` for the
subprocess.

## Use this when

- Your handler shells out to a CLI.
- Users configure the handler via YAML/JSON.
- You want operation-scoped overrides (e.g. embed uses a different
  model than prompt).

## Result

`BuildFlagsFromConfig(cfg, operation, skip)` returns a `[]string` of
`--key value` flags. Sub-maps named after an `operation` override
top-level keys. Keys in `skip` are dropped (used for internal config
like `dir`).

## Steps

Flat config:

```go
cfg := map[string]any{"model": "sonnet", "verbose": true}
flags := cxr.BuildFlagsFromConfig(cfg, "", nil)
// ["--model", "sonnet", "--verbose"]
```

Operation override:

```go
cfg := map[string]any{
    "model":  "sonnet",
    "prompt": map[string]any{"model": "opus"},
}
flags := cxr.BuildFlagsFromConfig(cfg, "prompt", nil)
// flags contains "opus" (override), not "sonnet"
// "--prompt" itself is NOT emitted as a flag
```

Skip list:

```go
cfg := map[string]any{"model": "sonnet", "dir": "/tmp"}
flags := cxr.BuildFlagsFromConfig(cfg, "", []string{"dir"})
// "--dir" omitted
```

## Verify

```bash
go test -run TestBuildFlagsFromConfig ./...
```

Expected: 3 sub-tests PASS.

## How it works

`BuildFlagsFromConfig` walks the map:

1. Skip keys in the `skip` list.
2. If a value is a sub-map AND its key matches the `operation`,
   recurse into it (overrides win).
3. Sub-map keys that don't match the operation are dropped (they
   belong to other operations).
4. Bool `true` emits just `--key`; bool `false` emits nothing.
5. Other scalars emit `--key value`.

## Tests

- [`process_test.go:TestBuildFlagsFromConfig_Flat`](../../process_test.go)
- [`process_test.go:TestBuildFlagsFromConfig_OperationOverride`](../../process_test.go)
- [`process_test.go:TestBuildFlagsFromConfig_Skip`](../../process_test.go)
