package aiginterp

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestExactCraigInterpolants(t *testing.T) {
	for shape := 0; shape < 4; shape++ {
		on, off := aig.NewGraph(), aig.NewGraph()
		a, b := on.NewInput(), on.NewInput()
		c, d := off.NewInput(), off.NewInput()
		var fa, fb aig.Edge
		switch shape {
		case 0:
			fa = on.AND(a, b)
			fb = off.OR(c, d).Not()
		case 1:
			fa = on.NewXor(a, b)
			fb = off.NewXor(c, d).Not()
		case 2:
			fa = on.Const0()
			fb = off.Const1()
		case 3:
			fa = on.OR(a, b)
			fb = off.OR(c, d).Not()
		}
		on.AddOutput(fa)
		off.AddOutput(fb)
		interpolant := Aig_ManInter(on, off, 0, 0)
		if shape == 2 {
			if interpolant != nil {
				t.Fatal("source returns nil during inconsistent partition insertion")
			}
			continue
		}
		if interpolant == nil {
			t.Fatal("disjoint predicates rejected")
		}
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			av, _ := on.Evaluate(inputs)
			bv, _ := off.Evaluate(inputs)
			iv, err := interpolant.Evaluate(inputs)
			if err != nil || av[0] && !iv[0] || iv[0] && bv[0] {
				t.Fatal(shape, mask, av, bv, iv, err)
			}
		}
	}
}
func TestRelationPartitionAndFirstVariable(t *testing.T) {
	on, off := aig.NewGraph(), aig.NewGraph()
	a, b := on.NewInput(), on.NewInput()
	c, d := off.NewInput(), off.NewInput()
	// Last A output is complemented: the shared extra variable still denotes its
	// REGULAR child, as Sto_ManChangeLastClause extracts literal>>1.
	regular := on.AND(a, b)
	on.AddOutput(on.OR(a, b))
	on.AddOutput(regular.Not())
	off.AddOutput(off.AND(c, d))
	result := Aig_ManInter(on, off, 1, 0)
	if result == nil || result.InputCount() != 3 {
		t.Fatal("missing relation variable")
	}
	for mask := 0; mask < 8; mask++ {
		extra, x, y := mask&1 != 0, mask&2 != 0, mask&4 != 0
		iv, err := result.Evaluate([]bool{extra, x, y})
		if err != nil {
			t.Fatal(err)
		}
		projectedA := (x || y) && (extra == (x && y))
		if projectedA && !iv[0] {
			t.Fatal("interpolant is not implied by partition A", mask, iv)
		}
		assertedB := !extra && (x && y)
		if iv[0] && assertedB {
			t.Fatal("relation intersects B")
		}
	}
}
func TestInterpolationOverlapAndFastPairs(t *testing.T) {
	on, off := aig.NewGraph(), aig.NewGraph()
	a := on.NewInput()
	b := off.NewInput()
	on.AddOutput(a)
	off.AddOutput(b)
	if Aig_ManInter(on, off, 0, 0) != nil {
		t.Fatal("overlap accepted")
	}
	on.AddOutput(a.Not())
	off.AddOutput(b)
	results := Aig_ManInterFast(on, off, 0)
	if len(results) != 2 || results[0] || !results[1] {
		t.Fatal(results)
	}
	if err := on.Check(); err != nil {
		t.Fatal(err)
	}
	if err := off.Check(); err != nil {
		t.Fatal(err)
	}
}
