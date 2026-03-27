package appletools

// Config controls which Apple tool categories are enabled.
// Each field corresponds to a tool category that can be independently toggled.
// Disabled categories are not registered with the host and consume zero tokens.
type Config struct {
	Enabled      bool `json:"enabled"`
	Calendar     bool `json:"calendar"`
	Reminders    bool `json:"reminders"`
	Contacts     bool `json:"contacts"`
	Notes        bool `json:"notes"`
	Mail         bool `json:"mail"`
	Messages     bool `json:"messages"`
	Music        bool `json:"music"`
	Safari       bool `json:"safari"`
	Shortcuts    bool `json:"shortcuts"`
	System       bool `json:"system"`
	Clipboard    bool `json:"clipboard"`
	Notification bool `json:"notification"`
	Finder       bool `json:"finder"`
	Spotlight    bool `json:"spotlight"`
}

// DefaultConfig returns the default configuration.
// Messages is disabled by default because it requires Full Disk Access permission.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		Calendar:     true,
		Reminders:    true,
		Contacts:     true,
		Notes:        true,
		Mail:         true,
		Messages:     false,
		Music:        true,
		Safari:       true,
		Shortcuts:    true,
		System:       true,
		Clipboard:    true,
		Notification: true,
		Finder:       true,
		Spotlight:    true,
	}
}
