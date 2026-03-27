//go:build darwin

package shortcuts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_shortcuts_list",
		Description: "List all available Apple Shortcuts on this Mac. Returns JSON array of shortcut names. Use these names with apple_shortcuts_run.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler:     handleList,
	})

	r.Add(core.Tool{
		Name:        "apple_shortcuts_run",
		Description: "Run an Apple Shortcut by name. Returns JSON {ok, name, output}. Use apple_shortcuts_list to discover available shortcuts.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"name":{"type":"string","description":"Shortcut name (exact match, case-sensitive)"},
				"input":{"type":"string","description":"Optional text input to pass to the shortcut"}
			},
			"required":["name"]
		}`),
		Handler: handleRun,
	})
}

func handleList(ctx context.Context, input json.RawMessage) (string, error) {
	out, err := core.RunCommand(ctx, "shortcuts", "list")
	if err != nil {
		return "", fmt.Errorf("failed to list shortcuts: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	names := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}

	return core.MustFormatJSON(names), nil
}

func handleRun(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		return "", fmt.Errorf("%w: shortcut name is required", core.ErrInvalidInput)
	}

	args := []string{"run", params.Name}
	if params.Input != "" {
		args = append(args, "--input-type", "text", "--input", params.Input)
	}

	out, err := core.RunCommand(ctx, "shortcuts", args...)
	if err != nil {
		return "", fmt.Errorf("shortcut %q failed: %w", params.Name, err)
	}

	result := map[string]any{
		"ok":     true,
		"name":   params.Name,
		"output": strings.TrimSpace(string(out)),
	}
	return core.MustFormatJSON(result), nil
}
