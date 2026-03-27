//go:build darwin

// Package calendar provides Apple Calendar tools for listing, searching,
// creating, and opening calendar events via JXA automation.
package calendar

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

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("calendar: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptListEvents   = mustLoad("list_events.js")
	scriptSearchEvents = mustLoad("search_events.js")
	scriptCreateEvent  = mustLoad("create_event.js")
	scriptOpenEvent    = mustLoad("open_event.js")
)

// Register adds all calendar tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolListEvents())
	r.Add(toolSearchEvents())
	r.Add(toolCreateEvent())
	r.Add(toolOpenEvent())
}

// --- list events ---

type listParams struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Calendar string `json:"calendar"`
	Limit    int    `json:"limit"`
}

type calendarEvent struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Location  string `json:"location"`
	Notes     string `json:"notes"`
	Calendar  string `json:"calendar"`
	AllDay    bool   `json:"allDay"`
	URL       string `json:"url"`
}

func toolListEvents() core.Tool {
	return core.Tool{
		Name: "apple_calendar_list",
		Description: `List upcoming calendar events within a date range.

Returns events from Apple Calendar sorted by start date. If no date range is specified, returns events for the next 7 days. Use the calendar parameter to filter by a specific calendar name.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "from": {
      "type": "string",
      "description": "Start of date range in ISO 8601 format (e.g., '2025-01-15T00:00:00Z'). Defaults to now."
    },
    "to": {
      "type": "string",
      "description": "End of date range in ISO 8601 format. Defaults to 7 days after 'from'."
    },
    "calendar": {
      "type": "string",
      "description": "Filter by calendar name (e.g., 'Work', 'Personal'). Omit to search all calendars."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of events to return (1-100). Defaults to 20."
    }
  },
  "additionalProperties": false
}`),
		Handler: handleListEvents,
	}
}

