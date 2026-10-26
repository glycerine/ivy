package abc

// Standalone tests for the remaining src/proof/pdr boundaries. The reference
// checks below enumerate Boolean states directly and use no C implementation.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func pdrRemainingCube(t testing.TB, values ...int) *PdrCube {
	t.Helper()
	lits := make([]PdrLit, len(values))
	for i, value := range values {
		lit, err := MakePdrLit(uint32(value>>1), value&1 != 0)
		if err != nil {
			t.Fatal(err)
		}
		lits[i] = lit
	}
	cube, err := NewPdrCube(lits, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cube
}

func TestPdrRemainingInventoryAndEquivalence(t *testing.T) {
	units := PdrRemainingUnits()
	want := []string{"pdr.h", "pdrClass.c", "pdrCnf.c", "pdrIncr.c", "pdrTsim.c", "pdrTsim2.c", "pdrTsim3.c", "module.make"}
	for i := range want {
		if units[i].Source != want[i] || units[i].Disposition == "" {
			t.Fatalf("unit %d: %+v", i, units[i])
		}
	}
	if units[2].Portable {
		t.Fatal("native CNF construction must remain explicit")
	}
	rows := [][]bool{{false, true, false, true, true}, {true, false, true, false, false}, {true, true, true, false, true}}
	p, stats, err := NewPdrEquivalencePartition(rows, 5, PdrRemainingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Check(); err != nil {
		t.Fatal(err)
	}
	if err = p.CheckObservations(rows); err != nil {
		t.Fatal(err)
	}
	if got, expected := p.Representatives(), []int{0, 1, 0, 3, 1}; !slices.Equal(got, expected) {
		t.Fatalf("representatives %v", got)
	}
	if got, expected := p.Classes(), [][]int{{0, 2}, {1, 4}, {3}}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("classes %v", got)
	}
	classes := p.Classes()
	classes[0][0] = 4
	if p.Classes()[0][0] != 0 {
		t.Fatal("classes leaked ownership")
	}
	p.representative[2] = 2
	if err = p.CheckObservations(rows); !errors.Is(err, ErrInvariant) {
		t.Fatalf("observation mismatch: %v", err)
	}
	p.representative[2] = 0
	if err = p.CheckObservations(rows[:2]); !errors.Is(err, ErrInvariant) {
		t.Fatalf("incomplete observations changed equality: %v", err)
	}
	if err = p.CheckObservations([][]bool{{false}}); !errors.Is(err, ErrDimension) {
		t.Fatalf("observation dimension: %v", err)
	}
	if stats.Work == 0 || stats.Bytes == 0 {
		t.Fatalf("stats %+v", stats)
	}
	if _, _, err = NewPdrEquivalencePartition(rows, 5, PdrRemainingLimits{MaxWork: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("work limit: %v", err)
	}
	if _, _, err = NewPdrEquivalencePartition(rows, 5, PdrRemainingLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, _, err = NewPdrEquivalencePartition(rows[:1], 4, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension: %v", err)
	}
}

func TestPdrClauseArchiveTransactionsOwnershipAndSubsumption(t *testing.T) {
	if _, err := NewPdrClauseArchive(4, PdrRemainingLimits{MaxFrames: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("frame limit: %v", err)
	}
	a, err := NewPdrClauseArchive(4, PdrRemainingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	specific := pdrRemainingCube(t, 0, 2)
	strong := pdrRemainingCube(t, 0)
	stateLit, _ := MakePdrLit(0, false)
	inputLit, _ := MakePdrLit(0, false)
	withInput, _ := NewPdrCube([]PdrLit{stateLit}, []PdrLit{inputLit})
	if err = a.Add(1, withInput, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("archive accepted input literal: %v", err)
	}
	if err = a.Add(1, specific, PdrRemainingLimits{}); err != nil {
		t.Fatal(err)
	}
	if err = a.Add(1, strong, PdrRemainingLimits{}); err != nil {
		t.Fatal(err)
	}
	if got := a.Stats(); got.Clauses != 1 || got.Cubes != 1 {
		t.Fatalf("live archive stats compounded: %+v", got)
	}
	if got := a.Frames()[1]; len(got) != 1 || !reflect.DeepEqual(got[0].StateLiterals(), strong.StateLiterals()) {
		t.Fatalf("subsumption: %v", got)
	}
	// A weaker duplicate clause cannot replace the retained stronger clause.
	if err = a.Add(1, specific, PdrRemainingLimits{}); err != nil || len(a.Frames()[1]) != 1 {
		t.Fatalf("redundant add: %v", err)
	}
	before := a.Frames()
	beforeStats := a.Stats()
	if err = a.Add(2, strong, PdrRemainingLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("add cancellation: %v", err)
	}
	if !reflect.DeepEqual(a.Frames(), before) || a.Stats() != beforeStats {
		t.Fatal("cancelled add was not atomic")
	}
	if err = a.Push(1, 3, strong, PdrRemainingLimits{MaxClauses: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("push limit: %v", err)
	}
	if !reflect.DeepEqual(a.Frames(), before) || a.Stats() != beforeStats {
		t.Fatal("failed push was not atomic")
	}
	if err = a.Push(1, 3, strong, PdrRemainingLimits{}); err != nil {
		t.Fatal(err)
	}
	if err = a.Check(); err != nil {
		t.Fatal(err)
	}
	saved, err := a.Save(true, PdrRemainingLimits{})
	if err != nil || len(saved) != 3 || len(saved[1]) != 1 || len(saved[2]) != 1 {
		t.Fatalf("drop-last save: %v %v", saved, err)
	}
	saved[1][0].state[0] ^= 1
	if err = a.Check(); err != nil {
		t.Fatalf("save leaked ownership: %v", err)
	}
	if _, err = a.Save(false, PdrRemainingLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("save cancellation: %v", err)
	}
	snapshot := a.Frames()
	snapshot[1][0].state[0] ^= 1
	if err = a.Restore(snapshot, PdrRemainingLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("tampered restore: %v", err)
	}
	if err = a.Restore([][]*PdrCube{{nil}}, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil restore cube: %v", err)
	}
	if err = a.Restore([][]*PdrCube{{withInput}}, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("input-bearing restore cube: %v", err)
	}
	if err = a.Restore(before, PdrRemainingLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("restore limit: %v", err)
	}
	if len(a.Frames()[3]) != 1 {
		t.Fatal("failed restore mutated archive")
	}
	owned := a.Frames()
	owned[1] = nil
	if len(a.Frames()[1]) != 1 {
		t.Fatal("frames leaked ownership")
	}
}

func pdrRemainingSystem() *PdrExplicitTransitionSystem {
	// state' = state XOR input; bad is asserted by state=1,input=0.
	return &PdrExplicitTransitionSystem{StateBits: 1, InputBits: 1, Initial: []uint64{0}, Next: []uint64{0, 1, 1, 0}, Bad: []bool{false, false, true, false}}
}

func TestPdrExplicitInvariantWitnessAndTampering(t *testing.T) {
	m := pdrRemainingSystem()
	falseState, err := NewPdrCube([]PdrLit{1}, nil) // blocks state zero, therefore invalid initiation
	if err != nil {
		t.Fatal(err)
	}
	inv, _ := NewPdrInvariant([]*PdrCube{falseState})
	result, err := m.VerifyInvariant(inv, PdrRemainingLimits{})
	if err != nil || result.Valid || result.Phase != PdrExplicitInitiation {
		t.Fatalf("initiation %+v %v", result, err)
	}
	blockOne := pdrRemainingCube(t, 0)
	inv, _ = NewPdrInvariant([]*PdrCube{blockOne})
	result, err = m.VerifyInvariant(inv, PdrRemainingLimits{})
	if err != nil || result.Valid || result.Phase != PdrExplicitInduction {
		t.Fatalf("induction %+v %v", result, err)
	}
	trueInv, _ := NewPdrInvariant(nil)
	result, err = m.VerifyInvariant(trueInv, PdrRemainingLimits{})
	if err != nil || result.Valid || result.Phase != PdrExplicitSafety {
		t.Fatalf("safety %+v %v", result, err)
	}
	safe := &PdrExplicitTransitionSystem{StateBits: 1, InputBits: 0, Initial: []uint64{0}, Next: []uint64{0, 1}, Bad: []bool{false, true}}
	result, err = safe.VerifyInvariant(inv, PdrRemainingLimits{})
	if err != nil || !result.Valid || result.Phase != PdrExplicitComplete {
		t.Fatalf("valid %+v %v", result, err)
	}
	tampered, _ := NewPdrInvariant([]*PdrCube{blockOne})
	tampered.clauses[0].signature ^= 1
	if _, err = safe.VerifyInvariant(tampered, PdrRemainingLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("tampered invariant: %v", err)
	}
	tampered, _ = NewPdrInvariant([]*PdrCube{blockOne})
	tampered.clauses[0].inputs = []PdrLit{0}
	if _, err = safe.VerifyInvariant(tampered, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("input-bearing invariant: %v", err)
	}
	valid, stats, err := m.VerifyWitness([]uint64{0, 1, 1}, []uint64{1, 0}, PdrRemainingLimits{})
	if err != nil || !valid || stats.Depth != 2 {
		t.Fatalf("witness %v %+v %v", valid, stats, err)
	}
	for _, trace := range [][]uint64{{1, 1, 1}, {0, 0, 1}} {
		valid, _, err = m.VerifyWitness(trace, []uint64{1, 0}, PdrRemainingLimits{})
		if err != nil || valid {
			t.Fatalf("tamper %v: %v %v", trace, valid, err)
		}
	}
	valid, _, err = m.VerifyWitness([]uint64{0, 1, 1}, []uint64{0, 0}, PdrRemainingLimits{})
	if err != nil || valid {
		t.Fatalf("input tamper: %v %v", valid, err)
	}
	if _, _, err = m.VerifyWitness([]uint64{0, 1, 1}, []uint64{1, 0}, PdrRemainingLimits{MaxDepth: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("depth: %v", err)
	}
	if _, _, err = m.VerifyWitness([]uint64{0, 1, 1}, []uint64{1, 0}, PdrRemainingLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, _, err = m.VerifyWitness([]uint64{2, 1}, []uint64{0}, PdrRemainingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("out-of-range first state: %v", err)
	}
}

func TestPdrExplicitGeneralizationExactAndLimits(t *testing.T) {
	m := &PdrExplicitTransitionSystem{StateBits: 2, InputBits: 1, Initial: []uint64{0}, Next: make([]uint64, 8), Bad: make([]bool, 8)}
	// next bit zero is state bit zero; bit one is always zero. Input and state
	// bit one are therefore don't-cares for target next bit zero = one.
	for state := 0; state < 4; state++ {
		for input := 0; input < 2; input++ {
			m.Next[state*2+input] = uint64(state & 1)
		}
	}
	target := pdrRemainingCube(t, 0)
	cube, stats, err := m.GeneralizePredecessor(1, 1, target, PdrRemainingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got := cube.StateLiterals(); len(got) != 1 || got[0] != 0 || len(cube.InputLiterals()) != 0 {
		t.Fatalf("generalized %v %v", got, cube.InputLiterals())
	}
	if stats.Work == 0 || stats.Cubes != 4 {
		t.Fatalf("stats %+v", stats)
	}
	for state := uint64(0); state < 4; state++ {
		matches, _ := cube.Matches(pdrBits(state, 2))
		if matches != (state&1 != 0) {
			t.Fatalf("state %d match %v", state, matches)
		}
	}
	if _, _, err = m.GeneralizePredecessor(0, 0, target, PdrRemainingLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("non-predecessor: %v", err)
	}
	if _, _, err = m.GeneralizePredecessor(1, 1, target, PdrRemainingLimits{MaxWork: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("work: %v", err)
	}
	if _, _, err = m.GeneralizePredecessor(1, 1, target, PdrRemainingLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("bytes: %v", err)
	}
	for name, limits := range map[string]PdrRemainingLimits{
		"nodes":       {MaxNodes: 3},
		"transitions": {MaxTransitions: 7},
		"cubes":       {MaxCubes: 1},
	} {
		if _, _, err = m.GeneralizePredecessor(1, 1, target, limits); !errors.Is(err, ErrLimit) {
			t.Fatalf("%s limit: %v", name, err)
		}
	}
}

// This deterministic campaign compares generalized cubes with an independent
// direct truth-table predicate over 5,000 generated transition systems.
func TestPdrExplicitGeneralizationDeterministic5000(t *testing.T) {
	seed := uint64(0x9e3779b97f4a7c15)
	for trial := 0; trial < 5000; trial++ {
		m := &PdrExplicitTransitionSystem{StateBits: 2, InputBits: 1, Initial: []uint64{0}, Next: make([]uint64, 8), Bad: make([]bool, 8)}
		for i := range m.Next {
			seed = seed*6364136223846793005 + 1442695040888963407
			m.Next[i] = seed & 3
		}
		state, input := uint64(trial&3), uint64((trial>>2)&1)
		next := m.Next[int(state)*2+int(input)]
		lit, _ := MakePdrLit(0, next&1 == 0)
		target, _ := NewPdrCube([]PdrLit{lit}, nil)
		cube, _, err := m.GeneralizePredecessor(state, input, target, PdrRemainingLimits{MaxWork: 128})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		for s := uint64(0); s < 4; s++ {
			stateMatch, _ := cube.Matches(pdrBits(s, 2))
			if !stateMatch {
				continue
			}
			for in := uint64(0); in < 2; in++ {
				if !pdrInputMatches(cube.inputs, in) {
					continue
				}
				got := m.Next[int(s)*2+int(in)]&1 != 0
				if got != (next&1 != 0) {
					t.Fatalf("trial %d cube accepts %d/%d", trial, s, in)
				}
			}
		}
		// Determinism includes literal-removal order and result ownership.
		again, _, err := m.GeneralizePredecessor(state, input, target, PdrRemainingLimits{MaxWork: 128})
		if err != nil || !reflect.DeepEqual(cube, again) {
			t.Fatalf("trial %d nondeterministic", trial)
		}
	}
}

func FuzzPdrExplicitGeneralization(f *testing.F) {
	for i := uint64(0); i < 32; i++ {
		f.Add(i, i*0x9e3779b97f4a7c15)
	}
	f.Fuzz(func(t *testing.T, selector, table uint64) {
		m := &PdrExplicitTransitionSystem{StateBits: 2, InputBits: 1, Initial: []uint64{0}, Next: make([]uint64, 8), Bad: make([]bool, 8)}
		for i := range m.Next {
			m.Next[i] = (table >> (2 * i)) & 3
		}
		state, input := selector&3, (selector>>2)&1
		expected := m.Next[int(state)*2+int(input)]&1 != 0
		lit, _ := MakePdrLit(0, !expected)
		target, _ := NewPdrCube([]PdrLit{lit}, nil)
		cube, _, err := m.GeneralizePredecessor(state, input, target, PdrRemainingLimits{MaxWork: 128})
		if err != nil {
			t.Fatal(err)
		}
		for s := uint64(0); s < 4; s++ {
			match, _ := cube.Matches(pdrBits(s, 2))
			for in := uint64(0); match && in < 2; in++ {
				if pdrInputMatches(cube.inputs, in) && (m.Next[int(s)*2+int(in)]&1 != 0) != expected {
					t.Fatalf("accepted counterexample %d/%d", s, in)
				}
			}
		}
	})
}
