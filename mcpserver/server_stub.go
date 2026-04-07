//go:build !darwin

package mcpserver

import (
	"context"
	"fmt"

	appletools "github.com/openelf-labs/apple-tools"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Server is a placeholder type for non-macOS platforms.
type Server struct{}

// New is a stub for non-macOS platforms. Returns nil.
func New(_ appletools.Config) *Server { return nil }

// ToolCount is a stub for non-macOS platforms. Always returns 0.
func ToolCount(_ appletools.Config) int { return 0 }

// Run is a stub for non-macOS platforms. Returns an error.
func Run(_ context.Context, _ *Server) error {
	return fmt.Errorf("apple tools MCP server is only available on macOS")
}
