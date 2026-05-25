package ivy2go

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// --- M10: hygiene — every TODO/DEFER references AUDIT_TODO_LIVE.md ---
//
// Mirrors ivy2cpp/comments_test.go's discipline (item 044): each
// deferral marker in source must point at a tracked audit item. For
// M10 we accept any reference of the form "M<N>", "M9.1", "M11+",
// or a literal "AUDIT_TODO_LIVE.md" / "deferred". The intent isn't
// to enforce a precise numbering scheme — it's to catch the lazy
// "TODO: fix this later" with no follow-up.

func TestCommentsReferenceAuditOrMilestone(t *testing.T) {
	// Files we audit: every production .go in ivy2go (test files
	// excluded — they're free to spell whatever).
	files := listProductionGoFiles(t)
	if len(files) == 0 {
		t.Fatal("no production .go files found in ivy2go/")
	}

	fset := token.NewFileSet()
	var violations []string
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				text := strings.ToLower(c.Text)
				if !hasMarker(text) {
					continue
				}
				if hasAuditAnchor(text) {
					continue
				}
				violations = append(violations,
					filepath.Base(path)+":"+positionLine(fset, c.Slash)+" — "+strings.TrimSpace(c.Text))
			}
		}
	}
	if len(violations) > 0 {
		t.Fatalf("%d TODO/DEFER comments lack an M<N> / M<N>.<N> / 'deferred'\n"+
			"audit anchor. Add one or document the gap in AUDIT_TODO_LIVE.md.\n%s",
			len(violations), strings.Join(violations, "\n"))
	}
}

// hasMarker returns true when text contains a deferral marker as a
// standalone token. "ARCHITECTURE_TODO.md" / "ivyAssert" / "Defer"
// (as a verb in a sentence) all share substrings with our markers
// but aren't deferrals — only the explicit `TODO(...)`, `TODO:`,
// `FIXME`, `XXX`, or `DEFER(...)` shapes count.
func hasMarker(text string) bool {
	// Note: text is already lowercased by the caller.
	patterns := []string{
		"todo(", "todo:", "todo ",
		"fixme(", "fixme:", "fixme ",
		"xxx(", "xxx:", "xxx ",
		"defer(", "defer:",
	}
	// Strip the leading "//" so we don't confuse comment markers
	// with content.
	stripped := strings.TrimSpace(strings.TrimPrefix(text, "//"))
	for _, p := range patterns {
		if strings.Contains(stripped, p) {
			return true
		}
	}
	return false
}

// hasAuditAnchor returns true when text contains an audit anchor: a
// milestone reference (M<N> or M<N>.<N>), a literal "deferred"
// (past-tense, paired with a destination), or a citation of
// AUDIT_TODO_LIVE.md / ARCHITECTURE_TODO.md / a §-section reference.
func hasAuditAnchor(text string) bool {
	if strings.Contains(text, "audit_todo_live") {
		return true
	}
	if strings.Contains(text, "architecture_todo") {
		return true
	}
	if strings.Contains(text, "deferred") {
		return true
	}
	// "Mn" with n in [0,9] preceded by whitespace / parenthesis /
	// bracket. Avoids matching arbitrary lowercase 'm<digit>' tokens
	// inside identifiers like "memo" or names that happen to embed
	// "m0".
	for i := 0; i+1 < len(text); i++ {
		c := text[i]
		if i > 0 {
			prev := text[i-1]
			if !(prev == ' ' || prev == '(' || prev == '[' || prev == '\t' || prev == ',' || prev == '/') {
				continue
			}
		}
		if c == 'm' && text[i+1] >= '0' && text[i+1] <= '9' {
			return true
		}
	}
	return false
}

func listProductionGoFiles(t *testing.T) []string {
	t.Helper()
	root := "."
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}

func positionLine(fset *token.FileSet, pos token.Pos) string {
	p := fset.Position(pos)
	return intStr(p.Line)
}

func intStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

var _ = ast.Print // keep ast imported until a future audit needs deeper analysis
