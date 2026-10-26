package abc

// Stand-alone tests for both satInterA and satInterB AIG interpolation rules.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satAigInterpolationCheckCraig(t testing.TB, store *SatProofStore, result SatAigInterpolation) {
	t.Helper()
	if store.variables > 20 {
		t.Fatal("test model is too large")
	}
	for assignment := 0; assignment < 1<<store.variables; assignment++ {
		inputs := make([]bool, len(result.Variables))
		for index, variable := range result.Variables {
			inputs[index] = assignment&(1<<variable) != 0
		}
		interpolant, err := result.Evaluate(inputs)
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

func TestSatAigInterpolationModesLearnedReplay(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	store := satTruthInterpolationStore(t,
		[][]SatLit{{x, y}, {x.Not(), y}},
		[][]SatLit{{x, y.Not()}, {x.Not(), y.Not()}},
		[][]SatLit{{y}},
	)
	for _, mode := range []SatAigInterpolationMode{SatAigInterpolationStandard, SatAigInterpolationMux} {
		t.Run([]string{"standard", "mux"}[mode], func(t *testing.T) {
			manager, err := NewSatAigInterpolator(SatAigInterpolationOptions{Mode: mode})
			if err != nil {
				t.Fatal(err)
			}
			result, err := manager.Interpolate(store)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Variables, []SatVar{0, 1}) || result.Graph.InputCount() != 2 || result.Graph.OutputCount() != 1 || result.Mode != mode {
				t.Fatalf("result shape: vars=%v inputs=%d outputs=%d mode=%d", result.Variables, result.Graph.InputCount(), result.Graph.OutputCount(), result.Mode)
			}
			for xv := 0; xv < 2; xv++ {
				for yv := 0; yv < 2; yv++ {
					got, evalErr := result.Evaluate([]bool{xv != 0, yv != 0})
					if evalErr != nil || got != (yv != 0) {
						t.Fatalf("evaluate(%d,%d) = %v, %v", xv, yv, got, evalErr)
					}
				}
			}
			satAigInterpolationCheckCraig(t, store, result)
			if err = result.Graph.Check(); err != nil {
				t.Fatal(err)
			}
			stats := manager.Stats()
			if stats.Variables != 2 || stats.Roots != 4 || stats.Clauses != 6 || stats.Globals != 2 || stats.ResolutionSteps != 3 || stats.AigObjects < 4 {
				t.Fatalf("stats = %+v", stats)
			}
		})
	}
}

func TestSatAigInterpolationLocalPivotRules(t *testing.T) {
	a := satStoreTestLit(t, 0, false)
	b := satStoreTestLit(t, 1, false)
	g := satStoreTestLit(t, 2, false)
	tests := []struct {
		name string
		a, b [][]SatLit
	}{
		{"A-local OR", [][]SatLit{{a, g}, {a.Not(), g}}, [][]SatLit{{g.Not()}}},
		{"B-local AND", [][]SatLit{{g}}, [][]SatLit{{b, g.Not()}, {b.Not(), g.Not()}}},
	}
	for _, test := range tests {
		for _, mode := range []SatAigInterpolationMode{SatAigInterpolationStandard, SatAigInterpolationMux} {
			t.Run(test.name+[]string{"/standard", "/mux"}[mode], func(t *testing.T) {
				store := satTruthInterpolationStore(t, test.a, test.b, nil)
				manager, _ := NewSatAigInterpolator(SatAigInterpolationOptions{Mode: mode})
				result, err := manager.Interpolate(store)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(result.Variables, []SatVar{2}) {
					t.Fatalf("globals = %v", result.Variables)
				}
				for value, want := range []bool{false, true} {
					got, evalErr := result.Evaluate([]bool{value != 0})
					if evalErr != nil || got != want {
						t.Fatalf("evaluate(%d) = %v, %v; want %v", value, got, evalErr, want)
					}
				}
				satAigInterpolationCheckCraig(t, store, result)
			})
		}
	}
}

func TestSatAigInterpolationConstantPartitions(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	for _, test := range []struct {
		name string
		a, b [][]SatLit
		want bool
	}{
		{"unsat A", [][]SatLit{{x}, {x.Not()}}, nil, false},
		{"unsat B", nil, [][]SatLit{{x}, {x.Not()}}, true},
	} {
		for _, mode := range []SatAigInterpolationMode{SatAigInterpolationStandard, SatAigInterpolationMux} {
			t.Run(test.name+[]string{"/standard", "/mux"}[mode], func(t *testing.T) {
				store := satTruthInterpolationStore(t, test.a, test.b, nil)
				manager, _ := NewSatAigInterpolator(SatAigInterpolationOptions{Mode: mode})
				result, err := manager.Interpolate(store)
				if err != nil {
					t.Fatal(err)
				}
				got, err := result.Evaluate(nil)
				if err != nil || got != test.want || len(result.Variables) != 0 {
					t.Fatalf("result=%v globals=%v err=%v want=%v", got, result.Variables, err, test.want)
				}
				satAigInterpolationCheckCraig(t, store, result)
			})
		}
	}
}

