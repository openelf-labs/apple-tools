//go:build darwin

package notification

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

//go:embed scripts/*.js
var scripts embed.FS

var scriptNotify = mustLoad("notify.js")

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("notification: embedded script %s not found: %v", name, err))
	}
	return data
}

func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_notification_send",
		Description: "Send a macOS system notification. Returns JSON {ok, title, message}. Useful for alerting the user when a long-running task completes.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"title":{"type":"string","description":"Notification title"},
				"message":{"type":"string","description":"Notification body text"},
				"subtitle":{"type":"string","description":"Optional subtitle"},
				"sound":{"type":"boolean","description":"Play notification sound (default true)"}
			},
			"required":["title","message"]
		}`),
		Handler: handleSend,
	})
}

func handleSend(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Title    string `json:"title"`
		Message  string `json:"message"`
		Subtitle string `json:"subtitle"`
		Sound    *bool  `json:"sound"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.Title = strings.TrimSpace(params.Title)
	params.Message = strings.TrimSpace(params.Message)
	if params.Title == "" {
		return "", fmt.Errorf("%w: title is required", core.ErrInvalidInput)
	}
	if params.Message == "" {
		return "", fmt.Errorf("%w: message is required", core.ErrInvalidInput)
	}

	jxaParams := map[string]any{
		"title":    params.Title,
		"message":  params.Message,
		"subtitle": params.Subtitle,
		"sound":    true,
	}
	if params.Sound != nil {
		jxaParams["sound"] = *params.Sound
	}

	raw, err := core.RunJXA(ctx, scriptNotify, jxaParams)
	if err != nil {
		return "", fmt.Errorf("failed to send notification: %w", err)
	}

	return string(raw), nil
}
