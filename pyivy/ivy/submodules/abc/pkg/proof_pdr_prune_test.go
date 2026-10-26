package abc

import "testing"

func pdrCheckPrune(t testing.TB, n0, n1, env, mask uint8) {
	t.Helper()
	model := bmcTruthModel(t, bmcTruthCase{next0: n0, next1: n1, constraint: env})
	cubes := []*PdrCube{}
	for code := 0; code < 8; code++ {
		if mask&(1<<code) == 0 {
			continue
		}
		value := code + 1
		literals := []PdrLit{}
		for bit := 0; bit < 2; bit++ {
			digit := value % 3
			value /= 3
			if digit != 0 {
				l := PdrLit(2 * bit)
				if digit == 2 {
					l = l.Not()
				}
				literals = append(literals, l)
			}
		}
		cube, _ := NewPdrCube(literals, nil)
		cubes = append(cubes, cube)
	}
	inv, _ := NewPdrInvariant(cubes)
	r, err := inv.PruneNonInductive(model, PdrPruneOptions{})
	if err != nil || !r.Complete || r.Invariant == nil {
		t.Fatal(r, err)
	}
	expected := append([]*PdrCube(nil), cubes...)
	passes := 0
	for {
		passes++
		keep := []*PdrCube{}
		for _, candidate := range expected {
			inductive := true
			for state := 0; state < 4; state++ {
				bits := []bool{state&1 != 0, state&2 != 0}
				allowed := true
				for _, cube := range expected {
					match, _ := cube.Matches(bits)
					if match {
						allowed = false
					}
				}
				if !allowed {
					continue
				}
				for input := 0; input < 2; input++ {
					index := 2*state + input
					if env&(1<<index) == 0 {
						continue
					}
					next := []bool{n0&(1<<index) != 0, n1&(1<<index) != 0}
					match, _ := candidate.Matches(next)
					if match {
						inductive = false
					}
				}
			}
			if inductive {
				keep = append(keep, candidate)
			}
		}
		if len(keep) == len(expected) {
			break
		}
		expected = keep
	}
	if r.Passes != passes || r.Removed != len(cubes)-len(expected) || len(r.Invariant.clauses) != len(expected) {
		t.Fatal("fixed point mismatch", r, passes, len(expected))
	}
	for i, cube := range expected {
		if cube.Compare(r.Invariant.clauses[i]) != 0 {
			t.Fatal("wrong retained clause")
		}
	}
	if len(inv.clauses) != len(cubes) {
		t.Fatal("source mutated")
	}
}
func TestPdrPruneIndependentFixedPoints(t *testing.T) {
	for i := 0; i < 256; i++ {
		pdrCheckPrune(t, uint8(i), uint8(i*73), uint8(i*17), uint8(i*111))
	}
}
func TestPdrPruneCascadeAndInterruption(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 240, next1: 255, constraint: 255})
	a, _ := NewPdrCube([]PdrLit{0}, nil)
	b, _ := NewPdrCube([]PdrLit{2}, nil)
	inv, _ := NewPdrInvariant([]*PdrCube{a, b})
	r, err := inv.PruneNonInductive(m, PdrPruneOptions{})
	if err != nil || !r.Complete || r.Removed != 2 || r.Passes != 3 || len(r.Invariant.clauses) != 0 {
		t.Fatal("dependent removal", r, err)
	}
	r, err = inv.PruneNonInductive(m, PdrPruneOptions{MaxQueries: 1})
	if err != nil || r.Complete || r.Invariant != nil || !r.QueryLimit {
		t.Fatal("incomplete pruning returned result", r, err)
	}
	r, err = inv.PruneNonInductive(m, PdrPruneOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || r.Complete || r.Invariant != nil || r.Reason != SatCancelled {
		t.Fatal(r, err)
	}
	for _, options := range []PdrPruneOptions{{MaxVariables: 1}, {MaxClauses: 1}, {MaxClauses: -1}, {MaxVariables: -1}} {
		if _, err := inv.PruneNonInductive(m, options); err == nil {
			t.Fatal("limit ignored", options)
		}
	}
}
func FuzzPdrPrune(f *testing.F) {
	f.Add(uint8(240), uint8(255), uint8(255), uint8(15))
	f.Fuzz(func(t *testing.T, n0, n1, env, mask uint8) { pdrCheckPrune(t, n0, n1, env, mask) })
}

func TestPdrPruneEmptyAndVacuousPremises(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 255, next1: 255, constraint: 255})
	empty, _ := NewPdrInvariant(nil)
	r, err := empty.PruneNonInductive(m, PdrPruneOptions{})
	if err != nil || !r.Complete || r.Passes != 1 || r.Queries != 0 {
		t.Fatal(r, err)
	}
	cube, _ := NewPdrCube(nil, nil)
	impossible, _ := NewPdrInvariant([]*PdrCube{cube})
	r, err = impossible.PruneNonInductive(m, PdrPruneOptions{})
	if err != nil || !r.Complete || len(r.Invariant.clauses) != 1 {
		t.Fatal("false predicate is transition closed", r, err)
	}
	check, err := r.Invariant.Verify(m, 0, PdrVerifyOptions{})
	if err != nil || check.Status != PdrInvariantInvalid || check.Phase != PdrInvariantInitiation {
		t.Fatal("closure confused with proof", check, err)
	}
	blocked, _ := NewPdrCube([]PdrLit{0}, nil)
	inv, _ := NewPdrInvariant([]*PdrCube{blocked})
	constrained := bmcTruthModel(t, bmcTruthCase{next0: 255, constraint: 0})
	r, err = inv.PruneNonInductive(constrained, PdrPruneOptions{})
	if err != nil || !r.Complete || r.Removed != 0 {
		t.Fatal("ignored environment", r, err)
	}
}
