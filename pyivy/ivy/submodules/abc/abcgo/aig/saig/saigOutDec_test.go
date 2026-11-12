package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSaigManFindPrimesFindsPairImplyingProperty(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.AND(a, c)
	root := g.AND(x, y)
	g.AddOutput(root)

	primes := Saig_ManFindPrimes(g, 2, 0)
	if got := vec.Vec_PtrSize(primes); got != 1 {
		t.Fatalf("prime count = %d, want one two-literal prime", got)
	}
	cube := vec.Vec_PtrEntry(primes, 0).(*vec.Vec_Int_t)
	if vec.Vec_IntSize(cube) != 2 {
		t.Fatalf("prime cube size = %d, want 2", vec.Vec_IntSize(cube))
	}
	want := map[int32]bool{int32(x.NodeID() << 1): false, int32(y.NodeID() << 1): false}
	for i := 0; i < vec.Vec_IntSize(cube); i++ {
		delete(want, vec.Vec_IntEntry(cube, i))
	}
	if len(want) != 0 {
		t.Fatalf("prime cube = %v, want positive literals for nodes %d and %d", cube, x.NodeID(), y.NodeID())
	}
}

func TestSaigManDecPropertyOutputPreservesFunction(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.AND(a, c)
	g.AddOutput(g.AND(x, y))
	decomposed := Saig_ManDecPropertyOutput(g, 2, 0)
	if got, want := len(decomposed.Outputs())-int(aig.Aig_ManRegNum(decomposed)), 2; got != want {
		t.Fatalf("PO count = %d, want original property plus one prime", got)
	}
	for mask := 0; mask < 1<<3; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := decomposed.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := inputs[0] && inputs[1] && inputs[2]
		if got[0] != want || got[1] != want {
			t.Fatalf("inputs %v: decomposed outputs %v, want both %v", inputs, got, want)
		}
	}
}
