package gia

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestMiniSupergateDerivation(t *testing.T) {
	p := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	x := miniaig.Mini_AigAnd(p, a, a)
	y := miniaig.Mini_AigAnd(p, x, b)
	miniaig.Mini_AigCreatePo(p, y)
	super := Gia_MiniAigSuperGates(p)
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(super, 0)); !reflect.DeepEqual(got, []int32{0, 0, 1}) {
		t.Fatal(got)
	}
	multi := Gia_MiniAigSuperPreprocess(p, super, 2, 0)
	if multi != 2 {
		t.Fatal(multi)
	}
	g := Gia_MiniAigSuperDeriveGia(super, 2, multi)
	for i := 0; i < 16; i++ {
		in := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0}
		out, err := g.Evaluate(in)
		if err != nil || len(out) != 1 || out[0] != (in[1] && in[2]) {
			t.Fatalf("input %v: %v %v", in, out, err)
		}
	}
}
