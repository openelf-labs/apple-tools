//go:build darwin

// Package mail provides Apple Mail tools for listing mailboxes, reading,
// searching, composing, replying, moving, and managing message status
// via JXA automation.
package mail

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openelf-labs/apple-tools/core"
)

// emptyJSONArray is the canonical empty-list response.
const emptyJSONArray = "[]"

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("mail: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptListMailboxes  = mustLoad("list_mailboxes.js")
	scriptListMessages   = mustLoad("list_messages.js")
	scriptReadMessage    = mustLoad("read_message.js")
	scriptSearchMessages = mustLoad("search_messages.js")
	scriptCompose        = mustLoad("compose.js")
	scriptReply          = mustLoad("reply.js")
	scriptMoveMessage    = mustLoad("move_message.js")
	scriptSetStatus      = mustLoad("set_status.js")
)

// Limits for message listing and search.
const (
	defaultListLimit   = 20
	maxListLimit       = 50
	defaultSearchLimit = 10
	maxSearchLimit     = 50
	defaultMaxLength   = 10000
	maxMaxLength       = 100000
)

// Register adds all mail tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolListMailboxes())
	r.Add(toolListMessages())
	r.Add(toolReadMessage())
	r.Add(toolSearchMessages())
	r.Add(toolCompose())
	r.Add(toolReply())
	r.Add(toolMoveMessage())
	r.Add(toolSetStatus())
}

// --- list mailboxes ---

type mailboxInfo struct {
	Path         string `json:"path"`
	Account      string `json:"account"`
	Name         string `json:"name"`
	UnreadCount  int    `json:"unreadCount"`
	MessageCount int    `json:"messageCount"`
}

func toolListMailboxes() core.Tool {
	return core.Tool{
		Name: "apple_mail_mailboxes",
		Description: `List all mailboxes from all configured mail accounts.

Returns mailbox paths (e.g., "iCloud/INBOX", "Gmail/[Gmail]/All Mail") with unread and total message counts. Use these paths as the mailbox parameter in other mail tools.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleListMailboxes,
	}
}

func handleListMailboxes(ctx context.Context, _ json.RawMessage) (string, error) {
	raw, err := core.RunJXA(ctx, scriptListMailboxes, struct{}{})
	if err != nil {
		return "", classifyMailError(err)
	}

	var mailboxes []mailboxInfo
	if err := json.Unmarshal(raw, &mailboxes); err != nil {
		return "", fmt.Errorf("failed to parse mailbox response: %v", err)
	}

	if len(mailboxes) == 0 {
		return emptyJSONArray, nil
	}

	return string(raw), nil
}

// --- list messages ---

type listParams struct {
	Mailbox    string `json:"mailbox"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	UnreadOnly bool   `json:"unreadOnly"`
	Since      string `json:"since"`
	From       string `json:"from"`
}

type messageSummary struct {
	MessageID      string `json:"messageId"`
	Subject        string `json:"subject"`
	Sender         string `json:"sender"`
	DateReceived   string `json:"dateReceived"`
	IsRead         bool   `json:"isRead"`
	IsFlagged      bool   `json:"isFlagged"`
	HasAttachments bool   `json:"hasAttachments"`
}

type listResult struct {
	Messages []messageSummary `json:"messages"`
	Total    int              `json:"total"`
	HasMore  bool             `json:"hasMore"`
}

func toolListMessages() core.Tool {
	return core.Tool{
		Name: "apple_mail_list",
		Description: `List messages in a specific mailbox.

Returns messages sorted by date (newest first). Use the mailbox path from apple_mail_mailboxes (e.g., "iCloud/INBOX"). Supports filtering by unread status, sender, and date.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "mailbox": {
      "type": "string",
      "description": "Mailbox path (e.g., 'iCloud/INBOX'). Use apple_mail_mailboxes to find available paths."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of messages to return (1-50). Defaults to 20."
    },
    "offset": {
      "type": "integer",
      "description": "Number of messages to skip for pagination. Defaults to 0."
    },
    "unreadOnly": {
      "type": "boolean",
      "description": "If true, only return unread messages. Defaults to false."
    },
    "since": {
      "type": "string",
      "description": "Only return messages received after this ISO 8601 date (e.g., '2025-01-15T00:00:00Z')."
    },
    "from": {
      "type": "string",
      "description": "Filter by sender (case-insensitive substring match on sender address/name)."
    }
  },
  "required": ["mailbox"],
  "additionalProperties": false
}`),
		Handler: handleListMessages,
	}
}

