//go:build darwin

package calendar

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

	if got := len(reg.Tools); got != 4 {
		t.Fatalf("expected 4 tools, got %d", got)
	}

	expected := []string{
		"apple_calendar_list",
		"apple_calendar_search",
		"apple_calendar_create",
		"apple_calendar_open",
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

func TestListEvents_InvalidFromDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_list", map[string]any{
		"from": "not-a-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid from date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestListEvents_InvalidToDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_list", map[string]any{
		"to": "2025-13-45",
	})
	if err == nil {
		t.Fatal("expected error for invalid to date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearchEvents_EmptyQuery(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_search", map[string]any{
		"query": "",
	})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearchEvents_MissingQuery(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_search", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearchEvents_InvalidFromDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_search", map[string]any{
		"query": "test",
		"from":  "bad-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid from date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateEvent_EmptyTitle(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_create", map[string]any{
		"title": "",
		"start": "2025-06-15T14:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateEvent_MissingTitle(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_create", map[string]any{
		"start": "2025-06-15T14:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateEvent_MissingStart(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_create", map[string]any{
		"title": "Test Event",
	})
	if err == nil {
		t.Fatal("expected error for missing start")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateEvent_InvalidStartDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_create", map[string]any{
		"title": "Test Event",
		"start": "next tuesday",
	})
	if err == nil {
		t.Fatal("expected error for invalid start date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCreateEvent_InvalidEndDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_create", map[string]any{
		"title": "Test Event",
		"start": "2025-06-15T14:00:00Z",
		"end":   "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid end date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestOpenEvent_EmptyEventID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_open", map[string]any{
		"eventId": "",
	})
	if err == nil {
		t.Fatal("expected error for empty eventId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestOpenEvent_MissingEventID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_calendar_open", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing eventId")
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
		{0, 20, 20},
		{-5, 10, 10},
		{50, 20, 50},
		{100, 20, 100},
		{101, 20, 100},
		{1, 20, 1},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input, tt.def)
		if got != tt.want {
			t.Errorf("clampLimit(%d, %d) = %d, want %d", tt.input, tt.def, got, tt.want)
		}
	}
}

// --- Integration test (read-only, safe to run) ---

func TestListEvents_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "apple_calendar_list", map[string]any{
		"limit": 5,
	})
	if err != nil {
		// Permission, app-not-running, and timeout errors are expected
		// in CI, sandboxed environments, or headless setups.
		if errors.Is(err, core.ErrPermissionDenied) ||
			errors.Is(err, core.ErrAppNotRunning) ||
			errors.Is(err, core.ErrTimeout) {
			t.Skipf("skipping: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Result must be valid JSON (array of events or empty array).
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}
	var events []map[string]any
	if err := json.Unmarshal([]byte(result), &events); err != nil {
		t.Fatalf("expected JSON array: %v\nraw: %s", err, result)
	}
	t.Logf("list result: %d event(s)", len(events))
}
