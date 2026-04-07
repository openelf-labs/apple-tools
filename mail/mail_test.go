//go:build darwin

package mail

import (
	"encoding/json"
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

	if got := len(reg.Tools); got != 8 {
		t.Fatalf("expected 8 tools, got %d", got)
	}

	expected := []string{
		"mail_mailboxes",
		"mail_list",
		"mail_read",
		"mail_search",
		"mail_compose",
		"mail_reply",
		"mail_move",
		"mail_set_status",
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

// --- Parameter validation: list messages ---

func TestListMessages_EmptyMailbox(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_list", map[string]any{
		"mailbox": "",
	})
	if err == nil {
		t.Fatal("expected error for empty mailbox")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestListMessages_MissingMailbox(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_list", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing mailbox")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestListMessages_InvalidSinceDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_list", map[string]any{
		"mailbox": "iCloud/INBOX",
		"since":   "not-a-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid since date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: read message ---

func TestReadMessage_EmptyMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_read", map[string]any{
		"messageId": "",
	})
	if err == nil {
		t.Fatal("expected error for empty messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReadMessage_MissingMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_read", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: search messages ---

// TestSearchMessages_EmptyQuery verifies that omitting or passing an empty query
// does not produce a validation error (list-all mode).
func TestSearchMessages_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: would invoke JXA")
	}

	reg := newRegistry()
	cases := []map[string]any{
		{"query": ""},
		{},
	}
	for _, params := range cases {
		_, err := testutil.CallTool(t, reg, "mail_search", params)
		// Should not produce ErrInvalidInput; JXA/timeout/permission errors are OK.
		if errors.Is(err, core.ErrInvalidInput) {
			t.Errorf("empty query should not produce ErrInvalidInput, params=%v, got: %v", params, err)
		}
	}
}

func TestSearchMessages_InvalidSinceDate(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_search", map[string]any{
		"query": "test",
		"since": "bad-date",
	})
	if err == nil {
		t.Fatal("expected error for invalid since date")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: compose ---

func TestCompose_EmptyTo(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_compose", map[string]any{
		"to":      []string{},
		"subject": "Test",
	})
	if err == nil {
		t.Fatal("expected error for empty to array")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCompose_MissingTo(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_compose", map[string]any{
		"subject": "Test",
	})
	if err == nil {
		t.Fatal("expected error for missing to")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCompose_EmptySubject(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_compose", map[string]any{
		"to":      []string{"test@example.com"},
		"subject": "",
	})
	if err == nil {
		t.Fatal("expected error for empty subject")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestCompose_EmptyAddressInTo(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_compose", map[string]any{
		"to":      []string{"valid@example.com", ""},
		"subject": "Test",
	})
	if err == nil {
		t.Fatal("expected error for empty address in to array")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: reply ---

func TestReply_EmptyMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_reply", map[string]any{
		"messageId": "",
		"body":      "Thanks!",
	})
	if err == nil {
		t.Fatal("expected error for empty messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReply_MissingMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_reply", map[string]any{
		"body": "Thanks!",
	})
	if err == nil {
		t.Fatal("expected error for missing messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReply_EmptyBody(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_reply", map[string]any{
		"messageId": "msg-123@example.com",
		"body":      "",
	})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestReply_MissingBody(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_reply", map[string]any{
		"messageId": "msg-123@example.com",
	})
	if err == nil {
		t.Fatal("expected error for missing body")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: move message ---

func TestMoveMessage_EmptyMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_move", map[string]any{
		"messageId":          "",
		"destinationMailbox": "iCloud/Archive",
	})
	if err == nil {
		t.Fatal("expected error for empty messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestMoveMessage_MissingMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_move", map[string]any{
		"destinationMailbox": "iCloud/Archive",
	})
	if err == nil {
		t.Fatal("expected error for missing messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestMoveMessage_EmptyDestination(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_move", map[string]any{
		"messageId":          "msg-123@example.com",
		"destinationMailbox": "",
	})
	if err == nil {
		t.Fatal("expected error for empty destinationMailbox")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestMoveMessage_MissingDestination(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_move", map[string]any{
		"messageId": "msg-123@example.com",
	})
	if err == nil {
		t.Fatal("expected error for missing destinationMailbox")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- Parameter validation: set status ---

func TestSetStatus_EmptyMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_set_status", map[string]any{
		"messageId": "",
		"isRead":    true,
	})
	if err == nil {
		t.Fatal("expected error for empty messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSetStatus_MissingMessageID(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_set_status", map[string]any{
		"isRead": true,
	})
	if err == nil {
		t.Fatal("expected error for missing messageId")
	}
	if !errors.Is(err, core.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSetStatus_NoFieldsSpecified(t *testing.T) {
	reg := newRegistry()
	_, err := testutil.CallTool(t, reg, "mail_set_status", map[string]any{
		"messageId": "msg-123@example.com",
	})
	if err == nil {
		t.Fatal("expected error when no status fields are specified")
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
		{0, 20, 50, 20},
		{-5, 10, 50, 10},
		{30, 20, 50, 30},
		{50, 20, 50, 50},
		{51, 20, 50, 50},
		{1, 20, 50, 1},
		{100, 10, 50, 50},
	}
	for _, tt := range tests {
		got := clampLimit(tt.input, tt.def, tt.max)
		if got != tt.want {
			t.Errorf("clampLimit(%d, %d, %d) = %d, want %d", tt.input, tt.def, tt.max, got, tt.want)
		}
	}
}

// --- Integration test (read-only, safe to run) ---

func TestListMailboxes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	reg := newRegistry()
	result, err := testutil.CallTool(t, reg, "mail_mailboxes", map[string]any{})
	if err != nil {
		// Permission, app-not-running, and timeout errors are expected
		// in CI, sandboxed environments, or headless setups.
		if errors.Is(err, core.ErrPermissionDenied) ||
			errors.Is(err, core.ErrAppNotRunning) ||
			errors.Is(err, core.ErrTimeout) {
			t.Skipf("skipping: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Result must be valid JSON (array of mailboxes or empty array).
	if !json.Valid([]byte(result)) {
		t.Fatalf("expected valid JSON, got: %s", result)
	}

	var mailboxes []mailboxInfo
	if err := json.Unmarshal([]byte(result), &mailboxes); err != nil {
		t.Fatalf("failed to unmarshal mailboxes: %v", err)
	}

	for _, mb := range mailboxes {
		if mb.Path == "" {
			t.Error("expected non-empty mailbox path")
		}
	}
	t.Logf("mailboxes result (%d items): %s", len(mailboxes), result)
}
