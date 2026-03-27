//go:build !darwin

package core

import "context"

// PermissionStatus represents the authorization state of an Apple tool category.
type PermissionStatus struct {
	Status     string `json:"status"`
	Permission string `json:"permission"`
}

// CategoryPermission describes a tool category's permission requirements.
type CategoryPermission struct {
	Type        string `json:"type"`
	SettingsURL string `json:"settings_url"`
}

// CategoryPermissions defines permission requirements per category (SSOT).
var CategoryPermissions = map[string]CategoryPermission{
	"calendar": {Type: "automation"}, "reminders": {Type: "automation"}, "contacts": {Type: "automation"},
	"notes": {Type: "automation"}, "mail": {Type: "automation"}, "messages": {Type: "full_disk_access"},
	"music": {Type: "automation"}, "safari": {Type: "automation"}, "finder": {Type: "automation"},
	"shortcuts": {Type: "none"}, "system": {Type: "none"}, "clipboard": {Type: "none"},
	"notification": {Type: "none"}, "spotlight": {Type: "none"},
}

func ProbePermission(_ context.Context, _ string) PermissionStatus {
	return PermissionStatus{Status: "unavailable", Permission: "none"}
}

func ProbeAll(_ context.Context, _ map[string]bool) map[string]PermissionStatus {
	return nil
}

func OpenSystemSettings(_ string) error { return nil }
