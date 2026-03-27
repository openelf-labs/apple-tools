package appletools

import "github.com/openelf-labs/apple-tools/core"

// Re-export exec functions from core.
var (
	RunJXA    = core.RunJXA
	RunCommand = core.RunCommand
)

// Re-export constants.
const (
	DefaultTimeout = core.DefaultTimeout
	ParamsEnvKey   = core.ParamsEnvKey
)
