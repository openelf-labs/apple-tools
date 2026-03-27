//go:build darwin

package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PermissionStatus represents the authorization state of an Apple tool category.
type PermissionStatus struct {
	Status     string `json:"status"`     // granted, denied, not_requested, no_permission
	Permission string `json:"permission"` // automation, full_disk_access, none
}

// CategoryPermissions defines the macOS permission type required by each tool category.
// This is the SSOT for permission requirements — both the backend API and frontend
// derive their permission display from this map.
var CategoryPermissions = map[string]string{
	"calendar":     "automation",
	"reminders":    "automation",
	"contacts":     "automation",
	"notes":        "automation",
	"mail":         "automation",
	"messages":     "full_disk_access",
	"music":        "automation",
	"safari":       "automation",
	"finder":       "automation",
	"shortcuts":    "none",
	"system":       "none",
	"clipboard":    "none",
	"notification": "none",
	"spotlight":    "none",
}

// probeScripts maps categories to minimal JXA scripts that test Automation permission.
// Each script does the absolute minimum to trigger (or confirm) TCC authorization.
var probeScripts = map[string]string{
	"calendar":  `Application("Calendar").calendars.length; "ok"`,
	"reminders": `Application("Reminders").lists.length; "ok"`,
	"contacts":  `Application("Contacts").people.length; "ok"`,
	"notes":     `Application("Notes").notes.length; "ok"`,
	"mail":      `Application("Mail").accounts.length; "ok"`,
	"music":     `Application("Music").name(); "ok"`,
	"safari":    `Application("Safari").name(); "ok"`,
	"finder":    `Application("Finder").name(); "ok"`,
}

// ProbePermission checks the macOS permission status for a given tool category.
// For categories that don't require permissions, it returns "no_permission" immediately.
// For Automation categories, it runs a minimal JXA probe.
// For Full Disk Access (messages), it checks file readability.
//
// WARNING: Probing an Automation category that hasn't been authorized yet WILL
// trigger a macOS permission prompt. Only call this when the user explicitly
// requests a permission test.
func ProbePermission(ctx context.Context, category string) PermissionStatus {
	permType, ok := CategoryPermissions[category]
	if !ok {
		return PermissionStatus{Status: "unknown", Permission: "unknown"}
	}

	if permType == "none" {
		return PermissionStatus{Status: "no_permission", Permission: "none"}
	}

	if permType == "full_disk_access" {
		return probeFullDiskAccess()
	}

	// Automation: run minimal JXA probe
	script, ok := probeScripts[category]
	if !ok {
		return PermissionStatus{Status: "not_requested", Permission: permType}
	}

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := RunJXA(probeCtx, []byte(script), nil)
	if err == nil {
		return PermissionStatus{Status: "granted", Permission: permType}
	}

	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "permission") || strings.Contains(errStr, "-1743") || strings.Contains(errStr, "not allowed") {
		return PermissionStatus{Status: "denied", Permission: permType}
	}

	// App not running or other transient error — treat as not_requested
	// (we can't distinguish "never asked" from "app crashed" via osascript)
	return PermissionStatus{Status: "not_requested", Permission: permType}
}

// ProbeAll checks all categories and returns a map of statuses.
// Only probes categories listed in the enabled set. Skips disabled categories.
func ProbeAll(ctx context.Context, enabled map[string]bool) map[string]PermissionStatus {
	result := make(map[string]PermissionStatus, len(CategoryPermissions))
	for category := range CategoryPermissions {
		if en, ok := enabled[category]; ok && !en {
			// Category is explicitly disabled — skip
			result[category] = PermissionStatus{
				Status:     "disabled",
				Permission: CategoryPermissions[category],
			}
			continue
		}
		result[category] = ProbePermission(ctx, category)
	}
	return result
}

func probeFullDiskAccess() PermissionStatus {
	home, err := os.UserHomeDir()
	if err != nil {
		return PermissionStatus{Status: "not_requested", Permission: "full_disk_access"}
	}

	chatDB := filepath.Join(home, "Library", "Messages", "chat.db")
	f, err := os.Open(chatDB)
	if err != nil {
		if os.IsPermission(err) {
			return PermissionStatus{Status: "denied", Permission: "full_disk_access"}
		}
		// File doesn't exist or other error
		return PermissionStatus{Status: "not_requested", Permission: "full_disk_access"}
	}
	f.Close()
	return PermissionStatus{Status: "granted", Permission: "full_disk_access"}
}
