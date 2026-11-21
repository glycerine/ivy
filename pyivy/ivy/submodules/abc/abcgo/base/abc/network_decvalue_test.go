package abc

import (
	"sync"
	"testing"
)

var decValueSink [6]any

func TestDecNetworkRepeatedLookupUsesNoScratch(t *testing.T) {
	n := NewNetwork("decomposition values")
	d := DecNetwork{n}
	x, y := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	and := n.AndEdges(x, y.Not())
	var a, b any = x, y.Not()
	allocs := testing.AllocsPerRun(100, func() {
		decValueSink[0], _ = d.LookupAndLevel(a, b, false, false)
		decValueSink[1] = d.LookupAnd(b, a)
		decValueSink[2] = d.And(a, b)
		decValueSink[3] = d.Complement(decValueSink[2])
		decValueSink[4] = d.Variable(1)
		decValueSink[5] = d.ConstantOne()
	})
	defer func() { clear(decValueSink[:]) }()
	for i, want := range []Edge{and, and, and, and.Not(), y, n.One()} {
		if got, ok := decValueSink[i].(Edge); !ok || got != want {
			t.Fatalf("value %d = %v, want %v", i, got, want)
		}
	}
	if allocs != 0 {
		t.Fatalf("six repeated decomposition queries allocated %.0f times", allocs)
	}
}

func TestDecompositionValuesSurviveObjectRenumbering(t *testing.T) {
	n := NewNetwork("renumber")
	d := DecNetwork{n}
	hole := n.NewObject(ObjectPI)
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	value := d.And(a, b)
	inv := d.Complement(value)
	oldID := value.(Edge).Object.ID
	if err := n.DeleteObject(hole); err != nil {
		t.Fatal(err)
	}
	Abc_NtkReassignIds(n)
	if value.(Edge).Object.ID == oldID || d.LookupAnd(b, a) != value || d.Complement(inv) != value || d.Variable(0) != a {
		t.Fatal("edge values lost object identity or polarity after renumbering")
	}
	if got, level := d.LookupAndLevel(b, b, true, false); got != n.One().Not() || level != 0 {
		t.Fatal("opposite-phase inputs did not simplify to zero")
	}
}

func TestConcurrentDecompositionValuesRetainBothPhases(t *testing.T) {
	n := NewNetwork("readers")
	d := DecNetwork{n}
	edge := n.NewObject(ObjectPI).Signal()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range 32 {
		wg.Go(func() {
			<-start
			want := edge
			if i%2 != 0 {
				want = want.Not()
			}
			for range 16 {
				value := want.DecValue()
				if value != want || d.Complement(value) != want.Not() {
					t.Error("wrong phased edge")
					return
				}
			}
		})
	}
	close(start)
	wg.Wait()
	var empty Edge
	if empty.DecValue() != empty || d.Complement(empty) != empty.Not() {
		t.Fatal("zero edge phase")
	}
}
