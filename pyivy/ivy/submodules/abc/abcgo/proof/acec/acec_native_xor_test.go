package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestCommonGraphPreservesNativeXor(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	common := Acec_CommonStart(nil, g)
	Acec_CommonFinish(common)
	if gia.Gia_ManCoNum(common) != 1 {
		t.Fatalf("outputs %d", gia.Gia_ManCoNum(common))
	}
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		got, err := common.Evaluate(input)
		want := input[0] != input[1]
		if gia.Gia_ObjValue(gia.Gia_ManObj(g, x>>1))&1 != 0 {
			want = !want
		}
		if err != nil || len(got) != 1 || got[0] != want {
			t.Fatalf("input=%v got=%v err=%v want=%v", input, got, err, want)
		}
	}
}

func TestArtificialBoxesPreserveNativeXor(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	boxed := Gia_ManDupWithArtificialBoxes(g, 10, 1, 8, 0, 0, 0, 1, 0)
	_, extra := boxed.BoxTiming()
	collapsed := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		got, err := collapsed.Evaluate(input)
		want := input[0] != input[1]
		if err != nil || len(got) != 1 || got[0] != want {
			t.Fatalf("input=%v got=%v err=%v want=%v", input, got, err, want)
		}
	}
}

func TestInsertBoxRecPreservesNativeXorAndMux(t *testing.T) {
	for _, mux := range []bool{false, true} {
		src := gia.Gia_ManStart(16)
		if mux {
			src.EnableMuxes()
		}
		a, b, c := gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src)
		var root int32
		if mux {
			root = gia.Gia_ManAppendMuxReal(src, a, b, c)
		} else {
			root = gia.Gia_ManAppendXorReal(src, a, b)
		}
		gia.Gia_ManAppendCo(src, root)
		dst := gia.Gia_ManStart(16)
		gia.Gia_ManFillValue(src)
		gia.Gia_ManConst0(src).Value = 0
		for i := int32(0); i < gia.Gia_ManCiNum(src); i++ {
			gia.Gia_ManCi(src, i).Value = uint32(gia.Gia_ManAppendCi(dst))
		}
		lit := Acec_InsertBox_rec(dst, src, gia.Gia_ManObj(src, root>>1))
		gia.Gia_ManAppendCo(dst, lit)
		for mask := 0; mask < 8; mask++ {
			input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, e1 := src.Evaluate(input)
			got, e2 := dst.Evaluate(input)
			if e1 != nil || e2 != nil || want[0] != got[0] {
				t.Fatalf("mux=%v input %v want %v got %v errors %v %v", mux, input, want, got, e1, e2)
			}
		}
	}
}

func TestArtificialBoxesPreserveNativeMux(t *testing.T) {
	g := gia.Gia_ManStart(16)
	g.EnableMuxes()
	c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendMuxReal(g, c, a, b)
	gia.Gia_ManAppendCo(g, x)
	boxed := Gia_ManDupWithArtificialBoxes(g, 10, 1, 8, 1, 0, 0, 1, 0)
	timing, extra := boxed.BoxTiming()
	if tim.Tim_ManBoxNum(timing) != 0 {
		t.Fatal("native mux was annotated as an AND carry gate")
	}
	collapsed := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err1 := g.Evaluate(in)
		got, err2 := collapsed.Evaluate(in)
		if err1 != nil || err2 != nil || len(got) != 1 || got[0] != want[0] {
			t.Fatalf("in=%v got=%v want=%v err=%v/%v", in, got, want, err1, err2)
		}
	}
}
