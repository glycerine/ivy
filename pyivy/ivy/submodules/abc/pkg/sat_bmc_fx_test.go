package abc

// Stand-alone divisor tests for src/sat/bmc/bmcFx.c.

import (
	"errors"
	"slices"
	"testing"
)

func fxTestLit(variable int, one bool) SatLit {
	literal, _ := MakeSatLit(SatVar(variable), !one)
	return literal
}

func TestFXCubeDifferenceAndClassification(t *testing.T) {
	left := []SatLit{fxTestLit(0, true), fxTestLit(1, false), fxTestLit(3, true)}
	right := []SatLit{fxTestLit(0, true), fxTestLit(2, true), fxTestLit(3, true)}
	shared, leftOnly, rightOnly, ok := fxCubeDifference(left, right)
	if !ok || shared != 2 || leftOnly != [2]SatLit{fxTestLit(1, false), InvalidSatLit} || rightOnly != [2]SatLit{fxTestLit(2, true), InvalidSatLit} {
		t.Fatal(shared, leftOnly, rightOnly, ok)
	}
	type_, inputs := fxClassifyDivisor([2]SatLit{fxTestLit(0, false), fxTestLit(1, true)}, [2]SatLit{fxTestLit(0, true), fxTestLit(1, false)})
	if type_ != FXDivisorXOR || inputs != [3]SatLit{fxTestLit(0, true), fxTestLit(1, true), InvalidSatLit} {
		t.Fatal("xor", type_, inputs)
	}
	type_, inputs = fxClassifyDivisor([2]SatLit{fxTestLit(0, true), fxTestLit(1, true)}, [2]SatLit{fxTestLit(0, false), fxTestLit(2, false)})
	if type_ != FXDivisorMux || inputs[0] != fxTestLit(0, true) {
		t.Fatal("mux", type_, inputs)
	}
	type_, inputs = fxClassifyDivisor([2]SatLit{fxTestLit(0, false), InvalidSatLit}, [2]SatLit{fxTestLit(1, true), InvalidSatLit})
	if type_ != FXDivisorAND || inputs != [3]SatLit{fxTestLit(0, true), fxTestLit(1, false), InvalidSatLit} {
		t.Fatal("or-as-inverted-and", type_, inputs)
	}
}

func TestFindFXDivisorsCostsOrderAndOwnership(t *testing.T) {
	cover := SOPCover{Variables: 4, Output: '1', Cubes: []string{"1100", "1101", "1110", "1111"}}
	divisors, err := FindFXDivisors(cover, FXDivisorOptions{MaxDivisors: 8})
	if err != nil || len(divisors) == 0 {
		t.Fatal(divisors, err)
	}
	want := FXDivisor{Type: FXDivisorAND, Inputs: [3]SatLit{fxTestLit(0, true), fxTestLit(1, true), InvalidSatLit}, Cost: 4}
	if divisors[0] != want {
		t.Fatal("best", divisors[0], want, divisors)
	}
	for index := 1; index < len(divisors); index++ {
		if divisors[index-1].Cost < divisors[index].Cost || divisors[index].Cost <= 1 {
			t.Fatal("order", divisors)
		}
	}
	again, err := FindFXDivisors(cover, FXDivisorOptions{MaxDivisors: 8})
	if err != nil || !slices.Equal(divisors, again) {
		t.Fatal("determinism", divisors, again, err)
	}
	divisors[0].Cost = 0
	fresh, _ := FindFXDivisors(cover, FXDivisorOptions{MaxDivisors: 8})
	if fresh[0].Cost != 4 {
		t.Fatal("result aliases state", fresh)
	}
}

