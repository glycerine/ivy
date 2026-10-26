package abc

import "testing"

func TestPdrFrameContainmentAndRecycling(t *testing.T) {
	model := bmcTruthModel(t, bmcTruthCase{next0: 204, next1: 240, constraint: 255})
	options, _ := (PdrOptions{RecycleVariables: 1}).normalized()
	p, err := newPdrEngine(model, options)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.newFrame(true); err != nil {
		t.Fatal(err)
	}
	if err = p.newFrame(false); err != nil {
		t.Fatal(err)
	}
	weak, _ := NewPdrCube([]PdrLit{0, 2}, nil)
	strong, _ := NewPdrCube([]PdrLit{0}, nil)
	unrelated, _ := NewPdrCube([]PdrLit{1, 2}, nil)
	for _, cube := range []*PdrCube{weak, unrelated, strong, weak} {
		if err = p.addBlocked(1, cube); err != nil {
			t.Fatal(err)
		}
	}
	if p.frames[1].clauses.Len() != 2 || p.stats.SubsumedClauses != 1 || p.stats.ContainedClauses != 1 || p.stats.ClauseAdds != 3 {
		t.Fatal(p.stats)
	}
	if _, exists := p.frames[1].clauses.Get2(pdrCubeKey(weak)); exists {
		t.Fatal("redundant frame entry retained")
	}
	for round := 0; round < 2; round++ {
		for state := 0; state < 4; state++ {
			literals := []SatLit{p.current[0], p.current[1]}
			if state&1 == 0 {
				literals[0] = literals[0].Not()
			}
			if state&2 == 0 {
				literals[1] = literals[1].Not()
			}
			result, err := p.query(1, literals)
			if err != nil {
				t.Fatal(err)
			}
			// !q0 and !( !q0 & q1 ) together allow only state 00.
			expected := SatUnsatisfiable
			if state == 0 {
				expected = SatSatisfiable
			}
			if result.Status != expected {
				t.Fatal("frame truth", round, state, result)
			}
		}
		if round == 0 {
			if _, err := p.queryRelativeCube(1, strong); err != nil {
				t.Fatal(err)
			}
			if err := p.recycleFrame(1); err != nil {
				t.Fatal(err)
			}
		}
	}
	if p.stats.SolverRecycles != 1 {
		t.Fatal("recycling not exercised")
	}
}
