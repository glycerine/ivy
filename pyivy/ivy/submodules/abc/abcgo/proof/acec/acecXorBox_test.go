package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestProduceBoxWithoutXors(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	defer func() {
		if recover() == nil {
			t.Fatal("expected C zero-root assertion")
		}
	}()
	Acec_ProduceBox(p, 0)
}

func TestProduceBoxReconstruction(t *testing.T) {
	p := gia.Gia_ManStart(128)
	var in [5]int32
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	var lo, hi [2]int32
	Acec_InsertFadd(p, in[:3], lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[3], in[4]}, hi[:])
	gia.Gia_ManAppendCo(p, lo[0])
	gia.Gia_ManAppendCo(p, hi[0])
	gia.Gia_ManAppendCo(p, hi[1])
	box := Acec_ProduceBox(p, 0)
	if box == nil {
		t.Fatal("no box")
	}
	normalized := Acec_InsertBox(box, 1)
	for i := 0; i < 32; i++ {
		input := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0}
		want, e1 := p.Evaluate(input)
		got, e2 := normalized.Evaluate(input)
		if e1 != nil || e2 != nil || len(want) != len(got) {
			t.Fatalf("input %v outputs %v %v errors %v %v", input, want, got, e1, e2)
		}
		for j := range want {
			if want[j] != got[j] {
				t.Fatalf("input %v want %v got %v", input, want, got)
			}
		}
	}
}
func TestProduceBoxWithoutAdderTree(t *testing.T) {
	p := gia.Gia_ManStart(64)
	in := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, in, out)
	gia.Gia_ManAppendCo(p, out[0])
	gia.Gia_ManAppendCo(p, out[1])
	if box := Acec_ProduceBox(p, 0); box != nil {
		t.Fatal("a single output rank is not an adder tree")
	}
}
