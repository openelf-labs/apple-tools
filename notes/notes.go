//go:build darwin

// Package notes provides Apple Notes tools for listing, searching,
// and creating notes via JXA automation.
package notes

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
		panic(fmt.Sprintf("notes: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptList   = mustLoad("list.js")
	scriptSearch = mustLoad("search.js")
	scriptCreate = mustLoad("create.js")
)

// Register adds all notes tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolList())
	r.Add(toolSearch())
	r.Add(toolCreate())
}

// --- list notes ---

type listParams struct {
	Folder string `json:"folder"`
	Limit  int    `json:"limit"`
}

type noteEntry struct {
	Name             string `json:"name"`
	Folder           string `json:"folder"`
	Snippet          string `json:"snippet"`
	CreationDate     string `json:"creationDate"`
	ModificationDate string `json:"modificationDate"`
}

func toolList() core.Tool {
	return core.Tool{
		Name: "notes_list",
		Description: `List notes from Apple Notes.

Returns notes sorted by modification date (most recent first). Use the folder parameter to filter by a specific folder name. Each note includes its name, folder, a snippet of the first 200 characters, and timestamps.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "folder": {
      "type": "string",
      "description": "Filter by folder name (e.g., 'OpenELF'). Omit to list notes from all folders."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of notes to return (1-200). Defaults to 50."
    }
  },
  "additionalProperties": false
}`),
		Handler: handleList,
	}
}

func handleList(ctx context.Context, input json.RawMessage) (string, error) {
	var p listParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Limit = clampLimit(p.Limit, 50)

	raw, err := core.RunJXA(ctx, scriptList, p)
	if err != nil {
		return "", classifyNotesError(err)
	}

	var notes []noteEntry
	if err := json.Unmarshal(raw, &notes); err != nil {
		return "", fmt.Errorf("failed to parse notes response: %v", err)
	}

	return string(raw), nil
}

// --- search notes ---

type searchParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func toolSearch() core.Tool {
	return core.Tool{
		Name: "notes_search",
		Description: `Search notes in Apple Notes by text.

Performs a case-insensitive search across note names and body content in all folders. Returns matching notes sorted by modification date. Omit query to list all notes.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search text to match against note names and body content (case-insensitive)."
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of notes to return (1-200). Defaults to 50."
    }
  },
  "additionalProperties": false
}`),
		Handler: handleSearch,
	}
}

func handleSearch(ctx context.Context, input json.RawMessage) (string, error) {
	var p searchParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Query = strings.TrimSpace(p.Query)

	p.Limit = clampLimit(p.Limit, 50)

	raw, err := core.RunJXA(ctx, scriptSearch, p)
	if err != nil {
		return "", classifyNotesError(err)
	}

	var notes []noteEntry
	if err := json.Unmarshal(raw, &notes); err != nil {
		return "", fmt.Errorf("failed to parse notes response: %v", err)
	}

	return string(raw), nil
}

// --- create note ---

type createParams struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Folder string `json:"folder"`
}

type createResult struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Message string `json:"message"`
}

func toolCreate() core.Tool {
	return core.Tool{
		Name: "notes_create",
		Description: `Create a new note in Apple Notes.

Creates a note with the specified title and body content. If no folder is specified, the note is created in the "OpenELF" folder (created automatically if it doesn't exist).`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "description": "Note title (required)."
    },
    "body": {
      "type": "string",
      "description": "Note body content (required)."
    },
    "folder": {
      "type": "string",
      "description": "Folder name to create the note in. Defaults to 'OpenELF'. Created if it doesn't exist."
    }
  },
  "required": ["title", "body"],
  "additionalProperties": false
}`),
		Handler: handleCreate,
	}
}

func handleCreate(ctx context.Context, input json.RawMessage) (string, error) {
	var p createParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return "", fmt.Errorf("%w: 'title' is required and must not be empty", core.ErrInvalidInput)
	}
	p.Body = strings.TrimSpace(p.Body)
	if p.Body == "" {
		return "", fmt.Errorf("%w: 'body' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.Folder == "" {
		p.Folder = "OpenELF"
	}

	raw, err := core.RunJXA(ctx, scriptCreate, p)
	if err != nil {
		return "", classifyNotesError(err)
	}

	var result createResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse create note response: %v", err)
	}

	if !result.Success {
		return "", fmt.Errorf("failed to create note: %s", result.Message)
	}

	return core.MustFormatJSON(map[string]any{
		"ok":     true,
		"name":   result.Name,
		"folder": result.Folder,
	}), nil
}

// --- helpers ---

// clampLimit constrains limit to [1, 200], using defaultVal when 0.
func clampLimit(limit, defaultVal int) int {
	return core.ClampLimit(limit, defaultVal, 200)
}

// classifyNotesError wraps errors with Notes-specific context.
func classifyNotesError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Notes", "Automation")
	}
	return err
}
