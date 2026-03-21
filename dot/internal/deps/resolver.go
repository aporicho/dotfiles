package deps

import "fmt"

// Resolve performs a topological sort on the requested modules and their
// transitive dependencies. It returns a slice in install order (dependencies
// first). An error is returned if a dependency cycle is detected.
func Resolve(requested []string, allDeps map[string][]string) ([]string, error) {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int)
	var order []string

	var visit func(name string) error
	visit = func(name string) error {
		switch state[name] {
		case visiting:
			return fmt.Errorf("dependency cycle detected involving %q", name)
		case visited:
			return nil
		}

		state[name] = visiting

		for _, dep := range allDeps[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}

		state[name] = visited
		order = append(order, name)
		return nil
	}

	for _, name := range requested {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// ReverseDeps returns all module names that directly depend on the given module.
func ReverseDeps(name string, allDeps map[string][]string) []string {
	var result []string
	for mod, deps := range allDeps {
		for _, d := range deps {
			if d == name {
				result = append(result, mod)
				break
			}
		}
	}
	return result
}
