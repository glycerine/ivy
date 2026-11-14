package aiginterp

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestProofInterpolationCraigConditions(t *testing.T) {
	for _, relation := range []int32{0, 1} {
		on, off := aig.NewGraph(), aig.NewGraph()
		a, b := on.NewInput(), on.NewInput()
		c, d := off.NewInput(), off.NewInput()
		on.AddOutput(on.AND(a, b))
		off.AddOutput(off.OR(c, d).Not())
		result := Aig_ManInter(on, off, relation, 0)
		if result == nil {
			t.Fatal("disjoint nonconstant partitions rejected")
		}
		if result.InputCount() != 2+int(relation) || result.OutputCount() != 1 {
			t.Fatal("incorrect common-variable interface")
		}
		for mask := 0; mask < 1<<result.InputCount(); mask++ {
			values := make([]bool, result.InputCount())
			for i := range values {
				values[i] = mask>>i&1 != 0
			}
			got, err := result.Evaluate(values)
			if err != nil {
				t.Fatal(err)
			}
			x, y := values[len(values)-2], values[len(values)-1]
			partA, partB := x && y, !x && !y
			if relation != 0 {
				partA = values[0] == (x && y)
				partB = values[0] && partB
			}
			if partA && !got[0] || got[0] && partB {
				t.Fatalf("relation=%d mask=%d A=%v I=%v B=%v", relation, mask, partA, got, partB)
			}
		}
	}
}

func TestInconsistentPartitionReturnsNilAsSource(t *testing.T) {
	on, off := aig.NewGraph(), aig.NewGraph()
	on.AddOutput(on.Const0())
	off.AddOutput(off.Const1())
	if result := Aig_ManInter(on, off, 0, 0); result != nil {
		t.Fatal("source exits during inconsistent partition insertion")
	}
}

func predicate(truth uint32) *aig.Graph {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput()}
	root := g.Const0()
	for bits := 0; bits < 4; bits++ {
		if truth>>bits&1 == 0 {
			continue
		}
		term := g.Const1()
		for i, e := range inputs {
			if bits>>i&1 == 0 {
				e = e.Not()
			}
			term = g.AND(term, e)
		}
		root = g.OR(root, term)
	}
	g.AddOutput(root)
	return g
}

func TestProofInterpolationAllTwoInputPredicatePairs(t *testing.T) {
	for a := uint32(0); a < 16; a++ {
		for b := uint32(0); b < 16; b++ {
			on, off := predicate(a), predicate(b)
			result := Aig_ManInter(on, off, 0, 0)
			if a == 0 || b == 0 || a&b != 0 {
				if result != nil {
					t.Fatalf("A=%x B=%x expected source failure", a, b)
				}
				continue
			}
			if result == nil {
				t.Fatalf("A=%x B=%x missing interpolant", a, b)
			}
			if err := result.Check(); err != nil {
				t.Fatal(err)
			}
			for bits := 0; bits < 4; bits++ {
				values, err := result.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
				if err != nil {
					t.Fatal(err)
				}
				av, bv := a>>bits&1 != 0, b>>bits&1 != 0
				if av && !values[0] || values[0] && bv {
					t.Fatalf("A=%x B=%x bits=%d I=%v", a, b, bits, values)
				}
			}
		}
	}
}

func TestIncrementalSATAllTwoInputOutputPairs(t *testing.T) {
	for a := uint32(0); a < 16; a++ {
		for b := uint32(0); b < 16; b++ {
			on, off := predicate(a), predicate(b)
			original := on.Node(on.Outputs()[0]).Fanins[0]
			on.AddOutput(original.Not())
			on.AddOutput(on.Const1())
			on.AddOutput(on.Const0())
			other := off.Node(off.Outputs()[0]).Fanins[0]
			for i := 0; i < 3; i++ {
				off.AddOutput(other)
			}
			result := Aig_ManInterFast(on, off, 0)
			want := []bool{a&b == 0, (^a&15)&b == 0, b == 0, true}
			if len(result) != len(want) {
				t.Fatalf("A=%x B=%x missing output results %v", a, b, result)
			}
			for i := range want {
				if result[i] != want[i] {
					t.Fatalf("A=%x B=%x output=%d got=%v want=%v", a, b, i, result, want)
				}
			}
			if on.RegisterCount() != 0 || off.RegisterCount() != 0 {
				t.Fatal("CNF derivation changed register metadata")
			}
			if err := on.Check(); err != nil {
				t.Fatal(err)
			}
			if err := off.Check(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
