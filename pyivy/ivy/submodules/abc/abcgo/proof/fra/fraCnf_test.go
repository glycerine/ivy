package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestStructuralCNFTruthTables(t *testing.T) {
	for _, kind := range []string{"and", "xor", "mux"} {
		t.Run(kind, func(t *testing.T) {
			g := aig.NewGraph()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			var out aig.Edge
			switch kind {
			case "and":
				out = g.AND(a.Not(), b)
			case "xor":
				out = g.XOR(a.Not(), b)
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
				case "xor":
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
