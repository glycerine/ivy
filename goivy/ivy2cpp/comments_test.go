package ivy2cpp

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var staleCommentSweepFiles = []string{
	"action_gen.go",
	"solver_emit.go",
	"thunk.go",
}

func TestNoStaleDeferralComments(t *testing.T) {
	stale := []string{
		"deferred to milestone 5",
		"TODO 014/026",
		"derived/function filtering is a follow-up",
		"the z3 / gen-mode path (Python lines 538-602) is currently stubbed",
	}
	for _, name := range staleCommentSweepFiles {
		raw := readIvy2CPPSourceForCommentTest(t, name)
		for _, bad := range stale {
			if strings.Contains(raw, bad) {
				t.Fatalf("%s still contains stale comment string %q", name, bad)
			}
		}
	}
}

func TestEveryDeferralCommentReferencesAudit2(t *testing.T) {
	for _, name := range staleCommentSweepFiles {
		name := name
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(".", name)
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse %s: %v", name, err)
			}
			for _, group := range file.Comments {
				text := group.Text()
				if !mentionsDeferralOrTODO(text) {
					continue
				}
				if hasAuditCrossReference(text) {
					continue
				}
				pos := fset.Position(group.Pos())
				t.Fatalf("%s comment mentions deferral/TODO without audit cross-reference:\n%s", pos, text)
			}
		})
	}
}

func readIvy2CPPSourceForCommentTest(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(".", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(raw)
}

func mentionsDeferralOrTODO(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "deferred") ||
		strings.Contains(lower, "deferral") ||
		strings.Contains(text, "TODO")
}

func hasAuditCrossReference(text string) bool {
	return strings.Contains(text, "AUDIT2 Item ") ||
		strings.Contains(text, "AUDIT2 DONE ") ||
		strings.Contains(text, "TODO_AUDIT2026")
}
