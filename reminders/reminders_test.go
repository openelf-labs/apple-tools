//go:build darwin

package reminders_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/reminders"
	"github.com/openelf-labs/apple-tools/testutil"
)

func newRegistry() *testutil.MockRegistry {
	return testutil.NewRegistryWith(func(r core.Registry) {
		reminders.Register(r)
	})
}

// --- Registration ---

func TestRegister_ToolCount(t *testing.T) {
	reg := newRegistry()
	if got := len(reg.Tools); got != 5 {
		t.Fatalf("expected 5 tools, got %d", got)
	}
}

func TestRegister_ToolNames(t *testing.T) {
	reg := newRegistry()
	expected := []string{
		"apple_reminders_list",
		"apple_reminders_search",
		"apple_reminders_create",
		"apple_reminders_complete",
		"apple_reminders_lists",
	}

	names := reg.ToolNames()
	for _, want := range expected {
		found := false
		for _, got := range names {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tool %q not found in %v", want, names)
		}
	}
}

func TestRegister_ToolsHaveValidJSONSchema(t *testing.T) {
	reg := newRegistry()
	for _, tool := range reg.Tools {
		if tool.Parameters == nil {
			t.Errorf("tool %q has nil Parameters", tool.Name)
			continue
		}
		var schema map[string]any
		if err := json.Unmarshal(tool.Parameters, &schema); err != nil {
			t.Errorf("tool %q has invalid JSON Schema: %v", tool.Name, err)
		}
		if schema["type"] != "object" {
			t.Errorf("tool %q schema type is %v, want object", tool.Name, schema["type"])
		}
	}
}

func TestRegister_ToolsHaveDescriptions(t *testing.T) {
	reg := newRegistry()
	for _, tool := range reg.Tools {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
}

// --- Validation ---

func TestList_InvalidStatus(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_reminders_list", map[string]any{
		"status": "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestList_ValidStatuses(t *testing.T) {
	// Pure validation test: ensure valid statuses don't trigger ErrInvalidInput.
	// We don't call the handler (which would invoke JXA); instead we verify that
	// the validation logic accepts all documented status values.
	for _, status := range []string{"incomplete", "completed", "all"} {
		if !reminders.IsValidStatus(status) {
			t.Errorf("status %q should be accepted as valid", status)
		}
	}
}

func TestList_InvalidStatusValues(t *testing.T) {
	for _, status := range []string{"bogus", "done", "pending", ""} {
		if reminders.IsValidStatus(status) {
			t.Errorf("status %q should be rejected as invalid", status)
		}
	}
}

// TestSearch_EmptyQuery verifies that omitting or passing an empty query
// does not produce a validation error (list-all mode).
func TestSearch_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: would invoke JXA")
	}

	reg := newRegistry()

	cases := []map[string]any{
		{},
		{"query": ""},
		{"query": "   "},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "apple_reminders_search", params)
		// Should not produce ErrInvalidInput; JXA/timeout/permission errors are OK.
		if errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("empty query should not produce ErrInvalidInput for params %v, got: %v", params, err)
		}
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	reg := newRegistry()

	cases := []map[string]any{
		{},
		{"title": ""},
		{"title": "   "},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "apple_reminders_create", params)
		if err == nil {
			t.Fatalf("expected error for params %v", params)
		}
		if !errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for params %v, got: %v", params, err)
		}
	}
}

func TestCreate_InvalidPriority(t *testing.T) {
	reg := newRegistry()

	for _, prio := range []int{-1, 10, 100} {
		params := map[string]any{"title": "test", "priority": prio}
		_, err := testutil.CallTool(t, reg, "apple_reminders_create", params)
		if err == nil {
			t.Fatalf("expected error for priority %d", prio)
		}
		if !errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for priority %d, got: %v", prio, err)
		}
	}
}

func TestCreate_ValidPriorityZero(t *testing.T) {
	// priority=0 means "no priority" and is valid (zero value).
	// This test only verifies it doesn't produce ErrInvalidInput;
	// the handler will proceed to JXA which requires integration setup.
	if testing.Short() {
		t.Skip("skipping: would invoke JXA")
	}

	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_reminders_create", map[string]any{
		"title": "test", "priority": 0,
	})
	if errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("priority 0 should not produce ErrInvalidInput, got: %v", err)
	}
}

func TestComplete_EmptyID(t *testing.T) {
	reg := newRegistry()

	cases := []map[string]any{
		{},
		{"id": ""},
		{"id": "   "},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "apple_reminders_complete", params)
		if err == nil {
			t.Fatalf("expected error for params %v", params)
		}
		if !errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for params %v, got: %v", params, err)
		}
	}
}

func TestList_InvalidJSON(t *testing.T) {
	reg := newRegistry()
	tool := reg.FindTool("apple_reminders_list")
	if tool == nil {
		t.Fatal("tool not found")
	}
	_, err := tool.Handler(nil, json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Integration (read-only, safe) ---

func TestIntegration_GetLists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "apple_reminders_lists", map[string]any{})
	if err != nil {
		// Permission errors are expected on CI or sandboxed environments.
		if errors.Is(err, core.ErrPermissionDenied) ||
			errors.Is(err, core.ErrAppNotRunning) ||
			errors.Is(err, core.ErrTimeout) {
			t.Skipf("skipping: %v", err)
		}
		// Also skip generic osascript errors that indicate sandboxing.
		if strings.Contains(err.Error(), "osascript") {
			t.Skipf("skipping: osascript unavailable: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// The result should be a JSON array.
	var lists []map[string]any
	if err := json.Unmarshal([]byte(result), &lists); err != nil {
		t.Fatalf("failed to parse result as JSON array: %v\nraw: %s", err, result)
	}

	// Reminders always has at least one list (the default list).
	if len(lists) == 0 {
		t.Error("expected at least one reminder list")
	}

	// Each entry should have id, name fields.
	for i, l := range lists {
		if _, ok := l["id"]; !ok {
			t.Errorf("list[%d] missing 'id' field", i)
		}
		if _, ok := l["name"]; !ok {
			t.Errorf("list[%d] missing 'name' field", i)
		}
	}
}
