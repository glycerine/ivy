package abc

// Stand-alone tests for SAT-guided SOP collapse from src/sat/bmc/bmcClp.c.

import (
	"errors"
	"slices"
	"testing"
)

func giaCollapseGraph(t testing.TB) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	giaTestHash(t, manager)
	a, b, c := giaTestCI(t, manager), giaTestCI(t, manager), giaTestCI(t, manager)
	ab := giaTestAnd(t, manager, a, b)
	ac := giaTestAnd(t, manager, a, c)
	root, err := manager.HashOr(ab, ac)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, manager, root)
	giaTestCheck(t, manager)
	return manager
}

func TestGiaOutputSensitivityAndStatistics(t *testing.T) {
	manager := giaCollapseGraph(t)
	sample, err := manager.OutputSensitivity(0, []bool{true, false, false}, 0, nil)
	if err != nil || sample.Value || !slices.Equal(sample.Differences, []bool{false, true, true}) {
		t.Fatal(sample, err)
	}
	patterns := make([][]bool, 8)
	for assignment := range patterns {
		patterns[assignment] = []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
	}
	stats, err := manager.AnalyzeOutputSensitivity(0, patterns, 8, 4, nil)
	if err != nil || stats.Samples != 8 || !slices.Equal(stats.DifferenceCounts, []int{6, 2, 2}) {
		t.Fatal(stats, err)
	}
	truth := func(values []bool) bool { return values[0] && (values[1] || values[2]) }
	wantPatterns := make([][2][2]int, 3)
	wantPairs := [2][][]int{make([][]int, 3), make([][]int, 3)}
	for phase := range wantPairs {
		for variable := range wantPairs[phase] {
			wantPairs[phase][variable] = make([]int, 3)
		}
	}
	for _, pattern := range patterns {
		phase := 0
		if truth(pattern) {
			phase = 1
		}
		differences := make([]bool, 3)
		for variable, value := range pattern {
			bit := 0
			if value {
				bit = 1
			}
			wantPatterns[variable][phase][bit]++
			changed := slices.Clone(pattern)
			changed[variable] = !changed[variable]
			differences[variable] = truth(changed) != truth(pattern)
		}
		for left := range differences {
			for right := left + 1; right < len(differences); right++ {
				if differences[left] && differences[right] {
					wantPairs[phase][left][right]++
				}
			}
		}
	}
	if !slices.Equal(stats.PatternCounts, wantPatterns) {
		t.Fatal("pattern counts", stats.PatternCounts, wantPatterns)
	}
	for phase := range wantPairs {
		for variable := range wantPairs[phase] {
			if !slices.Equal(stats.CoDifferenceCounts[phase][variable], wantPairs[phase][variable]) {
				t.Fatal("pair counts", phase, variable, stats.CoDifferenceCounts, wantPairs)
			}
		}
	}
	if _, err = manager.OutputSensitivity(1, patterns[0], 0, nil); !errors.Is(err, ErrDimension) {
		t.Fatal("output bounds", err)
	}
	if _, err = manager.OutputSensitivity(0, patterns[0], 3, nil); !errors.Is(err, ErrLimit) {
		t.Fatal("evaluation limit", err)
	}
	if _, err = manager.AnalyzeOutputSensitivity(0, patterns, 7, 0, nil); !errors.Is(err, ErrLimit) {
		t.Fatal("sample limit", err)
	}
	if _, err = manager.AnalyzeOutputSensitivity(0, patterns, 8, 4, func() bool { return true }); !errors.Is(err, ErrCancelled) {
		t.Fatal("sensitivity cancellation", err)
	}
}

func verifyGiaCollapseCover(t testing.TB, manager *GiaManager, result GiaCollapseResult) {
	t.Helper()
	if result.Status != SatSatisfiable || result.Reason != SatFinished {
		t.Fatal("collapse status", result.Status, result.Reason)
	}
	inputs := len(manager.cis)
	for assignment := 0; assignment < 1<<inputs; assignment++ {
		values := make([]bool, inputs)
		for variable := range values {
			values[variable] = assignment&(1<<variable) != 0
		}
		outputs, err := manager.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		var got bool
		if result.Cover.Variables == 0 {
			got, err = result.Cover.Evaluate(nil)
		} else {
			got, err = result.Cover.Evaluate(values)
		}
		if err != nil || got != outputs[0] {
			t.Fatalf("assignment %03b got=%v want=%v cover=%+v err=%v", assignment, got, outputs[0], result.Cover, err)
		}
	}
}

func giaCollapseCubeMatches(cube string, values []bool) bool {
	for variable := range cube {
		if cube[variable] == '0' && values[variable] || cube[variable] == '1' && !values[variable] {
			return false
		}
	}
	return true
}

func TestGiaCollapseSOPSemanticsOrderAndOwnership(t *testing.T) {
	manager := giaCollapseGraph(t)
	before := slices.Clone(manager.objects)
	forward, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: true})
	if err != nil {
		t.Fatal(err)
	}
	verifyGiaCollapseCover(t, manager, forward)
	if !slices.Equal(before, manager.objects) || len(forward.Cover.Cubes) == 0 || forward.EnumeratedOn+forward.EnumeratedOff == 0 {
		t.Fatal("ownership or enumeration", forward)
	}
	again, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: true})
	if err != nil || !slices.Equal(forward.Cover.Cubes, again.Cover.Cubes) || forward.Cover.Output != again.Cover.Output {
		t.Fatal("canonical determinism", forward.Cover, again.Cover, err)
	}
	reverse, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: true, Reverse: true})
	if err != nil {
		t.Fatal(err)
	}
	verifyGiaCollapseCover(t, manager, reverse)
	for _, cube := range reverse.Cover.Cubes {
		if len(cube) != len(manager.cis) {
			t.Fatal("reverse changed external column order", cube)
		}
	}
	noncanonical, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	verifyGiaCollapseCover(t, manager, noncanonical)
	forward.Cover.Cubes[0] = "---"
	fresh, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: true})
	if err != nil || fresh.Cover.Cubes[0] == "---" {
		t.Fatal("returned cover aliases future call", fresh.Cover, err)
	}
}

