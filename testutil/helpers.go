package testutil

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
)

// CallTool dispatches a tool call by name with the given params map.
func CallTool(t *testing.T, reg *MockRegistry, toolName string, params map[string]any) (string, error) {
	t.Helper()
	tool := reg.FindTool(toolName)
	if tool == nil {
		t.Fatalf("tool %q not registered", toolName)
	}
	input, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal params: %v", err)
	}
	return tool.Handler(context.Background(), json.RawMessage(input))
}

// MustCallTool is like CallTool but fails on error.
func MustCallTool(t *testing.T, reg *MockRegistry, toolName string, params map[string]any) string {
	t.Helper()
	result, err := CallTool(t, reg, toolName, params)
	if err != nil {
		t.Fatalf("tool %q error: %v", toolName, err)
	}
	return result
}

// NewRegistryWith creates a MockRegistry and calls the register function.
func NewRegistryWith(registerFn func(r core.Registry)) *MockRegistry {
	reg := &MockRegistry{}
	registerFn(reg)
	return reg
}
