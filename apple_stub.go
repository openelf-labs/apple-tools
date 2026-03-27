//go:build !darwin

package appletools

// RegisterAll is a no-op on non-macOS platforms.
// Apple tools are only available on macOS where osascript and JXA are present.
func RegisterAll(_ Registry, _ Config) {}