func TestFindFXDivisorsLimitsCancellationAndMalformed(t *testing.T) {
	cover := SOPCover{Variables: 3, Output: '1', Cubes: []string{"110", "111"}}
	bad := []FXDivisorOptions{{MaxDivisors: -1}, {MaxCubes: -1}, {MaxCubeEntries: -1}, {MaxPairs: -1}}
	for _, options := range bad {
		if _, err := FindFXDivisors(cover, options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	for _, options := range []FXDivisorOptions{{MaxCubes: 1}, {MaxCubeEntries: 5}, {MaxPairs: 1}} {
		if _, err := FindFXDivisors(cover, options); !errors.Is(err, ErrLimit) {
			t.Fatal(options, err)
		}
	}
	if _, err := FindFXDivisors(cover, FXDivisorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	for _, malformed := range []SOPCover{{Variables: -1, Cubes: []string{""}}, {Variables: 2, Cubes: []string{"1"}}, {Variables: 2, Cubes: []string{"1x"}}} {
		if result, err := FindFXDivisors(malformed, FXDivisorOptions{}); result != nil || err == nil {
			t.Fatal(malformed, result, err)
		}
	}
}

func FuzzFindFXDivisorsDeterministic(f *testing.F) {
	for _, seed := range []uint64{0, 1, 0x123456789abcdef, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		cubes := make([]string, 1+int(seed%8))
		state := seed
		for index := range cubes {
			cube := make([]byte, 6)
			for variable := range cube {
				state = state*6364136223846793005 + 1442695040888963407
				cube[variable] = []byte{'-', '0', '1'}[(state>>61)%3]
			}
			cubes[index] = string(cube)
		}
		cover := SOPCover{Variables: 6, Cubes: cubes, Output: '1'}
		first, err := FindFXDivisors(cover, FXDivisorOptions{MaxDivisors: 32, MaxPairs: 10000})
		if err != nil {
			t.Fatal(err)
		}
		second, err := FindFXDivisors(cover, FXDivisorOptions{MaxDivisors: 32, MaxPairs: 10000})
		if err != nil || !slices.Equal(first, second) {
			t.Fatal(first, second, err)
		}
		for index, divisor := range first {
			if divisor.Cost <= 1 || index > 0 && first[index-1].Cost < divisor.Cost {
				t.Fatal(first)
			}
		}
	})
}

func fxSharingGraph(t testing.TB) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	giaTestHash(t, manager)
	a, b, c, d := giaTestCI(t, manager), giaTestCI(t, manager), giaTestCI(t, manager), giaTestCI(t, manager)
	ab := giaTestAnd(t, manager, a, b)
	xor, err := manager.HashXor(c, d)
	if err != nil {
		t.Fatal(err)
	}
	root := giaTestAnd(t, manager, ab, xor)
	giaTestCO(t, manager, root)
	giaTestCheck(t, manager)
	return manager
}

func TestFXGateClausesAndCandidateEvaluation(t *testing.T) {
	relation := &AigCNF{variables: 3}
	candidates := []SatLit{0, 2, 4}
	selected := []FXDivisor{
		{Type: FXDivisorAND, Inputs: [3]SatLit{0, 2, InvalidSatLit}, Cost: 3},
		{Type: FXDivisorXOR, Inputs: [3]SatLit{0, 2, InvalidSatLit}, Cost: 2},
		{Type: FXDivisorMux, Inputs: [3]SatLit{0, 2, 4}, Cost: 2},
	}
	options, err := normalizeFXExtractionOptions(FXExtractionOptions{Iterations: 1, MaxVariables: 32, MaxClauses: 32, MaxLiterals: 128, MaxCandidates: 16})
	if err != nil {
		t.Fatal(err)
	}
	candidates, appended, err := fxAddClauses(relation, candidates, selected, options)
	if err != nil || len(candidates) != 6 || !slices.Equal(appended, []FXDivisor{selected[2], selected[1], selected[0]}) {
		t.Fatal(candidates, appended, err)
	}
	model := FXExtractionResult{BaseInputs: 3, Divisors: appended}
	for assignment := 0; assignment < 8; assignment++ {
		base := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want := []bool{base[0], base[1], base[2], func() bool {
			if base[0] {
				return base[1]
			}
			return base[2]
		}(), base[0] != base[1], base[0] && base[1]}
		values, evalErr := model.EvaluateCandidates(base)
		if evalErr != nil || !slices.Equal(values, want) {
			t.Fatal(assignment, values, want, evalErr)
		}
		solver, solveErr := NewSatSolverWithLimit(32)
		if solveErr != nil {
			t.Fatal(solveErr)
		}
		if ok, loadErr := relation.LoadInto(solver, 0); loadErr != nil || !ok {
			t.Fatal(ok, loadErr)
		}
		assumptions := make([]SatLit, len(want))
		for index, value := range want {
			assumptions[index] = SatLit(index) << 1
			if !value {
				assumptions[index] = assumptions[index].Not()
			}
		}
		answer, solveErr := solver.Solve(assumptions, SatLimits{})
		if solveErr != nil || answer.Status != SatSatisfiable {
			t.Fatal(assignment, answer.Status, solveErr)
		}
		for index := 3; index < len(want); index++ {
			wrong := slices.Clone(assumptions)
			wrong[index] = wrong[index].Not()
			answer, solveErr = solver.Solve(wrong, SatLimits{})
			if solveErr != nil || answer.Status != SatUnsatisfiable {
				t.Fatal("gate not functional", assignment, index, answer.Status, solveErr)
			}
		}
	}
}

func TestExtractFXSharingIterationsAndSemantics(t *testing.T) {
	manager := fxSharingGraph(t)
	before := slices.Clone(manager.objects)
	result, err := manager.ExtractFXSharing(FXExtractionOptions{Output: 0, Iterations: 3, DivisorsPerIteration: 2})
	if err != nil || result.Status != SatSatisfiable || result.Reason != SatFinished || len(result.Iterations) != 3 || len(result.Divisors) == 0 {
		t.Fatal(result, err)
	}
	if !slices.Equal(before, manager.objects) {
		t.Fatal("source mutated")
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		outputs, evalErr := manager.Evaluate(inputs)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		for iteration := range result.Iterations {
			got, coverErr := result.EvaluateIteration(iteration, inputs)
			if coverErr != nil || got != outputs[0] {
				t.Fatal(assignment, iteration, got, outputs[0], coverErr, result.Iterations[iteration].Cover)
			}
		}
	}
	for iteration, step := range result.Iterations {
		if step.Cover.Variables != step.Candidates || iteration > 0 && step.Candidates < result.Iterations[iteration-1].Candidates {
			t.Fatal("candidate progression", result.Iterations)
		}
	}
	values, err := result.EvaluateCandidates([]bool{true, true, false, true})
	if err != nil || len(values) != result.BaseInputs+len(result.Divisors) {
		t.Fatal(values, err)
	}
	values[0] = false
	fresh, err := result.EvaluateCandidates([]bool{true, true, false, true})
	if err != nil || !fresh[0] {
		t.Fatal("candidate result aliases", fresh, err)
	}
}

func TestAnalyzeFXOutputs(t *testing.T) {
	manager := NewGiaManager()
	giaTestHash(t, manager)
	a, b := giaTestCI(t, manager), giaTestCI(t, manager)
	root := giaTestAnd(t, manager, a, b)
	giaTestCO(t, manager, root)
	giaTestCO(t, manager, root.Not())
	analysis, err := manager.AnalyzeFXOutputs(FXExtractionOptions{Iterations: 1})
	if err != nil || len(analysis) != 2 {
		t.Fatal(analysis, err)
	}
	for output, item := range analysis {
		if item.Output != output || item.RepeatedCubes[0] != item.RepeatedCubes[1] {
			t.Fatal(item)
		}
		for assignment := 0; assignment < 4; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0}
			outputs, _ := manager.Evaluate(inputs)
			on, onErr := item.Onset.Cover.Evaluate(inputs)
			off, offErr := item.Offset.Cover.Evaluate(inputs)
			if onErr != nil || offErr != nil || on != outputs[output] || off == outputs[output] {
				t.Fatal(output, assignment, on, off, outputs[output], onErr, offErr)
			}
		}
	}
}

