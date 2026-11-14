package core

import (
	"bytes"
	"math"
	"testing"
)

func TestZddMiscCountsAndSubtablePrint(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 2))
	if got := Cudd_zddDagSize(f); got <= 0 || int(got) > len(dd.zddUnique) {
		t.Fatalf("ZDD DAG size=%d", got)
	}
	if got := Cudd_zddCountMinterm(dd, f, 3); math.Abs(got-2) > 1e-9 {
		t.Fatalf("minterm count=%g, want 2", got)
	}
	var output bytes.Buffer
	dd.stdout = &output
	Cudd_zddPrintSubtable(dd)
	if output.Len() == 0 || !bytes.Contains(output.Bytes(), []byte("subtable[")) {
		t.Fatal("subtable printer produced no output")
	}
}
