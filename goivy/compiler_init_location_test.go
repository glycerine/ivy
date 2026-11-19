package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python p_labeledfmla_fmla gives an unlabeled init's LabeledFormula the
// formula's lineno, and ARGSetup.init appends the compiled formula to
// labeled_inits, so check prints its location ("line 7: (no name)",
// test/somemin.ivy), not "(internal)".
func TestUnlabeledInitKeepsLocation(t *testing.T) {
	src := "#lang ivy1.6\n\ntype t\n\nrelation p(X:t)\ninit p(X)\n"
	fn := filepath.Join(t.TempDir(), "init.ivy")
	if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	if len(mod.LabeledInits) != 1 {
		t.Fatalf("inits %d", len(mod.LabeledInits))
	}
	if got := mod.LabeledInits[0].GetLineno().Line; got != 6 {
		t.Fatalf("init line %d, want 6", got)
	}
}
