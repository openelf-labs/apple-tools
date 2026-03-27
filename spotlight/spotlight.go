//go:build darwin

package spotlight

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

// Register adds the Spotlight search tool to the registry.
func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_spotlight_search",
		Description: "Search for files on macOS using Spotlight (mdfind). Supports content queries, scoped directory search, and content type filtering (e.g. com.adobe.pdf, public.image).",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Spotlight search query (file name, content, or metadata)"
				},
				"directory": {
					"type": "string",
					"description": "Limit search to this directory path (optional)"
				},
				"contentType": {
					"type": "string",
					"description": "Filter by UTI content type, e.g. com.adobe.pdf, public.image (optional)"
				},
				"limit": {
					"type": "integer",
					"description": "Maximum number of results (default 20)",
					"default": 20
				}
			},
			"required": ["query"]
		}`),
		Handler: handleSearch,
	})
}

func handleSearch(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Query       string `json:"query"`
		Directory   string `json:"directory"`
		ContentType string `json:"contentType"`
		Limit       int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.Query = strings.TrimSpace(params.Query)
	if params.Query == "" {
		return "", fmt.Errorf("%w: query must not be empty", core.ErrInvalidInput)
	}

	params.Directory = strings.TrimSpace(params.Directory)
	if params.Directory != "" && core.ContainsTraversal(params.Directory) {
		return "", fmt.Errorf("%w: directory path must not contain '..'", core.ErrInvalidInput)
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	// Build mdfind arguments.
	var args []string

	if params.Directory != "" {
		args = append(args, "-onlyin", params.Directory)
	}

	// Build the query expression, optionally filtering by content type.
	query := params.Query
	if params.ContentType != "" {
		// Validate contentType to prevent injection into mdfind query.
		if strings.ContainsAny(params.ContentType, `'"()&|!`) {
			return "", fmt.Errorf("%w: contentType contains invalid characters", core.ErrInvalidInput)
		}
		query = fmt.Sprintf("(%s) && (kMDItemContentType == '%s')", params.Query, params.ContentType)
	}
	args = append(args, query)

	out, err := core.RunCommand(ctx, "mdfind", args...)
	if err != nil {
		return "", err
	}

	// Parse and limit results.
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "[]", nil
	}

	lines := strings.Split(raw, "\n")
	if len(lines) > params.Limit {
		lines = lines[:params.Limit]
	}

	return core.MustFormatJSON(lines), nil
}
