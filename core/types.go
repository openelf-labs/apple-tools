// Package core provides shared types and utilities for apple-tools.
// Subpackages (calendar, reminders, etc.) import core to avoid circular dependencies
// with the root package which orchestrates registration.
package core

import (
	"context"
	"encoding/json"
	"strings"
)

// Handler is the function signature for tool execution.
type Handler func(ctx context.Context, input json.RawMessage) (string, error)

// Tool describes a single tool that can be registered with a host application.
type Tool struct {
	Name        string          // Unique tool identifier
	Description string          // Human-readable description for the LLM
	Parameters  json.RawMessage // JSON Schema describing input parameters
	Handler     Handler         // The implementation function
}

// Registry is the interface that host applications must implement
// to receive tool registrations.
type Registry interface {
	Add(tool Tool)
}

// ClampLimit constrains limit to [1, max], using defaultVal when 0 or negative.
func ClampLimit(limit, defaultVal, max int) int {
	if limit <= 0 {
		return defaultVal
	}
	if limit > max {
		return max
	}
	return limit
}

// ContainsTraversal checks for path traversal patterns ("..") in the given path.
func ContainsTraversal(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// FormatJSON marshals v to a compact JSON string for tool output.
// This is the standard way to return structured data from tool handlers.
// LLMs consume JSON natively; let the LLM handle presentation to the user.
func FormatJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MustFormatJSON is like FormatJSON but returns the raw JSON on marshal error.
// Use when the input is known to be marshalable (e.g., already-parsed structs).
func MustFormatJSON(v any) string {
	s, err := FormatJSON(v)
	if err != nil {
		// Fallback: try to return something useful
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	return s
}
