package ivyutils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// IvyHavePolymorphism controls whether polymorphic symbol handling is active.
// Set by SetStringVersion: true for language versions > 1.2, false otherwise.
// Corresponds to Python's iu.ivy_have_polymorphism.
var IvyHavePolymorphism = true

// IvyUsePolymorphicMacros controls whether macro expansion is active.
// Set by SetStringVersion: true for language versions > 1.5.
// Corresponds to Python's iu.ivy_use_polymorphic_macros.
var IvyUsePolymorphicMacros = false

// IvyForbidGhostInit controls whether ghost initialization is forbidden.
// Set by SetStringVersion: true for language versions > 1.6.
// Corresponds to Python's iu.ivy_forbid_ghost_init.
var IvyForbidGhostInit = false

// IvyLatestLanguageVersion is the latest supported language version.
// Corresponds to Python's ivy_latest_language_version = '1.7'.
var IvyLatestLanguageVersion = "1.7"

// SymbolCharsParser is a regex matching valid symbol characters (excludes brackets and compose char).
// Updated by SetStringVersion.
// Corresponds to Python's symbol_chars_parser.
var SymbolCharsParser = regexp.MustCompile(`[^\[\]\.]*`)

// GetStringVersion returns the current Ivy language version string.
// Corresponds to Python's get_string_version().
func GetStringVersion() string {
	return ivyLanguageVersion
}

// SetStringVersion sets the current Ivy language version string.
// Also updates ComposeCharacter and version-dependent flags.
// Corresponds to Python's set_string_version(version) in ivy_utils.py lines 567-578.
func SetStringVersion(version string) {
	ivyLanguageVersion = strings.TrimSpace(version)
	nv := GetNumericVersion()
	// Python: ivy_compose_character = ':' if get_numeric_version() <= [1,1] else '.'
	if versionLESlice(nv, []int{1, 1}) {
		ComposeCharacter = ":"
	} else {
		ComposeCharacter = "."
	}
	// Python: symbol_chars_parser = re.compile(r'[^\[\]' + ivy_compose_character + r']*')
	SymbolCharsParser = regexp.MustCompile(`[^\[\]` + regexp.QuoteMeta(ComposeCharacter) + `]*`)
	// Python: ivy_have_polymorphism = not get_numeric_version() <= [1,2]
	IvyHavePolymorphism = !versionLESlice(nv, []int{1, 2})
	// Python: ivy_use_polymorphic_macros = not get_numeric_version() <= [1,5]
	IvyUsePolymorphicMacros = !versionLESlice(nv, []int{1, 5})
	// Python: ivy_forbid_ghost_init = not get_numeric_version() <= [1,6]
	IvyForbidGhostInit = !versionLESlice(nv, []int{1, 6})
}

// versionLESlice compares two numeric version slices using Python's list <= semantics.
func versionLESlice(v1, v2 []int) bool {
	maxLen := len(v1)
	if len(v2) > maxLen {
		maxLen = len(v2)
	}
	for i := 0; i < maxLen; i++ {
		a, b := 0, 0
		if i < len(v1) {
			a = v1[i]
		}
		if i < len(v2) {
			b = v2[i]
		}
		if a < b {
			return true
		}
		if a > b {
			return false
		}
	}
	return true // equal
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

// incDirPat matches version directory names like "1.7", "1.5".
// Corresponds to Python's inc_dir_pat = re.compile(r'[0-9]*\.[0-9]*')
var incDirPat = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)

// GetStdIncludeDir returns the standard Ivy include directory by scanning
// for the smallest version subdirectory >= the current language version.
// Corresponds to Python's get_std_include_dir() in ivy_utils.py lines 594-604.
func GetStdIncludeDir() string {
	if stdIncludeDir != "" {
		return stdIncludeDir
	}
	incBaseDir := getIncludeBaseDir()

	entries, err := os.ReadDir(incBaseDir)
	if err != nil {
		// Fallback: if the base dir doesn't exist, try plain "include"
		if info, statErr := os.Stat("include"); statErr == nil && info.IsDir() {
			stdIncludeDir = "include"
			return stdIncludeDir
		}
		return ""
	}

	var bestDir string
	for _, entry := range entries {
		d := entry.Name()
		if !entry.IsDir() {
			continue
		}
		if !incDirPat.MatchString(d) {
			continue
		}
		// Python: version_le(ivy_language_version, d) — current version <= directory version
		if !VersionLE(ivyLanguageVersion, d) {
			continue
		}
		// Pick smallest qualifying version
		if bestDir == "" || VersionLE(d, bestDir) {
			bestDir = d
		}
	}
	if bestDir == "" {
		// Python: raise IvyError(None, 'cannot find standard library for language version ...')
		panic(NewIvyError(nil, fmt.Sprintf(
			"cannot find standard library for language version %s", ivyLanguageVersion)))
	}
	stdIncludeDir = filepath.Join(incBaseDir, bestDir)
	return stdIncludeDir
}

// getIncludeBaseDir returns the base directory containing version subdirectories.
// Tries: executable dir + "/include", then CWD + "/include".
func getIncludeBaseDir() string {
	// Try relative to executable (like Python's os.path.dirname(os.path.abspath(__file__)))
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "include")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// Try CWD
	if info, err := os.Stat("include"); err == nil && info.IsDir() {
		return "include"
	}
	return "include" // fallback
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

// VersionLE returns true if version a <= version b, comparing
// dot-separated numeric components left to right.
// Corresponds to Python's ivy_utils.version_le(a, b).
func VersionLE(a, b string) bool {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		va, vb := 0, 0
		if i < len(pa) {
			va, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			vb, _ = strconv.Atoi(pb[i])
		}
		if va < vb {
			return true
		}
		if va > vb {
			return false
		}
	}
	return true // equal
}

// StringVersionToNumericVersion converts a version string like "1.7" to []int{1, 7}.
// Corresponds to Python's string_version_to_numeric_version(v).
func StringVersionToNumericVersion(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		result[i], _ = strconv.Atoi(p)
	}
	return result
}

// GetNumericVersion returns the current language version as a numeric slice.
// Corresponds to Python's get_numeric_version().
func GetNumericVersion() []int {
	return StringVersionToNumericVersion(ivyLanguageVersion)
}

// ParseIntSubscripts parses a name like "f[1][2]" into ("f", [1, 2]).
// Corresponds to Python's parse_int_subscripts(name) in ivy_utils.py lines 758-765.
func ParseIntSubscripts(name string) (string, []int, error) {
	things := strings.Split(name, "[")
	thy := things[0]
	things = things[1:]
	for _, t := range things {
		if !strings.HasSuffix(t, "]") {
			return "", nil, fmt.Errorf("bad subscript syntax: %s", name)
		}
	}
	prms := make([]int, len(things))
	for i, t := range things {
		val, err := strconv.Atoi(t[:len(t)-1])
		if err != nil {
			return "", nil, fmt.Errorf("bad subscript syntax: %s", name)
		}
		prms[i] = val
	}
	return thy, prms, nil
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
