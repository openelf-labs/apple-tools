//go:build darwin

package main

import (
	"testing"

	appletools "github.com/openelf-labs/apple-tools"
	"github.com/openelf-labs/apple-tools/mcpserver"
	"github.com/openelf-labs/apple-tools/testutil"
)

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"apple_calendar_list", "calendar"},
		{"apple_system_battery", "system"},
		{"apple_music_now_playing", "music"},
		{"apple_messages_send", "messages"},
		{"apple_shortcuts_run", "shortcuts"},
	}
	for _, tt := range tests {
		got := extractCategory(tt.input)
		if got != tt.want {
			t.Errorf("extractCategory(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListRegistersTools(t *testing.T) {
	cfg := appletools.DefaultConfig()
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, cfg)

	if len(reg.Tools) == 0 {
		t.Fatal("expected tools to be registered with default config")
	}

	// Every tool should have apple_ prefix.
	for _, tool := range reg.Tools {
		if tool.Name[:6] != "apple_" {
			t.Errorf("tool %q missing apple_ prefix", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
		if tool.Handler == nil {
			t.Errorf("tool %q has nil handler", tool.Name)
		}
	}
}

func TestMCPServerCreation(t *testing.T) {
	cfg := appletools.DefaultConfig()
	server := mcpserver.New(cfg)
	if server == nil {
		t.Fatal("mcpserver.New returned nil")
	}
}

func TestToolCountConsistency(t *testing.T) {
	cfg := appletools.DefaultConfig()

	// ToolCount should match MockRegistry count.
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, cfg)
	regCount := len(reg.Tools)

	mcpCount := mcpserver.ToolCount(cfg)

	if regCount != mcpCount {
		t.Errorf("MockRegistry count (%d) != ToolCount (%d)", regCount, mcpCount)
	}
}

func TestCallToolNotFound(t *testing.T) {
	cfg := appletools.DefaultConfig()
	reg := &testutil.MockRegistry{}
	appletools.RegisterAll(reg, cfg)

	tool := reg.FindTool("nonexistent_tool")
	if tool != nil {
		t.Error("expected nil for nonexistent tool")
	}
}