func handleListEvents(ctx context.Context, input json.RawMessage) (string, error) {
	var p listParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	if p.From != "" {
		if _, err := time.Parse(time.RFC3339, p.From); err != nil {
			return "", fmt.Errorf("%w: invalid 'from' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}
	if p.To != "" {
		if _, err := time.Parse(time.RFC3339, p.To); err != nil {
			return "", fmt.Errorf("%w: invalid 'to' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}

	p.Limit = clampLimit(p.Limit, 20)

	raw, err := core.RunJXA(ctx, scriptListEvents, p)
	if err != nil {
		return "", classifyCalendarError(err)
	}

	var events []calendarEvent
	if err := json.Unmarshal(raw, &events); err != nil {
		return "", fmt.Errorf("failed to parse calendar response: %v", err)
	}

	return string(raw), nil
}

// --- search events ---

type searchParams struct {
	Query string `json:"query"`
	From  string `json:"from"`
	To    string `json:"to"`
	Limit int    `json:"limit"`
}

func toolSearchEvents() core.Tool {
	return core.Tool{
		Name: "apple_calendar_search",
		Description: `Search calendar events by title.

Performs a case-insensitive search across all calendars. By default searches events from 30 days ago to 90 days in the future. Use from/to parameters to narrow the search window.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search text to match against event titles (case-insensitive)."
    },
    "from": {
      "type": "string",
      "description": "Start of search range in ISO 8601 format. Defaults to 30 days ago."
    },
    "to": {
      "type": "string",
      "description": "End of search range in ISO 8601 format. Defaults to 90 days from now."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of events to return (1-100). Defaults to 10."
    }
  },
  "required": ["query"],
  "additionalProperties": false
}`),
		Handler: handleSearchEvents,
	}
}

func handleSearchEvents(ctx context.Context, input json.RawMessage) (string, error) {
	var p searchParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Query = strings.TrimSpace(p.Query)
	if p.Query == "" {
		return "", fmt.Errorf("%w: 'query' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.From != "" {
		if _, err := time.Parse(time.RFC3339, p.From); err != nil {
			return "", fmt.Errorf("%w: invalid 'from' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}
	if p.To != "" {
		if _, err := time.Parse(time.RFC3339, p.To); err != nil {
			return "", fmt.Errorf("%w: invalid 'to' date: must be ISO 8601 format (e.g., 2025-01-15T00:00:00Z)", core.ErrInvalidInput)
		}
	}

	p.Limit = clampLimit(p.Limit, 10)

	raw, err := core.RunJXA(ctx, scriptSearchEvents, p)
	if err != nil {
		return "", classifyCalendarError(err)
	}

	var events []calendarEvent
	if err := json.Unmarshal(raw, &events); err != nil {
		return "", fmt.Errorf("failed to parse calendar response: %v", err)
	}

	return string(raw), nil
}

// --- create event ---

type createParams struct {
	Title        string `json:"title"`
	Start        string `json:"start"`
	End          string `json:"end"`
	Location     string `json:"location"`
	Notes        string `json:"notes"`
	Calendar     string `json:"calendar"`
	AllDay       bool   `json:"allDay"`
	AlertMinutes *int   `json:"alertMinutes"`
}

type createResult struct {
	Success   bool   `json:"success"`
	EventID   string `json:"eventId"`
	Title     string `json:"title"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func toolCreateEvent() core.Tool {
	return core.Tool{
		Name: "apple_calendar_create",
		Description: `Create a new event in Apple Calendar.

Creates an event with the specified details. If no calendar is specified, the default calendar is used. If no end time is given, the event defaults to 1 hour duration (or next day for all-day events). You can optionally set an alert that fires N minutes before the event.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "description": "Event title (required)."
    },
    "start": {
      "type": "string",
      "description": "Event start time in ISO 8601 format (required, e.g., '2025-01-15T14:00:00Z')."
    },
    "end": {
      "type": "string",
      "description": "Event end time in ISO 8601 format. Defaults to 1 hour after start."
    },
    "location": {
      "type": "string",
      "description": "Event location."
    },
    "notes": {
      "type": "string",
      "description": "Event notes or description."
    },
    "calendar": {
      "type": "string",
      "description": "Calendar name to add event to. Uses default calendar if omitted."
    },
    "allDay": {
      "type": "boolean",
      "description": "Whether this is an all-day event. Defaults to false."
    },
    "alertMinutes": {
      "type": "integer",
      "description": "Minutes before the event to trigger an alert (e.g., 15 for a 15-minute reminder)."
    }
  },
  "required": ["title", "start"],
  "additionalProperties": false
}`),
		Handler: handleCreateEvent,
	}
}

func handleCreateEvent(ctx context.Context, input json.RawMessage) (string, error) {
	var p createParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return "", fmt.Errorf("%w: 'title' is required and must not be empty", core.ErrInvalidInput)
	}
	p.Start = strings.TrimSpace(p.Start)
	if p.Start == "" {
		return "", fmt.Errorf("%w: 'start' is required", core.ErrInvalidInput)
	}
	if _, err := time.Parse(time.RFC3339, p.Start); err != nil {
		return "", fmt.Errorf("%w: invalid 'start' date: must be ISO 8601 format (e.g., 2025-01-15T14:00:00Z)", core.ErrInvalidInput)
	}
	if p.End != "" {
		if _, err := time.Parse(time.RFC3339, p.End); err != nil {
			return "", fmt.Errorf("%w: invalid 'end' date: must be ISO 8601 format (e.g., 2025-01-15T15:00:00Z)", core.ErrInvalidInput)
		}
	}

	raw, err := core.RunJXA(ctx, scriptCreateEvent, p)
	if err != nil {
		return "", classifyCalendarError(err)
	}

	var result createResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse create event response: %v", err)
	}

	return core.MustFormatJSON(result), nil
}

// --- open event ---

type openParams struct {
	EventID string `json:"eventId"`
}

type openResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func toolOpenEvent() core.Tool {
	return core.Tool{
		Name: "apple_calendar_open",
		Description: `Open a specific event in the Apple Calendar app.

Opens Calendar.app and navigates to the date of the specified event. Use the event ID returned by apple_calendar_list or apple_calendar_search.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "eventId": {
      "type": "string",
      "description": "The unique ID of the event to open (from list or search results)."
    }
  },
  "required": ["eventId"],
  "additionalProperties": false
}`),
		Handler: handleOpenEvent,
	}
}

func handleOpenEvent(ctx context.Context, input json.RawMessage) (string, error) {
	var p openParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.EventID = strings.TrimSpace(p.EventID)
	if p.EventID == "" {
		return "", fmt.Errorf("%w: 'eventId' is required and must not be empty", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptOpenEvent, p)
	if err != nil {
		return "", classifyCalendarError(err)
	}

	var result openResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse open event response: %v", err)
	}

	return core.MustFormatJSON(result), nil
}

// --- helpers ---

// clampLimit constrains limit to [1, 100], using defaultVal when 0.
func clampLimit(limit, defaultVal int) int {
	return core.ClampLimit(limit, defaultVal, 100)
}

// classifyCalendarError wraps errors with Calendar-specific context.
func classifyCalendarError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Calendar", "Automation")
	}
	return err
}
