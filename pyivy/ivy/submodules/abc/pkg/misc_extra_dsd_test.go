package abc

// Go-only tests for src/misc/extra/extraUtilDsd.c behavior.

import (
	"errors"
	"slices"
	"testing"
)

func extraDSDTruthFromConfig(classTruth uint64, permutation [6]int, phase, outputComplement int) uint64 {
	var truth uint64
	for oldAssignment := 0; oldAssignment < 64; oldAssignment++ {
		classAssignment := 0
		for oldVariable := 0; oldVariable < 6; oldVariable++ {
			bit := (oldAssignment >> uint(oldVariable)) & 1
			bit ^= (phase >> uint(oldVariable)) & 1
			classAssignment |= bit << uint(permutation[oldVariable])
		}
		value := classTruth >> uint(classAssignment) & 1
		if outputComplement != 0 {
			value ^= 1
		}
		truth |= value << uint(oldAssignment)
	}
	return truth
}

func TestExtraDSDClassDatasetAndFormulas(t *testing.T) {
	manager, err := NewExtraDSDManager(ExtraDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	classes := manager.Classes()
	if len(classes) != 595 {
		t.Fatal(len(classes))
	}
	want := []ExtraDSDClass{{0, 0, 1, 0, "0"}, {1, 0, 2, 0xaaaaaaaaaaaaaaaa, "a"}, {2, 1, 3, 0x8888888888888888, "(ab)"}, {2, 3, 4, 0x6666666666666666, "[ab]"}}
	if !slices.Equal(classes[:4], want) {
		t.Fatalf("prefix = %+v", classes[:4])
	}
	if classes[594] != (ExtraDSDClass{6, 9, 8, 0xffff0000f066f066, "<<[ab]cd>ef>"}) {
		t.Fatalf("last = %+v", classes[594])
	}
	for index, class := range classes {
		support := 0
		for variable := 0; variable < 6; variable++ {
			if extraDSDHasVariable(class.Truth, variable) {
				support++
			}
		}
		if support != class.Variables {
			t.Fatalf("class %d support = %d want %d", index, support, class.Variables)
		}
		config, checkErr := manager.Check(class.Truth)
		if checkErr != nil || config>>17 != index {
			t.Fatalf("class %d config = %x %v", index, config, checkErr)
		}
	}
	costs := manager.CNFCosts()
	if len(costs) != 595 || costs[0] != 0 || costs[1] != 0 || costs[2] != 3 || costs[594] != 8 {
		t.Fatalf("costs endpoints %v %v", costs[:3], costs[594])
	}
}

func TestExtraDSDSchedulesAndPermutationResolution(t *testing.T) {
	gray, err := ExtraGrayCodeSchedule(3)
	if err != nil || !slices.Equal(gray, []int{0, 1, 0, 2, 0, 1, 0, 2}) {
		t.Fatalf("gray = %v %v", gray, err)
	}
	schedule, err := ExtraPermutationSchedule(4)
	if err != nil || len(schedule) != 24 {
		t.Fatalf("schedule = %v %v", schedule, err)
	}
	permutation := []int{0, 1, 2, 3}
	seen := map[[4]int]struct{}{}
	for _, swap := range schedule {
		permutation[swap], permutation[swap+1] = permutation[swap+1], permutation[swap]
		seen[[4]int(permutation)] = struct{}{}
	}
	if len(seen) != 24 || !slices.Equal(permutation, []int{0, 1, 2, 3}) {
		t.Fatalf("permutation cycle = %v %d", permutation, len(seen))
	}
	manager, _ := NewExtraDSDManager(ExtraDSDOptions{})
	mask := 0x3ffff ^ ((7 ^ 0) << 6) ^ ((7 ^ 1) << 9)
	formatted, err := manager.FormatPermutation(mask)
	if err != nil || formatted != "2 3 0 1 4 5  " {
		t.Fatalf("formatted = %q %v", formatted, err)
	}
	partial := 7 | (2 << 3) | (7 << 6) | (0 << 9) | (7 << 12) | (7 << 15)
	resolved, index, err := manager.resolvePermutation(partial)
	if err != nil || resolved != ([6]int{1, 2, 3, 0, 4, 5}) || index < 0 {
		t.Fatalf("partial = %v %d %v", resolved, index, err)
	}
}

func TestExtraDSDCheckTransformedAndCompute(t *testing.T) {
	manager, _ := NewExtraDSDManager(ExtraDSDOptions{})
	probe := uint64(0x123456789abcdef0)
	for index, permutation := range manager.perms {
		permutedProbe := extraDSDPermuteTruth(probe, permutation)
		if direct := extraDSDTransformToClass(probe, permutation, 0); direct != permutedProbe {
			t.Fatalf("permutation direction %d: %x != %x (%v)", index, direct, permutedProbe, permutation)
		}
	}
	class, _ := manager.Class(130)
	permutation := manager.perms[317]
	truth := extraDSDTruthFromConfig(class.Truth, permutation, 0x2d, 1)
	config, err := manager.Check(truth)
	if err != nil || config < 0 || config>>17 != 130 {
		t.Fatalf("config = %x %v", config, err)
	}
	foundClass := config >> 17
	foundOutput := (config >> 16) & 1
	foundPerm := (config >> 6) & 0x3ff
	foundPhase := config & 63
	replayed := extraDSDTruthFromConfig(manager.classes[foundClass].Truth, manager.perms[foundPerm], foundPhase, foundOutput)
	if replayed != truth {
		t.Fatalf("replay %x want %x", replayed, truth)
	}
	mask := 0x3ffff ^ ((7 ^ 0) << 6) ^ ((7 ^ 1) << 9)
	result, err := manager.Compute(4, 6, []int{4, 0, 2, 4, 6}, mask, false)
	if err != nil || result.Literal < 0 || len(result.Cut) != result.Cut[0]+1 {
		t.Fatalf("compute = %+v %v", result, err)
	}
	outputClass := manager.classes[result.Class]
	var fromCut uint64
	for assignment := 0; assignment < 64; assignment++ {
		classAssignment := 0
		for variable, literal := range result.Cut[1:] {
			bit := (assignment >> uint(literal>>1)) & 1
			if literal&1 != 0 {
				bit ^= 1
			}
			classAssignment |= bit << uint(variable)
		}
		value := outputClass.Truth >> uint(classAssignment) & 1
		if result.Literal&1 != 0 {
			value ^= 1
		}
		fromCut |= value << uint(assignment)
	}
	if fromCut != result.Truth {
		t.Fatalf("cut truth %x want %x", fromCut, result.Truth)
	}
	stats := manager.Statistics()
	if stats.All != 1 || stats.NonDSD != 0 || stats.ClassCounts[result.Class] != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	report, _ := manager.StatisticsReport(true)
	if report.Unused != 594 || len(report.Rows) != 1 {
		t.Fatalf("report = %+v", report)
	}
	manager.ResetStatistics()
	if manager.Statistics().All != 0 {
		t.Fatal("reset")
	}
}

func TestExtraDSDDivisorsLimitsAndErrors(t *testing.T) {
	manager, _ := NewExtraDSDManager(ExtraDSDOptions{})
	and := truthVariablePattern(0) & truthVariablePattern(1)
	divisors, err := manager.CollectDivisors(and)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(divisors[2], []uint64{and}) || len(divisors[3])+len(divisors[4]) != 0 {
		t.Fatalf("divisors = %v", divisors)
	}
	if _, err = NewExtraDSDManager(ExtraDSDOptions{MaxChecks: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("options = %v", err)
	}
	cancelled, _ := NewExtraDSDManager(ExtraDSDOptions{Stop: func() bool { return false }})
	cancelled.options.Stop = func() bool { return true }
	if _, err = cancelled.Check(0x1234); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %v", err)
	}
	limited, _ := NewExtraDSDManager(ExtraDSDOptions{MaxChecks: 1})
	if _, err = limited.Check(0x123456789abcdef0); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit = %v", err)
	}
	if _, err = manager.Compute(-1, 0, nil, 2, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("compute dimension = %v", err)
	}
	if _, _, err = manager.resolvePermutation(6); !errors.Is(err, ErrDimension) {
		t.Fatalf("perm dimension = %v", err)
	}
}

