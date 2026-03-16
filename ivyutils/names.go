package ivyutils

import (
	"fmt"
	"os"
	"strings"
)

// ComposeCharacter is the separator used in qualified names ("." by default).
var ComposeCharacter = "."

// ComposeNames joins names with ComposeCharacter, skipping "this".
func ComposeNames(names ...string) string {
	if len(names) > 0 && names[0] == "this" {
		names = names[1:]
	}
	return strings.Join(names, ComposeCharacter)
}

// SplitName splits a qualified name into components.
// e.g., "a.b[1].c" -> ["a", "b[1]", "c"]
func SplitName(name string) []string {
	if strings.HasPrefix(name, "\"") {
		return []string{name}
	}
	var result []string
	start := 0
	i := 0
	for i < len(name) {
		if string(name[i]) == ComposeCharacter {
			result = append(result, name[start:i])
			start = i + len(ComposeCharacter)
			i = start
		} else if name[i] == '[' {
			i = skipSubscript(name, i+1) + 1
		} else {
			i++
		}
	}
	result = append(result, name[start:i])
	return result
}

func skipSubscript(name string, pos int) int {
	for pos < len(name) && name[pos] != ']' {
		if string(name[pos]) == ComposeCharacter {
			pos += len(ComposeCharacter)
		} else if name[pos] == '[' {
			pos = skipSubscript(name, pos+1) + 1
		} else {
			pos++
		}
	}
	return pos
}

// BaseName returns the first component of a split name.
func BaseName(name string) string {
	parts := SplitName(name)
	if len(parts) == 0 {
		return name
	}
	return parts[0]
}

// ParentChildName splits name into [parent, child].
// Returns ["this", name] if no separator found.
func ParentChildName(name string) [2]string {
	idx := strings.LastIndex(name, ComposeCharacter)
	if idx >= 0 {
		return [2]string{name[:idx], name[idx+len(ComposeCharacter):]}
	}
	return [2]string{"this", name}
}

// ExtractParametersName extracts subscripts from name.
// e.g., "f[1][2]" -> ("f", ["1", "2"])
func ExtractParametersName(name string) (string, []string) {
	var parms []string
	pos := len(name) - 1
	for pos >= 0 && name[pos] == ']' {
		end := pos
		pos--
		count := 1
		for pos >= 0 && count > 0 {
			if name[pos] == '[' {
				count--
			} else if name[pos] == ']' {
				count++
			}
			pos--
		}
		if pos >= 0 {
			parms = append(parms, name[pos+2:end])
		}
	}
	if pos >= 0 {
		// Reverse parms
		for i, j := 0, len(parms)-1; i < j; i, j = i+1, j-1 {
			parms[i], parms[j] = parms[j], parms[i]
		}
		return name[:pos+1], parms
	}
	return name, nil
}

// AddParamsName adds parameter subscripts to a name.
func AddParamsName(name string, parms []string) string {
	var b strings.Builder
	b.WriteString(name)
	for _, p := range parms {
		b.WriteByte('[')
		b.WriteString(p)
		b.WriteByte(']')
	}
	return b.String()
}

// -----------------------------------------------------------------------
// Language version management.
// Corresponds to Python's ivy_utils.py string version functions.
// -----------------------------------------------------------------------

// ivyLanguageVersion holds the current Ivy language version string.
var ivyLanguageVersion = "1.7"

// GetStringVersion returns the current Ivy language version string.
// Corresponds to Python's get_string_version().
func GetStringVersion() string {
	return ivyLanguageVersion
}

// SetStringVersion sets the current Ivy language version string.
// Also updates ComposeCharacter for version-dependent behavior.
// Corresponds to Python's set_string_version().
func SetStringVersion(version string) {
	ivyLanguageVersion = strings.TrimSpace(version)
	// Version-dependent compose character: pre-1.3 uses "__", 1.3+ uses "."
	parts := strings.SplitN(ivyLanguageVersion, ".", 2)
	if len(parts) == 2 {
		major := 0
		minor := 0
		if n, err := parseIntSafe(parts[0]); err == nil {
			major = n
		}
		if n, err := parseIntSafe(parts[1]); err == nil {
			minor = n
		}
		if major < 1 || (major == 1 && minor < 3) {
			ComposeCharacter = "__"
		} else {
			ComposeCharacter = "."
		}
	}
}

func parseIntSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// -----------------------------------------------------------------------
// Standard include directory.
// Corresponds to Python's get_std_include_dir().
// -----------------------------------------------------------------------

// stdIncludeDir caches the standard include directory path.
var stdIncludeDir string

// GetStdIncludeDir returns the standard Ivy include directory.
// This looks for an 'include' subdirectory relative to the executable
// or the package source.
// Corresponds to Python's get_std_include_dir().
func GetStdIncludeDir() string {
	if stdIncludeDir != "" {
		return stdIncludeDir
	}
	// Try relative to current working directory
	if info, err := os.Stat("include"); err == nil && info.IsDir() {
		stdIncludeDir = "include"
		return stdIncludeDir
	}
	// Fallback to empty
	return ""
}

// SetStdIncludeDir sets the standard include directory.
func SetStdIncludeDir(dir string) {
	stdIncludeDir = dir
}

// Distinct returns true if all elements in the slice are unique.
func Distinct[T comparable](l []T) bool {
	seen := make(map[T]struct{}, len(l))
	for _, v := range l {
		if _, ok := seen[v]; ok {
			return false
		}
		seen[v] = struct{}{}
	}
	return true
}

// PolymorphicSymbols are operator names that can be polymorphic.
var PolymorphicSymbols = map[string]struct{}{
	"<": {}, "<=": {}, ">": {}, ">=": {},
	"+": {}, "*": {}, "-": {}, "/": {},
	"*>":     {},
	"bvand":  {},
	"bvor":   {},
	"bvnot":  {},
	"cast":   {},
	"arrsel": {},
	"arrupd": {},
	"arrcst": {},
}
