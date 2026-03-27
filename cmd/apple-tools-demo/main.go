//go:build darwin

// apple-tools-demo is a standalone CLI for testing and debugging apple-tools.
// Contributors can use this to verify their tools without running OpenELF.
//
// Usage:
//
//	go run ./cmd/apple-tools-demo list                               # List all tools
//	go run ./cmd/apple-tools-demo call apple_calendar_list           # Call with empty params
//	go run ./cmd/apple-tools-demo call apple_calendar_list '{"limit":3}'  # Call with JSON params
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	appletools "github.com/openelf-labs/apple-tools"
	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/testutil"
)

func main() {
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, appletools.DefaultConfig())

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listTools(reg)
	case "call":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: apple-tools-demo call <tool-name> [json-params]\n")
			os.Exit(1)
		}
		toolName := os.Args[2]
		params := "{}"
		if len(os.Args) >= 4 {
			params = os.Args[3]
		}
		callTool(reg, toolName, params)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `apple-tools-demo — standalone Apple tools CLI

Usage:
  apple-tools-demo list                          List all registered tools
  apple-tools-demo call <tool> [json-params]     Call a tool with optional JSON params

Examples:
  apple-tools-demo list
  apple-tools-demo call apple_calendar_list
  apple-tools-demo call apple_calendar_list '{"limit":3}'
  apple-tools-demo call apple_shortcuts_list
  apple-tools-demo call apple_system_battery
`)
}

func listTools(reg *testutil.MockRegistry) {
	// Group by category
	categories := map[string][]core.Tool{}
	for _, t := range reg.Tools {
		parts := strings.SplitN(strings.TrimPrefix(t.Name, "apple_"), "_", 2)
		cat := parts[0]
		categories[cat] = append(categories[cat], t)
	}

	// Sort category names
	cats := make([]string, 0, len(categories))
	for c := range categories {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	fmt.Printf("Registered %d Apple tools:\n\n", len(reg.Tools))
	for _, cat := range cats {
		tools := categories[cat]
		fmt.Printf("  %s (%d):\n", strings.ToUpper(cat), len(tools))
		for _, t := range tools {
			fmt.Printf("    %-35s %s\n", t.Name, truncate(t.Description, 60))
		}
		fmt.Println()
	}
}

func callTool(reg *testutil.MockRegistry, name, paramsJSON string) {
	tool := reg.FindTool(name)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool %q not found. Use 'list' to see available tools.\n", name)
		os.Exit(1)
	}

	// Validate JSON
	if !json.Valid([]byte(paramsJSON)) {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON params: %s\n", paramsJSON)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Calling %s...\n", name)
	result, err := tool.Handler(context.Background(), json.RawMessage(paramsJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
