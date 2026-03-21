package platform

import (
	"runtime"
	"testing"
)

func TestCurrent(t *testing.T) {
	got := Current()
	if got != runtime.GOOS {
		t.Errorf("Current() = %q, want %q", got, runtime.GOOS)
	}
}

func TestMatchesPlatform_NilPlatforms(t *testing.T) {
	if !MatchesPlatform(nil) {
		t.Error("MatchesPlatform(nil) = false, want true")
	}
}

func TestMatchesPlatform_EmptyPlatforms(t *testing.T) {
	if !MatchesPlatform([]string{}) {
		t.Error("MatchesPlatform(empty) = false, want true")
	}
}

func TestMatchesPlatform_Contains(t *testing.T) {
	platforms := []string{"linux", "darwin"}
	if !MatchesPlatform(platforms) {
		t.Errorf("MatchesPlatform(%v) = false, want true (current: %s)", platforms, Current())
	}
}

func TestMatchesPlatform_NotContains(t *testing.T) {
	platforms := []string{"fakeos"}
	if MatchesPlatform(platforms) {
		t.Errorf("MatchesPlatform(%v) = true, want false (current: %s)", platforms, Current())
	}
}

func TestPackageManager(t *testing.T) {
	got := PackageManager()
	switch runtime.GOOS {
	case "darwin":
		if got != "brew" {
			t.Errorf("PackageManager() = %q, want %q", got, "brew")
		}
	case "linux":
		if got != "apt" {
			t.Errorf("PackageManager() = %q, want %q", got, "apt")
		}
	default:
		t.Logf("PackageManager() = %q on %s (no assertion for this OS)", got, runtime.GOOS)
	}
}
