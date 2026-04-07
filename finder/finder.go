//go:build darwin

// Package finder provides a Finder tool for revealing files via JXA automation.
package finder

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("finder: embedded script %s not found: %v", name, err))
	}
	return data
}

var scriptReveal = mustLoad("reveal.js")

// Register adds all finder tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolReveal())
}

// --- reveal in Finder ---

type revealParams struct {
	Path string `json:"path"`
}

type revealResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func toolReveal() core.Tool {
	return core.Tool{
		Name: "finder_reveal",
		Description: `Reveal a file or folder in Finder.

Opens Finder and selects the specified file or folder. The path must be an absolute POSIX path (starting with /).`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Absolute POSIX path to the file or folder to reveal (e.g., '/Users/username/Documents/file.txt')."
    }
  },
  "required": ["path"],
  "additionalProperties": false
}`),
		Handler: handleReveal,
	}
}

func handleReveal(ctx context.Context, input json.RawMessage) (string, error) {
	var p revealParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Path = strings.TrimSpace(p.Path)
	if p.Path == "" {
		return "", fmt.Errorf("%w: 'path' is required and must not be empty", core.ErrInvalidInput)
	}
	if !strings.HasPrefix(p.Path, "/") {
		return "", fmt.Errorf("%w: 'path' must be an absolute path (starting with /)", core.ErrInvalidInput)
	}
	if core.ContainsTraversal(p.Path) {
		return "", fmt.Errorf("%w: 'path' must not contain path traversal sequences (..)", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptReveal, p)
	if err != nil {
		return "", classifyFinderError(err)
	}

	var result revealResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse reveal response: %v", err)
	}

	return core.MustFormatJSON(map[string]any{
		"ok":   true,
		"path": result.Path,
	}), nil
}

// --- helpers ---

// classifyFinderError wraps errors with Finder-specific context.
func classifyFinderError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Finder", "Automation")
	}
	return err
}
