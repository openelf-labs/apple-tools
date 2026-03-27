//go:build darwin

package notes

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/testutil"
)

func newRegistry() *testutil.MockRegistry {
	return testutil.NewRegistryWith(func(r core.Registry) {
		Register(r)
	})
}

func TestRegister(t *testing.T) {
	reg := newRegistry()

	if got := len(reg.Tools); got != 3 {
		t.Fatalf("expected 3 tools, got %d", got)
	}

	expected := []string{
		"apple_notes_list",
		"apple_notes_search",
		"apple_notes_create",
	}
	names := reg.ToolNames()
	sort.Strings(names)
	sort.Strings(expected)
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected tool %q at index %d, got %q", name, i, names[i])
		}
	}
}

func TestRegister_ToolsHaveSchemas(t *testing.T) {
	reg := newRegistry()
	for _, tool := range reg.Tools {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
		if len(tool.Parameters) == 0 {
			t.Errorf("tool %q has empty parameters schema", tool.Name)
		}
		if tool.Handler == nil {
			t.Errorf("tool %q has nil handler", tool.Name)
		}
	}
}

// --- Parameter validation tests ---

// TestSearch_EmptyQuery verifies that omitting or passing an empty query
// does not produce a validation error (list-all mode).
func TestSearch_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: would invoke JXA")
	}

	reg := newRegistry()
	cases := []map[string]any{
		{"query": ""},
		{},
		{"query": "   "},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "apple_notes_search", params)
		// Should not produce ErrInvalidInput; JXA/timeout/permission errors are OK.
		if errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("empty query should not produce ErrInvalidInput, params=%v, got: %v", params, err)
		}
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_notes_create", map[string]any{
		"title": "",
		"body":  "some content",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreate_MissingTitle(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_notes_create", map[string]any{
		"body": "some content",
	})
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreate_EmptyBody(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_notes_create", map[string]any{
		"title": "Test Note",
		"body":  "",
	})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreate_MissingBody(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_notes_create", map[string]any{
		"title": "Test Note",
	})
	if err == nil {
		t.Fatal("expected error for missing body")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Helpers ---

func TestClampLimit(t *testing.T) {
	tests := []struct {
		input, def, want int
	}{
		{0, 50, 50},
		{-5, 50, 50},
		{50, 50, 50},
		{200, 50, 200},
		{201, 50, 200},
		{1, 50, 1},
		{100, 50, 100},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input, tt.def)
		if got != tt.want {
			t.Errorf("clampLimit(%d, %d) = %d, want %d", tt.input, tt.def, got, tt.want)
		}
	}
}

// --- Integration test (read-only, safe to run) ---

func TestList_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "apple_notes_list", map[string]any{
		"limit": 1,
	})
	if err != nil {
		if errors.Is(err, core.ErrPermissionDenied) {
			t.Skip("skipping: Notes permission not granted")
		}
		if errors.Is(err, core.ErrAppNotRunning) {
			t.Skip("skipping: Notes app not available")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}

	var notes []map[string]any
	if err := json.Unmarshal([]byte(result), &notes); err != nil {
		t.Fatalf("expected JSON array, got unmarshal error: %v\nresult: %s", err, result)
	}

	if len(notes) > 0 {
		// Verify expected fields are present.
		for _, field := range []string{"name", "folder"} {
			if _, ok := notes[0][field]; !ok {
				t.Errorf("expected field %q in note entry", field)
			}
		}
	}
	t.Logf("list result (truncated): %.200s", result)
}
