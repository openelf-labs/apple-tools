//go:build !darwin

package core

import (
	"context"
	"encoding/json"
	"time"
)

// DefaultTimeout is the default timeout for Apple tool operations.
const DefaultTimeout = 30 * time.Second

// ParamsEnvKey is the environment variable used to pass JSON parameters.
const ParamsEnvKey = "APPLE_TOOLS_PARAMS"

// RunJXA is a stub for non-macOS platforms.
func RunJXA(_ context.Context, _ []byte, _ any) (json.RawMessage, error) {
	return nil, ErrNotMacOS
}

// RunCommand is a stub for non-macOS platforms.
func RunCommand(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, ErrNotMacOS
}
