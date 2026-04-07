//go:build darwin

package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
		{"calendar_list", "calendar"},
		{"system_battery", "system"},
		{"messages_send", "messages"},
		{"music_now_playing", "music"},
		{"unknown", "unknown"},
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
	result := wrapError(err, "calendar_list")

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
	result := wrapError(err, "system_battery")

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

func TestPermissionWaitTimeout_Value(t *testing.T) {
	// Ensure the timeout is reasonable (10-30 seconds range).
	if permissionWaitTimeout < 10*time.Second || permissionWaitTimeout > 30*time.Second {
		t.Errorf("permissionWaitTimeout = %v, want between 10s and 30s", permissionWaitTimeout)
	}
}

func TestShouldAutoWaitForPermission_DefaultDisabled(t *testing.T) {
	t.Setenv("APPLE_AUTO_WAIT_FOR_PERMISSION", "")
	if shouldAutoWaitForPermission() {
		t.Error("expected auto wait to be disabled by default")
	}
}

func TestShouldAutoWaitForPermission_Enabled(t *testing.T) {
	t.Setenv("APPLE_AUTO_WAIT_FOR_PERMISSION", "true")
	if !shouldAutoWaitForPermission() {
		t.Error("expected auto wait to be enabled")
	}
}

func TestEnvBool(t *testing.T) {
	key := "TEST_APPLE_TOOLS_BOOL"
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetenv: %v", err)
	}
	if !envBoolDefault(key, true) {
		t.Error("expected default true when unset")
	}

	t.Setenv(key, "off")
	if envBoolDefault(key, true) {
		t.Error("expected off to parse as false")
	}

	t.Setenv(key, "1")
	if !envBoolDefault(key, false) {
		t.Error("expected 1 to parse as true")
	}
}

func TestPermissionState_NeedsProbe(t *testing.T) {
	clearNeedsProbe("contacts")

	if categoryNeedsProbe("contacts") {
		t.Error("expected contacts not flagged initially")
	}

	markNeedsProbe("contacts")
	if !categoryNeedsProbe("contacts") {
		t.Error("expected contacts flagged after marking")
	}

	// Other categories unaffected.
	if categoryNeedsProbe("calendar") {
		t.Error("expected calendar not flagged")
	}

	clearNeedsProbe("contacts")
	if categoryNeedsProbe("contacts") {
		t.Error("expected contacts cleared")
	}
}

func TestRequiresPermission(t *testing.T) {
	tests := []struct {
		category string
		want     bool
	}{
		{"contacts", true},
		{"calendar", true},
		{"mail", true},
		{"messages", true},   // full_disk_access
		{"shortcuts", false}, // none
		{"system", false},
		{"clipboard", false},
		{"nonexistent", false},
	}
	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			if got := requiresPermission(tt.category); got != tt.want {
				t.Errorf("requiresPermission(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

// TestRegisterTool_PermissionDenied_MarksNeedsProbe verifies that a tool
// returning ErrPermissionDenied sets the needs-probe flag and returns the
// structured permission_denied JSON.
func TestRegisterTool_PermissionDenied_MarksNeedsProbe(t *testing.T) {
	clearNeedsProbe("contacts")
	defer clearNeedsProbe("contacts")

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	callCount := 0
	registerTool(server, core.Tool{
		Name:        "contacts_search",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			callCount++
			return "", core.NewPermissionError("Contacts", "contacts")
		},
	})

	result := callMCPTool(t, server, "contacts_search", `{}`)
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if !categoryNeedsProbe("contacts") {
		t.Error("expected needs-probe flag set after permission denied")
	}
	// Verify structured JSON
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, `"permission_denied"`) {
		t.Errorf("expected permission_denied in response, got: %s", text)
	}
}

// TestRegisterTool_Timeout_MarksNeedsProbe verifies that a timeout in a
// permission-requiring category is treated as a permission error.
func TestRegisterTool_Timeout_MarksNeedsProbe(t *testing.T) {
	clearNeedsProbe("contacts")
	defer clearNeedsProbe("contacts")

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	registerTool(server, core.Tool{
		Name:        "contacts_search",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			return "", fmt.Errorf("%w: osascript killed after timeout", core.ErrTimeout)
		},
	})

	result := callMCPTool(t, server, "contacts_search", `{}`)
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	if !categoryNeedsProbe("contacts") {
		t.Error("expected needs-probe flag set after timeout")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, `"permission_denied"`) {
		t.Errorf("expected permission_denied in response for timeout, got: %s", text)
	}
}

