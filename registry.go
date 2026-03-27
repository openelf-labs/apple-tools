// Package appletools provides macOS Apple application integration tools.
//
// Tools are registered via the Registry interface which host applications implement.
// On non-macOS platforms, RegisterAll is a no-op.
//
// Subpackages use github.com/openelf-labs/apple-tools/core for shared types
// to avoid circular imports with this root orchestration package.
package appletools

import "github.com/openelf-labs/apple-tools/core"

// Re-export core types so external consumers can use the root package.
type (
	Handler  = core.Handler
	Tool     = core.Tool
	Registry = core.Registry
)

// Re-export utility functions from core.
var (
	ClampLimit        = core.ClampLimit
	ContainsTraversal = core.ContainsTraversal
	ProbePermission   = core.ProbePermission
	ProbeAll          = core.ProbeAll
)

// Re-export permission types from core (SSOT for permission requirements).
type PermissionStatus = core.PermissionStatus

var CategoryPermissions = core.CategoryPermissions
