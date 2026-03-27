package appletools

import "github.com/openelf-labs/apple-tools/core"

// Re-export error types from core.
var (
	ErrPermissionDenied = core.ErrPermissionDenied
	ErrAppNotRunning    = core.ErrAppNotRunning
	ErrNotFound         = core.ErrNotFound
	ErrInvalidInput     = core.ErrInvalidInput
	ErrTimeout          = core.ErrTimeout
	ErrNotMacOS         = core.ErrNotMacOS
)

// Re-export error constructors.
var (
	NewPermissionError = core.NewPermissionError
	ClassifyError      = core.ClassifyError
)

// Re-export PermissionError type.
type PermissionError = core.PermissionError
