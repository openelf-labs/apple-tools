//go:build !darwin

package core

import (
	"context"
	"encoding/json"
)

// RunJXA is a stub for non-macOS platforms.
func RunJXA(_ context.Context, _ []byte, _ any) (json.RawMessage, error) {
	return nil, ErrNotMacOS
}

// RunCommand is a stub for non-macOS platforms.
func RunCommand(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, ErrNotMacOS
}
