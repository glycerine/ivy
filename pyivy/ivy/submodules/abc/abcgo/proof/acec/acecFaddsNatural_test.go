package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"reflect"
	"testing"
)

func TestNaturalAdderBoxesCollapse(t *testing.T) {
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
	boxed := Gia_ManDupWithNaturalBoxes(p, 2, 0)
	if boxed == nil {
		t.Fatal("missing boxed graph")
	}
	timing, extra := boxed.BoxTiming()
	if timing == nil || extra == nil || tim.Tim_ManBoxNum(timing) != 2 {
		t.Fatal("missing two timing boxes")
	}
	flat := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
	for i := 0; i < 32; i++ {
		input := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0}
		want, e1 := p.Evaluate(input)
		got, e2 := flat.Evaluate(input)
		if e1 != nil || e2 != nil || len(want) != len(got) {
			t.Fatalf("input %v results %v %v errors %v %v", input, want, got, e1, e2)
		}
		for k := range want {
			if want[k] != got[k] {
				t.Fatalf("input %v want %v got %v", input, want, got)
			}
		}
	}
}

// Two dependent full-adder boxes exercise phase propagation through a box
// boundary and the auxiliary AIG's replicated outputs.
func TestNaturalBoxesWithInputPolarities(t *testing.T) {
	for phase := 0; phase < 32; phase++ {
		p := gia.Gia_ManStart(128)
		var in [5]int32
		for i := range in {
			in[i] = gia.Gia_ManAppendCi(p)
			if phase&(1<<i) != 0 {
				in[i] ^= 1
			}
		}
		var lo, hi [2]int32
		Acec_InsertFadd(p, in[:3], lo[:])
		Acec_InsertFadd(p, []int32{lo[1], in[3], in[4]}, hi[:])
		for _, out := range []int32{lo[0], hi[0], hi[1]} {
			gia.Gia_ManAppendCo(p, out)
		}
		boxed := Gia_ManDupWithNaturalBoxes(p, 2, 0)
		if boxed == nil {
			t.Fatalf("phase=%d missing boxed graph", phase)
		}
		timing, extra := boxed.BoxTiming()
		if timing == nil || extra == nil || tim.Tim_ManBoxNum(timing) != 2 {
			t.Fatalf("phase=%d missing boxes", phase)
		}
		flat := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
		for bits := 0; bits < 32; bits++ {
			inputs := make([]bool, 5)
			for i := range inputs {
				inputs[i] = bits&(1<<i) != 0
			}
			want, e1 := p.Evaluate(inputs)
			got, e2 := flat.Evaluate(inputs)
			if e1 != nil || e2 != nil || !reflect.DeepEqual(want, got) {
				t.Fatalf("phase=%d input=%v want=%v got=%v errors=%v/%v", phase, inputs, want, got, e1, e2)
			}
		}
	}
}
