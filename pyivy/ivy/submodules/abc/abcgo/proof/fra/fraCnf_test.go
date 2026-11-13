package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestStructuralCNFTruthTables(t *testing.T) {
	for _, kind := range []string{"and", "xor", "native-xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			g := aig.NewGraph()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			var out aig.Edge
			switch kind {
			case "and":
				out = g.AND(a.Not(), b)
			case "xor":
				out = g.XOR(a.Not(), b)
			case "native-xor":
				out = g.NewXor(a.Not(), b)
			case "mux":
				out = g.AND(g.AND(a, b).Not(), g.AND(a.Not(), c).Not())
			}
			p := &Fra_Man_t{pManFraig: g, pSat: bsat.NewSolver(), nSatVars: 1}
			Fra_ManClean(p, int32(len(g.Nodes())))
			if !p.pSat.AddClause(bsat.Positive(0)) {
				t.Fatal("constant setup")
			}
			Fra_CnfNodeAddToSolver(p, aig.Aig_Regular(out), aig.InvalidEdge)
			for mask := 0; mask < 8; mask++ {
				av, bv, cv := mask&1 != 0, mask&2 != 0, mask&4 != 0
				expected := false
				switch kind {
				case "and":
					expected = !av && bv
				case "xor", "native-xor":
					expected = !av != bv
				case "mux":
					expected = !((av && bv) || (!av && cv))
				}
				assumptions := []bsat.Literal{fraLit(Fra_ObjSatNum(p, a), !av), fraLit(Fra_ObjSatNum(p, b), !bv)}
				if kind == "mux" {
					assumptions = append(assumptions, fraLit(Fra_ObjSatNum(p, c), !cv))
				}
				result := p.pSat.Solve(append(assumptions, fraLit(Fra_ObjSatNum(p, out), !expected != out.Complemented()))...)
				if result != bsat.Satisfiable {
					t.Fatalf("inputs %03b, correct result %v: %v", mask, expected, result)
				}
				result = p.pSat.Solve(append(assumptions, fraLit(Fra_ObjSatNum(p, out), expected != out.Complemented()))...)
				if result != bsat.Unsatisfiable {
					t.Fatalf("inputs %03b, wrong result %v: %v", mask, !expected, result)
				}
			}
		})
	}
}

func TestStructuralCNFSupergateAndSharedNode(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c, d := graph.NewInput(), graph.NewInput(), graph.NewInput(), graph.NewInput()
	shared := graph.AND(a, b)
	root := graph.AND(graph.AND(shared, c.Not()), d)
	graph.AddOutput(shared) // Give the shared node a second reference.
	p := &Fra_Man_t{pManFraig: graph, pSat: bsat.NewSolver(), nSatVars: 1}
	Fra_ManClean(p, int32(len(graph.Nodes())))
	if !p.pSat.AddClause(bsat.Positive(0)) {
		t.Fatal("constant setup")
	}
	Fra_CnfNodeAddToSolver(p, root, aig.InvalidEdge)
	if Fra_ObjFaninVec(p, shared) == nil {
		t.Fatal("shared node was flattened")
	}
	for mask := 0; mask < 16; mask++ {
		var assumptions []bsat.Literal
		for i, edge := range []aig.Edge{a, b, c, d} {
			assumptions = append(assumptions, fraLit(Fra_ObjSatNum(p, edge), mask&(1<<i) == 0))
		}
		expected := mask&1 != 0 && mask&2 != 0 && mask&4 == 0 && mask&8 != 0
		if p.pSat.Solve(append(assumptions, fraLit(Fra_ObjSatNum(p, root), !expected))...) != bsat.Satisfiable {
			t.Fatalf("correct result rejected for %04b", mask)
		}
		if p.pSat.Solve(append(assumptions, fraLit(Fra_ObjSatNum(p, root), expected))...) != bsat.Unsatisfiable {
			t.Fatalf("wrong result accepted for %04b", mask)
		}
	}
}

func TestStructuralCNFIncrementalSharedCone(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c, d := graph.NewInput(), graph.NewInput(), graph.NewInput(), graph.NewInput()
	shared := graph.AND(a, b)
	left, right := graph.AND(shared, c), graph.AND(shared, d.Not())
	p := &Fra_Man_t{pManFraig: graph, pSat: bsat.NewSolver(), nSatVars: 1}
	Fra_ManClean(p, int32(len(graph.Nodes())))
	if !p.pSat.AddClause(bsat.Positive(0)) {
		t.Fatal("constant setup")
	}
	Fra_CnfNodeAddToSolver(p, left, aig.InvalidEdge)
	if p.pSat.Solve() != bsat.Satisfiable {
		t.Fatal("first cone was unsatisfiable")
	}
	sharedVar := Fra_ObjSatNum(p, shared)
	Fra_CnfNodeAddToSolver(p, right, aig.InvalidEdge)
	if Fra_ObjSatNum(p, shared) != sharedVar {
		t.Fatal("shared node got a new variable")
	}
	for mask := 0; mask < 16; mask++ {
		var assumptions []bsat.Literal
		for i, edge := range []aig.Edge{a, b, c, d} {
			assumptions = append(assumptions, fraLit(Fra_ObjSatNum(p, edge), mask&(1<<i) == 0))
		}
		sharedValue := mask&1 != 0 && mask&2 != 0
		leftValue := sharedValue && mask&4 != 0
		rightValue := sharedValue && mask&8 == 0
		assumptions = append(assumptions, fraLit(Fra_ObjSatNum(p, left), !leftValue), fraLit(Fra_ObjSatNum(p, right), !rightValue))
		if p.pSat.Solve(assumptions...) != bsat.Satisfiable {
			t.Fatalf("consistent shared-cone result rejected for %04b", mask)
		}
		assumptions[len(assumptions)-1] ^= 1
		if p.pSat.Solve(assumptions...) != bsat.Unsatisfiable {
			t.Fatalf("incorrect shared-cone result accepted for %04b", mask)
		}
	}
}
