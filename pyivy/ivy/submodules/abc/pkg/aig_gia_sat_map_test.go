package abc

import (
	"errors"
	"slices"
	"testing"
)

func giaSatMapFixture() GiaSatMapProblem {
	return GiaSatMapProblem{
		LogN:           3,
		ObjectLiterals: 4,
		Inputs:         1,
		Roots:          []int{2},
		Cuts: []GiaSatMapCut{
			{Root: 0, Gate: 10, Area: 2},
			{Root: 0, Fanins: []GiaSatMapFanin{{Kind: GiaSatMapFaninInputInverter, Index: 0}}, Gate: 11, Area: 1},
			{Root: 1, Gate: 12, Area: 1},
			{Root: 2, Fanins: []GiaSatMapFanin{{Kind: GiaSatMapFaninObjectLiteral, Index: 0}}, Gate: 20, Area: 3},
			{Root: 3, Fanins: []GiaSatMapFanin{{Kind: GiaSatMapFaninObjectLiteral, Index: 1}}, Gate: 21, Area: 3},
		},
		Initial: []GiaSatMapSelection{
			{Kind: GiaSatMapSelectCut, Index: 0},
			{Kind: GiaSatMapSelectCut, Index: 3},
			{Kind: GiaSatMapSelectInputInverter, Index: 0},
		},
		InverterArea: 7,
	}
}

