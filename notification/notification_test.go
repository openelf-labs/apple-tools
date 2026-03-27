//go:build darwin

package notification

import (
	"encoding/json"
	"testing"

	"github.com/openelf-labs/apple-tools/testutil"
)

func TestRegister(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	if len(reg.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(reg.Tools))
	}

	tool := reg.FindTool("apple_notification_send")
	if tool == nil {
		t.Fatal("tool not registered")
	}
	if !json.Valid(tool.Parameters) {
		t.Error("invalid JSON schema")
	}
}

func TestSendValidation(t *testing.T) {
	reg := &testutil.MockRegistry{}
	Register(reg)

	tests := []struct {
		name   string
		params map[string]any
	}{
		{"empty title", map[string]any{"title": "", "message": "test"}},
		{"empty message", map[string]any{"title": "test", "message": ""}},
		{"whitespace title", map[string]any{"title": "   ", "message": "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testutil.CallTool(t, reg, "apple_notification_send", tt.params)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
