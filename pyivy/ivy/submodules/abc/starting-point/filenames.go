package startingpoint

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PlanGoFilenames chooses names for a complete destination inventory. Unique
// stems lose the C extension; collisions retain it. Recheck after promotion:
// foo.c + foo.h + foo.c.cpp must not overwrite the latter's initial foo.c.go.
// Paths are relative to the output root; the returned order matches the input.
func PlanGoFilenames(sources []string) ([]string, error) {
	names := make([]string, len(sources))
	promoted := make([]bool, len(sources))
	seen := map[string]bool{}
	dirs := map[string]bool{}
	for i, s := range sources {
		if filepath.IsAbs(s) || s == "." || s == ".." || strings.HasPrefix(filepath.Clean(s), ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("expected relative source filename: %s", s)
		}
		if seen[s] {
			return nil, fmt.Errorf("duplicate source filename: %s", s)
		}
		seen[s] = true
		names[i] = strings.TrimSuffix(s, filepath.Ext(s)) + ".go"
		for dir := filepath.Dir(s); dir != "."; dir = filepath.Dir(dir) {
			dirs[dir] = true
		}
	}
	for {
		counts := map[string]int{}
		for _, name := range names {
			counts[name]++
		}
		changed := false
		for i, name := range names {
			if counts[name] < 2 && !dirs[name] {
				continue
			}
			if !promoted[i] {
				names[i] = sources[i] + ".go"
				promoted[i] = true
				changed = true
			}
		}
		if changed {
			continue
		}
		for _, name := range names {
			if counts[name] > 1 || dirs[name] {
				return nil, fmt.Errorf("cannot allocate distinct Go filename: %s", name)
			}
		}
		return names, nil
	}
}
