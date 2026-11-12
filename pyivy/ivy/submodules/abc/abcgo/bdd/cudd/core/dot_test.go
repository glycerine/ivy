package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteDOTBDDAndADD(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	bdd := Cudd_bddAnd(dd, Cudd_Not(dd.Vars[0]), dd.Vars[1])
	for _, root := range []*DdNode{bdd, Cudd_BddToAdd(dd, bdd)} {
		var out bytes.Buffer
		if err := WriteDOT(&out, dd, []*DdNode{root}, []string{"input\"a", "b"}, []string{"out"}); err != nil {
			t.Fatal(err)
		}
		for _, expected := range []string{"digraph DD", `input\"a`, `label="out"`, "out0 ->"} {
			if !strings.Contains(out.String(), expected) {
				t.Fatalf("missing %q in %s", expected, out.String())
			}
		}
	}
	other := NewBDDManager(1)
	defer Cudd_Quit(other)
	if err := WriteDOT(&bytes.Buffer{}, dd, []*DdNode{other.Vars[0]}, nil, nil); err == nil {
		t.Fatal("foreign root was accepted")
	}
}
