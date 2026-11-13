package msat

import "testing"

func TestClauseReasonBumpsLearnedActivity(t *testing.T) {
	p := New(2)
	c := &Msat_Clause_t{learned: true, lits: []int{0, 2}}
	p.pAssigns[0], p.pAssigns[1] = 0, 3
	out := Msat_IntVecAlloc(2)
	Msat_ClauseCalcReason(p, c, 0, out)
	if c.activity != float32(p.dClaInc) {
		t.Fatalf("C bumps learned reason activity, got %v", c.activity)
	}
	if len(out.data) != 1 || out.data[0] != 3 {
		t.Fatalf("reason = %v", out.data)
	}
}

func TestClauseReasonRequiresFirstImpliedLiteral(t *testing.T) {
	p := New(2)
	c := &Msat_Clause_t{lits: []int{0, 2}}
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts implied literal is first")
		}
	}()
	Msat_ClauseCalcReason(p, c, 2, Msat_IntVecAlloc(2))
}
