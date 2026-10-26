package abc

// Go-only orbit, frontier, multiplicity and fuzz tests for DTT enumeration.

import (
	"errors"
	"slices"
	"testing"
)

func independentDttOrbit(truth uint64, variables int) []uint64 {
	mask := dauTruthMask(variables)
	top := uint64(1) << uint((1<<uint(variables))-1)
	seen := make(map[uint64]bool)
	result := []uint64{}
	for _, permutation := range dauNPNPermutations(variables) {
		for phase := 0; phase < 1<<uint(variables); phase++ {
			candidate := transformDauNPNTruth(truth&mask, variables, permutation, phase)
			if candidate&top != 0 {
				candidate = (^candidate) & mask
			}
			if !seen[candidate] {
				seen[candidate] = true
				result = append(result, candidate)
			}
		}
	}
	slices.Sort(result)
	return result
}

func TestDttOrbitSourceWalk(t *testing.T) {
	walker, err := newDttOrbitWalker(3, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, truth := range []uint64{0xaa, 0x80, 0x96, 0xe8} {
		actual, err := walker.orbit(truth)
		if err != nil {
			t.Fatal(err)
		}
		sorted := slices.Clone(actual)
		slices.Sort(sorted)
		if !slices.Equal(sorted, independentDttOrbit(truth, 3)) {
			t.Fatal(truth, actual, sorted, independentDttOrbit(truth, 3))
		}
	}
	limited, _ := newDttOrbitWalker(3, 47, nil)
	if orbit, err := limited.orbit(0xaa); orbit != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(orbit, err)
	}
	stopped, _ := newDttOrbitWalker(3, 0, func() bool { return true })
	if orbit, err := stopped.orbit(0xaa); orbit != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(orbit, err)
	}
}

func checkDttEnumeration(t *testing.T, result *DttEnumeration, delay bool) {
	t.Helper()
	functions := result.Functions()
	classes := result.Classes()
	classCounts := make(map[uint64]int)
	classFrontier := make(map[uint64]int)
	for index, function := range functions {
		classCounts[function.Canonical]++
		if previous, found := classFrontier[function.Canonical]; !found || function.Frontier < previous {
			classFrontier[function.Canonical] = function.Frontier
		}
		if index < 2 {
			if function.Fanin0 != -1 || function.Fanin1 != -1 || function.Nodes != 0 || function.Frontier != 0 {
				t.Fatal("initial", index, function)
			}
			continue
		}
		if function.Fanin0 < 0 || function.Fanin0 >= index || function.Fanin1 < 0 || function.Fanin1 >= index || function.Type > 4 {
			t.Fatal("fanin", index, function)
		}
		left, right := functions[function.Fanin0], functions[function.Fanin1]
		if function.Nodes != left.Nodes+right.Nodes+1 {
			t.Fatal("nodes", index, function, left, right)
		}
		wantFrontier := left.Frontier + right.Frontier + 1
		if delay {
			wantFrontier = max(left.Frontier, right.Frontier) + 1
		}
		if function.Frontier != wantFrontier {
			t.Fatal("frontier", index, function, wantFrontier)
		}
		transformed, err := function.LeftTransform.ApplyTruth(left.Truth)
		if err != nil {
			t.Fatal("transform", index, err)
		}
		if function.LeftOutputComplement {
			transformed = (^transformed) & dauTruthMask(result.variables)
		}
		if dttCandidate(function.Type, transformed, right.Truth, dauTruthMask(result.variables)) != function.Truth {
			t.Fatal("retained transform", index, function, transformed)
		}
		walker, _ := newDttOrbitWalker(result.variables, 0, nil)
		orbit, _ := walker.orbit(left.Truth)
		found := false
		for _, transformed := range orbit {
			if dttCandidate(function.Type, transformed, right.Truth, walker.mask) == function.Truth {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("unreplayable", index, function)
		}
	}
	seen := make(map[uint64]bool)
	for _, class := range classes {
		if seen[class.Canonical] || class.Implementations != classCounts[class.Canonical] || class.Frontier != classFrontier[class.Canonical] {
			t.Fatal("class", class, classCounts[class.Canonical], classFrontier[class.Canonical])
		}
		seen[class.Canonical] = true
	}
	levels := result.Levels()
	for index, level := range levels {
		if level.Frontier != index || level.Start < 0 || level.Start > level.End || level.End > len(functions) {
			t.Fatal("level", index, level)
		}
		for _, function := range functions[level.Start:level.End] {
			if function.Frontier != level.Frontier {
				t.Fatal("level member", level, function)
			}
		}
	}
}

func TestEnumerateDttFunctionsAllThreeVariableClasses(t *testing.T) {
	result, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, MaxFrontier: 8})
	if err != nil {
		t.Fatal(err)
	}
	checkDttEnumeration(t, result, false)
	if result.Variables() != 3 || len(result.Classes()) != 14 || len(result.Functions()) != 14 {
		t.Fatal(result.Variables(), len(result.Classes()), len(result.Functions()), result.Levels())
	}
	got := make([]uint64, len(result.Classes()))
	for index, class := range result.Classes() {
		got[index] = class.Canonical
		if class.Implementations != 1 {
			t.Fatal(class)
		}
	}
	slices.Sort(got)
	exact, _ := EnumerateDauNPNClasses(3, DauNPNClassOptions{})
	if !slices.Equal(got, exact.Representatives()) {
		t.Fatal(got, exact.Representatives())
	}
	functions, levels := result.Functions(), result.Levels()
	functions[0].Truth = 1
	levels[0].End = 0
	if result.Functions()[0].Truth != 0 || result.Levels()[0].End != 2 {
		t.Fatal("accessors alias storage")
	}
}

