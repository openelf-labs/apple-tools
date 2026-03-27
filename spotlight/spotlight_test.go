//go:build darwin

package spotlight_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/spotlight"
	"github.com/openelf-labs/apple-tools/testutil"
)

func newRegistry() *testutil.MockRegistry {
	return testutil.NewRegistryWith(spotlight.Register)
}

func TestRegister(t *testing.T) {
	reg := newRegistry()

	if len(reg.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(reg.Tools))
	}

	if reg.Tools[0].Name != "apple_spotlight_search" {
		t.Errorf("tool name = %q, want %q", reg.Tools[0].Name, "apple_spotlight_search")
	}

	if !json.Valid(reg.Tools[0].Parameters) {
		t.Error("tool has invalid JSON Schema parameters")
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	reg := newRegistry()

	_, err := testutil.CallTool(t, reg, "apple_spotlight_search", map[string]any{
		"query": "",
	})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearch_WhitespaceQuery(t *testing.T) {
	reg := newRegistry()

	_, err := testutil.CallTool(t, reg, "apple_spotlight_search", map[string]any{
		"query": "   ",
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearch_PathTraversal(t *testing.T) {
	reg := newRegistry()

	cases := []string{
		"/tmp/../etc/passwd",
		"../secret",
		"/home/user/../../root",
	}
	for _, dir := range cases {
		_, err := testutil.CallTool(t, reg, "apple_spotlight_search", map[string]any{
			"query":     "test",
			"directory": dir,
		})
		if err == nil {
			t.Errorf("expected error for directory %q", dir)
			continue
		}
		if !errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("directory %q: expected ErrInvalidInput, got: %v", dir, err)
		}
	}
}

func TestSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()

	// Search for Go files — there should be at least one on any dev machine.
	result, err := testutil.CallTool(t, reg, "apple_spotlight_search", map[string]any{
		"query": "kMDItemFSName == '*.go'",
		"limit": 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Found") {
		t.Errorf("expected result to contain 'Found', got: %s", result)
	}
}
