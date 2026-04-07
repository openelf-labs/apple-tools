//go:build darwin

package mcpserver

import (
	"encoding/json"
	"errors"
	"testing"

	appletools "github.com/openelf-labs/apple-tools"
	"github.com/openelf-labs/apple-tools/core"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolCount_DefaultConfig(t *testing.T) {
	cfg := appletools.DefaultConfig()
	count := ToolCount(cfg)

	// DefaultConfig has Messages=false, so 45 total minus 3 messages tools = 42.
	// If the actual count differs, update this test when tools are added/removed.
	if count == 0 {
		t.Fatal("expected non-zero tool count with default config")
	}
	t.Logf("Default config: %d tools registered", count)
}

func TestToolCount_AllEnabled(t *testing.T) {
	cfg := appletools.DefaultConfig()
	cfg.Messages = true
	countAll := ToolCount(cfg)

	cfg.Messages = false
	countDefault := ToolCount(cfg)

	if countAll <= countDefault {
		t.Errorf("all-enabled (%d) should have more tools than default (%d)", countAll, countDefault)
	}
	t.Logf("All enabled: %d, Default: %d, Messages adds: %d", countAll, countDefault, countAll-countDefault)
}

func TestToolCount_DisableCategory(t *testing.T) {
	full := appletools.DefaultConfig()
	full.Messages = true
	fullCount := ToolCount(full)

	// Disable calendar.
	partial := full
	partial.Calendar = false
	partialCount := ToolCount(partial)

	if partialCount >= fullCount {
		t.Errorf("disabling calendar should reduce tool count: full=%d, partial=%d", fullCount, partialCount)
	}
}

func TestToolCount_DisabledMaster(t *testing.T) {
	cfg := appletools.DefaultConfig()
	cfg.Enabled = false
	count := ToolCount(cfg)

	if count != 0 {
		t.Errorf("Enabled=false should register 0 tools, got %d", count)
	}
}

func TestNew_CreatesServer(t *testing.T) {
	cfg := appletools.DefaultConfig()
	server := New(cfg)

	if server == nil {
		t.Fatal("New() returned nil server")
	}
}

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		toolName string
		want     string
	}{
		{"apple_calendar_list", "calendar"},
		{"apple_system_battery", "system"},
		{"apple_messages_send", "messages"},
		{"apple_music_now_playing", "music"},
		{"unknown", "unknown"},
		{"apple_", ""},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := extractCategory(tt.toolName)
			if got != tt.want {
				t.Errorf("extractCategory(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestWrapError_PermissionDenied(t *testing.T) {
	err := core.NewPermissionError("Calendar", "Automation")
	result := wrapError(err, "apple_calendar_list")

	if !result.IsError {
		t.Error("expected IsError=true for permission denied")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var guidance permissionGuidance
	if err := json.Unmarshal([]byte(text), &guidance); err != nil {
		t.Fatalf("expected valid JSON guidance, got: %s", text)
	}

	if guidance.Error != "permission_denied" {
		t.Errorf("guidance.Error = %q, want %q", guidance.Error, "permission_denied")
	}
	if guidance.Category != "calendar" {
		t.Errorf("guidance.Category = %q, want %q", guidance.Category, "calendar")
	}
	if guidance.Permission != "automation" {
		t.Errorf("guidance.Permission = %q, want %q", guidance.Permission, "automation")
	}
	if guidance.SettingsURL == "" {
		t.Error("expected non-empty settings_url for automation permission")
	}
}

func TestWrapError_GenericError(t *testing.T) {
	err := errors.New("something went wrong")
	result := wrapError(err, "apple_system_battery")

	if !result.IsError {
		t.Error("expected IsError=true")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if text != "something went wrong" {
		t.Errorf("expected error message, got: %s", text)
	}
}

func TestBuildPermissionGuidance_Automation(t *testing.T) {
	err := core.NewPermissionError("Calendar", "Automation")
	g := buildPermissionGuidance("calendar", err)

	if g.Permission != "automation" {
		t.Errorf("permission = %q, want %q", g.Permission, "automation")
	}
	if g.SettingsURL == "" {
		t.Error("expected settings URL for automation")
	}
}

func TestBuildPermissionGuidance_FullDiskAccess(t *testing.T) {
	err := core.NewPermissionError("Messages", "Full Disk Access")
	g := buildPermissionGuidance("messages", err)

	if g.Permission != "full_disk_access" {
		t.Errorf("permission = %q, want %q", g.Permission, "full_disk_access")
	}
	if g.SettingsURL == "" {
		t.Error("expected settings URL for FDA")
	}
}

func TestBuildPermissionGuidance_UnknownCategory(t *testing.T) {
	err := errors.New("permission denied: unknown")
	g := buildPermissionGuidance("nonexistent", err)

	if g.Permission != "unknown" {
		t.Errorf("permission = %q, want %q", g.Permission, "unknown")
	}
}
