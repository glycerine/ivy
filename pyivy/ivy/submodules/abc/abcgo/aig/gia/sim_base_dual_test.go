package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDualPolarityInputsAndAnd(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	inputs := vec.Vec_WrdAlloc(4)
	for _, v := range []uint64{0xaa, 0xff, 0xcc, 0x0f} {
		vec.Vec_WrdPush(inputs, v)
	}
	sims := vec.Vec_WrdStart(int(2 * p.nObjs * 2))
	Gia_ManSimPatAssignInputs2(p, 2, sims, inputs)
	Gia_ManSimPatSimAnd2(p, x>>1, Gia_ManObj(p, x>>1), 2, sims)
	data := vec.Vec_WrdArray(sims)
	for w := 0; w < 2; w++ {
		want := vec.Vec_WrdEntry(inputs, w) &^ vec.Vec_WrdEntry(inputs, 2+w)
		if data[int(x)*2+w] != want || data[(int(x)+1)*2+w] != ^want {
			t.Fatalf("word %d incorrect", w)
		}
	}
}
func TestSimulationCombineWords(t *testing.T) {
	base, extra := vec.Vec_WrdAlloc(4), vec.Vec_WrdAlloc(4)
	for _, v := range []uint64{1, 2, 3, 4} {
		vec.Vec_WrdPush(base, v)
	}
	for _, v := range []uint64{5, 6, 7, 8} {
		vec.Vec_WrdPush(extra, v)
	}
	got := vec.Vec_WrdArray(Gia_ManSimCombine(2, base, extra, 1))
	want := []uint64{1, 2, 5, 3, 4, 7}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("combined %v", got)
		}
	}
}
