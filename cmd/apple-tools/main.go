//go:build darwin

// apple-tools is the Apple Tools MCP server and CLI for macOS.
//
// As an MCP server (stdio transport):
//
//	apple-tools mcp
//
// As a standalone CLI:
//
//	apple-tools list [--json]
//	apple-tools call <tool> [json-params]
//	apple-tools permissions [--json] [--probe <category|all>]
//	apple-tools doctor [--json]
//	apple-tools version
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
	"github.com/openelf-labs/apple-tools/mcpserver"
	"github.com/openelf-labs/apple-tools/testutil"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	mcpserver.Version = version

	rootCmd := &cobra.Command{
		Use:   "apple-tools",
		Short: "Apple Tools — macOS application integration for OpenELF",
		Long:  "MCP server and CLI for interacting with macOS applications (Calendar, Reminders, Notes, Mail, etc.) via Apple automation.",
	}

	rootCmd.AddCommand(mcpCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(callCmd())
	rootCmd.AddCommand(permissionsCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- mcp ---

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server (stdio transport)",
		Long:  "Start the Apple Tools MCP server using stdin/stdout JSON-RPC transport. Used by MCP clients like OpenELF.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mcpserver.ConfigFromEnv()
			server := mcpserver.New(cfg)
			return mcpserver.Run(context.Background(), server)
		},
	}
}

// --- list ---

func listCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered Apple tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mcpserver.ConfigFromEnv()
			reg := &testutil.MockRegistry{}
			appletools.RegisterAll(reg, cfg)

			if jsonOutput {
				return printListJSON(reg)
			}
			printListHuman(reg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func printListJSON(reg *testutil.MockRegistry) error {
	type toolInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}
	tools := make([]toolInfo, len(reg.Tools))
	for i, t := range reg.Tools {
		tools[i] = toolInfo{
			Name:        t.Name,
			Description: t.Description,
			Category:    extractCategory(t.Name),
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{
		"total": len(tools),
		"tools": tools,
	})
}

func printListHuman(reg *testutil.MockRegistry) {
	categories := map[string][]core.Tool{}
	for _, t := range reg.Tools {
		cat := extractCategory(t.Name)
		categories[cat] = append(categories[cat], t)
	}

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
			desc := strings.ReplaceAll(t.Description, "\n", " ")
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("    %-35s %s\n", t.Name, desc)
		}
		fmt.Println()
	}
}

// --- call ---

func callCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "call <tool-name> [json-params]",
		Short: "Call an Apple tool directly",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mcpserver.ConfigFromEnv()
			reg := &testutil.MockRegistry{}
			appletools.RegisterAll(reg, cfg)

			toolName := args[0]
			params := "{}"
			if len(args) >= 2 {
				params = args[1]
			}

			tool := reg.FindTool(toolName)
			if tool == nil {
				return fmt.Errorf("tool %q not found. Use 'list' to see available tools", toolName)
			}

			if !json.Valid([]byte(params)) {
				return fmt.Errorf("invalid JSON params: %s", params)
			}

			result, err := tool.Handler(context.Background(), json.RawMessage(params))
			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}
}

// --- permissions ---

