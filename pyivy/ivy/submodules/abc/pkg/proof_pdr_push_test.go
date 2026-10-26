package abc

import "testing"

func TestPdrPushCoreStrengthensCumulativeFrames(t *testing.T) {
	model := bmcTruthModel(t, bmcTruthCase{next0: 0, next1: 240, constraint: 255})
	options, _ := (PdrOptions{}).normalized()
	p, err := newPdrEngine(model, options)
	if err != nil {
		t.Fatal(err)
	}
	for _, initial := range []bool{true, false, false} {
		if err = p.newFrame(initial); err != nil {
			t.Fatal(err)
		}
	}
	weak, _ := NewPdrCube([]PdrLit{0, 2}, nil)
	if err = p.addBlocked(1, weak); err != nil {
		t.Fatal(err)
	}
	other, _ := NewPdrCube([]PdrLit{0, 3}, nil)
	if err = p.addBlocked(1, other); err != nil {
		t.Fatal(err)
	}
	frame, err := p.propagate(1)
	if err != nil || frame != 1 {
		t.Fatal("did not converge", frame, err)
	}
	strong, _ := NewPdrCube([]PdrLit{0}, nil)
	for i := 1; i <= 2; i++ {
		if p.frames[i].clauses.Len() != 1 {
			t.Fatal("redundant clauses", i)
		}
		if _, exists := p.frames[i].clauses.Get2(pdrCubeKey(strong)); !exists {
			t.Fatal("pushed core not retained", i)
		}
		for state := 0; state < 4; state++ {
			assumptions := []SatLit{p.current[0], p.current[1]}
			if state&1 == 0 {
				assumptions[0] = assumptions[0].Not()
			}
			if state&2 == 0 {
				assumptions[1] = assumptions[1].Not()
			}
			result, err := p.query(i, assumptions)
			if err != nil {
				t.Fatal(err)
			}
			expected := SatSatisfiable
			if state&1 != 0 {
				expected = SatUnsatisfiable
			}
			if result.Status != expected {
				t.Fatal("cumulative strengthening", i, state, result)
			}
		}
	}
	if p.stats.CoreReductions == 0 || p.stats.SubsumedClauses < 2 || p.stats.ClausePushes != 1 {
		t.Fatal(p.stats)
	}
}
