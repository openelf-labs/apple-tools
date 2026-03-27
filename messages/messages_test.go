//go:build darwin

package messages

import (
	"errors"
	"sort"
	"testing"

	"github.com/openelf-labs/apple-tools/core"
	"github.com/openelf-labs/apple-tools/testutil"
)

func newRegistry() *testutil.MockRegistry {
	return testutil.NewRegistryWith(func(r core.Registry) {
		Register(r)
	})
}

func TestRegister(t *testing.T) {
	reg := newRegistry()

	if got := len(reg.Tools); got != 3 {
		t.Fatalf("expected 3 tools, got %d", got)
	}

	expected := []string{
		"apple_messages_send",
		"apple_messages_read",
		"apple_messages_unread",
	}
	names := reg.ToolNames()
	sort.Strings(names)
	sort.Strings(expected)
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected tool %q at index %d, got %q", name, i, names[i])
		}
	}
}

func TestRegister_ToolsHaveSchemas(t *testing.T) {
	reg := newRegistry()
	for _, tool := range reg.Tools {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
		if len(tool.Parameters) == 0 {
			t.Errorf("tool %q has empty parameters schema", tool.Name)
		}
		if tool.Handler == nil {
			t.Errorf("tool %q has nil handler", tool.Name)
		}
	}
}

// --- Send validation tests ---

func TestSend_EmptyPhoneNumber(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_send", map[string]any{
		"phoneNumber": "",
		"message":     "hello",
	})
	if err == nil {
		t.Fatal("expected error for empty phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSend_MissingPhoneNumber(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_send", map[string]any{
		"message": "hello",
	})
	if err == nil {
		t.Fatal("expected error for missing phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSend_InvalidPhoneNumber(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_send", map[string]any{
		"phoneNumber": "abc",
		"message":     "hello",
	})
	if err == nil {
		t.Fatal("expected error for invalid phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSend_EmptyMessage(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_send", map[string]any{
		"phoneNumber": "+15551234567",
		"message":     "",
	})
	if err == nil {
		t.Fatal("expected error for empty message")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSend_MissingMessage(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_send", map[string]any{
		"phoneNumber": "+15551234567",
	})
	if err == nil {
		t.Fatal("expected error for missing message")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSend_EmailAsPhoneNumber(t *testing.T) {
	// Email should be accepted as a valid phoneNumber (for iMessage).
	// Only verify the pattern matches; calling the handler would invoke JXA.
	if !phonePattern.MatchString("user@example.com") {
		t.Error("email should be accepted by phonePattern")
	}
}

// --- Read validation tests ---

func TestRead_EmptyPhoneNumber(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_read", map[string]any{
		"phoneNumber": "",
	})
	if err == nil {
		t.Fatal("expected error for empty phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestRead_InvalidPhoneNumber(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "apple_messages_read", map[string]any{
		"phoneNumber": "not-a-phone",
	})
	if err == nil {
		t.Fatal("expected error for invalid phoneNumber")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Helpers ---

func TestClampLimit(t *testing.T) {
	tests := []struct {
		input, def, max, want int
	}{
		{0, 10, 50, 10},
		{-5, 10, 50, 10},
		{25, 10, 50, 25},
		{50, 10, 50, 50},
		{51, 10, 50, 50},
		{1, 10, 50, 1},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input, tt.def, tt.max)
		if got != tt.want {
			t.Errorf("clampLimit(%d, %d, %d) = %d, want %d", tt.input, tt.def, tt.max, got, tt.want)
		}
	}
}

func TestPhonePattern(t *testing.T) {
	valid := []string{
		"+15551234567",
		"5551234567",
		"(555) 123-4567",
		"+86 138 0000 0000",
		"user@example.com",
		"test.user+tag@gmail.com",
	}
	for _, v := range valid {
		if !phonePattern.MatchString(v) {
			t.Errorf("expected %q to match phone pattern", v)
		}
	}

	invalid := []string{
		"abc",
		"12",
		"",
		"not a phone",
	}
	for _, v := range invalid {
		if phonePattern.MatchString(v) {
			t.Errorf("expected %q to NOT match phone pattern", v)
		}
	}
}
