//go:build darwin

// Package music provides Apple Music tools for playback control,
// track information, library search, and playlist management via JXA automation.
package music

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
		panic(fmt.Sprintf("music: embedded script %s not found: %v", name, err))
	}
	return data
}

var (
	scriptNowPlaying    = mustLoad("now_playing.js")
	scriptPlay          = mustLoad("play.js")
	scriptPause         = mustLoad("pause.js")
	scriptNext          = mustLoad("next.js")
	scriptPrevious      = mustLoad("previous.js")
	scriptSearchPlay    = mustLoad("search_play.js")
	scriptVolume        = mustLoad("volume.js")
	scriptListPlaylists = mustLoad("list_playlists.js")
)

// Register adds all music tools to the provided registry.
func Register(r core.Registry) {
	r.Add(toolNowPlaying())
	r.Add(toolPlay())
	r.Add(toolPause())
	r.Add(toolNext())
	r.Add(toolPrevious())
	r.Add(toolSearchPlay())
	r.Add(toolVolume())
	r.Add(toolPlaylists())
}

// --- now playing ---

func toolNowPlaying() core.Tool {
	return core.Tool{
		Name: "music_now_playing",
		Description: `Get information about the currently playing track in Apple Music.

Returns the track name, artist, album, duration (seconds), playback position (seconds), and player state (playing/paused/stopped). Returns a message if Music is not running or no track is playing.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleNowPlaying,
	}
}

func handleNowPlaying(ctx context.Context, _ json.RawMessage) (string, error) {
	raw, err := core.RunJXA(ctx, scriptNowPlaying, struct{}{})
	if err != nil {
		return "", classifyMusicError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse now playing response: invalid JSON")
	}

	return string(raw), nil
}

// --- play ---

func toolPlay() core.Tool {
	return core.Tool{
		Name: "music_play",
		Description: `Start or resume playback in Apple Music.

Resumes playback of the current track, or starts playing if stopped.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleSimpleCommand(scriptPlay),
	}
}

// --- pause ---

func toolPause() core.Tool {
	return core.Tool{
		Name: "music_pause",
		Description: `Pause playback in Apple Music.

Pauses the currently playing track.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleSimpleCommand(scriptPause),
	}
}

// --- next ---

func toolNext() core.Tool {
	return core.Tool{
		Name: "music_next",
		Description: `Skip to the next track in Apple Music.

Advances to the next track in the current queue.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleSimpleCommand(scriptNext),
	}
}

// --- previous ---

func toolPrevious() core.Tool {
	return core.Tool{
		Name: "music_previous",
		Description: `Go to the previous track in Apple Music.

Returns to the previous track in the current queue.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`),
		Handler: handleSimpleCommand(scriptPrevious),
	}
}

// handleSimpleCommand returns a handler for play/pause/next/previous scripts
// that all share the same {success, message} response shape.
func handleSimpleCommand(script []byte) core.Handler {
	return func(ctx context.Context, _ json.RawMessage) (string, error) {
		raw, err := core.RunJXA(ctx, script, struct{}{})
		if err != nil {
			return "", classifyMusicError(err)
		}

		if !json.Valid(raw) {
			return "", fmt.Errorf("failed to parse music response: invalid JSON")
		}

		return string(raw), nil
	}
}

// --- search and play ---

type searchPlayParams struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

func toolSearchPlay() core.Tool {
	return core.Tool{
		Name: "music_search_play",
		Description: `Search Apple Music library and play the first match.

Searches the local Music library by song name, artist, or album (for songs) or by playlist name. Plays the first matching result.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search text to match against songs (name, artist, album) or playlist names."
    },
    "type": {
      "type": "string",
      "enum": ["song", "playlist"],
      "description": "Type of item to search for. Defaults to 'song'."
    }
  },
  "required": ["query"],
  "additionalProperties": false
}`),
		Handler: handleSearchPlay,
	}
}

func handleSearchPlay(ctx context.Context, input json.RawMessage) (string, error) {
	var p searchPlayParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Query = strings.TrimSpace(p.Query)
	if p.Query == "" {
		return "", fmt.Errorf("%w: 'query' is required and must not be empty", core.ErrInvalidInput)
	}

	if p.Type != "" && p.Type != "song" && p.Type != "playlist" {
		return "", fmt.Errorf("%w: 'type' must be 'song' or 'playlist'", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptSearchPlay, p)
	if err != nil {
		return "", classifyMusicError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse search play response: invalid JSON")
	}

	return string(raw), nil
}

// --- volume ---

type volumeParams struct {
	Level *int `json:"level"`
}

func toolVolume() core.Tool {
	return core.Tool{
		Name: "music_volume",
		Description: `Set the volume level in Apple Music.

Sets the sound volume to a value between 0 (mute) and 100 (maximum).`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "level": {
      "type": "integer",
      "description": "Volume level from 0 (mute) to 100 (maximum).",
      "minimum": 0,
      "maximum": 100
    }
  },
  "required": ["level"],
  "additionalProperties": false
}`),
		Handler: handleVolume,
	}
}

func handleVolume(ctx context.Context, input json.RawMessage) (string, error) {
	var p volumeParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	if p.Level == nil {
		return "", fmt.Errorf("%w: 'level' is required", core.ErrInvalidInput)
	}
	if *p.Level < 0 || *p.Level > 100 {
		return "", fmt.Errorf("%w: 'level' must be between 0 and 100", core.ErrInvalidInput)
	}

	raw, err := core.RunJXA(ctx, scriptVolume, p)
	if err != nil {
		return "", classifyMusicError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse volume response: invalid JSON")
	}

	return string(raw), nil
}

// --- playlists ---

type playlistsParams struct {
	Limit int `json:"limit"`
}

func toolPlaylists() core.Tool {
	return core.Tool{
		Name: "music_playlists",
		Description: `List playlists in the Apple Music library.

Returns the names of playlists in the local Music library, up to the specified limit.`,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "limit": {
      "type": "integer",
      "description": "Maximum number of playlists to return (1-100). Defaults to 25."
    }
  },
  "additionalProperties": false
}`),
		Handler: handlePlaylists,
	}
}

func handlePlaylists(ctx context.Context, input json.RawMessage) (string, error) {
	var p playlistsParams
	if err := json.Unmarshal(input, &p); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	p.Limit = clampLimit(p.Limit, 25)

	raw, err := core.RunJXA(ctx, scriptListPlaylists, p)
	if err != nil {
		return "", classifyMusicError(err)
	}

	if !json.Valid(raw) {
		return "", fmt.Errorf("failed to parse playlists response: invalid JSON")
	}

	return string(raw), nil
}

// --- helpers ---

// clampLimit constrains limit to [1, 100], using defaultVal when 0.
func clampLimit(limit, defaultVal int) int {
	return core.ClampLimit(limit, defaultVal, 100)
}

// classifyMusicError wraps errors with Music-specific context.
func classifyMusicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Music", "Automation")
	}
	return err
}
