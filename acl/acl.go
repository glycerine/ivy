// Package acl implements check/nocheck theorem filtering for Ivy.
// This is a port of Python's ivy_acl.py.
//
// It allows users to specify lists of theorem names (or regex patterns)
// that should be skipped during verification or assumed without proof.
package acl

import (
	"regexp"
	"strings"
	"sync"
)

var (
	mu            sync.RWMutex
	ignores       map[string]bool
	ignoresRegex  *regexp.Regexp
	assumes       map[string]bool
	assumesRegex  *regexp.Regexp
)

func init() {
	ignores = make(map[string]bool)
	assumes = make(map[string]bool)
}

// RegisterIgnores sets the list of theorem names to skip.
// Names starting with "regex(" are compiled as regular expressions.
func RegisterIgnores(ignList []string) error {
	mu.Lock()
	defer mu.Unlock()

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

	ignores = newIgnores

	if len(regexParts) > 0 {
		combined := strings.Join(regexParts, "|")
		var err error
		ignoresRegex, err = regexp.Compile(combined)
		if err != nil {
			return err
		}
	} else {
		ignoresRegex = nil
	}
	return nil
}

// RegisterAssumes sets the list of theorem names to assume without proof.
func RegisterAssumes(assList []string) error {
	mu.Lock()
	defer mu.Unlock()

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

	assumes = newAssumes

	if len(regexParts) > 0 {
		combined := strings.Join(regexParts, "|")
		var err error
		assumesRegex, err = regexp.Compile(combined)
		if err != nil {
			return err
		}
	} else {
		assumesRegex = nil
	}
	return nil
}

// Register sets both ignore and assume lists.
func Register(ignList, assList []string) error {
	if err := RegisterIgnores(ignList); err != nil {
		return err
	}
	return RegisterAssumes(assList)
}

// IsIgnored returns true if the named theorem should be skipped.
func IsIgnored(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if ignores[name] {
		return true
	}
	if ignoresRegex != nil && ignoresRegex.MatchString(name) {
		return true
	}
	return false
}

// IsAssumed returns true if the named theorem should be assumed.
func IsAssumed(name string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if assumes[name] {
		return true
	}
	if assumesRegex != nil && assumesRegex.MatchString(name) {
		return true
	}
	return false
}

// ShouldSkip returns true if the named theorem should be skipped (alias for IsIgnored).
func ShouldSkip(name string) bool {
	return IsIgnored(name)
}

// ShouldAssume returns true if the named theorem should be assumed (alias for IsAssumed).
func ShouldAssume(name string) bool {
	return IsAssumed(name)
}

// GetIgnoresList returns the current set of ignored names.
func GetIgnoresList() map[string]bool {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]bool, len(ignores))
	for k, v := range ignores {
		result[k] = v
	}
	return result
}

// GetAssumesList returns the current set of assumed names.
func GetAssumesList() map[string]bool {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]bool, len(assumes))
	for k, v := range assumes {
		result[k] = v
	}
	return result
}

// Clear resets all ACL state.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	ignores = make(map[string]bool)
	assumes = make(map[string]bool)
	ignoresRegex = nil
	assumesRegex = nil
}
