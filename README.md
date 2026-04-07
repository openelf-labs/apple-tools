# apple-tools

macOS Apple application integration tools for Go. Provides 45 tools across 15 categories that let AI agents interact with native Apple apps via JXA (JavaScript for Automation).

Built for [OpenELF](https://github.com/openelf-labs/openelf), but usable by any Go application.

## Features

- **Zero external dependencies** — uses macOS built-in `osascript` and CLI tools
- **45 tools** across Calendar, Reminders, Contacts, Notes, Mail, Messages, Music, Safari, Shortcuts, Spotlight, Finder, System Info, Clipboard, and Notifications
- **Secure parameter passing** — parameters passed via environment variables, not string concatenation
- **Transparent & auditable** — every JXA script is readable in the `scripts/` directories
- **Platform-safe** — `//go:build darwin` tags ensure zero impact on Linux/Windows builds

## Quick Start

```go
import (
    appletools "github.com/openelf-labs/apple-tools"
    "github.com/openelf-labs/apple-tools/core"
)

// Implement the Registry interface
type myRegistry struct{}
func (r *myRegistry) Add(t core.Tool) {
    // Register tool with your system
}

// Register all tools
reg := &myRegistry{}
appletools.RegisterAll(reg, appletools.DefaultConfig())
```

## Demo CLI

Test tools without any host application:

```bash
go run ./cmd/apple-tools-demo list
go run ./cmd/apple-tools-demo call calendar_list '{"limit":3}'
go run ./cmd/apple-tools-demo call shortcuts_list
go run ./cmd/apple-tools-demo call system_battery
```

## Tool Catalog

| Category | Tools | Description |
|----------|-------|-------------|
| Calendar | 4 | List, search, create events, open in Calendar app |
| Reminders | 5 | List, search, create, complete reminders; list reminder lists |
| Contacts | 3 | Search by name, get details, find by phone number |
| Notes | 3 | List, search, create notes in Apple Notes |
| Mail | 8 | List mailboxes, list/read/search messages, compose, reply, move, set status |
| Messages | 3 | Send iMessage, read history, get unread messages |
| Music | 8 | Now playing, play/pause/next/prev, search & play, volume, playlists |
| Safari | 4 | List tabs, get page content, bookmarks, reading list |
| Shortcuts | 2 | List and run Apple Shortcuts with input/output |
| System | 3 | Battery status, disk usage, network info |
| Clipboard | 2 | Read and write clipboard text |
| Notification | 1 | Send macOS system notifications |
| Finder | 1 | Reveal files in Finder |
| Spotlight | 1 | File search via mdfind |

## Data Access Declaration

All operations are local. No data is sent to any remote server.

| Tool | Reads | Writes | macOS Permission |
|------|-------|--------|-----------------|
| calendar_list | Calendar events | — | Automation (Calendar) |
| calendar_search | Calendar events | — | Automation (Calendar) |
| calendar_create | — | Calendar events | Automation (Calendar) |
| calendar_open | — | — | Automation (Calendar) |
| reminders_list | Reminders | — | Automation (Reminders) |
| reminders_search | Reminders | — | Automation (Reminders) |
| reminders_create | — | Reminders | Automation (Reminders) |
| reminders_complete | — | Reminders | Automation (Reminders) |
| reminders_lists | Reminder lists | — | Automation (Reminders) |
| contacts_search | Contacts | — | Automation (Contacts) |
| contacts_details | Contacts | — | Automation (Contacts) |
| contacts_find_by_phone | Contacts | — | Automation (Contacts) |
| notes_list | Notes content | — | Automation (Notes) |
| notes_search | Notes content | — | Automation (Notes) |
| notes_create | — | Notes content | Automation (Notes) |
| mail_* | Email messages | Email drafts | Automation (Mail) |
| messages_send | — | iMessages | Automation (Messages) |
| messages_read | Message history | — | Full Disk Access |
| messages_unread | Message history | — | Full Disk Access |
| music_* | Music library | Playback state | Automation (Music) |
| safari_tabs | Open tab URLs | — | Automation (Safari) |
| safari_get_page | Page content | — | Automation (Safari) |
| safari_bookmarks | Bookmarks file | — | None (reads plist) |
| safari_reading_list | Reading list | — | None (reads plist) |
| shortcuts_list | Shortcut names | — | None |
| shortcuts_run | — | Depends on shortcut | None (shortcut's own permissions) |
| system_* | System status | — | None |
| clipboard_read | Clipboard text | — | None |
| clipboard_write | — | Clipboard text | None |
| notification_send | — | Notification | None |
| finder_reveal | — | — | None |
| spotlight_search | File metadata | — | None |

## Configuration

Tools can be individually enabled/disabled:

```go
cfg := appletools.DefaultConfig()
cfg.Messages = false  // Disable Messages (requires Full Disk Access)
cfg.Mail = false      // Disable Mail
appletools.RegisterAll(reg, cfg)
```

Disabled tools are not registered — zero token overhead for the LLM.

## Permissions

macOS automatically prompts for permission when a tool first accesses an app. Permissions are granted to the host application (e.g., OpenELF.app), not to individual scripts.

If permission is denied, tools return actionable error messages:
```
permission denied: Calendar requires Automation access.
Grant access in System Settings > Privacy & Security > Automation > Allow control of "Calendar"
```

## Architecture

```
Host App (OpenELF)
    │
    │  RegisterAll(registry, config)
    │
    ├── calendar.Register(r)  ──▶  osascript -l JavaScript  ──▶  Calendar.app
    ├── reminders.Register(r) ──▶  osascript -l JavaScript  ──▶  Reminders.app
    ├── contacts.Register(r)  ──▶  osascript -l JavaScript  ──▶  Contacts.app
    ├── notes.Register(r)     ──▶  osascript -l JavaScript  ──▶  Notes.app
    ├── mail.Register(r)      ──▶  osascript -l JavaScript  ──▶  Mail.app
    ├── messages.Register(r)  ──▶  osascript -l JavaScript  ──▶  Messages.app
    ├── music.Register(r)     ──▶  osascript -l JavaScript  ──▶  Music.app
    ├── safari.Register(r)    ──▶  osascript -l JavaScript  ──▶  Safari.app
    ├── shortcuts.Register(r) ──▶  shortcuts CLI             ──▶  Shortcuts.app
    ├── spotlight.Register(r) ──▶  mdfind CLI                ──▶  Spotlight index
    ├── system.Register(r)    ──▶  pmset/df/networksetup     ──▶  System info
    ├── clipboard.Register(r) ──▶  pbpaste/JXA               ──▶  Pasteboard
    ├── notification.Register(r) ──▶ osascript               ──▶  Notification Center
    └── finder.Register(r)    ──▶  osascript -l JavaScript  ──▶  Finder.app
```

Parameters are passed via `APPLE_TOOLS_PARAMS` environment variable (JSON), never concatenated into scripts.

## License

MIT
