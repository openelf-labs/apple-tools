//go:build darwin

package music

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

	if got := len(reg.Tools); got != 8 {
		t.Fatalf("expected 8 tools, got %d", got)
	}

	expected := []string{
		"apple_music_now_playing",
		"apple_music_play",
		"apple_music_pause",
		"apple_music_next",
		"apple_music_previous",
		"apple_music_search_play",
		"apple_music_volume",
		"apple_music_playlists",
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

func TestSearchPlay_EmptyQuery(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_search_play", map[string]any{
		"query": "",
	})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearchPlay_MissingQuery(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_search_play", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing query")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSearchPlay_InvalidType(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_search_play", map[string]any{
		"query": "test",
		"type":  "album",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestVolume_MissingLevel(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_volume", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing level")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestVolume_BelowZero(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_volume", map[string]any{
		"level": -1,
	})
	if err == nil {
		t.Fatal("expected error for negative level")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestVolume_Above100(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_music_volume", map[string]any{
		"level": 101,
	})
	if err == nil {
		t.Fatal("expected error for level > 100")
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
		{0, 25, 25},
		{-5, 10, 10},
		{50, 25, 50},
		{100, 25, 100},
		{101, 25, 100},
		{1, 25, 1},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input, tt.def)
		if got != tt.want {
			t.Errorf("clampLimit(%d, %d) = %d, want %d", tt.input, tt.def, got, tt.want)
		}
	}
}

// --- Integration test (read-only, safe to run) ---

func TestNowPlaying_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "apple_music_now_playing", map[string]any{})
	if err != nil {
		if errors.Is(err, core.ErrPermissionDenied) {
			t.Skip("skipping: Music permission not granted")
		}
		if errors.Is(err, core.ErrAppNotRunning) {
			t.Skip("skipping: Music app not available")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Result must be valid JSON.
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}
	t.Logf("now playing result: %s", result)
}
