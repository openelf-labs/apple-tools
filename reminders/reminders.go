//go:build darwin

// Package reminders provides Apple Reminders tools for listing, searching,
// creating, and completing reminders via JXA automation.
package reminders

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
		panic("reminders: missing embedded script: " + name)
	}
	return data
}

var (
	scriptList     = mustLoad("list.js")
	scriptSearch   = mustLoad("search.js")
	scriptCreate   = mustLoad("create.js")
	scriptComplete = mustLoad("complete.js")
	scriptGetLists = mustLoad("get_lists.js")
)

// clampLimit constrains limit to [1, 200], using defaultVal when 0 or negative.
func clampLimit(limit, defaultVal int) int {
	return core.ClampLimit(limit, defaultVal, 200)
}

// validStatuses are the accepted values for the status parameter.
var validStatuses = map[string]bool{
	"incomplete": true,
	"completed":  true,
	"all":        true,
}

// IsValidStatus reports whether status is an accepted value for the list filter.
func IsValidStatus(status string) bool {
	return validStatuses[status]
}

// Register adds all Reminders tools to the given registry.
func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_reminders_list",
		Description: "List reminders from Apple Reminders. Filters by list, completion status, and due date range.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "list":      {"type": "string", "description": "Filter by reminder list name"},
    "status":    {"type": "string", "enum": ["incomplete", "completed", "all"], "default": "incomplete", "description": "Filter by completion status"},
    "limit":     {"type": "integer", "default": 50, "minimum": 1, "maximum": 200, "description": "Maximum number of reminders to return"},
    "dueAfter":  {"type": "string", "format": "date-time", "description": "Only return reminders due after this ISO 8601 date"},
    "dueBefore": {"type": "string", "format": "date-time", "description": "Only return reminders due before this ISO 8601 date"}
  },
  "additionalProperties": false
}`),
		Handler: handleList,
	})

	r.Add(core.Tool{
		Name:        "apple_reminders_search",
		Description: "Search reminders by title or notes content. Case-insensitive.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Search text to match against reminder title and notes"},
    "limit": {"type": "integer", "default": 20, "minimum": 1, "maximum": 200, "description": "Maximum number of results"}
  },
  "required": ["query"],
  "additionalProperties": false
}`),
		Handler: handleSearch,
	})

	r.Add(core.Tool{
		Name:        "apple_reminders_create",
		Description: "Create a new reminder in Apple Reminders.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "title":    {"type": "string", "description": "Reminder title"},
    "notes":    {"type": "string", "description": "Additional notes"},
    "list":     {"type": "string", "description": "Target reminder list name (uses default list if omitted)"},
    "dueDate":  {"type": "string", "format": "date-time", "description": "Due date in ISO 8601 format"},
    "priority": {"type": "integer", "minimum": 1, "maximum": 9, "description": "Priority (1=highest, 9=lowest)"}
  },
  "required": ["title"],
  "additionalProperties": false
}`),
		Handler: handleCreate,
	})

	r.Add(core.Tool{
		Name:        "apple_reminders_complete",
		Description: "Mark a reminder as completed. Accepts reminder ID or exact title.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "Reminder ID or exact title"}
  },
  "required": ["id"],
  "additionalProperties": false
}`),
		Handler: handleComplete,
	})

	r.Add(core.Tool{
		Name:        "apple_reminders_lists",
		Description: "Get all reminder lists from Apple Reminders.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleGetLists,
	})
}

// --- list ---

type listParams struct {
	List      string `json:"list"`
	Status    string `json:"status"`
	Limit     int    `json:"limit"`
	DueAfter  string `json:"dueAfter"`
	DueBefore string `json:"dueBefore"`
}

func handleList(ctx context.Context, input json.RawMessage) (string, error) {
	var p listParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	// Defaults.
	if p.Status == "" {
		p.Status = "incomplete"
	}

	// Validate status.
	if !validStatuses[p.Status] {
		return "", fmt.Errorf("%w: status must be one of: incomplete, completed, all", core.ErrInvalidInput)
	}

	p.Limit = clampLimit(p.Limit, 50)

	out, err := core.RunJXA(ctx, scriptList, p)
	if err != nil {
		return "", classifyRemindersError(err)
	}
	return string(out), nil
}

// --- search ---

type searchParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func handleSearch(ctx context.Context, input json.RawMessage) (string, error) {
	var p searchParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Query = strings.TrimSpace(p.Query)
	if p.Query == "" {
		return "", fmt.Errorf("%w: query must not be empty", core.ErrInvalidInput)
	}

	p.Limit = clampLimit(p.Limit, 20)

	out, err := core.RunJXA(ctx, scriptSearch, p)
	if err != nil {
		return "", classifyRemindersError(err)
	}
	return string(out), nil
}

// --- create ---

type createParams struct {
	Title    string `json:"title"`
	Notes    string `json:"notes"`
	List     string `json:"list"`
	DueDate  string `json:"dueDate"`
	Priority int    `json:"priority"`
}

func handleCreate(ctx context.Context, input json.RawMessage) (string, error) {
	var p createParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return "", fmt.Errorf("%w: title must not be empty", core.ErrInvalidInput)
	}

	if p.Priority != 0 && (p.Priority < 1 || p.Priority > 9) {
		return "", fmt.Errorf("%w: priority must be between 1 and 9", core.ErrInvalidInput)
	}

	out, err := core.RunJXA(ctx, scriptCreate, p)
	if err != nil {
		return "", classifyRemindersError(err)
	}
	return string(out), nil
}

// --- complete ---

type completeParams struct {
	ID string `json:"id"`
}

func handleComplete(ctx context.Context, input json.RawMessage) (string, error) {
	var p completeParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.ID = strings.TrimSpace(p.ID)
	if p.ID == "" {
		return "", fmt.Errorf("%w: id must not be empty", core.ErrInvalidInput)
	}

	out, err := core.RunJXA(ctx, scriptComplete, p)
	if err != nil {
		return "", classifyRemindersError(err)
	}
	return string(out), nil
}

// --- get lists ---

func handleGetLists(ctx context.Context, _ json.RawMessage) (string, error) {
	out, err := core.RunJXA(ctx, scriptGetLists, struct{}{})
	if err != nil {
		return "", classifyRemindersError(err)
	}
	return string(out), nil
}

// classifyRemindersError wraps errors with Reminders-specific context.
func classifyRemindersError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Reminders", "Reminders")
	}
	return err
}
