package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Python compile_expr_vocab raises when sort_infer_list fails, inside
// ASTContext(expr), so the proof's match f(V)=V+1 reports
// "line 18: error: cannot infer sort of V in V + 1" (test/proving2.ivy)
// instead of continuing to "Match is inconsistent".
func TestCompileExprVocabSortErrorPropagates(t *testing.T) {
	fn := filepath.Join(examplesDir(), "test", "proving2.ivy")
	if _, err := os.Stat(fn); err != nil {
		t.Skip(err)
	}
	_, err := IvyLoadFile(fn)
	if err == nil {
		t.Fatal("no error")
	}
	got := pyErrorString(err)
	if !strings.HasSuffix(got, "proving2.ivy: line 18: error: cannot infer sort of V in V + 1") {
		t.Fatalf("%q", got)
	}
}
