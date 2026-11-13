package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func sopValue(sop []byte, in []bool) bool {
	for len(sop) > 1 && sop[0] != 0 {
		match := true
		for i, v := range in {
			if sop[i] != '-' && (sop[i] == '1') != v {
				match = false
			}
		}
		if match {
			return sop[len(in)+1] == '1'
		}
		sop = sop[len(in)+3:]
	}
	return false
}
func TestCollapseOneCover(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	for _, canon := range []int32{0, 1} {
		for _, rev := range []int32{0, 1} {
			sop := Bmc_CollapseOneInt2(g, 0, 0, canon, rev, 0, 0)
			if sop == nil {
				t.Fatalf("canon=%d rev=%d no cover", canon, rev)
			}
			for bits := 0; bits < 4; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0}
				if sopValue(vec.Vec_StrArray(sop), in) != (in[0] && in[1]) {
					t.Fatalf("sop=%q input=%v", vec.Vec_StrArray(sop), in)
				}
			}
		}
	}
}
func TestCollapseConstantCover(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 0)
	for _, complement := range []int32{0, 1} {
		sop := Bmc_CollapseOneInt2(g, 0, 0, 0, 0, 0, complement)
		want := " 0\n\x00"
		if complement != 0 {
			want = " 1\n\x00"
		}
		if sop == nil || string(vec.Vec_StrArray(sop)) != want {
			t.Fatalf("got %q want %q", vec.Vec_StrArray(sop), want)
		}
	}
}

func TestCollapseCoverXorAndLimit(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendXor(g, a, b))
	if Bmc_CollapseOneInt2(g, 1, 0, 0, 0, 0, 0) != nil {
		t.Fatal("accepted a cover exceeding the cube limit")
	}
	for _, canon := range []int32{0, 1} {
		sop := Bmc_CollapseOneInt2(g, 0, 0, canon, 0, 0, 0)
		if sop == nil {
			t.Fatal("missing XOR cover")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			if sopValue(vec.Vec_StrArray(sop), in) != (in[0] != in[1]) {
				t.Fatalf("sop=%q input=%v", vec.Vec_StrArray(sop), in)
			}
		}
	}
}
