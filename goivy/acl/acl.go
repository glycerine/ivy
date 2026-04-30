// Package acl implements check/nocheck theorem filtering for Ivy.
// This is a port of Python's ivy_acl.py.
//
// It allows users to specify lists of theorem names (or regex patterns)
// that should be skipped during verification or assumed without proof.
package acl

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
)

// Config holds per-session ACL state (ignore/assume lists).
type Config struct {
	mu           sync.RWMutex
	ignores      map[string]bool
	ignoresRegex *regexp.Regexp
	assumes      map[string]bool
	assumesRegex *regexp.Regexp
}

// NewConfig creates a new ACL Config with empty ignore/assume lists.
func NewConfig() *Config {
	return &Config{
		ignores: make(map[string]bool),
		assumes: make(map[string]bool),
	}
}

// RegisterIgnores sets the list of theorem names to skip.
// Names starting with "regex(" are compiled as regular expressions.
func (cfg *Config) RegisterIgnores(ignList []string) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	newIgnores := make(map[string]bool)
	var regexParts []string

	for _, x := range ignList {
		if strings.HasPrefix(x, "regex(") && strings.HasSuffix(x, ")") {
			regexStr := x[6 : len(x)-1]
			regexParts = append(regexParts, regexStr)
		} else {
			newIgnores[x] = true
		}
	}

	cfg.ignores = newIgnores

	if len(regexParts) > 0 {
		// Anchor with ^(?:...) to match Python's re.match() (start-of-string only)
		combined := "^(?:" + strings.Join(regexParts, "|") + ")"
		var err error
		cfg.ignoresRegex, err = regexp.Compile(combined)
		if err != nil {
			return err
		}
	} else {
		cfg.ignoresRegex = nil
	}
	return nil
}

// RegisterAssumes sets the list of theorem names to assume without proof.
func (cfg *Config) RegisterAssumes(assList []string) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	newAssumes := make(map[string]bool)
	var regexParts []string

	for _, x := range assList {
		if strings.HasPrefix(x, "regex(") && strings.HasSuffix(x, ")") {
			regexStr := x[6 : len(x)-1]
			regexParts = append(regexParts, regexStr)
		} else {
			newAssumes[x] = true
		}
	}

	cfg.assumes = newAssumes

	if len(regexParts) > 0 {
		// Anchor with ^(?:...) to match Python's re.match() (start-of-string only)
		combined := "^(?:" + strings.Join(regexParts, "|") + ")"
		var err error
		cfg.assumesRegex, err = regexp.Compile(combined)
		if err != nil {
			return err
		}
	} else {
		cfg.assumesRegex = nil
	}
	return nil
}

// Register sets both ignore and assume lists.
func (cfg *Config) Register(ignList, assList []string) error {
	if err := cfg.RegisterIgnores(ignList); err != nil {
		return err
	}
	return cfg.RegisterAssumes(assList)
}

// IsIgnored returns true if the named theorem should be skipped.
func (cfg *Config) IsIgnored(name string) bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	if cfg.ignores[name] {
		return true
	}
	if cfg.ignoresRegex != nil && cfg.ignoresRegex.MatchString(name) {
		return true
	}
	return false
}

// IsAssumed returns true if the named theorem should be assumed.
func (cfg *Config) IsAssumed(name string) bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	if cfg.assumes[name] {
		return true
	}
	if cfg.assumesRegex != nil && cfg.assumesRegex.MatchString(name) {
		return true
	}
	return false
}

// ShouldSkip returns true if the named theorem should be skipped (alias for IsIgnored).
func (cfg *Config) ShouldSkip(name string) bool {
	return cfg.IsIgnored(name)
}

// ShouldAssume returns true if the named theorem should be assumed (alias for IsAssumed).
func (cfg *Config) ShouldAssume(name string) bool {
	return cfg.IsAssumed(name)
}

// GetIgnoresList returns the current set of ignored names.
func (cfg *Config) GetIgnoresList() map[string]bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	result := make(map[string]bool, len(cfg.ignores))
	for k, v := range cfg.ignores {
		result[k] = v
	}
	return result
}

// GetAssumesList returns the current set of assumed names.
func (cfg *Config) GetAssumesList() map[string]bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	result := make(map[string]bool, len(cfg.assumes))
	for k, v := range cfg.assumes {
		result[k] = v
	}
	return result
}

// RegisterFromFile loads ACL rules from a YAML file.
// The file should have optional keys "ignores" and "assumes", each a list of
// strings. Matches Python ivy_acl.register_from_file (ivy_acl.py:55-71).
func RegisterFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("acl: reading %s: %w", path, err)
	}
	var doc struct {
		Ignores []string `yaml:"ignores"`
		Assumes []string `yaml:"assumes"`
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("acl: parsing %s: %w", path, err)
		}
	}
	cfg := NewConfig()
	if err := cfg.Register(doc.Ignores, doc.Assumes); err != nil {
		return nil, fmt.Errorf("acl: %s: %w", path, err)
	}
	return cfg, nil
}

// Clear resets all ACL state.
func (cfg *Config) Clear() {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.ignores = make(map[string]bool)
	cfg.assumes = make(map[string]bool)
	cfg.ignoresRegex = nil
	cfg.assumesRegex = nil
}
