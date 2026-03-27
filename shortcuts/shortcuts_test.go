//go:build darwin

package shortcuts

import (
	"encoding/json"
	"testing"

	"github.com/openelf-labs/apple-tools/testutil"
)

func TestRegister(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	if len(reg.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(reg.Tools))
	}

	for _, name := range []string{"apple_shortcuts_list", "apple_shortcuts_run"} {
		tool := reg.FindTool(name)
		if tool == nil {
			t.Errorf("tool %q not registered", name)
			continue
		}
		if !json.Valid(tool.Parameters) {
			t.Errorf("tool %q has invalid JSON schema", name)
		}
	}
}

func TestRunValidation(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	// Empty name should fail
	_, err := testutil.CallTool(t, reg, "apple_shortcuts_run", map[string]any{"name": ""})
	if err == nil {
		t.Error("expected error for empty shortcut name")
	}

	// Whitespace-only name should fail
	_, err = testutil.CallTool(t, reg, "apple_shortcuts_run", map[string]any{"name": "   "})
	if err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestIntegrationList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "apple_shortcuts_list", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("shortcuts list: %s", result)
}
