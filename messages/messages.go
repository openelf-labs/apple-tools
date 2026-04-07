//go:build darwin

// Package messages provides Apple Messages (iMessage) tools for sending messages,
// reading conversation history, and checking unread messages via JXA automation.
package messages

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

// emptyJSONArray is the canonical empty-list response.
const emptyJSONArray = "[]"

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("messages: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptSend        = mustLoad("send.js")
	scriptReadHistory = mustLoad("read_history.js")
	scriptGetUnread   = mustLoad("get_unread.js")
)

// phonePattern matches common phone number formats:
// +1234567890, (123) 456-7890, 123-456-7890, 1234567890, etc.
// Also accepts email addresses for iMessage.
var phonePattern = regexp.MustCompile(`^(\+?[\d\s\-\(\)]{7,20}|[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})$`)

// Register adds all messages tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolSend())
	r.Add(toolRead())
	r.Add(toolUnread())
}

// --- send ---

type sendParams struct {
	PhoneNumber string `json:"phoneNumber"`
	Message     string `json:"message"`
}

func toolSend() core.Tool {
	return core.Tool{
		Name: "messages_send",
		Description: `Send an iMessage to a phone number or Apple ID.

Sends a text message via the Messages app. The recipient must be reachable via iMessage. Requires Automation permission for Messages.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "phoneNumber": {
      "type": "string",
      "description": "Recipient phone number (e.g., '+15551234567') or Apple ID email."
    },
    "message": {
      "type": "string",
      "description": "The message text to send."
    }
  },
  "required": ["phoneNumber", "message"],
  "additionalProperties": false
}`),
		Handler: handleSend,
	}
}

func handleSend(ctx context.Context, input json.RawMessage) (string, error) {
	var p sendParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.PhoneNumber = strings.TrimSpace(p.PhoneNumber)
	if p.PhoneNumber == "" {
		return "", fmt.Errorf("%w: 'phoneNumber' is required and must not be empty", core.ErrInvalidInput)
	}
	if !phonePattern.MatchString(p.PhoneNumber) {
		return "", fmt.Errorf("%w: 'phoneNumber' must be a valid phone number or email address", core.ErrInvalidInput)
	}
	p.Message = strings.TrimSpace(p.Message)
	if p.Message == "" {
		return "", fmt.Errorf("%w: 'message' is required and must not be empty", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptSend, p)
	if err != nil {
		return "", classifySendError(err)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse send response: %v", err)
	}

	if !result.Success {
		return "", fmt.Errorf("send failed: %s", result.Message)
	}

	return string(raw), nil
}

// --- read history ---

type readParams struct {
	PhoneNumber string `json:"phoneNumber"`
	Limit       int    `json:"limit"`
}

type historyMessage struct {
	Content  string `json:"content"`
	Date     string `json:"date"`
	IsFromMe bool   `json:"isFromMe"`
	Sender   string `json:"sender"`
}

func toolRead() core.Tool {
	return core.Tool{
		Name: "messages_read",
		Description: `Read message history with a specific contact.

Reads recent messages from the Messages chat database for a given phone number or Apple ID. Requires Full Disk Access permission in System Settings > Privacy & Security > Full Disk Access.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "phoneNumber": {
      "type": "string",
      "description": "Contact phone number (e.g., '+15551234567') or Apple ID email."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of messages to return (1-50). Defaults to 10."
    }
  },
  "required": ["phoneNumber"],
  "additionalProperties": false
}`),
		Handler: handleRead,
	}
}

func handleRead(ctx context.Context, input json.RawMessage) (string, error) {
	var p readParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.PhoneNumber = strings.TrimSpace(p.PhoneNumber)
	if p.PhoneNumber == "" {
		return "", fmt.Errorf("%w: 'phoneNumber' is required and must not be empty", core.ErrInvalidInput)
	}
	if !phonePattern.MatchString(p.PhoneNumber) {
		return "", fmt.Errorf("%w: 'phoneNumber' must be a valid phone number or email address", core.ErrInvalidInput)
	}

	p.Limit = clampLimit(p.Limit, 10, 50)

	raw, err := core.RunJXA(ctx, scriptReadHistory, p)
	if err != nil {
		return "", classifyReadError(err)
	}

	var messages []historyMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		return "", fmt.Errorf("failed to parse message history: %v", err)
	}

	if len(messages) == 0 {
		return emptyJSONArray, nil
	}

	return string(raw), nil
}

// --- unread ---

type unreadParams struct {
	Limit int `json:"limit"`
}

type unreadMessage struct {
	Content string `json:"content"`
	Date    string `json:"date"`
	Sender  string `json:"sender"`
	ChatID  string `json:"chatId"`
}

func toolUnread() core.Tool {
	return core.Tool{
		Name: "messages_unread",
		Description: `Get unread messages from Apple Messages.

Reads unread incoming messages from the Messages chat database. Requires Full Disk Access permission in System Settings > Privacy & Security > Full Disk Access.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "limit": {
      "type": "integer",
      "description": "Maximum number of unread messages to return (1-50). Defaults to 10."
    }
  },
  "additionalProperties": false
}`),
		Handler: handleUnread,
	}
}

func handleUnread(ctx context.Context, input json.RawMessage) (string, error) {
	var p unreadParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Limit = clampLimit(p.Limit, 10, 50)

	raw, err := core.RunJXA(ctx, scriptGetUnread, p)
	if err != nil {
		return "", classifyReadError(err)
	}

	var messages []unreadMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		return "", fmt.Errorf("failed to parse unread messages: %v", err)
	}

	if len(messages) == 0 {
		return emptyJSONArray, nil
	}

	return string(raw), nil
}

// --- helpers ---

// clampLimit constrains limit to [1, max], using defaultVal when 0.
func clampLimit(limit, defaultVal, max int) int {
	return core.ClampLimit(limit, defaultVal, max)
}

// classifySendError wraps errors for send operations (requires Automation permission).
func classifySendError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Messages", "Automation")
	}
	return err
}

// classifyReadError wraps errors for read/unread operations (requires Full Disk Access).
func classifyReadError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Messages", "Full Disk Access")
	}
	return err
}
