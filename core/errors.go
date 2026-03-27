package core

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for common failure modes.
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrAppNotRunning    = errors.New("application not running")
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrTimeout          = errors.New("operation timed out")
	ErrNotMacOS         = errors.New("apple tools are only available on macOS")
)

// PermissionError provides actionable guidance for resolving macOS permission issues.
type PermissionError struct {
	App      string
	Category string
	Guide    string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: %s requires %s access. %s", e.App, e.Category, e.Guide)
}

func (e *PermissionError) Unwrap() error {
	return ErrPermissionDenied
}

// NewPermissionError creates a PermissionError with a localized guidance message.
func NewPermissionError(app, category string) *PermissionError {
	guide := fmt.Sprintf("Grant access in System Settings > Privacy & Security > %s > Allow control of \"%s\"", category, app)
	return &PermissionError{App: app, Category: category, Guide: guide}
}

// ClassifyError inspects osascript exit code and stderr to return a typed error.
func ClassifyError(exitCode int, stderr string) error {
	lower := strings.ToLower(stderr)

	if strings.Contains(lower, "not allowed") ||
		strings.Contains(lower, "-1743") ||
		strings.Contains(lower, "not permitted") ||
		strings.Contains(lower, "assistive access") ||
		strings.Contains(lower, "permission") {
		return fmt.Errorf("%w: %s", ErrPermissionDenied, strings.TrimSpace(stderr))
	}

	if strings.Contains(lower, "not running") ||
		strings.Contains(lower, "can't get application") ||
		strings.Contains(lower, "connection is invalid") {
		return fmt.Errorf("%w: %s", ErrAppNotRunning, strings.TrimSpace(stderr))
	}

	if strings.Contains(lower, "can't get") ||
		strings.Contains(lower, "doesn't exist") ||
		strings.Contains(lower, "missing value") ||
		strings.Contains(lower, "no result") {
		return fmt.Errorf("%w: %s", ErrNotFound, strings.TrimSpace(stderr))
	}

	if exitCode != 0 {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = fmt.Sprintf("osascript exited with code %d", exitCode)
		}
		return fmt.Errorf("osascript error: %s", msg)
	}

	return nil
}
