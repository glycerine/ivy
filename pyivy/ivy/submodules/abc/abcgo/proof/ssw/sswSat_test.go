package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSatEquivalenceQueries(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		x := g.AND(a, b)
		same := g.AND(x, a)
		different := g.AND(a, c)
		if x == same {
			t.Fatal("test requires distinct nodes")
		}
		sat := Ssw_SatStart(flip)
		sat.pAig = g
		p := &Ssw_Man_t{pPars: &Ssw_Pars_t{FPolarFlip: flip}, pFrames: g, pMSat: sat, iOutputLit: -1}
		if got := Ssw_NodesAreEquiv(p, x, same); got != 1 {
			t.Fatalf("flip %d: equal result %d", flip, got)
		}
		if got := Ssw_NodesAreEquiv(p, x, different); got != 0 {
			t.Fatalf("flip %d: different result %d", flip, got)
		}
	}
}
func TestSatEquivalenceConstraint(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	sat := Ssw_SatStart(0)
	sat.pAig = g
	p := &Ssw_Man_t{pPars: &Ssw_Pars_t{FConstrs: 1}, pFrames: g, pMSat: sat, iOutputLit: -1}
	Ssw_NodesAreConstrained(p, a, b.Not())
	if sat.pSat.Solve(sswLit(sat, a, false), sswLit(sat, b, false)) != bsat.Unsatisfiable {
		t.Fatal("constraint permits equal inputs")
	}
	if sat.pSat.Solve(sswLit(sat, a, false), sswLit(sat, b, true)) != bsat.Satisfiable {
		t.Fatal("constraint rejects opposite inputs")
	}
}

func TestOutputConstraintUsesSourceFaninPolarity(t *testing.T) {
	// sswSat.c adds a unit with !fanin-complement, i.e. output=false.
	for _, flip := range []int32{0, 1} {
		for _, complemented := range []bool{false, true} {
			for _, inputValue := range []bool{false, true} {
				g := aig.NewGraph()
				x := g.NewInput()
				driver := x
				if complemented {
					driver = driver.Not()
				}
				out := g.AddOutput(driver)
				sat := Ssw_SatStart(flip)
				sat.pAig = g
				p := &Ssw_Man_t{pPars: &Ssw_Pars_t{FPolarFlip: flip}, pFrames: g, pMSat: sat, iOutputLit: -1}
				if Ssw_NodeIsConstrained(p, out) != 1 {
					t.Fatal("constraint rejected")
				}
				lit := sswQueryLit(p, x, !inputValue)
				status := sat.pSat.Solve(lit)
				if (status == bsat.Satisfiable) != (inputValue == complemented) {
					t.Fatalf("flip=%d complement=%t input=%t status=%v", flip, complemented, inputValue, status)
				}
				Ssw_SatStop(sat)
			}
		}
	}
}
