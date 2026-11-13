package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldTernaryTruthTables(t *testing.T) {
	values := []int32{SAIG_ZER, SAIG_ONE, SAIG_UND}
	want := [][]int32{{SAIG_ZER, SAIG_ZER, SAIG_ZER}, {SAIG_ZER, SAIG_ONE, SAIG_UND}, {SAIG_ZER, SAIG_UND, SAIG_UND}}
	for i, a := range values {
		if got, expected := Saig_ManSimInfoNot(a), []int32{SAIG_ONE, SAIG_ZER, SAIG_UND}[i]; got != expected {
			t.Fatalf("not(%d)=%d, want %d", a, got, expected)
		}
		for j, b := range values {
			if got := Saig_ManSimInfoAnd(a, b); got != want[i][j] {
				t.Fatalf("%d and %d = %d", a, b, got)
			}
		}
	}
}

func TestVisitedSimulationEncoding(t *testing.T) {
	for value := int32(0); value < 4; value++ {
		if Saig_ManSimInfo2Not(Saig_ManSimInfo2Not(value)) != value {
			t.Fatal("negation was not involutive", value)
		}
		if old := Saig_ManSimInfo2IsOld(value); (old != 0) != (value >= 2) {
			t.Fatal("visited flag", value, old)
		}
		if value < 2 && Saig_ManSimInfo2SetOld(value) != value+2 {
			t.Fatal("visited conversion", value)
		}
	}
	for a := int32(0); a < 2; a++ {
		for b := int32(0); b < 2; b++ {
			if got := Saig_ManSimInfo2And(a, b); got != a&b {
				t.Fatal("binary and", a, b, got)
			}
		}
	}
}

func TestPackedSimulationValuesAcrossWordBoundary(t *testing.T) {
	info := vec.Vec_PtrAllocSimInfo(2, 2)
	object := aig.Edge(2)
	for frame := int32(0); frame < 32; frame++ {
		Saig_ManSimInfo2Set(info, &object, frame, frame%4)
	}
	for frame := int32(0); frame < 32; frame++ {
		if got := Saig_ManSimInfo2Get(info, &object, frame); got != frame%4 {
			t.Fatalf("frame %d: got %d", frame, got)
		}
	}
	Saig_ManSimInfoSet(info, &object, 15, SAIG_ZER)
	Saig_ManSimInfoSet(info, &object, 16, SAIG_UND)
	if Saig_ManSimInfoGet(info, &object, 15) != SAIG_ZER || Saig_ManSimInfoGet(info, &object, 16) != SAIG_UND || Saig_ManSimInfoGet(info, &object, 17) != 1 {
		t.Fatal("setting adjacent frames corrupted packed values")
	}
}
