package abc

// Stand-alone tests for the satInterP proof replay and UNSAT-core port.

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func satProofCoreStore(t testing.TB, roots, learned [][]SatLit) *SatProofStore {
	t.Helper()
	store, err := NewSatProofStore(SatProofStoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, clause := range roots {
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

func TestSatProofCoreOppositeRootUnits(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	store := satProofCoreStore(t, [][]SatLit{{x}, {x.Not()}}, nil)
	manager, err := NewSatProofCoreManager(SatProofCoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	core, err := manager.UnsatCore(store, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{1, 0}; !reflect.DeepEqual(core, want) {
		t.Fatalf("root core = %v, want %v", core, want)
	}
	if err = VerifySatProofRootCore(store, core); err != nil {
		t.Fatalf("verify root core: %v", err)
	}
	learned, err := manager.UnsatCore(store, true)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{2}; !reflect.DeepEqual(learned, want) {
		t.Fatalf("learned core = %v, want %v", learned, want)
	}
	stats := manager.Stats()
	if stats.Variables != 1 || stats.Roots != 2 || stats.Clauses != 3 || stats.ResolutionSteps != 1 || stats.AntecedentReferences != 2 || stats.CoreSize != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestSatProofCoreLearnedUnitReplay(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	store := satProofCoreStore(t, [][]SatLit{
		{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()},
	}, [][]SatLit{{y}})
	manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
	core, err := manager.UnsatCore(store, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{3, 2, 1, 0}; !reflect.DeepEqual(core, want) {
		t.Fatalf("root core = %v, want %v", core, want)
	}
	if err = VerifySatProofRootCore(store, core); err != nil {
		t.Fatal(err)
	}
	learned, err := manager.UnsatCore(store, true)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{4, 5}; !reflect.DeepEqual(learned, want) {
		t.Fatalf("learned core = %v, want %v", learned, want)
	}
	formatted, err := FormatSatProofLearnedCoreForBMC(store, learned, []SatBMCVariable{{Object: 10, Frame: 2}, {Object: 7, Frame: 3}})
	if err != nil {
		t.Fatal(err)
	}
	want := "UNSAT contains 2 learned clauses:\n" +
		"     0 :      0 : 7(3) \n" +
		"     1 :      1 : Empty\n"
	if formatted != want {
		t.Fatalf("formatted core:\n%q\nwant:\n%q", formatted, want)
	}
}

func TestSatProofCoreStrongerResolventKeepsStoreOwned(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, false)
	z := satStoreTestLit(t, 2, false)
	store := satProofCoreStore(t, [][]SatLit{
		{x, y}, {x.Not(), y}, {x, y.Not()}, {x.Not(), y.Not()},
	}, [][]SatLit{{y, z}})
	manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
	core, err := manager.UnsatCore(store, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(core) != 4 {
		t.Fatalf("core = %v", core)
	}
	if err = VerifySatProofRootCore(store, core); err != nil {
		t.Fatal(err)
	}
	clause, err := store.Clause(4)
	if err != nil || !reflect.DeepEqual(clause.Literals, []SatLit{y, z}) {
		t.Fatalf("source clause was mutated: %v, err=%v", clause.Literals, err)
	}
}

func TestSatProofCoreIterativeTraversalDepth(t *testing.T) {
	const depth = 10000
	manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
	manager.stats = SatProofCoreStats{Roots: 1, Clauses: depth + 1}
	run := &satProofCoreRun{
		manager:       manager,
		clauses:       make([]satProofCoreClause, depth+1),
		antecedents:   make([][]int, depth),
		antecedentSet: make([]bool, depth),
		empty:         depth,
	}
	for clause := 1; clause <= depth; clause++ {
		run.antecedents[clause-1] = []int{clause - 1}
		run.antecedentSet[clause-1] = true
	}
	core, err := run.core(false)
	if err != nil || !reflect.DeepEqual(core, []int{0}) {
		t.Fatalf("root chain core = %v, err=%v", core, err)
	}
	core, err = run.core(true)
	if err != nil || len(core) != depth || core[0] != 1 || core[len(core)-1] != depth {
		t.Fatalf("learned chain len=%d endpoints=%v/%v err=%v", len(core), core[0], core[len(core)-1], err)
	}
}

func TestSatProofCoreExactDebugFormatting(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	y := satStoreTestLit(t, 1, true)
	clause := SatProofStoreClause{ID: 7, Literals: []SatLit{x, y}}
	formatted, err := FormatSatProofCoreClause(clause, 11)
	if err != nil || formatted != "Clause ID = 7. Proof = 11. { 0 3 }\n" {
		t.Fatalf("clause format = %q, err=%v", formatted, err)
	}
	formatted, err = FormatSatProofCoreResolvent([]SatLit{y, x})
	if err != nil || formatted != "Resolvent: { 3 0 }\n" {
		t.Fatalf("resolvent format = %q, err=%v", formatted, err)
	}
	formatted, err = FormatSatProofCoreInterpolantClause(7)
	if err != nil || formatted != "Clause  7 :  \n" {
		t.Fatalf("interpolant format = %q, err=%v", formatted, err)
	}
	if _, err = FormatSatProofCoreClause(SatProofStoreClause{ID: -1}, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad clause format: %v", err)
	}
	if _, err = FormatSatProofCoreInterpolantClause(-1); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad interpolant format: %v", err)
	}
}

func TestSatProofCoreErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	x := satStoreTestLit(t, 0, false)
	valid := satProofCoreStore(t, [][]SatLit{{x}, {x.Not()}}, nil)
	tests := []struct {
		name   string
		limits SatProofCoreLimits
		want   error
	}{
		{"variables", SatProofCoreLimits{MaxVariables: 0}, nil},
		{"variable limit", SatProofCoreLimits{MaxVariables: 0}, nil},
		{"clause limit", SatProofCoreLimits{MaxClauses: 2}, ErrLimit},
		{"literal limit", SatProofCoreLimits{MaxLiterals: 1}, ErrLimit},
		{"resolution limit", SatProofCoreLimits{MaxResolutionSteps: 1}, nil},
		{"antecedent limit", SatProofCoreLimits{MaxAntecedentReferences: 1}, ErrLimit},
		{"cancel", SatProofCoreLimits{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager, err := NewSatProofCoreManager(test.limits)
			if err != nil {
				t.Fatal(err)
			}
			_, err = manager.UnsatCore(valid, false)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if _, err := NewSatProofCoreManager(SatProofCoreLimits{MaxClauses: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}

	missingEmpty, _ := NewSatProofStore(SatProofStoreLimits{})
	_, _ = missingEmpty.AddClause([]SatLit{x})
	_ = missingEmpty.MarkRoots()
	manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
	if _, err := manager.UnsatCore(missingEmpty, false); !errors.Is(err, ErrInvariant) {
		t.Fatalf("missing empty: %v", err)
	}
	satisfiable := satProofCoreStore(t, [][]SatLit{{x}}, nil)
	if _, err := manager.UnsatCore(satisfiable, false); !errors.Is(err, ErrInvariant) {
		t.Fatalf("unproved empty: %v", err)
	}
	corrupt := satProofCoreStore(t, [][]SatLit{{x}, {x.Not()}}, nil)
	corrupt.clauses[1].Root = false
	if _, err := manager.UnsatCore(corrupt, false); !errors.Is(err, ErrInvariant) {
		t.Fatalf("non-prefix roots: %v", err)
	}
	if err := VerifySatProofRootCore(valid, []int{2}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid root core: %v", err)
	}
	if err := VerifySatProofRootCore(valid, []int{0}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("satisfiable root core: %v", err)
	}
	if _, err := FormatSatProofLearnedCoreForBMC(valid, []int{99}, []SatBMCVariable{{}}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid formatted core: %v", err)
	}
	if _, err := FormatSatProofLearnedCoreForBMC(valid, nil, nil); !errors.Is(err, ErrInvariant) {
		t.Fatalf("short map: %v", err)
	}
	manager.Stop()
	manager.Stop()
	if !manager.Stats().Stopped {
		t.Fatal("manager did not stop")
	}
	if _, err := manager.UnsatCore(valid, false); !errors.Is(err, ErrInvariant) {
		t.Fatalf("use after stop: %v", err)
	}
}

func FuzzSatProofCoreLearnedReplay(f *testing.F) {
	for _, seed := range []struct {
		first, second int
		mask          byte
	}{{0, 1, 0}, {2, 7, 3}, {31, 4, 15}} {
		f.Add(seed.first, seed.second, seed.mask)
	}
	f.Fuzz(func(t *testing.T, first, second int, mask byte) {
		first = int(uint(first) % 32)
		second = int(uint(second) % 31)
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
		roots := [][]SatLit{{a, b}, {a.Not(), b}, {a, b.Not()}, {a.Not(), b.Not()}}
		for i := 0; i < len(roots); i++ {
			j := int(mask>>uint(2*i)) % len(roots)
			roots[i], roots[j] = roots[j], roots[i]
		}
		store := satProofCoreStore(t, roots, [][]SatLit{{b}})
		manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
		core, err := manager.UnsatCore(store, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(core) != 4 {
			t.Fatalf("root core = %v", core)
		}
		seen := [4]bool{}
		for _, clause := range core {
			if clause < 0 || clause >= 4 || seen[clause] {
				t.Fatalf("invalid root core = %v", core)
			}
			seen[clause] = true
		}
		if err = VerifySatProofRootCore(store, core); err != nil {
			t.Fatal(err)
		}
		learned, err := manager.UnsatCore(store, true)
		if err != nil {
			t.Fatal(err)
		}
		if want := []int{4, 5}; !reflect.DeepEqual(learned, want) {
			t.Fatalf("learned core = %v, want %v", learned, want)
		}
	})
}

func FuzzSatProofCoreOppositeUnits(f *testing.F) {
	for _, seed := range []struct {
		variables int
		pivot     int
		negative  bool
	}{{1, 0, false}, {4, 2, true}, {16, 15, false}} {
		f.Add(seed.variables, seed.pivot, seed.negative)
	}
	f.Fuzz(func(t *testing.T, variables, pivot int, negative bool) {
		variables = 1 + int(uint(variables)%64)
		pivot = int(uint(pivot) % uint(variables))
		literal, err := MakeSatLit(SatVar(pivot), negative)
		if err != nil {
			t.Fatal(err)
		}
		roots := make([][]SatLit, 0, variables+2)
		for variable := 0; variable < variables; variable++ {
			if variable == pivot {
				continue
			}
			lit, _ := MakeSatLit(SatVar(variable), variable&1 != 0)
			roots = append(roots, []SatLit{lit, literal})
		}
		first := len(roots)
		roots = append(roots, []SatLit{literal}, []SatLit{literal.Not()})
		store := satProofCoreStore(t, roots, nil)
		manager, _ := NewSatProofCoreManager(SatProofCoreLimits{})
		core, err := manager.UnsatCore(store, false)
		if err != nil {
			t.Fatal(err)
		}
		want := []int{first + 1, first}
		if !reflect.DeepEqual(core, want) {
			t.Fatalf("vars=%d pivot=%d sign=%v core=%v want=%v", variables, pivot, negative, core, want)
		}
		if err = VerifySatProofRootCore(store, core); err != nil {
			t.Fatal(err)
		}
		mapping := make([]SatBMCVariable, variables)
		for i := range mapping {
			mapping[i] = SatBMCVariable{Object: i + 100, Frame: i % 5}
		}
		learned, err := manager.UnsatCore(store, true)
		if err != nil {
			t.Fatal(err)
		}
		text, err := FormatSatProofLearnedCoreForBMC(store, learned, mapping)
		if err != nil || !strings.Contains(text, fmt.Sprintf("UNSAT contains %d learned clauses", len(learned))) {
			t.Fatalf("format = %q, err=%v", text, err)
		}
	})
}
