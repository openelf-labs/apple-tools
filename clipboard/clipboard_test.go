//go:build darwin

package clipboard

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

	for _, name := range []string{"apple_clipboard_read", "apple_clipboard_write"} {
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

func TestWriteValidation(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	_, err := testutil.CallTool(t, reg, "apple_clipboard_write", map[string]any{"text": ""})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestIntegrationReadWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	reg := &testutil.MockRegistry{}
	Register(reg)

	// Write
	result, err := testutil.CallTool(t, reg, "apple_clipboard_write", map[string]any{"text": "apple-tools-test"})
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	t.Logf("write result: %s", result)

	// Read back
	result, err = testutil.CallTool(t, reg, "apple_clipboard_read", map[string]any{})
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	t.Logf("read result: %s", result)
}
