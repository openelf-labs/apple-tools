//go:build darwin

package finder

import (
	"encoding/json"
	"errors"
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

	if got := len(reg.Tools); got != 1 {
		t.Fatalf("expected 1 tool, got %d", got)
	}

	if reg.Tools[0].Name != "apple_finder_reveal" {
		t.Errorf("expected tool name %q, got %q", "apple_finder_reveal", reg.Tools[0].Name)
	}
}

func TestRegister_ToolHasSchema(t *testing.T) {
	reg := newRegistry()
	tool := reg.Tools[0]

	if tool.Description == "" {
		t.Error("tool has empty description")
	}
	if len(tool.Parameters) == 0 {
		t.Error("tool has empty parameters schema")
	}
	if tool.Handler == nil {
		t.Error("tool has nil handler")
	}
}

// --- Parameter validation tests ---

func TestReveal_EmptyPath(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{
		"path": "",
	})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReveal_MissingPath(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReveal_RelativePath(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{
		"path": "Documents/file.txt",
	})
	if err == nil {
		t.Fatal("expected error for relative path")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReveal_PathTraversal(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"basic traversal", "/Users/test/../../../etc/passwd"},
		{"double dot mid-path", "/Users/../root"},
		{"trailing double dot", "/Users/test/.."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newRegistry()
			_, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{
				"path": tt.path,
			})
			if err == nil {
				t.Fatal("expected error for path traversal")
			}
			if !errors.Is(err, core.ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got: %v", err)
			}
		})
	}
}

func TestReveal_WhitespacePath(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{
		"path": "   ",
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only path")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Integration test ---

func TestReveal_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "apple_finder_reveal", map[string]any{
		"path": "/Applications",
	})
	if err != nil {
		if errors.Is(err, core.ErrPermissionDenied) {
			t.Skip("skipping: Finder permission not granted")
		}
		if errors.Is(err, core.ErrAppNotRunning) {
			t.Skip("skipping: Finder not available")
		}
		if errors.Is(err, core.ErrTimeout) {
			t.Skip("skipping: timeout")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("expected JSON object: %v", err)
	}
	if m["ok"] != true {
		t.Errorf("expected ok=true, got %v", m["ok"])
	}
	if m["path"] == nil || m["path"] == "" {
		t.Errorf("expected non-empty path field")
	}
}

// --- Helpers ---

func TestContainsTraversal(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/Users/test/file.txt", false},
		{"/Users/test/../etc/passwd", true},
		{"/Users/test/..", true},
		{"/../root", true},
		{"/Users/test/...hidden", false},
		{"/Users/test/file..txt", false},
		{"/Users/test/..nameprefix", false},
	}
	for _, tt := range tests {
		got := core.ContainsTraversal(tt.path)
		if got != tt.want {
			t.Errorf("core.ContainsTraversal(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
