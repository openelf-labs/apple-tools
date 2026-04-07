//go:build darwin

package safari

import (
	"encoding/json"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/testutil"
)

func TestRegister(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	expected := []string{
		"safari_tabs",
		"safari_get_page",
		"safari_bookmarks",
		"safari_reading_list",
	}

	if len(reg.Tools) != len(expected) {
		t.Fatalf("expected %d tools, got %d", len(expected), len(reg.Tools))
	}

	for _, name := range expected {
		tool := reg.FindTool(name)
		if tool == nil {
			t.Errorf("tool %q not registered", name)
			continue
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", name)
		}
		if !json.Valid(tool.Parameters) {
			t.Errorf("tool %q has invalid JSON schema", name)
		}
	}
}

func TestIntegrationListTabs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "safari_tabs", map[string]any{})
	if err != nil {
		t.Logf("safari tabs returned error (may not be running): %v", err)
		return
	}
	// Output must be valid JSON
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got: %s", result)
	}
	t.Logf("safari tabs result: %s", result)
}

func TestIntegrationReadingList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	reg := &testutil.MockRegistry{}
	Register(reg)

	result, err := testutil.CallTool(t, reg, "safari_reading_list", map[string]any{"limit": 3})
	if err != nil {
		t.Logf("reading list returned error: %v", err)
		return
	}
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got: %s", result)
	}
	t.Logf("reading list result: %s", result)
}

var _ core.Registry = (*testutil.MockRegistry)(nil)