func permissionsCmd() *cobra.Command {
	var (
		jsonOutput bool
		probe      string
	)
	cmd := &cobra.Command{
		Use:   "permissions",
		Short: "Show Apple tool permission status",
		Long:  "Show the macOS TCC permission requirements for each tool category. Use --probe to actively test permission status (may trigger macOS permission dialogs).",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mcpserver.ConfigFromEnv()

			if probe != "" {
				return runProbe(cfg, probe, jsonOutput)
			}

			return showPermissionRequirements(cfg, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&probe, "probe", "", "Actively probe permission status (category name or 'all'). WARNING: may trigger macOS permission dialogs")
	return cmd
}

func showPermissionRequirements(cfg appletools.Config, jsonOutput bool) error {
	type catInfo struct {
		Category   string `json:"category"`
		Permission string `json:"permission_type"`
		Enabled    bool   `json:"enabled"`
	}

	enabledMap := categoryEnabledMap(cfg)

	var infos []catInfo
	for cat, cp := range core.CategoryPermissions {
		infos = append(infos, catInfo{
			Category:   cat,
			Permission: cp.Type,
			Enabled:    enabledMap[cat],
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Category < infos[j].Category })

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(infos)
	}

	fmt.Println("Apple tool permission requirements:")
	fmt.Println()
	for _, info := range infos {
		status := "enabled"
		if !info.Enabled {
			status = "disabled"
		}
		fmt.Printf("  %-15s %-20s %s\n", info.Category, info.Permission, status)
	}
	fmt.Println()
	fmt.Println("Use --probe <category|all> to actively test permission status.")
	fmt.Println("Note: probing may trigger macOS permission dialogs.")
	return nil
}

func runProbe(cfg appletools.Config, target string, jsonOutput bool) error {
	ctx := context.Background()

	if target == "all" {
		results := core.ProbeAll(ctx, categoryEnabledMap(cfg))

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(results)
		}

		// Sort categories for stable output.
		cats := make([]string, 0, len(results))
		for c := range results {
			cats = append(cats, c)
		}
		sort.Strings(cats)

		for _, cat := range cats {
			ps := results[cat]
			fmt.Printf("  %-15s %-15s %s\n", cat, ps.Permission, ps.Status)
		}
		return nil
	}

	// Single category.
	if !cfg.Enabled {
		ps := core.PermissionStatus{Status: "disabled", Permission: "none"}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(ps)
		}
		fmt.Printf("  %s: %s (%s)\n", target, ps.Status, ps.Permission)
		return nil
	}
	ps := core.ProbePermission(ctx, target)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(ps)
	}

	fmt.Printf("  %s: %s (%s)\n", target, ps.Status, ps.Permission)
	return nil
}

// --- doctor ---

func doctorCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose Apple tools environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

type checkResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, warn, fail
	Message string `json:"message"`
}

func runDoctor(jsonOutput bool) error {
	var checks []checkResult

	// Check osascript availability.
	if _, err := core.RunCommand(context.Background(), "which", "osascript"); err != nil {
		checks = append(checks, checkResult{"osascript", "fail", "osascript not found in PATH"})
	} else {
		checks = append(checks, checkResult{"osascript", "ok", "osascript available"})
	}

	// Check shortcuts CLI.
	if _, err := core.RunCommand(context.Background(), "which", "shortcuts"); err != nil {
		checks = append(checks, checkResult{"shortcuts_cli", "warn", "shortcuts CLI not found (Shortcuts tools may not work)"})
	} else {
		checks = append(checks, checkResult{"shortcuts_cli", "ok", "shortcuts CLI available"})
	}

	// Check mdfind (Spotlight).
	if _, err := core.RunCommand(context.Background(), "which", "mdfind"); err != nil {
		checks = append(checks, checkResult{"mdfind", "warn", "mdfind not found (Spotlight search may not work)"})
	} else {
		checks = append(checks, checkResult{"mdfind", "ok", "mdfind available"})
	}

	// Check tool registration.
	cfg := mcpserver.ConfigFromEnv()
	count := mcpserver.ToolCount(cfg)
	if count == 0 {
		checks = append(checks, checkResult{"tools", "fail", "no tools registered (check APPLE_ENABLED env)"})
	} else {
		checks = append(checks, checkResult{"tools", "ok", fmt.Sprintf("%d tools registered", count)})
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(checks)
	}

	hasError := false
	for _, c := range checks {
		icon := "✓"
		switch c.Status {
		case "warn":
			icon = "⚠"
		case "fail":
			icon = "✗"
			hasError = true
		}
		fmt.Printf("  %s %-20s %s\n", icon, c.Name, c.Message)
	}

	if hasError {
		return fmt.Errorf("some checks failed")
	}
	return nil
}

// --- version ---

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("apple-tools %s\n", version)
		},
	}
}

// --- helpers ---

// categoryEnabledMap builds a category→enabled map from Config,
// respecting the master Enabled flag.
func categoryEnabledMap(cfg appletools.Config) map[string]bool {
	m := map[string]bool{
		"calendar": cfg.Calendar, "reminders": cfg.Reminders, "contacts": cfg.Contacts,
		"notes": cfg.Notes, "mail": cfg.Mail, "messages": cfg.Messages,
		"music": cfg.Music, "safari": cfg.Safari, "shortcuts": cfg.Shortcuts,
		"system": cfg.System, "clipboard": cfg.Clipboard, "notification": cfg.Notification,
		"finder": cfg.Finder, "spotlight": cfg.Spotlight,
	}
	if !cfg.Enabled {
		for k := range m {
			m[k] = false
		}
	}
	return m
}

func extractCategory(toolName string) string {
	parts := strings.SplitN(toolName, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
