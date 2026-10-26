package abc

import "testing"

func TestPdrRecyclePreservesFramesAndLimits(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 204, next1: 240, constraint: 255})
	options, _ := (PdrOptions{RecycleVariables: 1}).normalized()
	p, err := newPdrEngine(m, options)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.newFrame(true); err != nil {
		t.Fatal(err)
	}
	if err = p.newFrame(false); err != nil {
		t.Fatal(err)
	}
	cube, _ := NewPdrCube([]PdrLit{0}, nil)
	other, _ := NewPdrCube([]PdrLit{2}, nil)
	if err = p.addBlocked(1, other); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		r, err := p.queryRelativeCube(1, cube)
		if err != nil || r.Status != SatUnsatisfiable {
			t.Fatal(r, err)
		}
		r, err = p.queryCube(1, cube, true)
		if err != nil || r.Status != SatSatisfiable {
			t.Fatal("lost unconstrained register", r, err)
		}
		r, err = p.queryCube(1, other, false)
		if err != nil || r.Status != SatUnsatisfiable {
			t.Fatal("lost frame clause", r, err)
		}
		r, err = p.queryRelativeCube(0, cube)
		if err != nil || r.Status != SatUnsatisfiable {
			t.Fatal(r, err)
		}
		r, err = p.queryCube(0, cube, false)
		if err != nil || r.Status != SatUnsatisfiable {
			t.Fatal("lost initial predicate", r, err)
		}
	}
	if p.stats.SolverRecycles != 40 {
		t.Fatal(p.stats)
	}
	for _, f := range p.frames {
		if f.solver.VariableCount() != p.relation.variables {
			t.Fatal("retired guards retained")
		}
		if f.solver.maxVariables != options.MaxVariables {
			t.Fatal("lost solver limit")
		}
		if err := f.solver.Check(); err != nil {
			t.Fatal(err)
		}
	}
	if p.stats.SAT.Propagations == 0 {
		t.Fatal("recycled statistics lost")
	}
}

func TestPdrReachabilityWithFrequentRecycling(t *testing.T) {
	for i := 0; i < 256; i++ {
		pdrVerifyTruthOptions(t, bmcTruthCase{bad: uint8(i), next0: uint8(i * 73), next1: uint8(i * 119), constraint: uint8(i * 17), initial: [2]LatchInit{LatchInit(i % 3), LatchInit((i / 3) % 3)}}, PdrOptions{RecycleVariables: 1, MaxFrames: 16, MaxQueries: 10000})
	}
}

func FuzzPdrRecycling(f *testing.F) {
	f.Add(uint8(192), uint8(170), uint8(204), uint8(255), uint8(0))
	f.Fuzz(func(t *testing.T, bad, n0, n1, env, init uint8) {
		pdrVerifyTruthOptions(t, bmcTruthCase{bad: bad, next0: n0, next1: n1, constraint: env, initial: [2]LatchInit{LatchInit(init % 3), LatchInit((init / 3) % 3)}}, PdrOptions{OrderAssumptions: true, RecycleVariables: 1, MaxFrames: 16, MaxQueries: 10000})
	})
}

func TestPdrRecycleCancellation(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{next0: 204, constraint: 255})
	options, _ := (PdrOptions{RecycleVariables: 1}).normalized()
	p, err := newPdrEngine(m, options)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.newFrame(false); err != nil {
		t.Fatal(err)
	}
	cube, _ := NewPdrCube([]PdrLit{0}, nil)
	if _, err = p.queryRelativeCube(0, cube); err != nil {
		t.Fatal(err)
	}
	calls := 0
	p.options.SAT.Stop = func() bool { calls++; return calls >= 2 }
	if _, err = p.queryCube(0, cube, false); err != ErrCancelled {
		t.Fatal("rebuild ignored cancellation", err)
	}
	// Rebuild was interrupted in relation loading. The engine caller treats
	// this as terminal UNKNOWN and must never continue a partially loaded frame.
	if p.stats.SolverRecycles != 1 {
		t.Fatal(p.stats)
	}
}
