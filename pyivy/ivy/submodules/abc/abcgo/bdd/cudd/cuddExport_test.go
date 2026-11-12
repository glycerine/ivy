package cudd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestDumpDotBlifAndFactoredForm(t *testing.T) {
	dd := NewBDDManager(2)
	x, y := core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddAnd(dd, x, Cudd_Not(y))
	roots := []*DdNode{f}
	var dot bytes.Buffer
	if Cudd_DumpDot(dd, roots, []string{"a", "b"}, []string{"out"}, &dot) != 1 {
		t.Fatal("DumpDot failed")
	}
	if !strings.Contains(dot.String(), "digraph DD") || !strings.Contains(dot.String(), "out") {
		t.Fatalf("unexpected DOT output: %s", dot.String())
	}
	var blif bytes.Buffer
	if Cudd_DumpBlif(dd, roots, []string{"a", "b"}, []string{"out"}, "model", &blif, 0) != 1 {
		t.Fatal("DumpBlif failed")
	}
	if !strings.Contains(blif.String(), ".model model") || !strings.Contains(blif.String(), ".outputs out") {
		t.Fatalf("unexpected BLIF output: %s", blif.String())
	}
	var factored bytes.Buffer
	if Cudd_DumpFactoredForm(dd, roots, []string{"a", "b"}, []string{"out"}, &factored) != 1 {
		t.Fatal("DumpFactoredForm failed")
	}
	if !strings.Contains(factored.String(), "out =") {
		t.Fatalf("unexpected factored form: %s", factored.String())
	}
}
