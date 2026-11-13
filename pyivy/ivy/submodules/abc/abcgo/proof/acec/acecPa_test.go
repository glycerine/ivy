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

func TestFullAdderPhasePropagation(t *testing.T) {
	p := gia.Gia_ManStart(64)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum := gia.Gia_ManHashXor(p, gia.Gia_ManHashXor(p, a, b), c)
	ab := gia.Gia_ManHashAnd(p, a, b)
	ac := gia.Gia_ManHashAnd(p, a, c)
	bc := gia.Gia_ManHashAnd(p, b, c)
	carry := gia.Gia_ManHashOr(p, gia.Gia_ManHashOr(p, ab, ac), bc)
	// Store regular node IDs and account for complemented output edges in sign bits.
	sign := int32(0)
	if sum&1 != 0 {
		sign |= 1 << 3
	}
	if carry&1 != 0 {
		sign |= 1 << 4
	}
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1, sum >> 1, carry >> 1, sign} {
		vec.Vec_IntPush(adds, v)
	}
	mapping := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	vec.Vec_IntWriteEntry(mapping, int(carry>>1), 0)
	roots := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(roots, carry>>1)
	var constants *Vec_Bit_t
	phases := Pas_ManPhase(p, adds, mapping, roots, &constants)
	for _, id := range []int32{a >> 1, b >> 1, c >> 1, carry >> 1} {
		if got := vec.Vec_BitEntry(phases, int(id)); got != 1 {
			t.Errorf("phase %d = %d, want 1", id, got)
		}
	}
	wantSum := int(1 ^ ((sign >> 3) & 1))
	if got := vec.Vec_BitEntry(phases, int(sum>>1)); got != wantSum {
		t.Errorf("sum phase = %d, want %d", got, wantSum)
	}
}
