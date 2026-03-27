//go:build darwin

package appletools

import (
	"github.com/openelf-labs/apple-tools/calendar"
	"github.com/openelf-labs/apple-tools/clipboard"
	"github.com/openelf-labs/apple-tools/contacts"
	"github.com/openelf-labs/apple-tools/finder"
	"github.com/openelf-labs/apple-tools/mail"
	"github.com/openelf-labs/apple-tools/messages"
	"github.com/openelf-labs/apple-tools/music"
	"github.com/openelf-labs/apple-tools/notes"
	"github.com/openelf-labs/apple-tools/notification"
	"github.com/openelf-labs/apple-tools/reminders"
	"github.com/openelf-labs/apple-tools/safari"
	"github.com/openelf-labs/apple-tools/shortcuts"
	"github.com/openelf-labs/apple-tools/spotlight"
	"github.com/openelf-labs/apple-tools/system"
)

// RegisterAll registers all enabled Apple tools with the provided registry.
// Tools are conditionally registered based on the Config flags.
// On non-darwin platforms, this is a no-op (see apple_stub.go).
func RegisterAll(r Registry, cfg Config) {
	if !cfg.Enabled {
		return
	}

	if cfg.Calendar {
		calendar.Register(r)
	}
	if cfg.Reminders {
		reminders.Register(r)
	}
	if cfg.Contacts {
		contacts.Register(r)
	}
	if cfg.Notes {
		notes.Register(r)
	}
	if cfg.Mail {
		mail.Register(r)
	}
	if cfg.Messages {
		messages.Register(r)
	}
	if cfg.Music {
		music.Register(r)
	}
	if cfg.Safari {
		safari.Register(r)
	}
	if cfg.Shortcuts {
		shortcuts.Register(r)
	}
	if cfg.System {
		system.Register(r)
	}
	if cfg.Clipboard {
		clipboard.Register(r)
	}
	if cfg.Notification {
		notification.Register(r)
	}
	if cfg.Finder {
		finder.Register(r)
	}
	if cfg.Spotlight {
		spotlight.Register(r)
	}
}