func TestGiaSatMapperClosureClausesSolveAndOptimize(t *testing.T) {
	problem := giaSatMapFixture()
	mapper, err := NewGiaSatMapper(problem, GiaWindowCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	prefix := len(mapper.cardinality.Clauses)
	got := mapper.Clauses()[prefix:]
	cut := func(index int, negative bool) SatLit { return giaSatMapLiteral(mapper.firstCut+index, negative) }
	want := [][]SatLit{
		{giaSatMapLiteral(0, true), cut(0, false), cut(1, false)},
		{giaSatMapLiteral(1, true), cut(2, false)},
		{cut(2, true), cut(1, true)},
		{giaSatMapLiteral(2, true), cut(3, false)},
		{giaSatMapLiteral(3, true), cut(4, false)},
		{cut(4, true), cut(3, true)},
		{giaSatMapLiteral(0, false), cut(0, true)},
		{giaSatMapLiteral(0, false), cut(1, true)},
		{giaSatMapLiteral(4, false), cut(1, true)},
		{giaSatMapLiteral(1, false), cut(2, true)},
		{giaSatMapLiteral(2, false), cut(3, true)},
		{giaSatMapLiteral(0, false), cut(3, true)},
		{giaSatMapLiteral(3, false), cut(4, true)},
		{giaSatMapLiteral(1, false), cut(4, true)},
	}
	if len(got) != len(want) {
		t.Fatal(len(got), len(want), got)
	}
	for index := range want {
		if !slices.Equal(got[index], want[index]) {
			t.Fatal(index, got[index], want[index])
		}
	}
	closed := []GiaSatMapSelection{{Kind: GiaSatMapSelectCut, Index: 0}, {Kind: GiaSatMapSelectCut, Index: 1}, {Kind: GiaSatMapSelectCut, Index: 3}, {Kind: GiaSatMapSelectInputInverter, Index: 0}}
	preferred, err := CheckGiaSatMapSolution(problem, closed, mapper.firstCut)
	wantPreferred := []SatVar{SatVar(mapper.firstCut), 0, SatVar(mapper.firstCut + 1), SatVar(mapper.firstCut + 3), 2, 4, SatVar(mapper.firstCut), SatVar(mapper.firstCut + 1), SatVar(mapper.firstCut + 3)}
	if err != nil || !slices.Equal(preferred, wantPreferred) {
		t.Fatal(preferred, wantPreferred, err)
	}
	assumptions, err := mapper.Assumptions(2)
	if err != nil || len(assumptions) != 1+(mapper.capacity-problem.ObjectLiterals-problem.Inputs)+len(problem.Roots) || assumptions[len(assumptions)-1] != giaSatMapLiteral(2, false) {
		t.Fatal(assumptions, err)
	}
	for bound, status := range []SatStatus{SatUnsatisfiable, SatUnsatisfiable, SatSatisfiable, SatSatisfiable} {
		result, solveErr := mapper.Solve(bound, SatLimits{})
		if solveErr != nil || result.SAT.Status != status {
			t.Fatal(bound, result, solveErr)
		}
		if status == SatSatisfiable {
			if !slices.Equal(result.SelectedObjects, []int{0, 2}) || !slices.Equal(result.SelectedCuts, []int{0, 3}) || !slices.Equal(result.SelectedGates, []int{10, 20}) || result.Area != 5 {
				t.Fatal(bound, result)
			}
		}
	}
	optimized, err := mapper.Optimize(SatLimits{})
	if err != nil || optimized.InitialCount != 3 || len(optimized.Attempts) != 2 || optimized.Attempts[0].SAT.Status != SatSatisfiable || optimized.Attempts[1].SAT.Status != SatUnsatisfiable || optimized.Best.Bound != 2 {
		t.Fatal(optimized, err)
	}
	problem.Roots[0] = 0
	problem.Cuts[0].Root = 3
	problem.Cuts[3].Fanins[0].Index = 3
	if result, solveErr := mapper.Solve(2, SatLimits{}); solveErr != nil || result.SAT.Status != SatSatisfiable || !slices.Equal(result.SelectedObjects, []int{0, 2}) {
		t.Fatal("caller mutation changed mapper", result, solveErr)
	}
	clauses := mapper.Clauses()
	clauses[prefix][0] = 0
	if slices.Equal(clauses[prefix], mapper.Clauses()[prefix]) {
		t.Fatal("clause snapshot aliases")
	}
}

func TestGiaSatMapperValidationLimitsCancellationAndLifecycle(t *testing.T) {
	base := giaSatMapFixture()
	for _, mutate := range []func(*GiaSatMapProblem){
		func(p *GiaSatMapProblem) { p.LogN = -1 },
		func(p *GiaSatMapProblem) { p.ObjectLiterals = 3 },
		func(p *GiaSatMapProblem) { p.Inputs = 5 },
		func(p *GiaSatMapProblem) { p.Roots[0] = 4 },
		func(p *GiaSatMapProblem) { p.Cuts[0].Root = 4 },
		func(p *GiaSatMapProblem) { p.Cuts[0].Gate = -1 },
		func(p *GiaSatMapProblem) { p.Cuts[3].Fanins[0].Index = 4 },
		func(p *GiaSatMapProblem) { p.Cuts = p.Cuts[:4] },
		func(p *GiaSatMapProblem) { p.Initial = append(p.Initial, p.Initial[0]) },
	} {
		problem := giaSatMapFixture()
		mutate(&problem)
		if mapper, err := NewGiaSatMapper(problem, GiaWindowCNFOptions{}); err == nil || mapper != nil {
			t.Fatal(problem, mapper, err)
		}
	}
	broken := giaSatMapFixture()
	broken.Initial = broken.Initial[1:]
	if mapper, err := NewGiaSatMapper(broken, GiaWindowCNFOptions{}); !errors.Is(err, ErrInvariant) || mapper != nil {
		t.Fatal(mapper, err)
	}
	broken = giaSatMapFixture()
	broken.Initial = broken.Initial[:1]
	if mapper, err := NewGiaSatMapper(broken, GiaWindowCNFOptions{}); !errors.Is(err, ErrInvariant) || mapper != nil {
		t.Fatal(mapper, err)
	}
	for _, options := range []GiaWindowCNFOptions{{MaxVariables: 4}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}, {MaxVariables: -1}} {
		if mapper, err := NewGiaSatMapper(base, options); err == nil || mapper != nil {
			t.Fatal(options, mapper, err)
		}
	}
	mapper, err := NewGiaSatMapper(base, GiaWindowCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = mapper.Assumptions(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = mapper.Assumptions(9); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	unknown, err := mapper.Solve(2, SatLimits{Stop: func() bool { return true }})
	if err != nil || unknown.SAT.Status != SatUnknown {
		t.Fatal(unknown, err)
	}
	mapper.Close()
	mapper.Close()
	if mapper.Clauses() != nil {
		t.Fatal("closed mapper exposes clauses")
	}
	if _, err = mapper.Solve(1, SatLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = (*GiaSatMapper)(nil).Optimize(SatLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzGiaSatMapper(f *testing.F) {
	for _, seed := range []struct {
		pairs, root, flags uint8
	}{{1, 0, 0}, {4, 3, 1}, {7, 6, 3}} {
		f.Add(seed.pairs, seed.root, seed.flags)
	}
	f.Fuzz(func(t *testing.T, pairByte, rootByte, flags uint8) {
		pairs := 1 + int(pairByte%8)
		objects := 2 * pairs
		logN := 0
		for 1<<uint(logN) < objects+2 {
			logN++
		}
		rootPair := int(rootByte) % pairs
		cuts := make([]GiaSatMapCut, objects)
		for object := range cuts {
			cuts[object] = GiaSatMapCut{Root: object, Gate: object, Area: uint64(object + 1)}
		}
		usedObjects := make([]int, 0, rootPair+1)
		initial := make([]GiaSatMapSelection, 0, rootPair+2)
		for pair := 0; pair <= rootPair; pair++ {
			object := 2 * pair
			usedObjects = append(usedObjects, object)
			initial = append(initial, GiaSatMapSelection{Kind: GiaSatMapSelectCut, Index: object})
			if pair > 0 {
				cuts[object].Fanins = append(cuts[object].Fanins, GiaSatMapFanin{Kind: GiaSatMapFaninObjectLiteral, Index: object - 2})
			}
		}
		inverter := flags&1 != 0
		if inverter {
			cuts[0].Fanins = append(cuts[0].Fanins, GiaSatMapFanin{Kind: GiaSatMapFaninInputInverter, Index: 0})
			initial = append(initial, GiaSatMapSelection{Kind: GiaSatMapSelectInputInverter, Index: 0})
		}
		problem := GiaSatMapProblem{LogN: logN, ObjectLiterals: objects, Inputs: 1, Roots: []int{2 * rootPair}, Cuts: cuts, Initial: initial, InverterArea: 5}
		mapper, err := NewGiaSatMapper(problem, GiaWindowCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		minimum := len(usedObjects)
		if inverter {
			minimum++
		}
		for bound := 0; bound <= minimum+1 && bound <= mapper.capacity; bound++ {
			result, solveErr := mapper.Solve(bound, SatLimits{})
			want := SatUnsatisfiable
			if bound >= minimum {
				want = SatSatisfiable
			}
			if solveErr != nil || result.SAT.Status != want {
				t.Fatal(pairs, rootPair, flags, bound, minimum, result, solveErr)
			}
			if want == SatSatisfiable && (!slices.Equal(result.SelectedObjects, usedObjects) || len(result.SelectedInverters) != boolInt(inverter)) {
				t.Fatal(result, usedObjects, inverter)
			}
		}
	})
}