func TestSatAigInterpolationExplicitGlobalsOwnershipAndCoverage(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	store := satTruthInterpolationStore(t,
		[][]SatLit{{x, y}, {x.Not(), y}},
		[][]SatLit{{x, y.Not()}, {x.Not(), y.Not()}},
		[][]SatLit{{y}},
	)
	globals := []SatVar{1, 0}
	manager, _ := NewSatAigInterpolator(SatAigInterpolationOptions{GlobalVariables: globals})
	globals[0] = 0
	result, err := manager.Interpolate(store)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Variables, []SatVar{1, 0}) {
		t.Fatalf("globals alias caller: %v", result.Variables)
	}
	result.Variables[0] = 0
	again, err := manager.Interpolate(store)
	if err != nil || !reflect.DeepEqual(again.Variables, []SatVar{1, 0}) {
		t.Fatalf("result aliases manager: %v, err=%v", again.Variables, err)
	}
	if err = manager.SetGlobalVariables([]SatVar{1}); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Interpolate(store); !errors.Is(err, ErrInvariant) {
		t.Fatalf("missing actual shared variable: %v", err)
	}
	if err = manager.SetGlobalVariables([]SatVar{3}); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Interpolate(store); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("out-of-range global: %v", err)
	}
}

func TestSatAigInterpolationDerivePartitionCNF(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	z := satStoreTestLit(t, 2, false)
	w := satStoreTestLit(t, 3, false)
	store := satTruthInterpolationStore(t,
		[][]SatLit{{x, y}, {z.Not()}},
		[][]SatLit{{x.Not(), y}, {w}},
		[][]SatLit{{y}},
	)
	manager, _ := NewSatAigInterpolator(SatAigInterpolationOptions{})
	for _, partA := range []bool{true, false} {
		formula, err := manager.DerivePartitionCNF(store, partA)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(formula.SharedVariables, []SatVar{0, 1}) || formula.Graph.OutputCount() != 1 || formula.InputForVariable[2] != 5 || formula.InputForVariable[3] != 6 {
			t.Fatalf("formula shape: shared=%v outputs=%d", formula.SharedVariables, formula.Graph.OutputCount())
		}
		for assignment := 0; assignment < 1<<store.variables; assignment++ {
			inputs := make([]bool, formula.Graph.InputCount())
			for variable, input := range formula.InputForVariable {
				if input < len(inputs) {
					inputs[input] = assignment&(1<<variable) != 0
				}
			}
			outputs, evalErr := formula.Graph.Evaluate(inputs)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			first, last := 0, store.clausesA
			if !partA {
				first, last = store.clausesA, store.roots
			}
			want := true
			for clauseID := first; clauseID < last; clauseID++ {
				clauseValue := false
				for _, literal := range store.clauses[clauseID].Literals {
					value := assignment&(1<<literal.Var()) != 0
					clauseValue = clauseValue || (value != literal.Negative())
				}
				want = want && clauseValue
			}
			if outputs[0] != want {
				t.Fatalf("partA=%v assignment=%d got=%v want=%v", partA, assignment, outputs[0], want)
			}
		}
	}
}

func TestSatAigInterpolationErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	store := satTruthInterpolationStore(t, [][]SatLit{{x}}, [][]SatLit{{x.Not()}}, nil)
	for _, test := range []struct {
		name    string
		options SatAigInterpolationOptions
		want    error
	}{
		{"mode", SatAigInterpolationOptions{Mode: 9}, ErrDimension},
		{"negative objects", SatAigInterpolationOptions{MaxAigObjects: -1}, ErrDimension},
		{"duplicate globals", SatAigInterpolationOptions{GlobalVariables: []SatVar{0, 0}}, ErrInvariant},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewSatAigInterpolator(test.options); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	for _, test := range []struct {
		name    string
		options SatAigInterpolationOptions
		want    error
	}{
		{"object limit", SatAigInterpolationOptions{MaxAigObjects: 2}, ErrLimit},
		{"replay limit", SatAigInterpolationOptions{Replay: SatProofCoreLimits{MaxClauses: 2}}, ErrLimit},
		{"cancel", SatAigInterpolationOptions{Replay: SatProofCoreLimits{Stop: func() bool { return true }}}, ErrCancelled},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager, err := NewSatAigInterpolator(test.options)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = manager.Interpolate(store); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if _, err := (SatAigInterpolation{}).Evaluate(nil); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty result: %v", err)
	}
	manager, _ := NewSatAigInterpolator(SatAigInterpolationOptions{})
	manager.Stop()
	manager.Stop()
	if !manager.Stats().Stopped {
		t.Fatal("manager did not stop")
	}
	if _, err := manager.Interpolate(store); !errors.Is(err, ErrInvariant) {
		t.Fatalf("use after stop: %v", err)
	}
}

func FuzzSatAigInterpolationCraig(f *testing.F) {
	for _, seed := range []struct {
		first, second int
		mask          byte
	}{{0, 1, 0}, {2, 5, 7}, {5, 0, 31}} {
		f.Add(seed.first, seed.second, seed.mask)
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
		options := SatAigInterpolationOptions{Mode: SatAigInterpolationMode(mask >> 7)}
		if mask&16 != 0 {
			options.GlobalVariables = []SatVar{b.Var(), a.Var()}
		}
		manager, err := NewSatAigInterpolator(options)
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.Interpolate(store)
		if err != nil {
			t.Fatal(err)
		}
		satAigInterpolationCheckCraig(t, store, result)
		position := slices.Index(result.Variables, b.Var())
		if position < 0 {
			t.Fatalf("shared variable absent from %v", result.Variables)
		}
		for assignment := 0; assignment < 1<<len(result.Variables); assignment++ {
			inputs := make([]bool, len(result.Variables))
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			got, evalErr := result.Evaluate(inputs)
			want := inputs[position]
			if b.Negative() {
				want = !want
			}
			if evalErr != nil || got != want {
				t.Fatalf("mode=%d assignment=%d got=%v err=%v want=%v", result.Mode, assignment, got, evalErr, want)
			}
		}
	})
}
