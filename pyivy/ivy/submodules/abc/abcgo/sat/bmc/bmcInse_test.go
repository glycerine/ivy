package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestInseKnownAndUnknownState(t *testing.T) {
	for _, hold := range []bool{false, true} {
		p := gia.Gia_ManStart(8)
		pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		gia.Gia_ManAppendCo(p, pi)
		next := pi
		if hold {
			next = ro
		}
		gia.Gia_ManAppendCo(p, next)
		gia.Gia_ManSetRegNum(p, 1)
		result := Gia_ManInsePerform(p, vec.Vec_IntStart(1), 3, 2, 0)
		x := vec.Vec_IntEntry(result, 0)
		if hold && x != 0 {
			t.Fatalf("hold: got %d", x)
		}
		if !hold && x&4 == 0 {
			t.Fatalf("input-driven register not known: %d", x)
		}
	}
}
func TestInseSimulateSequence(t *testing.T) {
	p := gia.Gia_ManStart(8)
	pi := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManAppendCo(p, pi^1)
	gia.Gia_ManSetRegNum(p, 1)
	initial := vec.Vec_IntStart(1)
	inputs := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	expected := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	result := Gia_ManInseSimulate(p, initial, inputs, expected)
	if got := vec.Vec_IntEntry(result, 0); got != 4 {
		t.Fatalf("got %d", got)
	}
}
