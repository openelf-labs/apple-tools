package core

import (
	"errors"
	"testing"
)

func TestClassifyError_PermissionDenied(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		stderr   string
	}{
		{"not allowed", 1, "execution error: Not allowed (-1743)"},
		{"not permitted", 1, "operation not permitted"},
		{"assistive access", 1, "osascript is not allowed assistive access"},
		{"permission keyword", 1, "System Events got an error: permission denied"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ClassifyError(tt.exitCode, tt.stderr)
			if !errors.Is(err, ErrPermissionDenied) {
				t.Errorf("expected ErrPermissionDenied, got: %v", err)
			}
		})
	}
}

func TestClassifyError_AppNotRunning(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
	}{
		{"not running", `application "Music" is not running`},
		{"cant get app", `can't get application "Calendar"`},
		{"invalid connection", "connection is invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ClassifyError(1, tt.stderr)
			if !errors.Is(err, ErrAppNotRunning) {
				t.Errorf("expected ErrAppNotRunning, got: %v", err)
			}
		})
	}
}

func TestClassifyError_NotFound(t *testing.T) {
	err := ClassifyError(1, `can't get note "foo"`)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestClassifyError_NoError(t *testing.T) {
	if err := ClassifyError(0, ""); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestPermissionError_Unwrap(t *testing.T) {
	err := NewPermissionError("Calendar", "Automation")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Error("should unwrap to ErrPermissionDenied")
	}
}
