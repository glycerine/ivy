package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// C Gia_ManToAig_rec uses Aig_And on the stored fanins for every internal
// object, including native XOR/MUX encodings; it does not use MUX control.
func TestNativeExportsUseSourceAndFaninSemantics(t *testing.T) {
	g := Gia_ManStart(24)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a, b^1)
	m := Gia_ManAppendMuxReal(g, c, b, d^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, m)
	for i, fn := range []func(*Gia_Man_t) *aig.Graph{func(p *Gia_Man_t) *aig.Graph { return Gia_ManToAig(p, 0) }, func(p *Gia_Man_t) *aig.Graph { return Gia_ManToAigSkip(p, 1) }, Gia_ManToAigSimple} {
		out := fn(g)
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			got, err := out.Evaluate(in)
			want := []bool{in[0] && !in[1], in[1] && !in[3]}
			if err != nil || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("mode %d mask %d got %v want %v err %v", i, mask, got, want, err)
			}
		}
	}
}
func TestSimpleExportPreservesInterleavedSourceObjectIDs(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	node := Gia_ManAppendAnd(g, a, b)
	late := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, node)
	Gia_ManAppendCo(g, late)
	out := Gia_ManToAigSimple(g)
	if Gia_ObjValue(Gia_ManObj(g, node>>1)) != node || Gia_ObjValue(Gia_ManObj(g, late>>1)) != late {
		t.Fatalf("source mapping node=%d input=%d", Gia_ObjValue(Gia_ManObj(g, node>>1)), Gia_ObjValue(Gia_ManObj(g, late>>1)))
	}
	if aig.Aig_ManCi(out, 2).NodeID() != int(late>>1) {
		t.Fatal("late CI ID changed")
	}
}
func TestSimpleExportRejectsCollapsedObjectID(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendAnd(g, a, b)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	defer func() {
		if got := recover(); got != "gia: simple export changed object ID" {
			t.Fatalf("panic %v", got)
		}
	}()
	Gia_ManToAigSimple(g)
}
