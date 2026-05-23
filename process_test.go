package cxr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/cxr"
)

func TestBuildFlagsFromConfig_Flat(t *testing.T) {
	cfg := map[string]any{"model": "sonnet", "verbose": true}
	flags := cxr.BuildFlagsFromConfig(cfg, "", nil)
	assert.Contains(t, flags, "--model")
	assert.Contains(t, flags, "sonnet")
	assert.Contains(t, flags, "--verbose")
}

func TestBuildFlagsFromConfig_OperationOverride(t *testing.T) {
	cfg := map[string]any{
		"model":  "sonnet",
		"prompt": map[string]any{"model": "opus"},
	}
	flags := cxr.BuildFlagsFromConfig(cfg, "prompt", nil)
	// "prompt" sub-map key itself must not appear as a flag
	assert.NotContains(t, flags, "--prompt")
	// operation override: opus wins
	assert.Contains(t, flags, "opus")
	assert.NotContains(t, flags, "sonnet")
}

func TestBuildFlagsFromConfig_Skip(t *testing.T) {
	cfg := map[string]any{"model": "sonnet", "dir": "/tmp"}
	flags := cxr.BuildFlagsFromConfig(cfg, "", []string{"dir"})
	assert.NotContains(t, flags, "--dir")
	assert.Contains(t, flags, "--model")
}

func TestProcessHandler_CanHandle(t *testing.T) {
	h := cxr.NewProcessHandler("claude", "", []string{"text", "code"}, nil)
	assert.True(t, h.CanHandle(cxr.Request{Capabilities: []string{"code"}}))
	assert.False(t, h.CanHandle(cxr.Request{Capabilities: []string{"image"}}))
}

func TestProcessHandler_CanHandle_Tools(t *testing.T) {
	h := cxr.NewProcessHandler("claude", "", nil, []string{"Bash", "Read"})
	assert.True(t, h.CanHandle(cxr.Request{Tools: []string{"Bash"}}))
	assert.False(t, h.CanHandle(cxr.Request{Tools: []string{"Write"}}))
}
