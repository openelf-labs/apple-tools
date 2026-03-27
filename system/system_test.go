//go:build darwin

package system

import (
	"encoding/json"
	"testing"

	"github.com/openelf-labs/apple-tools/testutil"
)

func TestRegister(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	expected := []string{"apple_system_battery", "apple_system_disk", "apple_system_network"}
	if len(reg.Tools) != len(expected) {
		t.Fatalf("expected %d tools, got %d", len(expected), len(reg.Tools))
	}
	for _, name := range expected {
		tool := reg.FindTool(name)
		if tool == nil {
			t.Errorf("tool %q not registered", name)
		}
		if !json.Valid(tool.Parameters) {
			t.Errorf("tool %q has invalid JSON schema", name)
		}
	}
}

func TestDiskPathTraversal(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	_, err := testutil.CallTool(t, reg, "apple_system_disk", map[string]any{"path": "/tmp/../etc/passwd"})
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestIntegrationBattery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "apple_system_battery", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if _, ok := parsed["available"]; !ok {
		t.Error("expected 'available' field in battery output")
	}
	t.Logf("battery: %s", result)
}

func TestIntegrationDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "apple_system_disk", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}
	var parsed map[string]any
	json.Unmarshal([]byte(result), &parsed)
	if parsed["path"] != "/" {
		t.Errorf("expected path '/', got %v", parsed["path"])
	}
	t.Logf("disk: %s", result)
}

func TestIntegrationNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "apple_system_network", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}
	t.Logf("network: %s", result)
}