func FuzzExtraDSDClassificationReplay(f *testing.F) {
	for _, seed := range []uint64{0, 1, 0x123456789abcdef, 0xffffffffffffffff} {
		f.Add(seed)
	}
	manager, err := NewExtraDSDManager(ExtraDSDOptions{})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		classIndex := int(seed % 595)
		permutationIndex := int(seed >> 10 % 720)
		phase := int(seed>>32) & 63
		output := int(seed >> 63)
		truth := extraDSDTruthFromConfig(manager.classes[classIndex].Truth, manager.perms[permutationIndex], phase, output)
		config, checkErr := manager.Check(truth)
		if checkErr != nil {
			t.Fatal(checkErr)
		}
		if config < 0 {
			t.Fatalf("not classified: class %d", classIndex)
		}
		foundClass := config >> 17
		foundOutput := (config >> 16) & 1
		foundPermutation := (config >> 6) & 0x3ff
		foundPhase := config & 63
		replayed := extraDSDTruthFromConfig(manager.classes[foundClass].Truth, manager.perms[foundPermutation], foundPhase, foundOutput)
		if replayed != truth {
			t.Fatalf("seed %x source class=%d perm=%d phase=%d out=%d; found class=%d perm=%d phase=%d out=%d; replay %x want %x", seed, classIndex, permutationIndex, phase, output, foundClass, foundPermutation, foundPhase, foundOutput, replayed, truth)
		}
	})
}
