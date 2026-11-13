package pdr

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestPdrRegisterMapAndRehash(t *testing.T) {
	g := aig.NewGraph()
	pi, lo0, lo1, lo2 := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(lo1, lo2.Not()))
	g.AddOutput(pi)
	g.AddOutput(pi)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 3)
	classes := Pdr_ManCreateMap(g)
	if got := vec.Vec_IntArray(classes); !reflect.DeepEqual(got, []int32{0, 0, -1}) {
		t.Fatalf("map = %v", got)
	}
	if got := Pdr_ManCountMap(classes); got != 2 {
		t.Fatalf("reduced = %d", got)
	}
	reduced := Pdr_ManRehashWithMap(g, classes)
	if reduced.RegisterCount() != 3 || reduced.InputCount() != 4 || reduced.OutputCount() != 4 {
		t.Fatalf("changed graph interface: inputs=%d outputs=%d regs=%d", reduced.InputCount(), reduced.OutputCount(), reduced.RegisterCount())
	}
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		wantInputs := append([]bool(nil), inputs...)
		wantInputs[2], wantInputs[3] = inputs[1], false
		want, err := g.Evaluate(wantInputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := reduced.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("mask=%04b got=%v want=%v err=%v", mask, got, want, err)
		}
	}
	_ = lo0
}

func TestPdrRegisterMapDistinguishesComplementedDrivers(t *testing.T) {
	g := aig.NewGraph()
	x, r0, r1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x.Not())
	aig.Aig_ManSetRegNum(g, 2)
	if got := vec.Vec_IntArray(Pdr_ManCreateMap(g)); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatalf("map = %v", got)
	}
	_, _ = r0, r1
}

func TestPdrRegisterRehashUsesUpdatedRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	r0, r1, r2 := g.NewInput(), g.NewInput(), g.NewInput()
	for _, signal := range []aig.Edge{r0, r1, r2, r0.Not(), r1, r2} {
		g.AddOutput(signal)
	}
	aig.Aig_ManSetRegNum(g, 3)
	mapping := vec.Vec_IntAllocArrayCopy([]int32{1, 0, 0}, 3)
	reduced := Pdr_ManRehashWithMap(g, mapping)
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := reduced.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		v := inputs[1]
		want := []bool{v, v, v, !v, v, v}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("input=%v got=%v want=%v", inputs, got, want)
		}
	}
}
