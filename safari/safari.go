//go:build darwin

package safari

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	"github.com/openelf-labs/apple-tools/core"
)

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("safari: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptListTabs    = mustLoad("list_tabs.js")
	scriptGetPage     = mustLoad("get_page.js")
	scriptBookmarks   = mustLoad("bookmarks.js")
	scriptReadingList = mustLoad("reading_list.js")
)

func Register(r core.Registry) {
	r.Add(core.Tool{
		Name: "safari_tabs",
		Description: "List all open tabs in Safari across all windows. Returns JSON array of {windowIndex, tabIndex, title, url}. Use tabIndex with safari_get_page.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler:     handleListTabs,
	})

	r.Add(core.Tool{
		Name: "safari_get_page",
		Description: "Get the URL, title, and HTML source of a Safari tab. Returns JSON {title, url, source}. Defaults to current tab; use tabIndex from safari_tabs to specify.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"tabIndex":{"type":"integer","description":"Tab index from safari_tabs output. Omit for current tab."}
			}
		}`),
		Handler: handleGetPage,
	})

	r.Add(core.Tool{
		Name: "safari_bookmarks",
		Description: "List Safari bookmarks. Returns JSON array of {title, url, folder}.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"limit":{"type":"integer","description":"Maximum bookmarks to return (default 50)"}
			}
		}`),
		Handler: handleBookmarks,
	})

	r.Add(core.Tool{
		Name: "safari_reading_list",
		Description: "List items in Safari's Reading List. Returns JSON array of {title, url, preview, dateAdded}.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"limit":{"type":"integer","description":"Maximum items to return (default 50)"}
			}
		}`),
		Handler: handleReadingList,
	})
}

func handleListTabs(ctx context.Context, input json.RawMessage) (string, error) {
	raw, err := core.RunJXA(ctx, scriptListTabs, nil)
	if err != nil {
		return "", err
	}
	// JXA already returns JSON array — pass through
	return string(raw), nil
}

func handleGetPage(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		TabIndex *int `json:"tabIndex"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	jxaParams := map[string]any{}
	if params.TabIndex != nil {
		jxaParams["tabIndex"] = *params.TabIndex
	}

	raw, err := core.RunJXA(ctx, scriptGetPage, jxaParams)
	if err != nil {
		return "", err
	}

	// Check for JXA-level error in response
	var page struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &page) == nil && page.Error != "" {
		return "", fmt.Errorf("%w: %s", core.ErrNotFound, page.Error)
	}

	return string(raw), nil
}

func handleBookmarks(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Limit int `json:"limit"`
	}
	_ = json.Unmarshal(input, &params)
	params.Limit = core.ClampLimit(params.Limit, 50, 500)

	raw, err := core.RunJXA(ctx, scriptBookmarks, params)
	if err != nil {
		return "", err
	}

	// Check for JXA-level error
	var items []struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &items) == nil && len(items) > 0 && items[0].Error != "" {
		return "", fmt.Errorf("bookmarks: %s", items[0].Error)
	}

	return string(raw), nil
}

func handleReadingList(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Limit int `json:"limit"`
	}
	_ = json.Unmarshal(input, &params)
	params.Limit = core.ClampLimit(params.Limit, 50, 500)

	raw, err := core.RunJXA(ctx, scriptReadingList, params)
	if err != nil {
		return "", err
	}

	// Check for JXA-level error
	var items []struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &items) == nil && len(items) > 0 && items[0].Error != "" {
		return "", fmt.Errorf("reading list: %s", items[0].Error)
	}

	return string(raw), nil
}