func TestExtractFXSharingLimitsCancellationUnknownAndDepth(t *testing.T) {
	manager := fxSharingGraph(t)
	bad := []FXExtractionOptions{
		{Iterations: 0}, {Iterations: 1, Output: -1}, {Iterations: 1, DivisorsPerIteration: -1},
		{Iterations: 1, MaxCandidates: -1}, {Iterations: 1, MaxCubes: -1}, {Iterations: 1, MaxCubeEntries: -1},
		{Iterations: 1, MaxPairs: -1}, {Iterations: 1, MaxObjects: -1}, {Iterations: 1, MaxMappingEntries: -1},
		{Iterations: 1, MaxVariables: -1}, {Iterations: 1, MaxClauses: -1}, {Iterations: 1, MaxLiterals: -1},
	}
	for _, options := range bad {
		if _, err := manager.ExtractFXSharing(options); !errors.Is(err, ErrDimension) {
			t.Fatal("bad options", options, err)
		}
	}
	for _, options := range []FXExtractionOptions{
		{Iterations: 1, MaxCandidates: 3}, {Iterations: 1, MaxCubes: 1}, {Iterations: 1, MaxCubeEntries: 1},
		{Iterations: 1, MaxPairs: 1, DivisorsPerIteration: 1}, {Iterations: 1, MaxObjects: 1},
		{Iterations: 1, MaxMappingEntries: 1}, {Iterations: 1, MaxVariables: 1},
		{Iterations: 1, MaxClauses: 1}, {Iterations: 1, MaxLiterals: 1},
	} {
		if _, err := manager.ExtractFXSharing(options); !errors.Is(err, ErrLimit) {
			t.Fatal("limit", options, err)
		}
	}
	if _, err := manager.ExtractFXSharing(FXExtractionOptions{Iterations: 1, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancel", err)
	}
	unknown, err := manager.ExtractFXSharing(FXExtractionOptions{Iterations: 1, SAT: SatLimits{Propagations: 1}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatPropagationLimit {
		t.Fatal("unknown", unknown, err)
	}
	if _, err = manager.ExtractFXSharing(FXExtractionOptions{Iterations: 1, Output: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("output", err)
	}
	deep := NewGiaManager()
	giaTestHash(t, deep)
	a, b := giaTestCI(t, deep), giaTestCI(t, deep)
	root := a
	for index := 0; index < 10000; index++ {
		var appendErr error
		root, appendErr = deep.AppendAnd(root, b)
		if appendErr != nil {
			t.Fatal(appendErr)
		}
	}
	giaTestCO(t, deep, root)
	deepResult, err := deep.ExtractFXSharing(FXExtractionOptions{Iterations: 1, DivisorsPerIteration: 1, MaxObjects: 20000, MaxMappingEntries: 20000, MaxVariables: 20000, MaxClauses: 50000, MaxLiterals: 150000})
	if err != nil || deepResult.Status != SatSatisfiable {
		t.Fatal(deepResult.Status, err)
	}
}

func FuzzExtractFXSharingSemantics(f *testing.F) {
	for _, seed := range []uint8{0, 0xff, 0x96, 0xe8, 0xca, 0x80} {
		f.Add(seed, uint8(1), uint8(2))
	}
	f.Fuzz(func(t *testing.T, truth, iterationSeed, divisorSeed uint8) {
		manager := NewGiaManager()
		giaTestHash(t, manager)
		inputs := []GiaLit{giaTestCI(t, manager), giaTestCI(t, manager), giaTestCI(t, manager)}
		root := GiaConst0
		for assignment := 0; assignment < 8; assignment++ {
			if truth&(1<<assignment) == 0 {
				continue
			}
			term := GiaConst1
			for variable, input := range inputs {
				term = giaTestAnd(t, manager, term, input.NotIf(assignment&(1<<variable) == 0))
			}
			var err error
			root, err = manager.HashOr(root, term)
			if err != nil {
				t.Fatal(err)
			}
		}
		giaTestCO(t, manager, root)
		iterations := 1 + int(iterationSeed%3)
		result, err := manager.ExtractFXSharing(FXExtractionOptions{Iterations: iterations, DivisorsPerIteration: int(divisorSeed % 4), MaxCandidates: 64, MaxCubes: 32, MaxCubeEntries: 2048, MaxPairs: 4096, MaxObjects: 1000, MaxMappingEntries: 1000, MaxVariables: 1000, MaxClauses: 10000, MaxLiterals: 30000})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			want := truth&(1<<assignment) != 0
			for iteration := range result.Iterations {
				got, evalErr := result.EvaluateIteration(iteration, values)
				if evalErr != nil || got != want {
					t.Fatal(truth, assignment, iteration, got, want, evalErr, result)
				}
			}
		}
	})
}
