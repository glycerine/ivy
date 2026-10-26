package abc

// Go-only combinatorial, injectivity, parser, and fuzz tests for extraUtilSupp.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestExtraSupportGenerationPairsAndDifference(t *testing.T) {
	masks, err := GenerateExtraSupportMasks(2, 4, ExtraSupportLimits{})
	want := []uint64{3, 5, 6, 9, 10, 12}
	if err != nil || !slices.Equal(masks, want) {
		t.Fatal(masks, err)
	}
	even, err := GenerateExtraSupportEvenMasks(1, 4, ExtraSupportLimits{})
	if err != nil || !slices.Equal(even, want) {
		t.Fatal(even, err)
	}
	pairs, err := GenerateExtraSupportPairs([]uint64{1, 2, 4, 1}, 3, ExtraSupportLimits{})
	if err != nil || !slices.Equal(pairs, []uint64{3, 5, 0, 6}) {
		t.Fatal(pairs, err)
	}
	difference, err := DifferenceExtraSupport([]uint64{1, 2, 4}, 3, ExtraSupportLimits{})
	if err != nil || !slices.Equal(difference, []uint64{3, 5, 6}) {
		t.Fatal(difference, err)
	}
	if _, err = GenerateExtraSupportMasks(5, 4, ExtraSupportLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = GenerateExtraSupportMasks(2, 20, ExtraSupportLimits{MaxOperations: 100}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = GenerateExtraSupportPairs(make([]uint64, 100), 3, ExtraSupportLimits{MaxOperations: 10}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = DifferenceExtraSupport(make([]uint64, 100), 3, ExtraSupportLimits{MaxPairs: 10}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestExtraSupportGreedyPrimitivesMinimizeAndVerify(t *testing.T) {
	rows := []uint64{3, 5, 6}
	state, err := NewExtraSupportGreedy(rows, 3, ExtraSupportLimits{})
	if err != nil {
		t.Fatal(err)
	}
	rows[0] = 0
	owned, _ := state.Rows()
	if !slices.Equal(owned, []uint64{3, 5, 6}) {
		t.Fatal(owned)
	}
	profile, err := state.Profile()
	best, merit, bestErr := state.BestVariable()
	if err != nil || bestErr != nil || !slices.Equal(profile, []int{2, 2, 2}) || best != 0 || merit != 2 {
		t.Fatal(profile, best, merit, err, bestErr)
	}
	if err = state.SelectVariable(0); err != nil {
		t.Fatal(err)
	}
	profile, _ = state.Profile()
	if !slices.Equal(profile, []int{0, 2, 2}) {
		t.Fatal(profile)
	}
	removed, err := state.FilterCovered()
	left, _ := state.Rows()
	if err != nil || removed != 2 || !slices.Equal(left, []uint64{6}) {
		t.Fatal(removed, left, err)
	}

	result, err := MinimizeExtraSupport([]uint64{3, 5, 6}, 3, ExtraSupportLimits{})
	if err != nil || !slices.Equal(result.Matrix, []uint64{1, 2}) || len(result.Steps) != 2 || result.Steps[0].Remaining != 1 || result.Steps[1].Remaining != 0 {
		t.Fatal(result, err)
	}
	verification, err := VerifyExtraSupport([]uint64{1, 2, 4}, result.Matrix, 3, ExtraSupportLimits{})
	if err != nil || verification.Collisions != 0 || !slices.Equal(verification.Codes, []uint64{1, 2, 0}) {
		t.Fatal(verification, err)
	}
	colliding, err := VerifyExtraSupport([]uint64{1, 2, 4}, []uint64{1}, 3, ExtraSupportLimits{})
	if err != nil || colliding.Collisions != 1 {
		t.Fatal(colliding, err)
	}
	mask, err := FormatExtraSupportMask(5, 3)
	if err != nil || mask != "101\n" {
		t.Fatal(mask, err)
	}
	text, err := FormatExtraSupportProfile([]uint64{3, 5, 6}, 3, ExtraSupportLimits{})
	wantText := " 0 :      2   66.67 %\n 1 :      2   66.67 %\n 2 :      2   66.67 %\n"
	if err != nil || text != wantText {
		t.Fatal(text, err)
	}
	zeroText, err := FormatExtraSupportProfile([]uint64{0}, 2, ExtraSupportLimits{})
	if err != nil || zeroText != " 0 :      0    0.00 %\n 1 :      0    0.00 %\n" {
		t.Fatal(zeroText, err)
	}
	if _, err = MinimizeExtraSupport([]uint64{0}, 3, ExtraSupportLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	empty, _ := NewExtraSupportGreedy(nil, 3, ExtraSupportLimits{})
	if _, err = empty.FindOne(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraSupportReductionAndProfile(t *testing.T) {
	result, err := SolveExtraSupport([]uint64{3, 5, 6}, 3, ExtraSupportLimits{})
	if err != nil || result.Remaining != 2 || result.Active != 6 || !slices.Equal(result.Removed, []int{0}) || !slices.Equal(result.Profile.Singles, []int{0, 1, 1}) {
		t.Fatal(result, err)
	}
	result, err = SolveExtraSupport([]uint64{1}, 3, ExtraSupportLimits{})
	if err != nil || result.Remaining != 1 || result.Active != 1 || !slices.Equal(result.Removed, []int{1, 2}) {
		t.Fatal(result, err)
	}
	if _, err = SolveExtraSupport([]uint64{0}, 3, ExtraSupportLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if _, err = SolveExtraSupport([]uint64{8}, 3, ExtraSupportLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraSupportMINParserFileAndWorkflow(t *testing.T) {
	data := []byte("header INPUT F-COVER 2, 10,01,cube, 01,10,cube, INPUT OUTPUT")
	variables, cubes, err := ParseExtraSupportMIN(data, ExtraSupportLimits{})
	if err != nil || variables != 2 || !slices.Equal(cubes, []uint64{1, 2}) {
		t.Fatal(variables, cubes, err)
	}
	problem, err := RunExtraSupportProblem(cubes, variables, ExtraSupportLimits{})
	if err != nil || !slices.Equal(problem.Pairs, []uint64{3}) || problem.Reduction.Remaining != 1 || len(problem.Greedy.Matrix) != 1 {
		t.Fatal(problem, err)
	}
	verification, err := VerifyExtraSupport(problem.Cubes, problem.Greedy.Matrix, variables, ExtraSupportLimits{})
	if err != nil || verification.Collisions != 0 {
		t.Fatal(verification, err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "cover.min")
	if err = os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	fileVariables, fileCubes, err := LoadExtraSupportMIN(path, ExtraSupportLimits{})
	if err != nil || fileVariables != variables || !slices.Equal(fileCubes, cubes) {
		t.Fatal(fileVariables, fileCubes, err)
	}
	for _, malformed := range [][]byte{
		[]byte("nothing"),
		[]byte("INPUT F-COVER 0"),
		[]byte("INPUT F-COVER 1,xx,cube,INPUT"),
		[]byte("INPUT F-COVER 2,10,cube,INPUT"),
		[]byte("INPUT F-COVER 2,10,cube,10,01,cube,INPUT"),
	} {
		if _, _, parseErr := ParseExtraSupportMIN(malformed, ExtraSupportLimits{}); !errors.Is(parseErr, ErrFormat) {
			t.Fatal(string(malformed), parseErr)
		}
	}
	if _, _, err = LoadExtraSupportMIN(path, ExtraSupportLimits{MaxOutputBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestExtraSupportLimitsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewExtraSupportGreedy([]uint64{1}, 1, ExtraSupportLimits{MaxRows: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	state, _ := NewExtraSupportGreedy([]uint64{1, 2}, 2, ExtraSupportLimits{MaxOperations: 1})
	before, _ := state.Rows()
	if _, err := state.Profile(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := state.Rows()
	if !slices.Equal(before, after) {
		t.Fatal("limit changed rows")
	}
	if _, err := NewExtraSupportGreedy([]uint64{1}, 1, ExtraSupportLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	calls := 0
	state, _ = NewExtraSupportGreedy([]uint64{1}, 1, ExtraSupportLimits{Stop: func() bool { calls++; return calls > 1 }})
	if _, err := state.Profile(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, _, err := ParseExtraSupportMIN([]byte("INPUT F-COVER 1,10,cube,INPUT"), ExtraSupportLimits{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzExtraSupport(f *testing.F) {
	for _, seed := range []struct {
		bits, count uint8
		values      uint64
	}{{2, 2, 1}, {4, 6, 0x123456}, {8, 16, ^uint64(0)}} {
		f.Add(seed.bits, seed.count, seed.values)
	}
	f.Fuzz(func(t *testing.T, bitsSeed, countSeed uint8, values uint64) {
		bitsCount := int(bitsSeed%7) + 2
		space := 1 << uint(bitsCount)
		count := int(countSeed%15) + 2
		if count > space {
			count = space
		}
		cubes := make([]uint64, 0, count)
		seen := make(map[uint64]bool)
		state := values | 1
		for len(cubes) < count {
			state = state*6364136223846793005 + 1442695040888963407
			cube := state & uint64(space-1)
			if seen[cube] {
				state++
				continue
			}
			seen[cube] = true
			cubes = append(cubes, cube)
		}
		limits := ExtraSupportLimits{MaxRows: 1 << 16, MaxPairs: 1 << 16, MaxOperations: 1 << 22}
		pairs, err := DifferenceExtraSupport(cubes, bitsCount, limits)
		greedy, err2 := MinimizeExtraSupport(pairs, bitsCount, limits)
		verification, err3 := VerifyExtraSupport(cubes, greedy.Matrix, bitsCount, limits)
		if err != nil || err2 != nil || err3 != nil || verification.Collisions != 0 {
			t.Fatal(bitsCount, cubes, err, err2, err3, verification)
		}
		reduction, err := SolveExtraSupport(pairs, bitsCount, limits)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range pairs {
			if row&reduction.Active == 0 {
				t.Fatal("uncovered row", row, reduction)
			}
		}
		for variable := 0; variable < bitsCount; variable++ {
			bit := uint64(1) << uint(variable)
			if reduction.Active&bit == 0 {
				continue
			}
			necessary := false
			for _, row := range pairs {
				if row&reduction.Active == bit {
					necessary = true
					break
				}
			}
			if !necessary {
				t.Fatal("non-irredundant active variable", variable, reduction)
			}
		}
		unique, pairErr := GenerateExtraSupportPairs(cubes, bitsCount, limits)
		if pairErr != nil {
			t.Fatal(pairErr)
		}
		pairSeen := make(map[uint64]bool)
		for _, row := range unique {
			if pairSeen[row] {
				t.Fatal("duplicate pair", row)
			}
			pairSeen[row] = true
		}
	})
}