func handleListMessages(ctx context.Context, input json.RawMessage) (string, error) {
	var p listParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Mailbox = strings.TrimSpace(p.Mailbox)
	if p.Mailbox == "" {
		return "", fmt.Errorf("%w: 'mailbox' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.Since != "" {
		if _, err := time.Parse(time.RFC3339, p.Since); err != nil {
			return "", fmt.Errorf("%w: invalid 'since' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}

	p.Limit = clampLimit(p.Limit, defaultListLimit, maxListLimit)
	if p.Offset < 0 {
		p.Offset = 0
	}

	raw, err := core.RunJXA(ctx, scriptListMessages, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	var result listResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse message list response: %v", err)
	}

	return string(raw), nil
}

// --- read message ---

type readParams struct {
	MessageID string `json:"messageId"`
	MaxLength int    `json:"maxLength"`
}

func toolReadMessage() core.Tool {
	return core.Tool{
		Name: "apple_mail_read",
		Description: `Read the full content of a specific email message.

Returns the message body, sender, recipients, and metadata. Use the messageId from apple_mail_list or apple_mail_search results. Body is truncated to maxLength characters (default 10000).`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "messageId": {
      "type": "string",
      "description": "The RFC Message-ID of the email (from list or search results)."
    },
    "maxLength": {
      "type": "integer",
      "description": "Maximum body length in characters (1-100000). Defaults to 10000."
    }
  },
  "required": ["messageId"],
  "additionalProperties": false
}`),
		Handler: handleReadMessage,
	}
}

func handleReadMessage(ctx context.Context, input json.RawMessage) (string, error) {
	var p readParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.MessageID = strings.TrimSpace(p.MessageID)
	if p.MessageID == "" {
		return "", fmt.Errorf("%w: 'messageId' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.MaxLength <= 0 {
		p.MaxLength = defaultMaxLength
	}
	if p.MaxLength > maxMaxLength {
		p.MaxLength = maxMaxLength
	}

	raw, err := core.RunJXA(ctx, scriptReadMessage, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse read message response: invalid JSON")
	}

	return string(raw), nil
}

// --- search messages ---

type searchParams struct {
	Query   string `json:"query"`
	Mailbox string `json:"mailbox"`
	Limit   int    `json:"limit"`
	Since   string `json:"since"`
}

func toolSearchMessages() core.Tool {
	return core.Tool{
		Name: "apple_mail_search",
		Description: `Search email messages by content, subject, or sender.

Performs a case-insensitive search across messages. Optionally restrict to a specific mailbox. Returns matching messages sorted by date (newest first).`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search text to match against subject, sender, and message content (case-insensitive)."
    },
    "mailbox": {
      "type": "string",
      "description": "Optional mailbox path to restrict search (e.g., 'iCloud/INBOX')."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of results to return (1-50). Defaults to 10."
    },
    "since": {
      "type": "string",
      "description": "Only search messages received after this ISO 8601 date."
    }
  },
  "required": ["query"],
  "additionalProperties": false
}`),
		Handler: handleSearchMessages,
	}
}

func handleSearchMessages(ctx context.Context, input json.RawMessage) (string, error) {
	var p searchParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Query = strings.TrimSpace(p.Query)
	if p.Query == "" {
		return "", fmt.Errorf("%w: 'query' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.Since != "" {
		if _, err := time.Parse(time.RFC3339, p.Since); err != nil {
			return "", fmt.Errorf("%w: invalid 'since' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}

	p.Limit = clampLimit(p.Limit, defaultSearchLimit, maxSearchLimit)

	raw, err := core.RunJXA(ctx, scriptSearchMessages, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	var result listResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse search response: %v", err)
	}

	return string(raw), nil
}

// --- compose ---

type composeParams struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	CC      []string `json:"cc"`
	BCC     []string `json:"bcc"`
	Send    bool     `json:"send"`
}

func toolCompose() core.Tool {
	return core.Tool{
		Name: "apple_mail_compose",
		Description: `Compose a new email message.

Creates a new email with the specified recipients and content. By default, the message is opened as a draft for review. Set send=true to send immediately.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "to": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Array of recipient email addresses (required, at least one)."
    },
    "subject": {
      "type": "string",
      "description": "Email subject line (required)."
    },
    "body": {
      "type": "string",
      "description": "Email body text."
    },
    "cc": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Array of CC recipient email addresses."
    },
    "bcc": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Array of BCC recipient email addresses."
    },
    "send": {
      "type": "boolean",
      "description": "If true, send the email immediately. If false (default), open as a draft for review."
    }
  },
  "required": ["to", "subject"],
  "additionalProperties": false
}`),
		Handler: handleCompose,
	}
}

func handleCompose(ctx context.Context, input json.RawMessage) (string, error) {
	var p composeParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	if len(p.To) == 0 {
		return "", fmt.Errorf("%w: 'to' is required and must contain at least one email address", core.ErrInvalidInput)
	}

	// Validate all To addresses are non-empty
	for i, addr := range p.To {
		if strings.TrimSpace(addr) == "" {
			return "", fmt.Errorf("%w: 'to[%d]' must not be empty", core.ErrInvalidInput, i)
		}
	}

	p.Subject = strings.TrimSpace(p.Subject)
	if p.Subject == "" {
		return "", fmt.Errorf("%w: 'subject' is required and must not be empty", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptCompose, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse compose response: invalid JSON")
	}

	return string(raw), nil
}

// --- reply ---

type replyParams struct {
	MessageID string `json:"messageId"`
	Body      string `json:"body"`
	ReplyAll  bool   `json:"replyAll"`
	Send      bool   `json:"send"`
}

func toolReply() core.Tool {
	return core.Tool{
		Name: "apple_mail_reply",
		Description: `Reply to an email message.

Creates a reply to the specified message. Set replyAll=true to reply to all recipients. By default, the reply is opened as a draft for review.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "messageId": {
      "type": "string",
      "description": "The RFC Message-ID of the email to reply to."
    },
    "body": {
      "type": "string",
      "description": "Reply body text (required)."
    },
    "replyAll": {
      "type": "boolean",
      "description": "If true, reply to all recipients. Defaults to false."
    },
    "send": {
      "type": "boolean",
      "description": "If true, send the reply immediately. If false (default), open as a draft for review."
    }
  },
  "required": ["messageId", "body"],
  "additionalProperties": false
}`),
		Handler: handleReply,
	}
}

func handleReply(ctx context.Context, input json.RawMessage) (string, error) {
	var p replyParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.MessageID = strings.TrimSpace(p.MessageID)
	if p.MessageID == "" {
		return "", fmt.Errorf("%w: 'messageId' is required and must not be empty", core.ErrInvalidInput)
	}

	p.Body = strings.TrimSpace(p.Body)
	if p.Body == "" {
		return "", fmt.Errorf("%w: 'body' is required and must not be empty", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptReply, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse reply response: invalid JSON")
	}

	return string(raw), nil
}

// --- move message ---

type moveParams struct {
	MessageID          string `json:"messageId"`
	DestinationMailbox string `json:"destinationMailbox"`
}

func toolMoveMessage() core.Tool {
	return core.Tool{
		Name: "apple_mail_move",
		Description: `Move an email message to a different mailbox.

Moves the specified message to the destination mailbox. Use mailbox paths from apple_mail_mailboxes (e.g., "iCloud/Archive", "Gmail/Trash").`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "messageId": {
      "type": "string",
      "description": "The RFC Message-ID of the email to move."
    },
    "destinationMailbox": {
      "type": "string",
      "description": "Destination mailbox path (e.g., 'iCloud/Archive'). Use apple_mail_mailboxes to find available paths."
    }
  },
  "required": ["messageId", "destinationMailbox"],
  "additionalProperties": false
}`),
		Handler: handleMoveMessage,
	}
}

func handleMoveMessage(ctx context.Context, input json.RawMessage) (string, error) {
	var p moveParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.MessageID = strings.TrimSpace(p.MessageID)
	if p.MessageID == "" {
		return "", fmt.Errorf("%w: 'messageId' is required and must not be empty", core.ErrInvalidInput)
	}

	p.DestinationMailbox = strings.TrimSpace(p.DestinationMailbox)
	if p.DestinationMailbox == "" {
		return "", fmt.Errorf("%w: 'destinationMailbox' is required and must not be empty", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptMoveMessage, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse move response: invalid JSON")
	}

	return string(raw), nil
}

// --- set status ---

type setStatusParams struct {
	MessageID string `json:"messageId"`
	IsRead    *bool  `json:"isRead"`
	IsFlagged *bool  `json:"isFlagged"`
	IsJunk    *bool  `json:"isJunk"`
}

func toolSetStatus() core.Tool {
	return core.Tool{
		Name: "apple_mail_set_status",
		Description: `Update the status flags of an email message.

Modify read, flagged, or junk status of a message. Only specified fields are changed; omitted fields are left unchanged. At least one status field must be provided.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "messageId": {
      "type": "string",
      "description": "The RFC Message-ID of the email to update."
    },
    "isRead": {
      "type": "boolean",
      "description": "Set read/unread status. Omit to leave unchanged."
    },
    "isFlagged": {
      "type": "boolean",
      "description": "Set flagged/unflagged status. Omit to leave unchanged."
    },
    "isJunk": {
      "type": "boolean",
      "description": "Set junk/not-junk status. Omit to leave unchanged."
    }
  },
  "required": ["messageId"],
  "additionalProperties": false
}`),
		Handler: handleSetStatus,
	}
}

func handleSetStatus(ctx context.Context, input json.RawMessage) (string, error) {
	var p setStatusParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.MessageID = strings.TrimSpace(p.MessageID)
	if p.MessageID == "" {
		return "", fmt.Errorf("%w: 'messageId' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.IsRead == nil && p.IsFlagged == nil && p.IsJunk == nil {
		return "", fmt.Errorf("%w: at least one of 'isRead', 'isFlagged', or 'isJunk' must be specified", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptSetStatus, p)
	if err != nil {
		return "", classifyMailError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse set status response: invalid JSON")
	}

	return string(raw), nil
}

// --- helpers ---

// clampLimit constrains limit to [1, max], using defaultVal when 0 or negative.
func clampLimit(limit, defaultVal, max int) int {
	return core.ClampLimit(limit, defaultVal, max)
}

// classifyMailError wraps errors with Mail-specific context.
func classifyMailError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Mail", "Automation")
	}
	return err
}
