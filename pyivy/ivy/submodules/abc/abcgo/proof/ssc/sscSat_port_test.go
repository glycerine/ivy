package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func sscSatFixture() (*Ssc_Man_t, []int32) {
	g := gia.Gia_ManStart(32)
	inputs := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
	p := &Ssc_Man_t{Aig: g, Params: DefaultParams()}
	Ssc_ManStartSolver(p)
	return p, inputs
}
func TestSatPortSupergateFrontier(t *testing.T) {
	p, in := sscSatFixture()
	first := gia.Gia_ManAppendAnd(p.Aig, in[0], in[1])
	root := gia.Gia_ManAppendAnd(p.Aig, first, in[2])
	leaves := p.collectSuper(root >> 1)
	if len(leaves) != 3 {
		t.Fatalf("supergate leaves=%v", leaves)
	}
	p.cnfNode(root >> 1)
	if p.satVar(first>>1) != 0 {
		t.Fatal("intermediate supergate was not flattened")
	}
	for mask := 0; mask < 8; mask++ {
		assum := make([]bsat.Literal, 0, 4)
		for i, lit := range in {
			v := p.satLit(lit)
			if mask&(1<<i) == 0 {
				v ^= 1
			}
			assum = append(assum, v)
		}
		want := mask == 7
		rootLit := p.satLit(root)
		if !want {
			rootLit ^= 1
		}
		if p.sat.Solve(append(append([]bsat.Literal{}, assum...), rootLit)...) != bsat.Satisfiable {
			t.Fatalf("correct assignment %d rejected", mask)
		}
		if p.sat.Solve(append(append([]bsat.Literal{}, assum...), rootLit^1)...) != bsat.Unsatisfiable {
			t.Fatalf("incorrect assignment %d accepted", mask)
		}
	}
}
func TestSatPortStructuralMuxFrontier(t *testing.T) {
	p, in := sscSatFixture()
	output := gia.Gia_ManAppendMux(p.Aig, in[0], in[1]^1, in[2])
	root := output ^ 1
	if gia.Gia_ObjIsMuxType(gia.Gia_ManObj(p.Aig, root>>1)) == 0 {
		t.Fatal("fixture is not structural mux")
	}
	p.cnfNode(root >> 1)
	for mask := 0; mask < 8; mask++ {
		assum := make([]bsat.Literal, 0, 4)
		for i, lit := range in {
			v := p.satLit(lit)
			if mask&(1<<i) == 0 {
				v ^= 1
			}
			assum = append(assum, v)
		}
		v := mask&4 != 0
		if mask&1 != 0 {
			v = mask&2 == 0
		}
		v = !v
		rootLit := p.satLit(root)
		if !v {
			rootLit ^= 1
		}
		if p.sat.Solve(append(append([]bsat.Literal{}, assum...), rootLit)...) != bsat.Satisfiable || p.sat.Solve(append(append([]bsat.Literal{}, assum...), rootLit^1)...) != bsat.Unsatisfiable {
			t.Fatalf("mux mask %d", mask)
		}
	}
}
func TestSatPortRejectsReverseNodeOrder(t *testing.T) {
	p, in := sscSatFixture()
	if got := Ssc_ManCheckEquivalence(p, in[1]>>1, in[0]>>1, 0); got != satoko.SATOKO_UNDEC {
		t.Fatalf("reverse order result %d", got)
	}
}
func TestSatPortConstantComparison(t *testing.T) {
	p, in := sscSatFixture()
	n := gia.Gia_ManAppendAnd(p.Aig, in[0], in[1])
	if got := Ssc_ManCheckEquivalence(p, 0, n>>1, 0); got != satoko.SATOKO_SAT {
		t.Fatalf("nonconstant gate result %d", got)
	}
	if len(p.Pattern) != 3 || !p.Pattern[0] || !p.Pattern[1] {
		t.Fatalf("bad pattern %v", p.Pattern)
	}
}

func TestSatPortOrderedImplicationAndReverseWitness(t *testing.T) {
	p, in := sscSatFixture()
	a, b := p.satLit(in[0]), p.satLit(in[1])
	if !p.sat.AddClause(a^1, b) {
		t.Fatal("cannot add care implication")
	}
	if got := Ssc_ManCheckEquivalence(p, in[0]>>1, in[1]>>1, 0); got != satoko.SATOKO_SAT {
		t.Fatalf("expected reverse counterexample, got %d", got)
	}
	if len(p.Pattern) != 3 || p.Pattern[0] || !p.Pattern[1] {
		t.Fatalf("wrong reverse witness: %v", p.Pattern)
	}
	if p.Stats.Calls != 0 || p.Stats.Counterexamples != 0 {
		t.Fatalf("direct SAT helper must not change sweep counters: %+v", p.Stats)
	}
}
func TestSatPortProvesBothImplicationsUnderCare(t *testing.T) {
	p, in := sscSatFixture()
	a, b := p.satLit(in[0]), p.satLit(in[1])
	if !p.sat.AddClause(a^1, b) || !p.sat.AddClause(a, b^1) {
		t.Fatal("cannot add care equivalence")
	}
	if got := Ssc_ManCheckEquivalence(p, in[0]>>1, in[1]>>1, 0); got != satoko.SATOKO_UNSAT {
		t.Fatalf("expected proof, got %d", got)
	}
	if p.Stats.Calls != 0 || p.Stats.Proofs != 0 {
		t.Fatalf("direct SAT helper must not change sweep counters: %+v", p.Stats)
	}
}
func TestSatPortConstantLiteralMapping(t *testing.T) {
	p, _ := sscSatFixture()
	if p.sat.Solve(bsat.Literal(p.encode(0, 0))) != bsat.Unsatisfiable {
		t.Fatal("GIA constant zero is satisfiable")
	}
	if p.sat.Solve(bsat.Literal(p.encode(0, 1))) != bsat.Satisfiable {
		t.Fatal("GIA constant one is unsatisfiable")
	}
}
