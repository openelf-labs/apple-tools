//go:build darwin

package contacts_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/contacts"
	"github.com/openelf-labs/apple-tools/testutil"
)

func newRegistry() *testutil.MockRegistry {
	return testutil.NewRegistryWith(contacts.Register)
}

func TestRegister(t *testing.T) {
	reg := newRegistry()

	if len(reg.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(reg.Tools))
	}

	expected := []string{
		"contacts_search",
		"contacts_details",
		"contacts_find_by_phone",
	}
	names := reg.ToolNames()
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("tool[%d] = %q, want %q", i, names[i], want)
		}
	}

	// Verify each tool has valid JSON Schema parameters.
	for _, tool := range reg.Tools {
		if !json.Valid(tool.Parameters) {
			t.Errorf("tool %q has invalid JSON Schema", tool.Name)
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
		{"query": ""},
		{"query": "   "},
		{},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "contacts_search", params)
		// Should not produce ErrInvalidInput; JXA/timeout/permission errors are OK.
		if errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("empty query should not produce ErrInvalidInput, params=%v, got: %v", params, err)
		}
	}
}

func TestDetails_EmptyName(t *testing.T) {
	reg := newRegistry()

	_, err := testutil.CallTool(t, reg, "contacts_details", map[string]any{
		"name": "",
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestFindByPhone_EmptyNumber(t *testing.T) {
	reg := newRegistry()

	_, err := testutil.CallTool(t, reg, "contacts_find_by_phone", map[string]any{
		"phoneNumber": "",
	})
	if err == nil {
		t.Fatal("expected error for empty phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestFindByPhone_InvalidFormat(t *testing.T) {
	reg := newRegistry()

	cases := []string{
		"abc",
		"12",
		"not-a-phone-number-at-all-way-too-long",
	}
	for _, phone := range cases {
		_, err := testutil.CallTool(t, reg, "contacts_find_by_phone", map[string]any{
			"phoneNumber": phone,
		})
		if err == nil {
			t.Errorf("expected error for phone %q", phone)
			continue
		}
		if !errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("phone %q: expected ErrInvalidInput, got: %v", phone, err)
		}
	}
}

func TestSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()

	// Search for an unlikely name to verify JXA executes without error.
	result, err := testutil.CallTool(t, reg, "contacts_search", map[string]any{
		"query": "zzzzznotarealcontact99999",
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

	// Should return an empty JSON array.
	var contacts []any
	if err := json.Unmarshal([]byte(result), &contacts); err != nil {
		t.Fatalf("failed to parse result as JSON array: %v\nresult: %s", err, result)
	}
	if len(contacts) != 0 {
		t.Errorf("expected empty result for unlikely query, got %d contacts", len(contacts))
	}
}
