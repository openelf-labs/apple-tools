//go:build darwin

package clipboard

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

var scriptReadImage = mustLoad("read_image.js")

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("clipboard: embedded script %s not found: %v", name, err))
	}
	return data
}

func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_clipboard_read",
		Description: "Read clipboard content. Returns JSON {text, hasImage, imageFormat}.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler:     handleRead,
	})

	r.Add(core.Tool{
		Name:        "apple_clipboard_write",
		Description: "Write text to the system clipboard. Returns JSON {ok, length}.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"text":{"type":"string","description":"Text to write to clipboard"}
			},
			"required":["text"]
		}`),
		Handler: handleWrite,
	})
}

func handleRead(ctx context.Context, input json.RawMessage) (string, error) {
	result := map[string]any{}

	out, err := core.RunCommand(ctx, "pbpaste")
	if err == nil {
		result["text"] = strings.TrimSpace(string(out))
	} else {
		result["text"] = ""
	}

	jxaResult, jxaErr := core.RunJXA(ctx, scriptReadImage, nil)
	if jxaErr == nil {
		var clipInfo struct {
			HasImage    bool   `json:"hasImage"`
			ImageFormat string `json:"imageFormat"`
		}
		if json.Unmarshal(jxaResult, &clipInfo) == nil {
			result["hasImage"] = clipInfo.HasImage
			if clipInfo.HasImage {
				result["imageFormat"] = clipInfo.ImageFormat
			}
		}
	}

	return core.MustFormatJSON(result), nil
}

func handleWrite(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}
	if params.Text == "" {
		return "", fmt.Errorf("%w: text is required", core.ErrInvalidInput)
	}

	script := []byte(`
		ObjC.import("AppKit");
		ObjC.import("Foundation");
		var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
		var params = JSON.parse(env.js);
		var pb = $.NSPasteboard.generalPasteboard;
		pb.clearContents;
		pb.setStringForType($(params.text), $.NSPasteboardTypeString);
		JSON.stringify({ok: true, length: params.text.length});
	`)

	raw, err := core.RunJXA(ctx, script, params)
	if err != nil {
		return "", fmt.Errorf("failed to write clipboard: %w", err)
	}

	return string(raw), nil
}
