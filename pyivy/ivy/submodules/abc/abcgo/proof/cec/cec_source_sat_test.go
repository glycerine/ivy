package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSourceSATFrontierAndClauses(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		node := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
		gia.Gia_ManSetPhase(g)
		p := new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(p)
		p.fPolarFlip = flip
		s := newCecSourceSAT(g, p)
		s.addNode(node >> 1)
		if s.cnf.IDs()[node>>1] != 2 || s.cnf.IDs()[a>>1] != 4 || s.cnf.IDs()[b>>1] != 5 || s.cnf.IDs()[c>>1] != 3 {
			t.Fatalf("frontier: %v", s.cnf.IDs())
		}
		if s.cnf.Solver().NumVars() != 6 {
			t.Fatalf("SAT variable count %d", s.cnf.Solver().NumVars())
		}
		if s.cnf.Solver().Solve(s.literal(node)) != bsat.Satisfiable {
			t.Fatal("source SAT query failed")
		}
		for _, lit := range []int32{a, b, c} {
			value, ok := s.cnf.Solver().Value(s.cnf.IDs()[lit>>1])
			if !ok || !value {
				t.Fatalf("source witness for %d: %v %v", lit, value, ok)
			}
		}
		if s.cnf.IDs()[node>>1-1] != 0 {
			t.Fatalf("supergate intermediate allocated: %v", s.cnf.IDs())
		}
		for mask := 0; mask < 8; mask++ {
			ass := []int32{s.literal(node)}
			for i, l := range []int32{a, b, c} {
				v := s.literal(l)
				if mask&(1<<i) == 0 {
					v ^= 1
				}
				ass = append(ass, v)
			}
			got := s.cnf.Solver().Solve(ass...)
			if (got == bsat.Satisfiable) != (mask == 7) {
				t.Fatalf("flip %d mask %d status %v", flip, mask, got)
			}
		}
	}
}

func TestSourceSATRecognizedMuxClauses(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := gia.Gia_ManStart(20)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a := gia.Gia_ManAppendAnd(g, x, y)
		b := gia.Gia_ManAppendAnd(g, x^1, z)
		root := gia.Gia_ManAppendAnd(g, a^1, b^1)
		gia.Gia_ManSetPhase(g)
		p := new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(p)
		p.fPolarFlip = flip
		s := newCecSourceSAT(g, p)
		s.addNode(root >> 1)
		if s.cnf.IDs()[a>>1] != 0 || s.cnf.IDs()[b>>1] != 0 {
			t.Fatalf("MUX intermediate variables: %v", s.cnf.IDs())
		}
		for mask := 0; mask < 16; mask++ {
			var ass []int32
			for i, l := range []int32{x, y, z, root} {
				v := s.literal(l)
				if mask&(1<<i) == 0 {
					v ^= 1
				}
				ass = append(ass, v)
			}
			want := !((mask&1 != 0 && mask&2 != 0) || (mask&1 == 0 && mask&4 != 0))
			if got := s.cnf.Solver().Solve(ass...); (got == bsat.Satisfiable) != (want == (mask&8 != 0)) {
				t.Fatalf("flip=%d mask=%d result=%v", flip, mask, got)
			}
		}
	}
}

// Cec_ManSatSolverRecycle only clears the solver and its variable map; it does
// not clear the collected counterexample vector or cumulative query counters.
func TestSourceSATRecycleRetainsCounterexampleAndCounters(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	p := Cec_ManSatCreate(g, nil)
	if Cec_ManSatCheckNode(p, gia.Gia_ManObj(g, x>>1)) != 0 {
		t.Fatal("expected SAT")
	}
	vec.Vec_IntPush(p.vCex, 17)
	calls := p.source.calls
	Cec_ManSatSolverRecycle(p)
	if p.source.calls != calls || p.source.cnf.IDs()[x>>1] != 0 || vec.Vec_IntSize(p.vCex) != 1 || vec.Vec_IntEntry(p.vCex, 0) != 17 {
		t.Fatalf("recycle lost state: calls=%d ids=%v cex=%v", p.source.calls, p.source.cnf.IDs(), vec.Vec_IntArray(p.vCex))
	}
}

// C reads the SAT variable's physical model bit, including variable zero for
// nodes omitted from a supergate, rather than evaluating the graph logically.
func TestSourceSATValueReadsPhysicalVariable(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	mid := gia.Gia_ManAppendAnd(g, a, b)
	root := gia.Gia_ManAppendAnd(g, mid, c)
	p := Cec_ManSatCreate(g, nil)
	if Cec_ManSatCheckNode(p, gia.Gia_ManObj(g, root>>1)) != 0 {
		t.Fatal("expected SAT")
	}
	if p.source.cnf.IDs()[mid>>1] != 0 {
		t.Fatal("intermediate should not have a SAT variable")
	}
	v, ok := p.source.cnf.Solver().Value(0)
	if !ok {
		t.Fatal("missing model")
	}
	want := int32(0)
	if v {
		want = 1
	}
	if got := Cec_ObjSatVarValue(p, gia.Gia_ManObj(g, mid>>1)); got != want {
		t.Fatalf("physical variable zero got %d want %d", got, want)
	}
}
