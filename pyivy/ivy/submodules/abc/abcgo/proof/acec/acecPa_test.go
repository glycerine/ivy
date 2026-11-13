package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPhasePropagationOnPrimaryInput(t *testing.T) {
	p := gia.Gia_ManStart(16)
	id := gia.Gia_ManAppendCi(p) >> 1
	adds := vec.Vec_IntAlloc(0)
	mapping := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	roots := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(roots, id)
	var constants *Vec_Bit_t
	phases := Pas_ManPhase(p, adds, mapping, roots, &constants)
	if vec.Vec_BitEntry(phases, int(id)) != 1 {
		t.Fatal("root phase not set")
	}
	if vec.Vec_IntSize(adds) != 0 {
		t.Fatal("modified adders")
	}
}
func TestPhaseTruthEvaluation(t *testing.T) {
	p := gia.Gia_ManStart(16)
	gia.Gia_ManHashStart(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	out := gia.Gia_ManHashAnd(p, a^1, b)
	gia.Gia_ManIncrementTravId(p)
	x, y := gia.Gia_ManObj(p, a>>1), gia.Gia_ManObj(p, b>>1)
	x.Value = 0xaa
	y.Value = 0xcc
	gia.Gia_ObjSetTravIdCurrent(p, x)
	gia.Gia_ObjSetTravIdCurrent(p, y)
	if got := Pas_ManVerifyPhaseOne_rec(p, gia.Gia_ManObj(p, out>>1)); got != 0x44 {
		t.Fatalf("truth = %x, want 44", got)
	}
}
