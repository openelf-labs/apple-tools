//go:build darwin

package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	appletools "github.com/openelf-labs/apple-tools"
	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/testutil"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// permissionWaitTimeout is how long to wait for the user to grant a macOS
// permission in System Settings before returning a permission_denied error.
const permissionWaitTimeout = 15 * time.Second

// permissionState tracks per-category permission state for the current MCP
// server session.
//
// When a tool call fails with ErrPermissionDenied or times out (likely TCC
// dialog blocking), the category is marked as "needs probe". On the next call
// to the same category, we run a lightweight ProbePermission instead of the
// full tool. The probe reads the cached TCC decision (no new dialog) and
// either clears the flag (user granted) or returns a structured error (still
// denied).
//
// This avoids two pitfalls:
//   - Permanent lockout: if the user grants permission in System Settings,
//     the next probe detects it and clears the flag.
//   - False positives: a one-off timeout doesn't permanently block a category;
//     the probe on retry confirms the actual TCC state.
var permissionState = struct {
	sync.RWMutex
	needsProbe map[string]bool // category → true if last attempt failed
}{needsProbe: make(map[string]bool)}

func markNeedsProbe(category string) {
	permissionState.Lock()
	permissionState.needsProbe[category] = true
	permissionState.Unlock()
}

func categoryNeedsProbe(category string) bool {
	permissionState.RLock()
	defer permissionState.RUnlock()
	return permissionState.needsProbe[category]
}

func clearNeedsProbe(category string) {
	permissionState.Lock()
	delete(permissionState.needsProbe, category)
	permissionState.Unlock()
}

// probeFn is the function used to check permission status.
// Package-level var for testability.
var probeFn = func(ctx context.Context, category string) core.PermissionStatus {
	return core.ProbePermission(ctx, category)
}

// New creates an MCP server with all enabled Apple tools registered.
// The returned server is ready to run via Run().
func New(cfg appletools.Config) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "apple-tools",
			Version: Version,
		},
		nil,
	)

	// Collect tools via MockRegistry, then bridge to MCP.
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, cfg)

	for _, t := range reg.Tools {
		registerTool(server, t)
	}

	return server
}

// ToolCount returns the number of tools that would be registered with the given config.
// Useful for testing and diagnostics without creating a full MCP server.
func ToolCount(cfg appletools.Config) int {
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, cfg)
	return len(reg.Tools)
}

// Run starts the MCP server using stdio transport (stdin/stdout).
// It blocks until the context is cancelled or the transport is closed.
func Run(ctx context.Context, server *mcp.Server) error {
	transport := &mcp.StdioTransport{}
	return server.Run(ctx, transport)
}

// registerTool bridges a core.Tool to an MCP tool registration.
func registerTool(server *mcp.Server, t core.Tool) {
	mcpTool := &mcp.Tool{
		Name:        t.Name,
		Description: t.Description,
	}

	// Set InputSchema from the tool's JSON Schema parameters.
	// The MCP SDK accepts any JSON-marshalable value for InputSchema.
	if len(t.Parameters) > 0 {
		var schema map[string]any
		if err := json.Unmarshal(t.Parameters, &schema); err == nil {
			mcpTool.InputSchema = schema
		}
	}

	// Capture handler for closure.
	handler := t.Handler

	mcp.AddTool(server, mcpTool, func(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
		// If args is empty or null, default to empty object.
		if len(args) == 0 || string(args) == "null" {
			args = json.RawMessage("{}")
		}

		category := extractCategory(t.Name)

		// ── Probe-on-retry: if a previous call failed for this category,
		// check the current TCC state with a lightweight probe before
		// spawning a full osascript (which could trigger another dialog).
		// ProbePermission reads the cached TCC decision — it does NOT show
		// a new dialog once the user has already responded.
		if requiresPermission(category) && categoryNeedsProbe(category) {
			status := probeFn(ctx, category)
			if status.Status == "granted" {
				// User has since granted permission — clear flag and proceed.
				clearNeedsProbe(category)
			} else {
				// Still denied / not yet granted — return structured error
				// immediately without running the tool.
				permErr := core.NewPermissionError(strings.Title(category), category)
				return wrapError(permErr, t.Name), nil, nil
			}
		}

		result, err := handler(ctx, args)
		if err != nil {
			isPermErr := errors.Is(err, core.ErrPermissionDenied)
			isTimeoutErr := errors.Is(err, core.ErrTimeout)

			// For permission-requiring categories, both explicit denial and
			// timeout (TCC dialog blocking osascript for 30s) indicate a
			// permission problem. Mark the category so the next call probes
			// first instead of spawning another blocking osascript.
			if requiresPermission(category) && (isPermErr || isTimeoutErr) {
				markNeedsProbe(category)

				// Optionally wait for the user to grant in System Settings.
				if isPermErr && shouldAutoWaitForPermission() {
					if core.WaitForPermission(ctx, category, permissionWaitTimeout) {
						clearNeedsProbe(category)
						if retryResult, retryErr := handler(ctx, args); retryErr == nil {
							return &mcp.CallToolResult{
								Content: []mcp.Content{&mcp.TextContent{Text: retryResult}},
							}, nil, nil
						} else {
							err = retryErr
						}
					}
				}

				// Convert to structured permission error for the agent.
				permErr := core.NewPermissionError(strings.Title(category), category)
				return wrapError(permErr, t.Name), nil, nil
			}

			return wrapError(err, t.Name), nil, nil
		}

		// Success — if there was a stale probe flag, clear it.
		clearNeedsProbe(category)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})
}

