//go:build darwin

package core

import (
	"context"
	"testing"
)

func TestCategoryPermissions_Complete(t *testing.T) {
	// Verify all expected categories have entries
	expected := []string{
		"calendar", "reminders", "contacts", "notes", "mail", "messages",
		"music", "safari", "shortcuts", "system", "clipboard", "notification",
		"finder", "spotlight",
	}
	for _, cat := range expected {
		if _, ok := CategoryPermissions[cat]; !ok {
			t.Errorf("CategoryPermissions missing entry for %q", cat)
		}
	}
}

func TestCategoryPermissions_ValidValues(t *testing.T) {
	validPerms := map[string]bool{"automation": true, "full_disk_access": true, "none": true}
	for cat, perm := range CategoryPermissions {
		if !validPerms[perm] {
			t.Errorf("CategoryPermissions[%q] = %q, not a valid permission type", cat, perm)
		}
	}
}

func TestProbePermission_NoPermissionCategories(t *testing.T) {
	noPerm := []string{"shortcuts", "system", "clipboard", "notification", "spotlight"}
	for _, cat := range noPerm {
		status := ProbePermission(context.Background(), cat)
		if status.Status != "no_permission" {
			t.Errorf("ProbePermission(%q) = %q, expected no_permission", cat, status.Status)
		}
		if status.Permission != "none" {
			t.Errorf("ProbePermission(%q).Permission = %q, expected none", cat, status.Permission)
		}
	}
}

func TestProbePermission_UnknownCategory(t *testing.T) {
	status := ProbePermission(context.Background(), "nonexistent")
	if status.Status != "unknown" {
		t.Errorf("expected unknown status for nonexistent category, got %q", status.Status)
	}
}

func TestProbePermission_FullDiskAccess(t *testing.T) {
	status := ProbePermission(context.Background(), "messages")
	// We can't know the result, but it should be one of the valid states
	validStates := map[string]bool{"granted": true, "denied": true, "not_requested": true}
	if !validStates[status.Status] {
		t.Errorf("ProbePermission(messages) = %q, not a valid state", status.Status)
	}
	if status.Permission != "full_disk_access" {
		t.Errorf("expected full_disk_access permission, got %q", status.Permission)
	}
}

func TestProbeAll_DisabledCategories(t *testing.T) {
	enabled := map[string]bool{
		"calendar": false, // explicitly disabled
		"shortcuts": true,
	}
	result := ProbeAll(context.Background(), enabled)
	if result["calendar"].Status != "disabled" {
		t.Errorf("expected disabled for calendar, got %q", result["calendar"].Status)
	}
	if result["shortcuts"].Status != "no_permission" {
		t.Errorf("expected no_permission for shortcuts, got %q", result["shortcuts"].Status)
	}
}
