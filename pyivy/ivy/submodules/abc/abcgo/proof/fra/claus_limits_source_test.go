package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// The pigeonhole constraints require more than one conflict. C's
// Fra_ClausRunSat asserts on undefined, while its BMC and Sat0 variants
// conservatively return zero. This is confirmed with fraClaus.c compiled C.
func TestClauseResourceLimitSourceStatusPaths(t *testing.T) {
	makeManager := func() *Clu_Man_t {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x)
		p := Fra_ClausAlloc(g, 0, 0, 4, 4, 1, 2, 1, 0, 0, 0, 0)
		p.pCnf = cnf.Cnf_DeriveSimple(g, 1)
		p.pSatMain = bsat.NewSolver()
		p.pSatBmc = p.pSatMain
		p.nBTLimit = 1
		base := int32(p.pCnf.Variables + 1)
		for i := int32(0); i < 8; i++ {
			row := make([]bsat.Literal, 7)
			for h := int32(0); h < 7; h++ {
				row[h] = bsat.Positive(base + i*7 + h)
			}
			if !p.pSatMain.AddClause(row...) {
				t.Fatal("fixture pigeon clause")
			}
		}
		for h := int32(0); h < 7; h++ {
			for i := int32(0); i < 8; i++ {
				for j := int32(0); j < i; j++ {
					if !p.pSatMain.AddClause(bsat.Negative(base+i*7+h), bsat.Negative(base+j*7+h)) {
						t.Fatal("fixture hole clause")
					}
				}
			}
		}
		return p
	}
	t.Run("negative budget on pigeonhole", func(t *testing.T) {
		p := makeManager()
		p.nBTLimit = -1
		if got := Fra_ClausRunSat0(p); got != 0 {
			t.Fatalf("status=%d; compiled C returns 0", got)
		}
	})
	t.Run("induction asserts", func(t *testing.T) {
		p := makeManager()
		defer func() {
			if recover() == nil {
				t.Fatal("source asserts on undefined SAT result")
			}
		}()
		Fra_ClausRunSat(p)
	})
	t.Run("single frame returns zero", func(t *testing.T) {
		if got := Fra_ClausRunSat0(makeManager()); got != 0 {
			t.Fatalf("status=%d", got)
		}
	})
	t.Run("BMC returns zero", func(t *testing.T) {
		p := makeManager()
		p.nFrames = 1
		if got := Fra_ClausRunBmc(p); got != 0 {
			t.Fatalf("status=%d", got)
		}
	})
}

// ABC passes the signed budget through to sat_solver_solve; negative values
// must not be rejected by the Go convenience SetLimits API.
func TestClauseNegativeResourceLimit(t *testing.T) {
	for _, run := range []struct {
		name string
		f    func(*Clu_Man_t) int32
	}{
		{"sat0", Fra_ClausRunSat0}, {"bmc", Fra_ClausRunBmc},
	} {
		t.Run(run.name, func(t *testing.T) {
			g := aig.NewGraph()
			x := g.NewInput()
			g.AddOutput(x)
			p := Fra_ClausAlloc(g, 1, 0, 4, 4, 1, 2, 1, 0, 0, 0, 0)
			p.pCnf = cnf.Cnf_DeriveSimple(g, 1)
			p.pSatMain = bsat.NewSolver()
			p.pSatBmc = p.pSatMain
			for i := 0; i < p.pCnf.Variables; i++ {
				p.pSatMain.NewVariable()
			}
			p.nBTLimit = -1
			if got := run.f(p); got != 0 {
				t.Fatalf("got %d; want C's zero", got)
			}
		})
	}
}