// TestRegisterTool_Timeout_NonPermissionCategory verifies that a timeout
// in a non-permission category is NOT treated as a permission error.
func TestRegisterTool_Timeout_NonPermissionCategory(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	registerTool(server, core.Tool{
		Name:        "system_battery",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			return "", fmt.Errorf("%w: osascript killed after timeout", core.ErrTimeout)
		},
	})

	result := callMCPTool(t, server, "system_battery", `{}`)
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	// Should be a plain error, not permission_denied
	if strings.Contains(text, `"permission_denied"`) {
		t.Errorf("non-permission category should not produce permission_denied, got: %s", text)
	}
}

// TestRegisterTool_ProbeOnRetry_Granted verifies that after a denial, if the
// user grants permission, the next call succeeds via probe-on-retry.
func TestRegisterTool_ProbeOnRetry_Granted(t *testing.T) {
	clearNeedsProbe("contacts")
	defer clearNeedsProbe("contacts")

	// Override probeFn to simulate "granted" on probe.
	origProbe := probeFn
	defer func() { probeFn = origProbe }()
	probeFn = func(_ context.Context, _ string) core.PermissionStatus {
		return core.PermissionStatus{Status: "granted", Permission: "automation"}
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	callCount := 0
	registerTool(server, core.Tool{
		Name:        "contacts_search",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			callCount++
			return `{"contacts":[]}`, nil
		},
	})

	// Simulate: previous call set the probe flag (e.g., permission was denied).
	markNeedsProbe("contacts")

	// Next call should probe → see "granted" → clear flag → run handler.
	result := callMCPTool(t, server, "contacts_search", `{}`)
	if result.IsError {
		t.Fatalf("expected success after probe-granted, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if categoryNeedsProbe("contacts") {
		t.Error("expected probe flag cleared after successful probe")
	}
}

// TestRegisterTool_ProbeOnRetry_StillDenied verifies that after a denial,
// if the user hasn't granted permission, the next call returns the error
// immediately without calling the handler (no osascript = no TCC dialog).
func TestRegisterTool_ProbeOnRetry_StillDenied(t *testing.T) {
	clearNeedsProbe("contacts")
	defer clearNeedsProbe("contacts")

	origProbe := probeFn
	defer func() { probeFn = origProbe }()
	probeFn = func(_ context.Context, _ string) core.PermissionStatus {
		return core.PermissionStatus{Status: "denied", Permission: "automation"}
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	callCount := 0
	registerTool(server, core.Tool{
		Name:        "contacts_search",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			callCount++
			return "", nil
		},
	})

	markNeedsProbe("contacts")

	result := callMCPTool(t, server, "contacts_search", `{}`)
	if !result.IsError {
		t.Fatal("expected error when probe returns denied")
	}
	if callCount != 0 {
		t.Errorf("handler should NOT be called when probe returns denied, got %d calls", callCount)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, `"permission_denied"`) {
		t.Errorf("expected permission_denied, got: %s", text)
	}
}

// TestRegisterTool_SuccessClearsProbeFlag verifies that a successful tool
// call clears any stale needs-probe flag.
func TestRegisterTool_SuccessClearsProbeFlag(t *testing.T) {
	clearNeedsProbe("system")
	defer clearNeedsProbe("system")

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	registerTool(server, core.Tool{
		Name:        "system_battery",
		Description: "test",
		Parameters:  []byte(`{"type":"object"}`),
		Handler: func(_ context.Context, _ json.RawMessage) (string, error) {
			return `{"level":85}`, nil
		},
	})

	// Even if somehow flagged, success should clear.
	markNeedsProbe("system")
	result := callMCPTool(t, server, "system_battery", `{}`)
	if result.IsError {
		t.Fatal("expected success")
	}
	if categoryNeedsProbe("system") {
		t.Error("expected probe flag cleared after success")
	}
}

// callMCPTool is a test helper that creates an in-memory client/server pair,
// initializes the session, and calls the named tool.
func callMCPTool(t *testing.T, server *mcp.Server, toolName, argsJSON string) *mcp.CallToolResult {
	t.Helper()
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()

	go func() { _ = server.Run(ctx, st) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: json.RawMessage(argsJSON),
	})
	if err != nil {
		t.Fatalf("CallTool(%q) transport error: %v", toolName, err)
	}
	return result
}
