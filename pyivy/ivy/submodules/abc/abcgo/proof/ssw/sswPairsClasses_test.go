package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestPairsBuildOrderedTransitiveClasses(t *testing.T) {
	pairs := vec.Vec_IntAlloc(12)
	for _, x := range []int32{3, 5, 1, 2, 2, 5, 0, 7, 7, 8} {
		vec.Vec_IntPush(pairs, x)
	}
	classes := Ssw_TransformPairsIntoTempClasses(pairs, 10)
	if !reflect.DeepEqual(vec.Vec_IntArray(classes[0]), []int32{0, 7, 8}) || !reflect.DeepEqual(vec.Vec_IntArray(classes[1]), []int32{1, 2, 3, 5}) || classes[3] != nil {
		t.Fatal("incorrect classes")
	}
	Ssw_FreeTempClasses(classes, 10)
	if classes[0] != nil || classes[1] != nil {
		t.Fatal("classes not freed")
	}
}
func TestMiterStatusConstantsAndUnknown(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	if Ssw_MiterStatus(g, 0) != 1 {
		t.Fatal("zero miter")
	}
	g.AddOutput(g.Const1())
	if Ssw_MiterStatus(g, 0) != 0 {
		t.Fatal("one miter")
	}
	s := aig.NewGraph()
	a, b := s.NewInput(), s.NewInput()
	gate := s.AND(a, b)
	s.AddOutput(gate)
	s.AddOutput(a)
	aig.Aig_ManSetRegNum(s, 1)
	aig.Aig_ManSetPhase(s)
	if Ssw_MiterStatus(s, 0) != -1 {
		t.Fatal("sequential miter should be undecided")
	}
}

func TestMiterStatusSourceClassificationAndPriority(t *testing.T) {
	t.Run("combinational inverted input is satisfiable", func(t *testing.T) {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x.Not())
		if got := Ssw_MiterStatus(g, 0); got != 0 {
			t.Fatalf("status = %d, want satisfiable", got)
		}
	})
	t.Run("sequential input is undecided", func(t *testing.T) {
		g := aig.NewGraph()
		x, latch := g.NewInput(), g.NewInput()
		g.AddOutput(x)
		g.AddOutput(latch)
		aig.Aig_ManSetRegNum(g, 1)
		aig.Aig_ManSetPhase(g)
		if got := Ssw_MiterStatus(g, 0); got != -1 {
			t.Fatalf("status = %d, want undecided", got)
		}
	})
	t.Run("one output phase proves satisfiability despite unknown", func(t *testing.T) {
		g := aig.NewGraph()
		x, latch := g.NewInput(), g.NewInput()
		g.AddOutput(x)
		g.AddOutput(x.Not())
		g.AddOutput(latch)
		aig.Aig_ManSetRegNum(g, 1)
		aig.Aig_ManSetPhase(g)
		if got := Ssw_MiterStatus(g, 0); got != 0 {
			t.Fatalf("status = %d, want satisfiable", got)
		}
	})
}
