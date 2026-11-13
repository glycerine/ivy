package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func TestCec4ChoiceSourcePathPreservesSnapshots(t *testing.T) {
	for _, native := range []bool{false, true} {
		g := gia.Gia_ManStart(24)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var x, y int32
		if native {
			x = gia.Gia_ManAppendXorReal(g, gia.Gia_ManAppendXorReal(g, a, b), c)
			y = gia.Gia_ManAppendXorReal(g, a, gia.Gia_ManAppendXorReal(g, b, c))
		} else {
			x = gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
			y = gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
		}
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, y)
		gia.Gia_ManAppendCo(g, x)
		out := Cec_ComputeChoicesNew(g, 0)
		if out == nil {
			t.Fatalf("native %v: no result", native)
		}
		if aig.Aig_ManChoiceNum(out) == 0 {
			t.Fatalf("native %v: missing choices", native)
		}
		sourceModel, err := core.NewModel(g)
		if err != nil {
			t.Fatal(err)
		}
		phases, err := sourceModel.Values(make([]bool, 3))
		if err != nil {
			t.Fatal(err)
		}
		for bits := 0; bits < 8; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, e1 := g.Evaluate(inputs)
			got, e2 := out.Evaluate(inputs)
			if e1 != nil || e2 != nil || len(got) != 1 || got[0] != want[0] {
				t.Fatalf("native %v bits %d want %v got %v errors %v %v", native, bits, want, got, e1, e2)
			}
			values, err := sourceModel.Values(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for id := int32(1); int(id) < g.ObjectCount(); id++ {
				repr := gia.Gia_ObjRepr(g, id)
				if repr != gia.GIA_VOID && (values[id] != values[repr]) != (phases[id] != phases[repr]) {
					t.Fatalf("native %v bits %d false source class %d -> %d", native, bits, id, repr)
				}
			}
		}
	}
}

func TestCec4ChoiceClassesAreTrueForAllAssignments(t *testing.T) {
	g := gia.Gia_ManStart(40)
	in := make([]int32, 8)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(g)
	}
	x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, in[0], in[1]), in[2])
	y := gia.Gia_ManAppendAnd(g, in[0], gia.Gia_ManAppendAnd(g, in[1], in[2]))
	rare := x
	for _, lit := range in[3:] {
		rare = gia.Gia_ManAppendAnd(g, rare, lit)
	}
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManAppendCo(g, rare)
	out := Cec_ComputeChoicesNew(g, 0)
	if out == nil {
		t.Fatal("choice: no result")
	}
	q := gia.Gia_ManFromAig(out)
	if !q.HasRepresentatives() || gia.Gia_ManCountChoices(q) == 0 {
		t.Fatal("missing candidate choices")
	}
	model, err := core.NewModel(q)
	if err != nil {
		t.Fatal(err)
	}
	phases, err := model.Values(make([]bool, 8))
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 256; bits++ {
		inputs := make([]bool, 8)
		for i := range inputs {
			inputs[i] = bits&(1<<i) != 0
		}
		want, e1 := g.Evaluate(inputs)
		got, e2 := q.Evaluate(inputs)
		if e1 != nil || e2 != nil || len(got) != 1 || got[0] != want[0] {
			t.Fatalf("bits %d want %v got %v errors %v %v", bits, want, got, e1, e2)
		}
		values, err := model.Values(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for id := int32(1); int(id) < q.ObjectCount(); id++ {
			repr := gia.Gia_ObjRepr(q, id)
			if repr != gia.GIA_VOID && (values[id] != values[repr]) != (phases[id] != phases[repr]) {
				t.Fatalf("bits %d false class %d -> %d", bits, id, repr)
			}
		}
	}
}

func TestCec4ChoicesPreserveSequentialSnapshotInterface(t *testing.T) {
	for _, native := range []bool{false, true} {
		g := gia.Gia_ManStart(24)
		a, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var x int32
		if native {
			x = gia.Gia_ManAppendXorReal(g, a, ro)
		} else {
			x = gia.Gia_ManAppendAnd(g, a, ro)
		}
		for i := 0; i < 3; i++ {
			gia.Gia_ManAppendCo(g, x)
		}
		for i := 0; i < 3; i++ {
			gia.Gia_ManAppendCo(g, ro^1)
		}
		gia.Gia_ManSetRegNum(g, 1)
		q := Cec_ComputeChoicesNew(g, 0)
		if q == nil || q.RegisterCount() != 1 {
			t.Fatalf("native %v: invalid register interface", native)
		}
		for bits := 0; bits < 4; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0}
			want, e1 := g.Evaluate(inputs)
			got, e2 := q.Evaluate(inputs)
			if e1 != nil || e2 != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[3] {
				t.Fatalf("native %v bits %d want %v got %v errors %v %v", native, bits, want, got, e1, e2)
			}
		}
	}
}
