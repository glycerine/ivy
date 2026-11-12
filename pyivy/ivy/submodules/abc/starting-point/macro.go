package startingpoint

import (
	"regexp"
	"strings"
)

type Options struct {
	IncludePaths   []string // Additional local header search directories for macro expansion.
	inlinePrepared bool
	TypeNames      map[string]bool  // Known typedef spellings from an input inventory.
	Defines        map[string]int64 // Explicit platform choices; zero means disabled.
	opaque         []opaqueRegion
	Package        string
}

var functionMacro = regexp.MustCompile(`^\s*#\s*define\s+([A-Za-z_]\w*)\(([^)]*)\)[ \t]*([\s\S]*)$`)
var identifier = regexp.MustCompile(`^[A-Za-z_]\w*$`)

// Function-like macro definitions are documentation, never Go functions.
// Supported invocations have already expanded before ordinary translation.
func macroDraft(t Token) (name, draft, message string, ok bool) {
	text := strings.ReplaceAll(strings.ReplaceAll(t.Text, "\\\r\n", ""), "\\\n", "")
	m := functionMacro.FindStringSubmatch(text)
	if m == nil {
		return
	}
	original := "// Original macro:\n// " + strings.ReplaceAll(t.Text, "\n", "\n// ") + "\n"
	return m[1], original, "Function-like macro retained as a comment; supported visible invocations are expanded and retained calls are reported separately.", true
}
