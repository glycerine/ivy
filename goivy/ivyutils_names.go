package goivy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/fileops"
)

// ComposeNames joins names with cfg.ComposeCharacter, skipping "this".
// All callers should migrate to cfg.ComposeNames(); this is kept
// only for callers inside the ivyutils package that already have cfg.
// External packages should use cfg.ComposeNames() directly.

// SplitName, BaseName, ParentChildName — see config.go method versions.
// External callers use cfg.SplitName(), cfg.BaseName(), cfg.ParentChildName().

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

// SetStringVersionOn sets the language version on the given config.
// Corresponds to Python's set_string_version(version) in ivy_utils.py lines 567-578.
func SetStringVersionOn(cfg *IvyUtilsConfig, version string) {
	cfg.SetStringVersion(version)
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

// incDirPat matches version directory names like "1.7", "1.5".
// Corresponds to Python's inc_dir_pat = re.compile(r'[0-9]*\.[0-9]*')
var incDirPat = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)

// GetStdIncludeDir returns the standard Ivy include directory by scanning
// for the smallest version subdirectory >= the current language version.
// Corresponds to Python's get_std_include_dir() in ivy_utils.py lines 594-604.
func (cfg *IvyUtilsConfig) GetStdIncludeDir() string {
	if cfg.StdIncludeDir != "" {
		return cfg.StdIncludeDir
	}
	incBaseDir := cfg.IncludeBaseDir
	if incBaseDir == "" {
		incBaseDir = getIncludeBaseDir()
	}

	entries, err := fileops.ReadDir(incBaseDir)
	if err != nil {
		// Fallback: if the base dir doesn't exist, try plain "include"
		if info, statErr := fileops.Stat("include"); statErr == nil && info.IsDir {
			cfg.StdIncludeDir = "include"
			return cfg.StdIncludeDir
		}
		return ""
	}

	var bestDir string
	for _, entry := range entries {
		d := entry.Name
		if !entry.IsDir {
			continue
		}
		if !incDirPat.MatchString(d) {
			continue
		}
		// Python: version_le(ivy_language_version, d) — current version <= directory version
		if !VersionLE(cfg.LanguageVersion, d) {
			continue
		}
		// Pick smallest qualifying version
		if bestDir == "" || VersionLE(d, bestDir) {
			bestDir = d
		}
	}
	if bestDir == "" {
		// Python: raise IvyError(None, 'cannot find standard library for language version ...')
		panic(fmt.Sprintf("cannot find standard library for language version %s", cfg.LanguageVersion))
	}
	cfg.StdIncludeDir = filepath.Join(incBaseDir, bestDir)
	return cfg.StdIncludeDir
}

// getIncludeBaseDir returns the base directory containing version subdirectories.
// Search order:
//  1. GOIVY_INCLUDE environment variable (if set)
//  2. Executable dir + "/include"
//  3. Executable dir + "/ivy-lang-examples/ivy/include"
//  4. Source tree: find go module root via runtime.Caller and look for ivy-lang-examples/ivy/include
//  5. CWD + "/include"
//  6. CWD + "/ivy-lang-examples/ivy/include"
//
// Corresponds to Python's os.path.join(os.path.dirname(os.path.abspath(__file__)),'include').
func getIncludeBaseDir() string {
	// 1. Environment variable
	if envDir := os.Getenv("GOIVY_INCLUDE"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			return envDir
		}
	}
	// 2-3. Relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, sub := range []string{"include", "ivy-lang-examples/ivy/include"} {
			dir := filepath.Join(exeDir, sub)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
	}
	// 4. Source tree: walk up from this source file's directory
	//    (for development, when running `go test` or `go run`)
	if srcDir := findSourceIncludeDir(); srcDir != "" {
		return srcDir
	}
	// 5-6. CWD
	for _, sub := range []string{"include", "ivy-lang-examples/ivy/include"} {
		if info, err := os.Stat(sub); err == nil && info.IsDir() {
			return sub
		}
	}
	return "include" // fallback
}

// findSourceIncludeDir looks for ivy-lang-examples/ivy/include relative to
// the Go source tree root (detected via runtime.Caller).
func findSourceIncludeDir() string {
	// Use runtime.Caller to find this source file's location
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// Walk up from the file's directory looking for ivy-lang-examples
	dir := filepath.Dir(thisFile)
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(dir, "ivy-lang-examples", "ivy", "include")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
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

// GetNumericVersionFrom returns the numeric version for a given version string.
func GetNumericVersionFrom(version string) []int {
	return StringVersionToNumericVersion(version)
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
