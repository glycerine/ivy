package amap

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
)

func TestAmapParseFormulaToHop(t *testing.T) {
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	root, err := Amap_ParseFormulaToHop(nil, "(A + B) * C'", []string{"A", "B", "C"}, manager, "AOI21")
	if err != nil {
		t.Fatal(err)
	}
	truth := hop.Hop_ManConvertAigToTruth(manager, root, 3, nil, 0)[0]
	for assignment := 0; assignment < 8; assignment++ {
		a, b, c := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		want := (a || b) && !c
		if got := truth&(1<<uint(assignment)) != 0; got != want {
			t.Fatalf("assignment %03b: got %v want %v", assignment, got, want)
		}
	}
}

func TestAmapParseFormulaToHopReportsError(t *testing.T) {
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	var output bytes.Buffer
	_, err := Amap_ParseFormulaToHop(&output, "A + missing", []string{"A"}, manager, "bad_gate")
	if err == nil || output.Len() == 0 {
		t.Fatalf("got error %v and diagnostic %q", err, output.String())
	}
	if _, err := Amap_ParseFormulaToHop(nil, "A", []string{"A"}, nil, "bad_manager"); err == nil {
		t.Fatal("nil Hop manager accepted")
	}
}
