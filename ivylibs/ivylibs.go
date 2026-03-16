// Package ivylibs manages native library specifications for the Ivy compiler.
// This corresponds to Python's ivy_libs.py.
//
// Libraries are stored as JSON in lib/specs within the Ivy directory.
// Each entry is [name, prefix, optional libdir].
package ivylibs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LibSpec represents a library specification: [name, prefix, optional libdir].
type LibSpec struct {
	Name   string
	Prefix string
	LibDir string // optional
}

// SpecFilePath returns the path to the library specs file.
func SpecFilePath(ivyDir string) string {
	return filepath.Join(ivyDir, "lib", "specs")
}

// LoadSpecs reads the library specs from disk.
func LoadSpecs(ivyDir string) ([]LibSpec, error) {
	path := SpecFilePath(ivyDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var raw [][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("bad format in %s: %w", path, err)
	}
	specs := make([]LibSpec, len(raw))
	for i, r := range raw {
		if len(r) < 2 {
			continue
		}
		specs[i] = LibSpec{Name: r[0], Prefix: r[1]}
		if len(r) >= 3 {
			specs[i].LibDir = r[2]
		}
	}
	return specs, nil
}

// SaveSpecs writes the library specs to disk.
func SaveSpecs(ivyDir string, specs []LibSpec) error {
	path := SpecFilePath(ivyDir)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var raw [][]string
	for _, s := range specs {
		entry := []string{s.Name, s.Prefix}
		if s.LibDir != "" {
			entry = append(entry, s.LibDir)
		}
		raw = append(raw, entry)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AddLib adds a library to the specs (replacing any existing entry with the same name).
func AddLib(ivyDir string, spec LibSpec) error {
	specs, err := LoadSpecs(ivyDir)
	if err != nil {
		return err
	}
	// Remove existing entry with same name
	filtered := make([]LibSpec, 0, len(specs)+1)
	for _, s := range specs {
		if s.Name != spec.Name {
			filtered = append(filtered, s)
		}
	}
	filtered = append(filtered, spec)
	return SaveSpecs(ivyDir, filtered)
}

// RemoveLib removes a library from the specs by name.
func RemoveLib(ivyDir string, name string) error {
	specs, err := LoadSpecs(ivyDir)
	if err != nil {
		return err
	}
	filtered := make([]LibSpec, 0, len(specs))
	for _, s := range specs {
		if s.Name != name {
			filtered = append(filtered, s)
		}
	}
	return SaveSpecs(ivyDir, filtered)
}

// DefaultPrefix returns the default prefix for a library name.
func DefaultPrefix(ivyDir, name string) string {
	return filepath.Join(ivyDir, "lib", name)
}