// wrapError converts a tool handler error into an MCP CallToolResult with
// IsError=true and structured guidance for the AI agent.
func wrapError(err error, toolName string) *mcp.CallToolResult {
	// Extract category from tool name (<category>_<action>).
	category := extractCategory(toolName)

	// Permission denied: provide actionable guidance with settings URL.
	if errors.Is(err, core.ErrPermissionDenied) {
		guidance := buildPermissionGuidance(category, err)
		body, _ := json.Marshal(guidance)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
			IsError: true,
		}
	}

	// All other errors: return error message.
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
		IsError: true,
	}
}

// permissionGuidance is the structured error payload for permission errors.
type permissionGuidance struct {
	Error       string `json:"error"`
	Category    string `json:"category"`
	Permission  string `json:"permission_type"`
	SettingsURL string `json:"settings_url,omitempty"`
	Guide       string `json:"guide"`
	Action      string `json:"action"`
}

// buildPermissionGuidance creates a structured guidance payload for permission errors.
func buildPermissionGuidance(category string, err error) permissionGuidance {
	g := permissionGuidance{
		Error:    "permission_denied",
		Category: category,
		Action:   "Open System Settings and enable the required permission, then retry.",
	}

	// Try to get detailed info from CategoryPermissions.
	if cp, ok := core.CategoryPermissions[category]; ok {
		g.Permission = cp.Type
		g.SettingsURL = cp.SettingsURL
		switch cp.Type {
		case "automation":
			g.Guide = fmt.Sprintf("Grant access in System Settings > Privacy & Security > Automation > Allow apple-tools to control the target application")
		case "full_disk_access":
			g.Guide = "Grant access in System Settings > Privacy & Security > Full Disk Access > Enable apple-tools"
		default:
			g.Guide = err.Error()
		}
	} else {
		g.Permission = "unknown"
		g.Guide = err.Error()
	}

	return g
}

// extractCategory derives the category from a tool name like "calendar_list".
func extractCategory(toolName string) string {
	name := canonicalToolName(toolName)
	parts := strings.SplitN(name, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// requiresPermission returns true if the given tool category needs a macOS
// permission (Automation or Full Disk Access) that may trigger a TCC dialog.
func requiresPermission(category string) bool {
	cp, ok := core.CategoryPermissions[category]
	return ok && cp.Type != "none"
}

func canonicalToolName(toolName string) string {
	return strings.TrimPrefix(toolName, "apple_")
}

func shouldAutoWaitForPermission() bool {
	return envBoolDefault("APPLE_AUTO_WAIT_FOR_PERMISSION", false)
}

func envBoolDefault(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	switch strings.ToLower(v) {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}