func TestGiaCollapseConstantsPhasesAndIrredundancy(t *testing.T) {
	for _, literal := range []GiaLit{GiaConst0, GiaConst1} {
		manager := NewGiaManager()
		giaTestCI(t, manager)
		giaTestCO(t, manager, literal)
		result, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{})
		if err != nil || result.Cover.Variables != 0 || len(result.Cover.Cubes) != 1 || result.Cover.Cubes[0] != "" {
			t.Fatal(literal, result, err)
		}
		verifyGiaCollapseCover(t, manager, result)
	}
	manager := giaCollapseGraph(t)
	result, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: true})
	if err != nil {
		t.Fatal(err)
	}
	for current, cube := range result.Cover.Cubes {
		others := cloneSOPCover(result.Cover)
		others.Cubes = append(slices.Clone(others.Cubes[:current]), others.Cubes[current+1:]...)
		if len(others.Cubes) == 0 {
			continue
		}
		essential := false
		for assignment := 0; assignment < 8; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			otherMatches := false
			for _, other := range others.Cubes {
				otherMatches = otherMatches || giaCollapseCubeMatches(other, values)
			}
			if giaCollapseCubeMatches(cube, values) && !otherMatches {
				essential = true
			}
		}
		if !essential {
			t.Fatal("redundant cube survived", current, result.Cover)
		}
	}
}

func TestGiaCollapseLimitsCancellationAndDepth(t *testing.T) {
	manager := giaCollapseGraph(t)
	bad := []GiaCollapseOptions{
		{Output: -1}, {Output: 1}, {MaxCubes: -1}, {MaxVariables: -1},
		{MaxClauses: -1}, {MaxLiterals: -1}, {MaxObjects: -1}, {MaxMappingEntries: -1},
	}
	for _, options := range bad {
		if _, err := manager.CollapseGiaOutputToSOP(options); err == nil {
			t.Fatal("accepted invalid options", options)
		}
	}
	limited := []GiaCollapseOptions{{MaxCubes: 1}, {MaxVariables: 2}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxObjects: 1}, {MaxMappingEntries: 1}}
	for _, options := range limited {
		if _, err := manager.CollapseGiaOutputToSOP(options); !errors.Is(err, ErrLimit) {
			t.Fatal("limit", options, err)
		}
	}
	if _, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancellation", err)
	}
	unknown, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{SAT: SatLimits{Propagations: 1}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatPropagationLimit {
		t.Fatal("logical SAT limit", unknown, err)
	}
	deep := NewGiaManager()
	giaTestHash(t, deep)
	input := giaTestCI(t, deep)
	second := giaTestCI(t, deep)
	root := input
	for index := 0; index < 10000; index++ {
		var appendErr error
		root, appendErr = deep.AppendAnd(root, second)
		if appendErr != nil {
			t.Fatal(appendErr)
		}
	}
	giaTestCO(t, deep, root)
	result, err := deep.CollapseGiaOutputToSOP(GiaCollapseOptions{MaxObjects: 20000, MaxMappingEntries: 20000, MaxVariables: 20000, MaxClauses: 50000, MaxLiterals: 150000})
	if err != nil {
		t.Fatal(err)
	}
	verifyGiaCollapseCover(t, deep, result)
}

func FuzzGiaCollapseSOP(f *testing.F) {
	for _, seed := range []uint8{0x00, 0xff, 0x80, 0x96, 0xe8, 0xca} {
		f.Add(seed, false, false)
	}
	f.Fuzz(func(t *testing.T, truth uint8, canonical, reverse bool) {
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
		result, err := manager.CollapseGiaOutputToSOP(GiaCollapseOptions{Canonical: canonical, Reverse: reverse, MaxCubes: 32, MaxVariables: 100, MaxClauses: 1000, MaxLiterals: 3000})
		if err != nil {
			t.Fatal(err)
		}
		verifyGiaCollapseCover(t, manager, result)
	})
}

func FuzzGiaOutputSensitivity(f *testing.F) {
	for _, seed := range []struct{ truth, pattern uint8 }{{0, 0}, {0xff, 7}, {0x96, 3}, {0xe8, 5}, {0xca, 6}} {
		f.Add(seed.truth, seed.pattern)
	}
	f.Fuzz(func(t *testing.T, truth, pattern uint8) {
		pattern &= 7
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
		values := []bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0}
		result, err := manager.OutputSensitivity(0, values, 4, nil)
		if err != nil {
			t.Fatal(err)
		}
		wantValue := truth&(1<<pattern) != 0
		wantDifferences := make([]bool, 3)
		for variable := range wantDifferences {
			wantDifferences[variable] = wantValue != (truth&(1<<(pattern^(1<<variable))) != 0)
		}
		if result.Value != wantValue || !slices.Equal(result.Differences, wantDifferences) {
			t.Fatal(truth, pattern, result, wantValue, wantDifferences)
		}
	})
}
