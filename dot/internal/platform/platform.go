// Package platform provides OS detection and platform-specific helpers.
package platform

import "runtime"

// Current returns the current operating system name (e.g. "darwin", "linux").
func Current() string {
	return runtime.GOOS
}

// MatchesPlatform reports whether the current OS is in the given list.
// A nil or empty list means "all platforms" and always returns true.
func MatchesPlatform(platforms []string) bool {
	if len(platforms) == 0 {
		return true
	}
	cur := Current()
	for _, p := range platforms {
		if p == cur {
			return true
		}
	}
	return false
}

// PackageManager returns the default system package manager for the current OS.
// It returns "brew" on darwin and "apt" on linux.
func PackageManager() string {
	switch Current() {
	case "darwin":
		return "brew"
	case "linux":
		return "apt"
	default:
		return ""
	}
}
