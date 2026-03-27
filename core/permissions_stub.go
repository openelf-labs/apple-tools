//go:build !darwin

package core

import "context"

// PermissionStatus represents the authorization state of an Apple tool category.
type PermissionStatus struct {
	Status     string `json:"status"`
	Permission string `json:"permission"`
}

// CategoryPermissions defines permission requirements per category (SSOT).
var CategoryPermissions = map[string]string{
	"calendar": "automation", "reminders": "automation", "contacts": "automation",
	"notes": "automation", "mail": "automation", "messages": "full_disk_access",
	"music": "automation", "safari": "automation", "finder": "automation",
	"shortcuts": "none", "system": "none", "clipboard": "none",
	"notification": "none", "spotlight": "none",
}

// ProbePermission is a no-op on non-macOS platforms.
func ProbePermission(_ context.Context, category string) PermissionStatus {
	return PermissionStatus{Status: "unavailable", Permission: "none"}
}

// ProbeAll is a no-op on non-macOS platforms.
func ProbeAll(_ context.Context, _ map[string]bool) map[string]PermissionStatus {
	return nil
}
