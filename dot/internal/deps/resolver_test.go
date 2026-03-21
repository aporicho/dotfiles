package deps

import (
	"sort"
	"testing"
)

func TestResolve_SimpleChain(t *testing.T) {
	allDeps := map[string][]string{
		"kitty": {"zsh"},
		"zsh":   {},
	}

	order, err := Resolve([]string{"kitty"}, allDeps)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	zshIdx := indexOf(order, "zsh")
	kittyIdx := indexOf(order, "kitty")

	if zshIdx == -1 {
		t.Fatal("zsh not found in result")
	}
	if kittyIdx == -1 {
		t.Fatal("kitty not found in result")
	}
	if zshIdx >= kittyIdx {
		t.Errorf("zsh (index %d) should come before kitty (index %d)", zshIdx, kittyIdx)
	}
}

func TestResolve_CycleDetection(t *testing.T) {
	allDeps := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}

	_, err := Resolve([]string{"a"}, allDeps)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestResolve_NoDeps(t *testing.T) {
	allDeps := map[string][]string{
		"git":  {},
		"curl": {},
	}

	order, err := Resolve([]string{"git", "curl"}, allDeps)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(order) != 2 {
		t.Fatalf("len = %d, want 2", len(order))
	}

	sorted := make([]string, len(order))
	copy(sorted, order)
	sort.Strings(sorted)

	if sorted[0] != "curl" || sorted[1] != "git" {
		t.Errorf("order = %v, want [curl git] (sorted)", sorted)
	}
}

func TestResolve_TransitiveDeps(t *testing.T) {
	allDeps := map[string][]string{
		"a": {},
		"b": {"a"},
		"c": {"b"},
	}

	order, err := Resolve([]string{"c"}, allDeps)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("len = %d, want 3", len(order))
	}

	aIdx := indexOf(order, "a")
	bIdx := indexOf(order, "b")
	cIdx := indexOf(order, "c")

	if aIdx == -1 || bIdx == -1 || cIdx == -1 {
		t.Fatalf("missing element in %v", order)
	}
	if !(aIdx < bIdx && bIdx < cIdx) {
		t.Errorf("order = %v, want a before b before c", order)
	}
}

func TestReverseDeps(t *testing.T) {
	allDeps := map[string][]string{
		"zsh":   {},
		"kitty": {"zsh"},
		"nvim":  {"zsh"},
		"git":   {},
	}

	result := ReverseDeps("zsh", allDeps)
	sort.Strings(result)

	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0] != "kitty" || result[1] != "nvim" {
		t.Errorf("ReverseDeps = %v, want [kitty nvim]", result)
	}
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return -1
}
