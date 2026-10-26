package abc

// Tests for the Go-only FXU cover flow. Oracles below evaluate the original
// covers directly and do not use production extraction or matrix helpers.

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func fxuFlowIndependentEvaluate(covers []FXUCover, inputs []bool) []bool {
	outputs := make([]bool, len(covers))
	for i, cover := range covers {
		for _, cube := range cover.Cubes {
			product := true
			for _, literal := range cube {
				if literal < 0 || literal >= len(inputs) || !inputs[literal] {
					product = false
					break
				}
			}
			if product {
				outputs[i] = true
				break
			}
		}
	}
	return outputs
}

func checkFXUFlowTruth(t *testing.T, inputs int, original []FXUCover, result FXUFlowResult) {
	t.Helper()
	for assignment := 0; assignment < 1<<inputs; assignment++ {
		values := make([]bool, inputs)
		for variable := range values {
			values[variable] = assignment&(1<<variable) != 0
		}
		want := fxuFlowIndependentEvaluate(original, values)
		got, err := result.Evaluate(values)
		if err != nil {
			t.Fatalf("assignment %d: %v", assignment, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("assignment %d: got %v want %v", assignment, got, want)
		}
	}
}

func TestFXUFlowSingleExtraction(t *testing.T) {
	input := []FXUCover{
		{Output: 7, Cubes: [][]int{{0, 1, 2}, {0, 1, 3}}},
		{Output: 9, Cubes: [][]int{{0, 1}, {2, 3}}},
	}
	result, err := ExtractFXU(4, input, FXUFlowOptions{OnlySingles: true, MaxExtractions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Kind != FXUStepSingle || result.Steps[0].Var1 != 0 || result.Steps[0].Var2 != 1 || result.Steps[0].Weight != 1 {
		t.Fatalf("steps=%+v", result.Steps)
	}
	if result.FinalLiterals != result.InitialLiterals-1 {
		t.Fatalf("literal counts %d -> %d", result.InitialLiterals, result.FinalLiterals)
	}
	checkFXUFlowTruth(t, 4, input, result)
	input[0].Cubes[0][0] = 99
	if result.Covers[0].Cubes[0][0] == 99 {
		t.Fatal("result aliases caller input")
	}
}

func TestFXUFlowDoubleExtraction(t *testing.T) {
	input := []FXUCover{{Output: 3, Cubes: [][]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}}}}
	result, err := ExtractFXU(4, input, FXUFlowOptions{OnlyDoubles: true, MaxExtractions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Kind != FXUStepDouble || result.Steps[0].Pairs != 2 || result.Steps[0].Weight != 2 {
		t.Fatalf("steps=%+v", result.Steps)
	}
	if !reflect.DeepEqual(result.Definitions[0].Cubes, [][]int{{2}, {3}}) {
		t.Fatalf("definition=%v", result.Definitions[0].Cubes)
	}
	checkFXUFlowTruth(t, 4, input, result)
	text, err := result.FormatFXUFlow()
	if err != nil || text != "fxu inputs=4 extracted=1 literals=8-6\nv4 w=2 double pairs=2 left=2,|3,\n" {
		t.Fatalf("format=%q err=%v", text, err)
	}
}

func TestFXUFlowOptionsAndFailures(t *testing.T) {
	input := []FXUCover{{Output: 0, Cubes: [][]int{{0, 1, 2}, {0, 1, 3}}}}
	if _, err := ExtractFXU(4, input, FXUFlowOptions{OnlySingles: true, OnlyDoubles: true}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("both modes: %v", err)
	}
	if _, err := ExtractFXU(4, input, FXUFlowOptions{UseComplement: true}); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("complement: %v", err)
	}
	if _, err := ExtractFXU(2, []FXUCover{{Output: 0, Cubes: [][]int{{0, 0}}}}, FXUFlowOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate literal: %v", err)
	}
	if _, err := ExtractFXU(2, []FXUCover{{Output: 0, Cubes: [][]int{{2}}}}, FXUFlowOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad literal: %v", err)
	}
	if _, err := ExtractFXU(4, input, FXUFlowOptions{OnlySingles: true, Kernel: FXUOptions{MaxOperations: 1}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("operation limit: %v", err)
	}
	if _, err := ExtractFXU(math.MaxInt, nil, FXUFlowOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("hostile variable count: %v", err)
	}
	if _, err := ExtractFXU(4, input, FXUFlowOptions{Kernel: FXUOptions{MaxEntries: 2}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("entry limit: %v", err)
	}
	calls := 0
	copyBefore := cloneFXUCovers(input)
	_, err := ExtractFXU(4, input, FXUFlowOptions{OnlySingles: true, Kernel: FXUOptions{Stop: func() bool { calls++; return calls > 3 }}})
	if !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(input, copyBefore) {
		t.Fatalf("cancel err=%v input=%v", err, input)
	}
}

func TestFXUFlowUsesOneWorkAndVariableBudget(t *testing.T) {
	operations := 2
	if _, _, err := buildFXUFlowMatrix(nil, 3, FXUFlowOptions{Kernel: FXUOptions{MaxOperations: 4}}, &operations); !errors.Is(err, ErrLimit) {
		t.Fatalf("shared operation budget: err=%v operations=%d", err, operations)
	}

	input := []FXUCover{{Output: 0, Cubes: [][]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}}}}
	_, err := ExtractFXU(4, input, FXUFlowOptions{
		OnlyDoubles:    true,
		MaxExtractions: 2,
		MinWeight:      -2,
		Kernel:         FXUOptions{MaxVariables: 6},
	})
	if !errors.Is(err, ErrLimit) {
		t.Fatalf("definition variable budget: %v", err)
	}
}

func TestFXUFlowNegativeThreshold(t *testing.T) {
	input := []FXUCover{{Output: 0, Cubes: [][]int{{0}, {1}}}}
	result, err := ExtractFXU(2, input, FXUFlowOptions{OnlyDoubles: true, MinWeight: -2, MaxExtractions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].Weight != -1 || result.FinalLiterals != result.InitialLiterals+1 {
		t.Fatalf("result=%+v", result)
	}
	checkFXUFlowTruth(t, 2, input, result)
}

func TestFXUFlowPairPreprocessingAndCanonical(t *testing.T) {
	input := []FXUCover{
		{Output: 0, Cubes: [][]int{{0, 2}, {1, 2}}},
		{Output: 1, Cubes: [][]int{{0, 3}, {1, 3}}},
		{Output: 2, Cubes: [][]int{{0, 1, 4}, {2, 3, 4}}},
	}
	limited, err := ExtractFXU(7, input, FXUFlowOptions{OnlyDoubles: true, MaxExtractions: 1, MaxPairs: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited.Steps) != 1 || !reflect.DeepEqual(limited.Steps[0].Left, []int{0}) || !reflect.DeepEqual(limited.Steps[0].Right, []int{1}) {
		t.Fatalf("limited step=%+v", limited.Steps)
	}
	checkFXUFlowTruth(t, 7, input, limited)
	canonical, err := ExtractFXU(7, input, FXUFlowOptions{OnlyDoubles: true, CanonicalOnly: true, MaxExtractions: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical.Steps) != 1 || len(canonical.Steps[0].Left) != 1 || len(canonical.Steps[0].Right) != 1 {
		t.Fatalf("canonical step=%+v", canonical.Steps)
	}
}

func TestFXUFlowDeterministicIndependentModel(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	for trial := 0; trial < 5000; trial++ {
		variables := 2 + int(next()%4)
		commonA := int(next() % uint64(variables))
		commonB := int(next() % uint64(variables))
		if commonA == commonB {
			commonB = (commonB + 1) % variables
		}
		covers := make([]FXUCover, 2)
		for output := range covers {
			covers[output].Output = output
			for cube := 0; cube < 2; cube++ {
				literals := []int{commonA, commonB}
				if variables > 2 && next()&1 != 0 {
					literals = append(literals, int(next()%uint64(variables)))
				}
				sort.Ints(literals)
				unique := literals[:0]
				for _, literal := range literals {
					if len(unique) == 0 || unique[len(unique)-1] != literal {
						unique = append(unique, literal)
					}
				}
				covers[output].Cubes = append(covers[output].Cubes, append([]int(nil), unique...))
			}
		}
		result, err := ExtractFXU(variables, covers, FXUFlowOptions{OnlySingles: true, MaxExtractions: 2})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		checkFXUFlowTruth(t, variables, covers, result)
		second, err := ExtractFXU(variables, covers, FXUFlowOptions{OnlySingles: true, MaxExtractions: 2})
		if err != nil || !reflect.DeepEqual(result, second) {
			t.Fatalf("trial %d nondeterministic err=%v", trial, err)
		}
	}
}

func FuzzFXUFlow(f *testing.F) {
	f.Add([]byte{4, 2, 0, 1, 2, 0, 1, 3})
	f.Add([]byte{3, 1, 0, 2, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := 2 + int(data[0]%4)
		covers := make([]FXUCover, 1+len(data)%3)
		position := 1
		for output := range covers {
			covers[output].Output = output
			cubeCount := 1 + (len(data)+output)%4
			for cubeIndex := 0; cubeIndex < cubeCount; cubeIndex++ {
				mask := byte(0)
				if len(data) != 0 {
					mask = data[position%len(data)]
					position++
				}
				cube := make([]int, 0, variables)
				for variable := 0; variable < variables; variable++ {
					if mask&(1<<variable) != 0 {
						cube = append(cube, variable)
					}
				}
				covers[output].Cubes = append(covers[output].Cubes, cube)
			}
		}
		options := FXUFlowOptions{MaxExtractions: 3, MaxPairs: 32, MaxSingles: 32}
		if len(data) > 1 {
			options.OnlySingles = data[1]&1 != 0
			options.OnlyDoubles = data[1]&2 != 0
			if options.OnlySingles && options.OnlyDoubles {
				options.OnlyDoubles = false
			}
		}
		result, err := ExtractFXU(variables, covers, options)
		if err != nil {
			if !errors.Is(err, ErrInvariant) && !errors.Is(err, ErrDimension) && !errors.Is(err, ErrLimit) {
				t.Fatalf("unexpected error: %v", err)
			}
			return
		}
		checkFXUFlowTruth(t, variables, covers, result)
		if result.FinalLiterals > result.InitialLiterals {
			t.Fatalf("literal cost increased: %d -> %d", result.InitialLiterals, result.FinalLiterals)
		}
		text, err := result.FormatFXUFlow()
		if err != nil || !strings.HasPrefix(text, "fxu inputs=") {
			t.Fatalf("format err=%v text=%q", err, text)
		}
	})
}
