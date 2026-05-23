package cxr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

// reservedConfigKeys are router-only keys that must never be forwarded
// as flags to the subprocess (they steer dispatch, not the CLI).
var reservedConfigKeys = []string{"handler"}

// ProcessHandler dispatches requests by running a subprocess.
// It implements Handler for any CLI-backed backend.
type ProcessHandler struct {
	id     string
	binary string   // resolved binary path (may be overridden at Exec time)
	caps   []string // declared capabilities for CanHandle matching
	tools  []string // declared tools for CanHandle matching

	mu         sync.Mutex
	probeOnce  sync.Once
	probedCaps *Capabilities
	probeErr   error // sticky error from the one-shot probe

	// BuildArgs constructs the argv slice from req (excluding binary).
	// Defaults to BuildFlagsFromConfig + prompt append if nil.
	BuildArgs func(req Request) []string

	// BuildEnv returns additional env vars (KEY=VALUE) to inject.
	BuildEnv func(req Request) []string

	// ParseOutput converts raw stdout into a Result.
	// Defaults to JSON-or-raw fallback if nil.
	ParseOutput func(raw []byte) (Result, error)
}

// NewProcessHandler returns a ProcessHandler with the given id, binary, and
// declared capability/tool sets used for CanHandle matching.
func NewProcessHandler(id, binary string, caps []string, tools []string) *ProcessHandler {
	return &ProcessHandler{id: id, binary: binary, caps: caps, tools: tools}
}

// ID implements Handler.
func (h *ProcessHandler) ID() string { return h.id }

// CanHandle implements Handler: true if any declared cap or tool intersects req.
func (h *ProcessHandler) CanHandle(req Request) bool {
	for _, c := range h.caps {
		for _, rc := range req.Capabilities {
			if c == rc {
				return true
			}
		}
	}
	for _, t := range h.tools {
		for _, rt := range req.Tools {
			if t == rt {
				return true
			}
		}
	}
	return false
}

// Exec implements Handler.
func (h *ProcessHandler) Exec(ctx context.Context, req Request) (Result, error) {
	bin := h.binary
	if bin == "" {
		var err error
		bin, err = exec.LookPath(h.id)
		if err != nil {
			return Result{}, fmt.Errorf("cxr: %s: binary not found: %w", h.id, err)
		}
	}

	var argv []string
	if h.BuildArgs != nil {
		argv = h.BuildArgs(req)
	} else {
		argv = BuildFlagsFromConfig(req.Config, req.Operation, reservedConfigKeys)
		if req.Prompt != "" {
			argv = append(argv, req.Prompt)
		}
	}

	cmd := exec.CommandContext(ctx, bin, argv...)
	if h.BuildEnv != nil {
		cmd.Env = append(cmd.Environ(), h.BuildEnv(req)...)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return Result{}, fmt.Errorf("cxr: %s: exec: %w: %s", h.id, err, strings.TrimSpace(stderr.String()))
	}

	if h.ParseOutput != nil {
		return h.ParseOutput(raw)
	}
	return defaultParseOutput(raw)
}

// Probe implements Handler: runs the binary's --help and caches the result
// (via sync.Once on this handler instance). Errors from binary resolution or
// the subprocess are captured and returned; later calls re-surface the same
// error rather than silently caching an empty Capabilities. A non-zero exit
// code from --help is NOT treated as a failure (many CLIs return 1 for help),
// but ctx cancellation and exec errors are.
func (h *ProcessHandler) Probe(ctx context.Context) (*Capabilities, error) {
	h.probeOnce.Do(func() {
		bin := h.binary
		if bin == "" {
			var err error
			bin, err = exec.LookPath(h.id)
			if err != nil {
				h.mu.Lock()
				h.probeErr = err
				h.mu.Unlock()
				return
			}
		}
		cmd := exec.CommandContext(ctx, bin, "--help")
		out, err := cmd.CombinedOutput()
		// Tolerate non-zero exit via ExitError (CLIs commonly exit 1 on --help).
		// Surface ctx cancellation and exec failures (LookPath-after-fork, etc).
		if err != nil {
			if _, isExit := err.(*exec.ExitError); !isExit {
				h.mu.Lock()
				h.probeErr = err
				h.mu.Unlock()
				return
			}
		}
		h.mu.Lock()
		h.probedCaps = parseHelpFlags(out)
		h.mu.Unlock()
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.probeErr != nil {
		return nil, fmt.Errorf("cxr: %s: probe: %w", h.id, h.probeErr)
	}
	return h.probedCaps, nil
}

// BuildFlagsFromConfig converts a config map into a []string argv fragment.
//
// Logic:
//  1. Collect flat string/bool/int keys from config (skip reserved).
//  2. If operation != "" and config[operation] is map[string]any: merge,
//     operation keys override flat values.
//  3. For each key/val: append "--key", "val" (bool true → flag only).
//  4. Keys in skip are omitted (sub-map keys, reserved names like "dir").
//
// Argv order is deterministic: keys are emitted in lexicographic order so
// xrr cassette recordings are stable across runs.
func BuildFlagsFromConfig(config map[string]any, operation string, skip []string) []string {
	if len(config) == 0 {
		return nil
	}

	skipSet := make(map[string]bool, len(skip))
	for _, s := range skip {
		skipSet[s] = true
	}

	// Merge: flat first, then operation override.
	merged := make(map[string]any, len(config))
	for k, v := range config {
		if _, isMap := v.(map[string]any); isMap {
			skipSet[k] = true // sub-map key — skip as flag
			continue
		}
		merged[k] = v
	}
	if operation != "" {
		if opMap, ok := config[operation].(map[string]any); ok {
			for k, v := range opMap {
				merged[k] = v
			}
		}
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		if skipSet[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var flags []string
	for _, k := range keys {
		switch val := merged[k].(type) {
		case bool:
			if val {
				flags = append(flags, "--"+k)
			}
		case string:
			flags = append(flags, "--"+k, val)
		default:
			flags = append(flags, "--"+k, fmt.Sprintf("%v", val))
		}
	}
	return flags
}

func defaultParseOutput(raw []byte) (Result, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err == nil {
		return Result{Output: m, Raw: raw}, nil
	}
	return Result{Output: map[string]any{"output": strings.TrimSpace(string(raw))}, Raw: raw}, nil
}

// parseHelpFlags does a best-effort parse of --help output to extract flag names.
func parseHelpFlags(out []byte) *Capabilities {
	caps := &Capabilities{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				flag := strings.TrimLeft(strings.Split(parts[0], "=")[0], "-")
				caps.Flags = append(caps.Flags, flag)
			}
		}
	}
	return caps
}
