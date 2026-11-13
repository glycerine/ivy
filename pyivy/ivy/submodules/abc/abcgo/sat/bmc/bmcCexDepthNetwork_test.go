package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCexInducedNetwork(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(0, 2, 2)
	c.IFrame = 1
	c.PData[0] = 9
	for start := int32(0); start < 2; start++ {
		q := Bmc_CexBuildNetwork2(g, c, start)
		if gia.Gia_ManCiNum(q) != 1 || gia.Gia_ManCoNum(q) != 1 {
			t.Fatal("wrong interface")
		}
		for _, x := range []bool{false, true} {
			out, err := q.Evaluate([]bool{x})
			want := start == 0 || x
			if err != nil || out[0] != want {
				t.Fatalf("start=%d x=%v out=%v err=%v", start, x, out, err)
			}
		}
	}
	q := Bmc_CexBuildNetwork2Test(g, c, 1)
	if q == nil || gia.Gia_ManRegNum(q) != 1 || gia.Gia_ManPoNum(q) != 2 {
		t.Fatal("failed to append cone")
	}
}
