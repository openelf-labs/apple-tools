//go:build darwin

package mcpserver

import (
	"os"
	"testing"
)

func TestConfigFromEnv_Defaults(t *testing.T) {
	// Clear all APPLE_ env vars.
	clearAppleEnv(t)

	cfg := ConfigFromEnv()

	if !cfg.Enabled {
		t.Error("Enabled should be true by default")
	}
	if !cfg.Calendar {
		t.Error("Calendar should be true by default")
	}
	if cfg.Messages {
		t.Error("Messages should be false by default (requires FDA)")
	}
	if !cfg.System {
		t.Error("System should be true by default")
	}
}

func TestConfigFromEnv_DisableCategory(t *testing.T) {
	clearAppleEnv(t)
	t.Setenv("APPLE_SAFARI", "false")
	t.Setenv("APPLE_MUSIC", "0")
	t.Setenv("APPLE_FINDER", "no")
	t.Setenv("APPLE_SPOTLIGHT", "off")

	cfg := ConfigFromEnv()

	if cfg.Safari {
		t.Error("Safari should be disabled with APPLE_SAFARI=false")
	}
	if cfg.Music {
		t.Error("Music should be disabled with APPLE_MUSIC=0")
	}
	if cfg.Finder {
		t.Error("Finder should be disabled with APPLE_FINDER=no")
	}
	if cfg.Spotlight {
		t.Error("Spotlight should be disabled with APPLE_SPOTLIGHT=off")
	}
	// Others should remain default.
	if !cfg.Calendar {
		t.Error("Calendar should still be true")
	}
}

func TestConfigFromEnv_EnableMessages(t *testing.T) {
	clearAppleEnv(t)
	t.Setenv("APPLE_MESSAGES", "true")

	cfg := ConfigFromEnv()

	if !cfg.Messages {
		t.Error("Messages should be enabled with APPLE_MESSAGES=true")
	}
}

func TestConfigFromEnv_DisableAll(t *testing.T) {
	clearAppleEnv(t)
	t.Setenv("APPLE_ENABLED", "false")

	cfg := ConfigFromEnv()

	if cfg.Enabled {
		t.Error("Enabled should be false with APPLE_ENABLED=false")
	}
}

func TestConfigFromEnv_CaseInsensitive(t *testing.T) {
	clearAppleEnv(t)
	t.Setenv("APPLE_CALENDAR", "FALSE")

	cfg := ConfigFromEnv()

	if cfg.Calendar {
		t.Error("Calendar should be disabled with APPLE_CALENDAR=FALSE (case-insensitive)")
	}
}

func TestConfigFromEnv_EmptyValueKeepsDefault(t *testing.T) {
	clearAppleEnv(t)
	t.Setenv("APPLE_CALENDAR", "")

	cfg := ConfigFromEnv()

	if !cfg.Calendar {
		t.Error("Calendar should keep default (true) with empty env var")
	}
}

func TestEnvBool_EdgeCases(t *testing.T) {
	tests := []struct {
		value      string
		defaultVal bool
		want       bool
	}{
		{"", true, true},
		{"", false, false},
		{"false", true, false},
		{"FALSE", true, false},
		{"False", true, false},
		{"0", true, false},
		{"no", true, false},
		{"off", true, false},
		{"true", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"anything", false, true},
	}

	for _, tt := range tests {
		name := "value=" + tt.value
		if tt.value == "" {
			name = "value=<empty>"
		}
		t.Run(name, func(t *testing.T) {
			key := "APPLE_TEST_ENVBOOL"
			if tt.value == "" {
				os.Unsetenv(key)
			} else {
				t.Setenv(key, tt.value)
			}
			got := envBool(key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("envBool(%q, %v) = %v, want %v", tt.value, tt.defaultVal, got, tt.want)
			}
		})
	}
}

// clearAppleEnv removes all APPLE_ environment variables for test isolation.
func clearAppleEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"APPLE_ENABLED", "APPLE_CALENDAR", "APPLE_REMINDERS", "APPLE_CONTACTS",
		"APPLE_NOTES", "APPLE_MAIL", "APPLE_MESSAGES", "APPLE_MUSIC",
		"APPLE_SAFARI", "APPLE_SHORTCUTS", "APPLE_SYSTEM", "APPLE_CLIPBOARD",
		"APPLE_NOTIFICATION", "APPLE_FINDER", "APPLE_SPOTLIGHT",
	} {
		os.Unsetenv(key)
	}
}
