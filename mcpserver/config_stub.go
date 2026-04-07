//go:build !darwin

package mcpserver

import appletools "github.com/openelf-labs/apple-tools"

// ConfigFromEnv is a stub for non-macOS platforms.
// Returns a disabled config since Apple tools are not available.
func ConfigFromEnv() appletools.Config {
	cfg := appletools.DefaultConfig()
	cfg.Enabled = false
	return cfg
}
