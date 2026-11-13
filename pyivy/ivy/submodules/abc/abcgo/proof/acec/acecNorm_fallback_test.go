package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestNormalizeWithoutArithmeticBox(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a^1)
	dst := Acec_Normalize(p, 0, 0)
	if dst == nil {
		t.Fatal("nil normalized graph")
	}
	for _, v := range []bool{false, true} {
		got, err := dst.Evaluate([]bool{v})
		if err != nil || len(got) != 1 || got[0] == v {
			t.Fatalf("input %v output %v error %v", v, got, err)
		}
	}
}
