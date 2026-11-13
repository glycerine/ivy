package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDeriveTwoAdderCarryBox(t *testing.T) {
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
	box := Acec_DeriveBox(p, nil, 0, 0, 0)
	if box == nil {
		t.Fatal("no carry box detected")
	}
	if vec.Vec_WecSizeSize(box.vAdds) < 2 {
		t.Fatalf("adder count %d", vec.Vec_WecSizeSize(box.vAdds))
	}
	if vec.Vec_WecSizeSize(box.vLeafLits) == 0 || vec.Vec_WecSizeSize(box.vRootLits) == 0 {
		t.Fatal("box lacks leaves or roots")
	}
	normalized := Acec_InsertBox(box, 1)
	for i := 0; i < 32; i++ {
		input := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0}
		want, e1 := p.Evaluate(input)
		got, e2 := normalized.Evaluate(input)
		if e1 != nil || e2 != nil || len(got) != len(want) {
			t.Fatalf("input %v results %v %v errors %v %v", input, want, got, e1, e2)
		}
		for j := range want {
			if want[j] != got[j] {
				t.Fatalf("input %v want %v got %v", input, want, got)
			}
		}
	}

}
