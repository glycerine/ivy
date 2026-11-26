package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMypyvyExportsMutableRelationsAndNewState(t *testing.T) {
	const source = `#lang ivy1.8
type item
individual exemplar : item
relation ready(X:item)
after init { ready(X) := false }
action inner(x:item) = { ready(x) := true }
export action step(x:item) = { call inner(x) }
invariant [safe] ready(exemplar) | ~ready(exemplar)
`
	mod := New()
	if err := SourceString("mypyvy.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	mod.Name = filepath.Join(t.TempDir(), "output")
	if err := MypyvyExport(mod); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(mod.Name + ".pyv")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"sort item", "mutable relation ready(item)", "immutable constant exemplar: item",
		"init (forall", "transition ext_c_step(", "modifies ready", "new(ready(", "invariant [safe]",
	} {
		if !strings.Contains(string(data), expected) {
			t.Errorf("mypyvy output is missing %q:\n%s", expected, data)
		}
	}
}
