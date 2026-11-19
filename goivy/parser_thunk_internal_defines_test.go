package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python ActionDecl.defines() includes each action's
// iter_internal_defines(): a thunk defines its label and label.run, so
// the parser's define() sees them (test/thunk2.ivy).
func TestActionDeclDefinesThunkInternals(t *testing.T) {
	fn := filepath.Join(examplesDir(), "test", "thunk2.ivy")
	if _, err := os.Stat(fn); err != nil {
		t.Skip(err)
	}
	res, err := ReadModule(fn, false, New().Cfg)
	if err != nil {
		t.Fatal(err)
	}
	if res.defined == nil {
		t.Fatal("no defined map")
	}
	if _, ok := res.defined["bar"]; !ok {
		t.Errorf("thunk label bar not defined: %v", keysOfDefined(res.defined))
	}
	if _, ok := res.defined["bar.run"]; !ok {
		t.Errorf("bar.run not defined: %v", keysOfDefined(res.defined))
	}
}

func keysOfDefined(d map[string][]definedEntry) []string {
	var ks []string
	for k := range d {
		ks = append(ks, k)
	}
	return ks
}
