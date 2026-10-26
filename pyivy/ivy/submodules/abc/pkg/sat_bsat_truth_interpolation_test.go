package abc

// Stand-alone tests for the satInter truth-table interpolation port.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satTruthInterpolationStore(t testing.TB, clausesA, clausesB, learned [][]SatLit) *SatProofStore {
	t.Helper()
	store, err := NewSatProofStore(SatProofStoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, clause := range clausesA {
		if _, err = store.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	if err = store.MarkClausesA(); err != nil {
		t.Fatal(err)
	}
	for _, clause := range clausesB {
		if _, err = store.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	if err = store.MarkRoots(); err != nil {
		t.Fatal(err)
	}
	for _, clause := range learned {
		if _, err = store.AddClause(clause); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = store.AddClause(nil); err != nil {
		t.Fatal(err)
	}
	return store
}

func satTruthInterpolationCheckCraig(t testing.TB, store *SatProofStore, result SatTruthInterpolation) {
	t.Helper()
	if store.variables > 20 {
		t.Fatal("test model is too large")
	}
	for assignment := 0; assignment < 1<<store.variables; assignment++ {
		index := uint64(0)
		for bit, variable := range result.Variables {
			if assignment&(1<<variable) != 0 {
				index |= uint64(1) << bit
			}
		}
		interpolant, err := result.Evaluate(index)
		if err != nil {
			t.Fatal(err)
		}
		satisfies := func(first, last int) bool {
			for clauseID := first; clauseID < last; clauseID++ {
				clauseTrue := false
				for _, literal := range store.clauses[clauseID].Literals {
					value := assignment&(1<<literal.Var()) != 0
					if literal.Negative() {
						value = !value
					}
					clauseTrue = clauseTrue || value
				}
				if !clauseTrue {
					return false
				}
			}
			return true
		}
		if satisfies(0, store.clausesA) && !interpolant {
			t.Fatalf("A does not imply interpolant at assignment %0*b", store.variables, assignment)
		}
		if satisfies(store.clausesA, store.roots) && interpolant {
			t.Fatalf("B and interpolant are satisfiable at assignment %0*b", store.variables, assignment)
		}
	}
}

func TestSatTruthInterpolationSharedUnit(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	store := satTruthInterpolationStore(t, [][]SatLit{{x}}, [][]SatLit{{x.Not()}}, nil)
	manager, err := NewSatTruthInterpolator(SatTruthInterpolationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := manager.Interpolate(store)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Variables, []SatVar{0}) || result.Bits != 2 || !reflect.DeepEqual(result.Words, []uint64{2}) {
		t.Fatalf("result = %+v", result)
	}
	for assignment, want := range []bool{false, true} {
		got, evalErr := result.Evaluate(uint64(assignment))
		if evalErr != nil || got != want {
			t.Fatalf("evaluate(%d) = %v, %v; want %v", assignment, got, evalErr, want)
		}
	}
	formatted, err := FormatSatTruthInterpolationClause(2, result)
	if err != nil || formatted != "Clause  2 :  10\n" {
		t.Fatalf("format = %q, err=%v", formatted, err)
	}
	satTruthInterpolationCheckCraig(t, store, result)
	stats := manager.Stats()
	if stats.Variables != 1 || stats.Roots != 2 || stats.Clauses != 3 || stats.Globals != 1 || stats.Words != 1 || stats.ResolutionSteps != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestSatTruthInterpolationLearnedReplayAndExplicitOrder(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	store := satTruthInterpolationStore(t,
		[][]SatLit{{x, y}, {x.Not(), y}},
		[][]SatLit{{x, y.Not()}, {x.Not(), y.Not()}},
		[][]SatLit{{y}},
	)
	manager, _ := NewSatTruthInterpolator(SatTruthInterpolationOptions{})
	result, err := manager.Interpolate(store)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Variables, []SatVar{0, 1}) || !reflect.DeepEqual(result.Words, []uint64{12}) {
		t.Fatalf("inferred result = %+v", result)
	}
	satTruthInterpolationCheckCraig(t, store, result)
	if err = manager.SetGlobalVariables([]SatVar{1, 0}); err != nil {
		t.Fatal(err)
	}
	explicit, err := manager.Interpolate(store)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(explicit.Variables, []SatVar{1, 0}) || !reflect.DeepEqual(explicit.Words, []uint64{10}) {
		t.Fatalf("explicit result = %+v", explicit)
	}
	satTruthInterpolationCheckCraig(t, store, explicit)
	explicit.Words[0] = 0
	again, err := manager.Interpolate(store)
	if err != nil || !reflect.DeepEqual(again.Words, []uint64{10}) {
		t.Fatalf("result aliases manager storage: %+v, err=%v", again, err)
	}
}

func TestSatTruthInterpolationConstantPartitions(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	tests := []struct {
		name string
		a, b [][]SatLit
		want uint64
	}{
		{"unsat A", [][]SatLit{{x}, {x.Not()}}, nil, 0},
		{"unsat B", nil, [][]SatLit{{x}, {x.Not()}}, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := satTruthInterpolationStore(t, test.a, test.b, nil)
			manager, _ := NewSatTruthInterpolator(SatTruthInterpolationOptions{})
			result, err := manager.Interpolate(store)
			if err != nil || len(result.Variables) != 0 || result.Bits != 1 || !reflect.DeepEqual(result.Words, []uint64{test.want}) {
				t.Fatalf("result = %+v, err=%v", result, err)
			}
			satTruthInterpolationCheckCraig(t, store, result)
		})
	}
}

func TestSatTruthInterpolationEightGlobalsAndFourWords(t *testing.T) {
	x0 := satStoreTestLit(t, 0, false)
	x7 := satStoreTestLit(t, 7, false)
	store := satTruthInterpolationStore(t, [][]SatLit{{x0, x7}, {x0}}, [][]SatLit{{x0.Not()}}, nil)
	globals := make([]SatVar, satTruthInterpolationMaxGlobals)
	for i := range globals {
		globals[i] = SatVar(i)
	}
	manager, err := NewSatTruthInterpolator(SatTruthInterpolationOptions{GlobalVariables: globals})
	if err != nil {
		t.Fatal(err)
	}
	result, err := manager.Interpolate(store)
	if err != nil {
		t.Fatal(err)
	}
	if result.Bits != 256 || len(result.Words) != 4 || !reflect.DeepEqual(result.Variables, globals) {
		t.Fatalf("result shape = %+v", result)
	}
	for _, assignment := range []uint64{0, 1, 2, 127, 128, 255} {
		got, evalErr := result.Evaluate(assignment)
		if evalErr != nil || got != (assignment&1 != 0) {
			t.Fatalf("evaluate(%d) = %v, %v", assignment, got, evalErr)
		}
	}
	formatted, err := FormatSatTruthBits(result.Words, result.Bits)
	if err != nil || len(formatted) != 256 || formatted[:4] != "1010" || formatted[len(formatted)-4:] != "1010" {
		t.Fatalf("formatted len=%d prefix/suffix=%q/%q err=%v", len(formatted), formatted[:4], formatted[len(formatted)-4:], err)
	}
	satTruthInterpolationCheckCraig(t, store, result)
}

func TestSatTruthInterpolationErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	store := satTruthInterpolationStore(t, [][]SatLit{{x}}, [][]SatLit{{x.Not()}}, nil)
	if _, err := NewSatTruthInterpolator(SatTruthInterpolationOptions{MaxTruthWords: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative words: %v", err)
	}
	if _, err := NewSatTruthInterpolator(SatTruthInterpolationOptions{GlobalVariables: []SatVar{0, 0}}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate globals: %v", err)
	}
	tooMany := make([]SatVar, satTruthInterpolationMaxGlobals+1)
	for i := range tooMany {
		tooMany[i] = SatVar(i)
	}
	if _, err := NewSatTruthInterpolator(SatTruthInterpolationOptions{GlobalVariables: tooMany}); !errors.Is(err, ErrLimit) {
		t.Fatalf("too many globals: %v", err)
	}
	for _, test := range []struct {
		name    string
		options SatTruthInterpolationOptions
		want    error
	}{
		{"truth words", SatTruthInterpolationOptions{MaxTruthWords: 2}, ErrLimit},
		{"replay clauses", SatTruthInterpolationOptions{Replay: SatProofCoreLimits{MaxClauses: 2}}, ErrLimit},
		{"cancel", SatTruthInterpolationOptions{Replay: SatProofCoreLimits{Stop: func() bool { return true }}}, ErrCancelled},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager, err := NewSatTruthInterpolator(test.options)
			if err != nil {
				t.Fatal(err)
			}
			_, err = manager.Interpolate(store)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	manager, _ := NewSatTruthInterpolator(SatTruthInterpolationOptions{})
	if err := manager.SetGlobalVariables([]SatVar{2}); !errors.Is(err, ErrInvalidLiteral) {
		// Bounds are checked against a particular store during interpolation.
		if err != nil {
			t.Fatalf("set globals: %v", err)
		}
		if _, err = manager.Interpolate(store); !errors.Is(err, ErrInvalidLiteral) {
			t.Fatalf("out-of-store global: %v", err)
		}
	}
	if _, err := (SatTruthInterpolation{}).Evaluate(0); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty evaluate: %v", err)
	}
	if _, err := FormatSatTruthBits(nil, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad bits: %v", err)
	}
	if _, err := FormatSatTruthInterpolationClause(-1, SatTruthInterpolation{Bits: 1, Words: []uint64{0}}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad clause: %v", err)
	}
	manager, _ = NewSatTruthInterpolator(SatTruthInterpolationOptions{})
	manager.Stop()
	manager.Stop()
	if !manager.Stats().Stopped {
		t.Fatal("manager did not stop")
	}
	if _, err := manager.Interpolate(store); !errors.Is(err, ErrInvariant) {
		t.Fatalf("use after stop: %v", err)
	}
}

func FuzzSatTruthInterpolationCraig(f *testing.F) {
	for _, seed := range []struct {
		a, b int
		mask byte
	}{{0, 1, 0}, {2, 5, 3}, {5, 0, 12}} {
		f.Add(seed.a, seed.b, seed.mask)
	}
	f.Fuzz(func(t *testing.T, first, second int, mask byte) {
		first = int(uint(first) % 6)
		second = int(uint(second) % 5)
		if second >= first {
			second++
		}
		a, err := MakeSatLit(SatVar(first), mask&1 != 0)
		if err != nil {
			t.Fatal(err)
		}
		b, err := MakeSatLit(SatVar(second), mask&2 != 0)
		if err != nil {
			t.Fatal(err)
		}
		clausesA := [][]SatLit{{a, b}, {a.Not(), b}}
		clausesB := [][]SatLit{{a, b.Not()}, {a.Not(), b.Not()}}
		if mask&4 != 0 {
			clausesA[0], clausesA[1] = clausesA[1], clausesA[0]
		}
		if mask&8 != 0 {
			clausesB[0], clausesB[1] = clausesB[1], clausesB[0]
		}
		store := satTruthInterpolationStore(t, clausesA, clausesB, [][]SatLit{{b}})
		options := SatTruthInterpolationOptions{}
		if mask&16 != 0 {
			options.GlobalVariables = []SatVar{b.Var(), a.Var()}
		}
		manager, err := NewSatTruthInterpolator(options)
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.Interpolate(store)
		if err != nil {
			t.Fatal(err)
		}
		satTruthInterpolationCheckCraig(t, store, result)
		position := slices.Index(result.Variables, b.Var())
		if position < 0 {
			t.Fatalf("shared variable %d absent from %v", b.Var(), result.Variables)
		}
		for assignment := uint64(0); assignment < uint64(result.Bits); assignment++ {
			got, evalErr := result.Evaluate(assignment)
			want := assignment&(uint64(1)<<position) != 0
			if b.Negative() {
				want = !want
			}
			if evalErr != nil || got != want {
				t.Fatalf("assignment=%d result=%v err=%v want=%v words=%x", assignment, got, evalErr, want, result.Words)
			}
		}
	})
}
