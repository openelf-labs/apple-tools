//go:build darwin

// Package mcpserver provides an MCP (Model Context Protocol) server bridge
// for apple-tools. It converts core.Tool registrations into MCP tool
// definitions and handles stdio transport for integration with MCP clients.
package mcpserver

import (
	"os"
	"strings"

	appletools "github.com/openelf-labs/apple-tools"
)

// ConfigFromEnv builds a Config from environment variables.
// Each category can be disabled by setting APPLE_<UPPER>=false or APPLE_<UPPER>=0.
// Unset variables use DefaultConfig() values (all enabled except Messages).
//
// Examples:
//
//	APPLE_MESSAGES=true   → enable Messages (requires Full Disk Access)
//	APPLE_SAFARI=false    → disable Safari tools
//	APPLE_ENABLED=false   → disable all Apple tools
func ConfigFromEnv() appletools.Config {
	cfg := appletools.DefaultConfig()

	cfg.Enabled = envBool("APPLE_ENABLED", cfg.Enabled)
	cfg.Calendar = envBool("APPLE_CALENDAR", cfg.Calendar)
	cfg.Reminders = envBool("APPLE_REMINDERS", cfg.Reminders)
	cfg.Contacts = envBool("APPLE_CONTACTS", cfg.Contacts)
	cfg.Notes = envBool("APPLE_NOTES", cfg.Notes)
	cfg.Mail = envBool("APPLE_MAIL", cfg.Mail)
	cfg.Messages = envBool("APPLE_MESSAGES", cfg.Messages)
	cfg.Music = envBool("APPLE_MUSIC", cfg.Music)
	cfg.Safari = envBool("APPLE_SAFARI", cfg.Safari)
	cfg.Shortcuts = envBool("APPLE_SHORTCUTS", cfg.Shortcuts)
	cfg.System = envBool("APPLE_SYSTEM", cfg.System)
	cfg.Clipboard = envBool("APPLE_CLIPBOARD", cfg.Clipboard)
	cfg.Notification = envBool("APPLE_NOTIFICATION", cfg.Notification)
	cfg.Finder = envBool("APPLE_FINDER", cfg.Finder)
	cfg.Spotlight = envBool("APPLE_SPOTLIGHT", cfg.Spotlight)

	return cfg
}

// envBool reads an environment variable as a boolean.
// Returns defaultVal if the variable is not set or empty.
// Returns false for "false", "0", "no", "off" (case-insensitive).
// Returns true for any other non-empty value.
func envBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	switch strings.ToLower(v) {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}