func TestEnumerateDttMultiplicityAndDelay(t *testing.T) {
	multiple, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, MaxFrontier: 2, Multiplicity: true, MaxFunctions: 10000})
	if err != nil {
		t.Fatal(err)
	}
	checkDttEnumeration(t, multiple, false)
	duplicate, counted := false, false
	for _, class := range multiple.Classes() {
		duplicate = duplicate || class.Implementations > 1
		counted = counted || class.Multiplicity > 0
	}
	if !duplicate || !counted {
		t.Fatal("multiplicity not retained", len(multiple.Functions()), multiple.Classes())
	}
	delay, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, MaxFrontier: 2, Delay: true})
	if err != nil {
		t.Fatal(err)
	}
	checkDttEnumeration(t, delay, true)
	wider := false
	for _, function := range delay.Functions() {
		wider = wider || function.Nodes > function.Frontier
	}
	if !wider {
		t.Fatal("delay frontier did not diverge from node count")
	}
}

func TestEnumerateDttLimits(t *testing.T) {
	for _, options := range []DttEnumerationOptions{{Variables: 1}, {Variables: 6}, {Variables: 3, MaxFrontier: -1}, {Variables: 3, MaxFrontier: 32}, {Variables: 3, MaxFunctions: -1}} {
		if result, err := EnumerateDttFunctions(options); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(options, result, err)
		}
	}
	for _, options := range []DttEnumerationOptions{
		{Variables: 3, MaxFrontier: 1, MaxFunctions: 2},
		{Variables: 3, MaxFrontier: 1, MaxCandidates: 1},
		{Variables: 3, MaxFrontier: 1, MaxOrbitTransforms: 47},
	} {
		if result, err := EnumerateDttFunctions(options); result != nil || !errors.Is(err, ErrLimit) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: 3, Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
}

func FuzzEnumerateDttFunctions(f *testing.F) {
	f.Add(uint8(3), uint8(2), false, false)
	f.Add(uint8(2), uint8(1), true, true)
	f.Fuzz(func(t *testing.T, width, depth uint8, delay, multiplicity bool) {
		variables := 2 + int(width%2)
		frontier := int(depth % 3)
		result, err := EnumerateDttFunctions(DttEnumerationOptions{Variables: variables, MaxFrontier: frontier, Delay: delay, Multiplicity: multiplicity, MaxFunctions: 10000, MaxCandidates: 1 << 20, MaxOrbitTransforms: 1 << 24})
		if err != nil {
			t.Fatal(variables, frontier, delay, multiplicity, err)
		}
		checkDttEnumeration(t, result, delay)
	})
}
