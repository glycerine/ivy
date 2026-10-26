package abc

// Go-only tests for src/proof/int, excluding proof/int2.

import (
	"errors"
	"slices"
	"testing"
)

func proofIntSequentialFixture(t testing.TB, safe bool) *AigManager {
	t.Helper()
	g := NewAigManager()
	pi, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	lo, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	property := lo
	if safe {
		property = AigConst0
	}
	if _, err = g.CreateCO(property); err != nil {
		t.Fatal(err)
	}
	next, err := g.Or(lo, pi)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func proofIntConstrainedFixture(t testing.TB, blocking bool) *AigManager {
	t.Helper()
	g := NewAigManager()
	pi, _ := g.CreateCI()
	lo, _ := g.CreateCI()
	_, _ = g.CreateCO(pi)
	constraint := AigConst0
	if blocking {
		constraint = pi
	}
	_, _ = g.CreateCO(constraint)
	_, _ = g.CreateCO(lo)
	if err := g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func proofIntComb(t testing.TB, operation byte) *AigManager {
	t.Helper()
	g := NewAigManager()
	x, _ := g.CreateCI()
	y, _ := g.CreateCI()
	var root AigLit
	var err error
	switch operation {
	case '&':
		root, err = g.And(x, y)
	case '|':
		root, err = g.Or(x, y)
	case '^':
		root, err = g.Xor(x, y)
	default:
		root = x
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	return g
}

func proofIntTruth2(t testing.TB, graph *AigManager, first, second AigLit, table uint8) AigLit {
	t.Helper()
	result := AigConst0
	for assignment := 0; assignment < 4; assignment++ {
		if table&(1<<assignment) == 0 {
			continue
		}
		term, err := graph.And(first.NotIf(assignment&1 == 0), second.NotIf(assignment&2 == 0))
		if err != nil {
			t.Fatal(err)
		}
		result, err = graph.Or(result, term)
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func proofIntTruthTransition(t testing.TB, property, next uint8) *AigManager {
	t.Helper()
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	_, _ = graph.CreateCO(proofIntTruth2(t, graph, input, state, property))
	_, _ = graph.CreateCO(proofIntTruth2(t, graph, input, state, next))
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestProofIntDefaultsDupFramesAndChecks(t *testing.T) {
	p := DefaultProofIntParams()
	if p.BacktrackLimit != 0 || p.MaxFrames != 0 || p.InductionK != 1 || !p.CheckKStep || p.UseBias || p.FrameReached != -1 {
		t.Fatal(p)
	}
	init, err := ProofIntInitialState(3, ProofIntLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, e := init.Evaluate(in)
		if e != nil || out[0] != (mask == 0) {
			t.Fatal(mask, out, e)
		}
	}
	g := proofIntSequentialFixture(t, false)
	dup, err := ProofIntDuplicateTransition(g, ProofIntLimits{})
	if err != nil || dup.RegisterCount() != 1 || dup.PrimaryInputCount() != 1 || dup.PrimaryOutputCount() != 0 {
		t.Fatal(dup, err)
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0}
		state := []bool{mask&2 != 0}
		_, n, e := g.Step(in, state)
		if e != nil {
			t.Fatal(e)
		}
		_, m, e := dup.Step(in, state)
		if e != nil || !slices.Equal(n, m) {
			t.Fatal(mask, n, m, e)
		}
	}
	frames, err := ProofIntFrames(g, 2, false, true, ProofIntLimits{})
	if err != nil || frames.RegisterCount() != 0 || frames.InputCount() != 3 || frames.OutputCount() != 1 {
		t.Fatal(frames, err)
	}
	bmc, err := ProofIntBMCFrames(g, 2, ProofIntLimits{})
	if err != nil || bmc.InputCount() != 2 || bmc.OutputCount() != 1 {
		t.Fatal(bmc, err)
	}
	if out, e := bmc.Evaluate([]bool{true, false}); e != nil || !out[0] {
		t.Fatal(out, e)
	}
	expanded, err := ProofIntExpand(proofIntComb(t, 0), frames, ProofIntLimits{})
	if err != nil || expanded.InputCount() != frames.InputCount() {
		t.Fatal(expanded, err)
	}
	initial, bad, err := ProofIntCheckInitial(g, ProofIntLimits{})
	if err != nil || bad || initial != nil {
		t.Fatal(initial, bad, err)
	}
	all, err := ProofIntCheckAllStates(g, ProofIntLimits{})
	if err != nil || all {
		t.Fatal(all, err)
	}
}

func TestProofIntContainmentEquivalenceAndCounterexample(t *testing.T) {
	andGraph, orGraph := proofIntComb(t, '&'), proofIntComb(t, '|')
	contained, err := ProofIntContainment(andGraph, orGraph, ProofIntLimits{})
	if err != nil || !contained {
		t.Fatal(contained, err)
	}
	contained, err = ProofIntContainment(orGraph, andGraph, ProofIntLimits{})
	if err != nil || contained {
		t.Fatal(contained, err)
	}
	equal, err := ProofIntEquivalent(andGraph, andGraph.Clone(), ProofIntLimits{})
	if err != nil || !equal {
		t.Fatal(equal, err)
	}
	g := proofIntSequentialFixture(t, false)
	cex, err := ProofIntCounterexample(g, 2, ProofIntLimits{})
	if err != nil || cex == nil || cex.Frames() != 2 {
		t.Fatal(cex, err)
	}
	if err = g.ReplayCounterexample(cex); err != nil {
		t.Fatal(err)
	}
	if cex2, err := ProofIntCounterexample(proofIntSequentialFixture(t, true), 4, ProofIntLimits{}); err != nil || cex2 != nil {
		t.Fatal(cex2, err)
	}
}

func TestProofIntConstraintsAndCumulativeLimits(t *testing.T) {
	blocked := proofIntConstrainedFixture(t, true)
	if counterexample, bad, err := ProofIntCheckInitial(blocked, ProofIntLimits{}); err != nil || bad || counterexample != nil {
		t.Fatal("blocked initial state", counterexample, bad, err)
	}
	if counterexample, err := ProofIntCounterexample(blocked, 3, ProofIntLimits{}); err != nil || counterexample != nil {
		t.Fatal("blocked trace", counterexample, err)
	}
	allowed := proofIntConstrainedFixture(t, false)
	if counterexample, bad, err := ProofIntCheckInitial(allowed, ProofIntLimits{}); err != nil || !bad || counterexample == nil {
		t.Fatal("allowed initial state", counterexample, bad, err)
	}
	if counterexample, err := ProofIntCounterexample(proofIntSequentialFixture(t, true), 3, ProofIntLimits{MaxAssignments: 2}); counterexample != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("cumulative assignment limit", counterexample, err)
	}
}

func TestProofIntInductiveAndIncrementalCheck(t *testing.T) {
	transition := proofIntSequentialFixture(t, true)
	invariant := NewAigManager()
	x, _ := invariant.CreateCI()
	_, _ = invariant.CreateCO(x)
	forward, err := ProofIntInductiveContainment(transition, invariant, 2, false, ProofIntLimits{})
	if err != nil || !forward {
		t.Fatal(forward, err)
	}
	backward, err := ProofIntInductiveContainment(transition, invariant, 1, true, ProofIntLimits{})
	if err != nil || backward {
		t.Fatal(backward, err)
	}
	checker, err := NewProofIntChecker(transition, 2, ProofIntLimits{})
	if err != nil {
		t.Fatal(err)
	}
	passed, err := checker.Perform(invariant)
	if err != nil || !passed {
		t.Fatal(passed, err)
	}
	checker.Close()
	checker.Close()
	if _, err = checker.Perform(invariant); err == nil {
		t.Fatal("closed checker accepted work")
	}
}

func proofIntProofFixture(t testing.TB) (*M114PProofSnapshot, []bool, []int) {
	t.Helper()
	s, err := NewM114PSolver(true, M114PSolverOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_ = s.SetVariableCount(2)
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	clauses := [][]SatLit{{x, y}, {x, y.Not()}, {x.Not(), y}, {x.Not(), y.Not()}}
	for _, c := range clauses {
		ok, e := s.AddClause(c)
		if e != nil || !ok {
			t.Fatal(ok, e)
		}
	}
	r, e := s.Solve(nil, 0)
	if e != nil || r.Status != SatUnsatisfiable {
		t.Fatal(r, e)
	}
	p, e := s.ProofSnapshot()
	if e != nil {
		t.Fatal(e)
	}
	return p, []bool{false, false, true, true}, []int{0, 1}
}

func TestProofIntM114PInterpolationReplayAndCorruption(t *testing.T) {
	proof, parts, mapping := proofIntProofFixture(t)
	for _, pudlak := range []bool{false, true} {
		inter, err := ProofIntInterpolateM114P(proof, parts, mapping, pudlak, ProofIntLimits{})
		if err != nil {
			t.Fatal(pudlak, err)
		}
		valid, err := ProofIntVerifyInterpolant(proof, parts, mapping, inter, ProofIntLimits{})
		if err != nil || !valid {
			t.Fatal(pudlak, valid, err)
		}
	}
	corrupt := cloneM114PProof(proof)
	corrupt.Chains[len(corrupt.Chains)-1].ClauseIDs[0] = 1 << 20
	if got, err := ProofIntInterpolateM114P(corrupt, parts, mapping, false, ProofIntLimits{}); got != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(got, err)
	}
	if got, err := ProofIntInterpolateM114P(proof, parts, []int{-3, 0}, false, ProofIntLimits{}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatal("invalid variable map", got, err)
	}
	if got, err := ProofIntInterpolateM114P(proof, parts, mapping, false, ProofIntLimits{MaxProofClauses: 1}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("proof clause limit", got, err)
	}
	if got, err := ProofIntInterpolateM114P(proof, parts, mapping, false, ProofIntLimits{MaxAssignments: 1}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("proof replay limit", got, err)
	}
}

func TestProofIntStoreInterpolationFoundation(t *testing.T) {
	store, err := NewSatProofStore(SatProofStoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	x, _ := MakeSatLit(0, false)
	y, _ := MakeSatLit(1, false)
	for _, clause := range [][]SatLit{{x, y}, {x, y.Not()}} {
		if _, err = store.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	if err = store.MarkClausesA(); err != nil {
		t.Fatal(err)
	}
	for _, clause := range [][]SatLit{{x.Not(), y}, {x.Not(), y.Not()}} {
		if _, err = store.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	if err = store.MarkRoots(); err != nil {
		t.Fatal(err)
	}
	if _, err = store.AddClause([]SatLit{x}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.AddClause(nil); err != nil {
		t.Fatal(err)
	}
	for _, pudlak := range []bool{false, true} {
		result, e := ProofIntInterpolateStore(store, pudlak, ProofIntLimits{})
		if e != nil || result.Graph == nil || result.Graph.OutputCount() != 1 {
			t.Fatal(pudlak, result, e)
		}
	}
}

func TestProofIntManagerLimitsCancellationAtomicityOwnership(t *testing.T) {
	g := proofIntSequentialFixture(t, true)
	before := g.Clone()
	m, err := NewProofIntManager(g, ProofIntParams{MaxFrames: 4}, ProofIntLimits{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Perform()
	if err != nil || !r.Proved || r.Invariant == nil {
		t.Fatal(r, err)
	}
	owned := m.Invariant()
	if owned == nil {
		t.Fatal("missing invariant")
	}
	owned.Name = "changed"
	if m.Invariant().Name == "changed" {
		t.Fatal("borrowed invariant")
	}
	if err = g.Check(); err != nil || g.ObjectCount() != before.ObjectCount() {
		t.Fatal("source mutated", err)
	}
	if err = m.Clean(); err != nil || m.Frame() != 0 || m.Invariant() != nil {
		t.Fatal(err)
	}
	m.Close()
	m.Close()
	if _, err = m.Perform(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = ProofIntContainment(proofIntComb(t, '&'), proofIntComb(t, '|'), ProofIntLimits{MaxAssignments: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = ProofIntCounterexample(proofIntSequentialFixture(t, false), 2, ProofIntLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if got, err := ProofIntFrames(g, 2, false, false, ProofIntLimits{MaxObjects: 1}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	if got, err := ProofIntDuplicateTransition(g, ProofIntLimits{MaxObjects: 1}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("duplicate object limit", got, err)
	}
}

func TestProofIntBoundedModelCampaign(t *testing.T) {
	for execution := 0; execution < 5000; execution++ {
		a, b := byte('&'), byte('|')
		if execution&1 != 0 {
			a, b = b, a
		}
		first, second := proofIntComb(t, a), proofIntComb(t, b)
		got, err := ProofIntContainment(first, second, ProofIntLimits{})
		if err != nil {
			t.Fatal(execution, err)
		}
		want := a == '&' && b == '|' || a == b
		if got != want {
			t.Fatal(execution, got, want)
		}
	}
}

func TestProofIntIndependentTransitionFrameModel5000(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		property, nextTable := uint8(state&15), uint8(state>>4&15)
		frames := 1 + int(state>>8&3)
		graph := proofIntTruthTransition(t, property, nextTable)
		reachable := []bool{true, false}
		wantBad := false
		for frame := 0; frame < frames; frame++ {
			nextReachable := []bool{false, false}
			for current, present := range reachable {
				if !present {
					continue
				}
				for input := 0; input < 2; input++ {
					index := input | current<<1
					wantBad = wantBad || property&(1<<index) != 0
					nextReachable[(nextTable>>index)&1] = true
				}
			}
			reachable = nextReachable
		}
		counterexample, err := ProofIntCounterexample(graph, frames, ProofIntLimits{MaxAssignments: 256, MaxStates: 8})
		if err != nil || (counterexample != nil) != wantBad {
			t.Fatalf("execution=%d property=%x next=%x frames=%d got=%v wantBad=%v err=%v", execution, property, nextTable, frames, counterexample, wantBad, err)
		}
		if counterexample != nil {
			if err = graph.ReplayCounterexample(counterexample); err != nil {
				t.Fatal(execution, err)
			}
		}
		bmc, err := ProofIntBMCFrames(graph, frames, ProofIntLimits{MaxFrames: 8})
		if err != nil {
			t.Fatal(execution, err)
		}
		inputs := make([]bool, frames)
		modelState := 0
		for frame := range inputs {
			inputs[frame] = state>>uint(16+frame)&1 != 0
			if frame+1 < frames {
				index := int(state>>uint(16+frame)&1) | modelState<<1
				modelState = int(nextTable>>index) & 1
			}
		}
		lastInput := 0
		if inputs[frames-1] {
			lastInput = 1
		}
		wantLast := property&(1<<(lastInput|modelState<<1)) != 0
		got, evalErr := bmc.Evaluate(inputs)
		if evalErr != nil || got[0] != wantLast {
			t.Fatalf("execution=%d BMC got=%v want=%v err=%v", execution, got, wantLast, evalErr)
		}
	}
}

func FuzzProofInt(f *testing.F) {
	f.Add(uint8(0), false)
	f.Add(uint8(255), true)
	f.Fuzz(func(t *testing.T, seed uint8, pudlak bool) {
		first := proofIntComb(t, []byte{'&', '|', '^'}[seed%3])
		second := first.Clone()
		equal, err := ProofIntEquivalent(first, second, ProofIntLimits{MaxAssignments: 16})
		if err != nil || !equal {
			t.Fatal(err, equal)
		}
		proof, parts, mapping := proofIntProofFixture(t)
		inter, err := ProofIntInterpolateM114P(proof, parts, mapping, pudlak, ProofIntLimits{})
		if err != nil {
			t.Fatal(err)
		}
		ok, err := ProofIntVerifyInterpolant(proof, parts, mapping, inter, ProofIntLimits{MaxAssignments: 16})
		if err != nil || !ok {
			t.Fatal(ok, err)
		}
	})
}
