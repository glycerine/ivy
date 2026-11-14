package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestTimingBoxOrder(t *testing.T) {
	p := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box)
	first := Gia_ManAppendCo(p, a)
	last := Gia_ManAppendCo(p, root)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	p.PManTime = timing
	got := vec.Vec_IntArray(Gia_ManOrderWithBoxes(p))
	want := []int32{0, a >> 1, first >> 1, box >> 1, root >> 1, last >> 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("boxed order %v", got)
	}
}

func TestDupUnnormalizeBoxOrderAndLogic(t *testing.T) {
	p := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, box^1)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, root^1)
	timing := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
	p.PManTime = timing
	p.SetIdentity("boxed", "source")
	q := Gia_ManDupUnnormalize(p)
	if q == nil || Gia_ManIsNormalized(q) != 0 {
		t.Fatal("expected unnormalized duplicate")
	}
	if name, spec := q.Identity(); name != "boxed" || spec != "source" {
		t.Fatalf("identity %q %q", name, spec)
	}
	if Gia_ManCiNum(q) != 2 || Gia_ManCoNum(q) != 2 {
		t.Fatal("box interface changed")
	}
	for mask := 0; mask < 4; mask++ {
		before, err := giaEvaluateNodes(p, []bool{mask&1 != 0, mask&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		after, err := giaEvaluateNodes(q, []bool{mask&1 != 0, mask&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		for i := int32(0); i < 2; i++ {
			if before[Gia_ManCo(p, i).id] != after[Gia_ManCo(q, i).id] {
				t.Fatalf("assignment %d output %d", mask, i)
			}
		}
	}
}
