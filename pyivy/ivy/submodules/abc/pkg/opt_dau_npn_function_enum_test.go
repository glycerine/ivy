package abc

// Go-only frontier, candidate-order and invariant checks for function enumeration.

import (
	"errors"
	"slices"
	"testing"
)

func TestEnumerateDauFunctionsFirstLevel(t *testing.T) {
	result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 3, TargetVariables: 3, MaxNodes: 1})
	if err != nil {
		t.Fatal(err)
	}
	classes := result.Classes()
	levels := result.Levels()
	if result.Inputs() != 3 || result.TargetVariables() != 3 || len(classes) != 4 || len(levels) != 1 {
		t.Fatal(result, classes, levels)
	}
	if classes[0] != (DauFunctionClass{Truth: 0, Nodes: 0, Support: 0}) || classes[1].Nodes != 0 || classes[1].Support != 1 {
		t.Fatal(classes)
	}
	if levels[0] != (DauFunctionLevel{Nodes: 1, Start: 2, End: 4, NewTarget: 2, Candidates: 3}) {
		t.Fatal(levels[0])
	}
	for _, class := range classes[2:] {
		if class.Nodes != 1 || class.Support != 2 {
			t.Fatal(class)
		}
	}
	and := uint64(0)
	xor := uint64(0)
	for assignment := 0; assignment < 8; assignment++ {
		a, b := assignment&1 != 0, assignment&2 != 0
		if a && b {
			and |= 1 << uint(assignment)
		}
		if a != b {
			xor |= 1 << uint(assignment)
		}
	}
	tables := make([]*DauNPNClasses, 4)
	for variables := range tables {
		tables[variables], _ = EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
	}
	and, _, _ = canonicalizeDauFunctionTruth(and, 3, tables)
	xor, _, _ = canonicalizeDauFunctionTruth(xor, 3, tables)
	truths := []uint64{classes[2].Truth, classes[3].Truth}
	if !slices.Contains(truths, and) || !slices.Contains(truths, xor) {
		t.Fatal(classes, and, xor)
	}
	classes[0].Truth = ^uint64(0)
	levels[0].End = 0
	if result.Classes()[0].Truth != 0 || result.Levels()[0].End != 4 {
		t.Fatal("result accessors alias storage")
	}
}

func TestEnumerateDauFunctionsReductionAndLimits(t *testing.T) {
	reduced, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 3, TargetVariables: 3, MaxNodes: 1, Reduce: true})
	if err != nil || len(reduced.Classes()) != 2 || len(reduced.Levels()) != 1 || reduced.Levels()[0].NewTarget != 0 {
		t.Fatal(reduced, err)
	}
	for _, options := range []DauFunctionEnumerationOptions{
		{Inputs: 2, TargetVariables: 2, MaxNodes: 1},
		{Inputs: 6, TargetVariables: 7, MaxNodes: 1},
		{Inputs: 7, TargetVariables: 5, MaxNodes: 1},
		{Inputs: 3, TargetVariables: 3, MaxNodes: -1},
		{Inputs: 5, TargetVariables: 5, MaxNodes: 1, MaxCanonicalScheduleSteps: -1},
	} {
		if result, err := EnumerateDauFunctions(options); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 3, TargetVariables: 3, MaxNodes: 1, MaxClasses: 3}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 3, TargetVariables: 3, MaxNodes: 1, MaxCandidates: 2}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 3, TargetVariables: 3, MaxNodes: 1, Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	for _, inputs := range []int{5, 6} {
		result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: inputs, TargetVariables: inputs, MaxNodes: 1})
		if err != nil || len(result.Classes()) != 4 {
			t.Fatal(inputs, result, err)
		}
	}
	if result, err := EnumerateDauFunctions(DauFunctionEnumerationOptions{Inputs: 5, TargetVariables: 5, MaxNodes: 4, MaxCanonicalCandidates: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("wide canonical budget", result, err)
	}
}

func testDauWideCanonicalizer(t *testing.T, variables int, truth uint64, phase int, output bool) {
	t.Helper()
	tables := make([]*DauNPNClasses, variables+1)
	for width := 0; width <= 4; width++ {
		tables[width], _ = EnumerateDauNPNClasses(width, DauNPNClassOptions{})
	}
	canonicalizer := dauFunctionCanonicalizer{inputs: variables, tables: tables}
	canonical, support, err := canonicalizer.canonicalize(truth & dauTruthMask(variables))
	if err != nil {
		t.Fatal(err)
	}
	permutation := make([]int, variables)
	for index := range permutation {
		permutation[index] = (index + 1) % variables
	}
	transformed := transformDauNPNTruth(truth&dauTruthMask(variables), variables, permutation, phase)
	if output {
		transformed = (^transformed) & dauTruthMask(variables)
	}
	other, otherSupport, err := canonicalizer.canonicalize(transformed)
	if err != nil || canonical != other || support != otherSupport {
		t.Fatal(variables, truth, transformed, canonical, other, support, otherSupport, err)
	}
}

func TestDauFunctionWideExactCanonicalization(t *testing.T) {
	testDauWideCanonicalizer(t, 5, 0x96a53c7e, 0x15, true)
	testDauWideCanonicalizer(t, 6, 0xd6a59c3e70b4f821, 0x2b, true)

	tables := make([]*DauNPNClasses, 7)
	for width := 0; width <= 4; width++ {
		tables[width], _ = EnumerateDauNPNClasses(width, DauNPNClassOptions{})
	}
	limited := dauFunctionCanonicalizer{inputs: 6, tables: tables, limited: true, remaining: 92159}
	if truth, support, err := limited.canonicalize(0xd6a59c3e70b4f821); truth != 0 || support != 0 || !errors.Is(err, ErrLimit) {
		t.Fatal(truth, support, err)
	}
	stopped := dauFunctionCanonicalizer{inputs: 6, tables: tables, stop: func() bool { return true }}
	if truth, support, err := stopped.canonicalize(0xd6a59c3e70b4f821); truth != 0 || support != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatal(truth, support, err)
	}
}

func checkDauFunctionEnumeration(t *testing.T, result *DauFunctionEnumeration) {
	t.Helper()
	classes := result.Classes()
	tables := make([]*DauNPNClasses, result.inputs+1)
	for variables := range tables {
		tables[variables], _ = EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
	}
	for index, class := range classes {
		if class.Support != len(dauLogicalSupport(class.Truth, result.inputs)) {
			t.Fatal("support", index, class)
		}
		canonical, support, err := canonicalizeDauFunctionTruth(class.Truth, result.inputs, tables)
		if err != nil || canonical != class.Truth || support != class.Support {
			t.Fatal("canonical", index, class, canonical, support, err)
		}
		for previous := 0; previous < index; previous++ {
			if classes[previous].Truth == class.Truth {
				t.Fatal("duplicate", previous, index, class.Truth)
			}
		}
		if index > 0 && class.Nodes < classes[index-1].Nodes {
			t.Fatal("node order", index, classes[index-1], class)
		}
	}
	for _, level := range result.Levels() {
		if level.Start < 0 || level.Start > level.End || level.End > len(classes) {
			t.Fatal(level)
		}
		for _, class := range classes[level.Start:level.End] {
			if class.Nodes != level.Nodes {
				t.Fatal(level, class)
			}
		}
	}
}

func TestEnumerateDauFunctionsFrontierInvariants(t *testing.T) {
	for _, options := range []DauFunctionEnumerationOptions{
		{Inputs: 3, TargetVariables: 3, MaxNodes: 3},
		{Inputs: 3, TargetVariables: 3, MaxNodes: 3, UseTwo: true},
		{Inputs: 4, TargetVariables: 3, MaxNodes: 2},
		{Inputs: 4, TargetVariables: 4, MaxNodes: 2, Reduce: true},
	} {
		first, err := EnumerateDauFunctions(options)
		if err != nil {
			t.Fatal(options, err)
		}
		checkDauFunctionEnumeration(t, first)
		second, err := EnumerateDauFunctions(options)
		if err != nil || !slices.Equal(first.Classes(), second.Classes()) || !slices.Equal(first.Levels(), second.Levels()) {
			t.Fatal("nondeterministic", options, err)
		}
	}
}

func FuzzEnumerateDauFunctions(f *testing.F) {
	f.Add(uint8(3), uint8(2), false, false)
	f.Add(uint8(4), uint8(1), true, true)
	f.Fuzz(func(t *testing.T, width, depth uint8, useTwo, reduce bool) {
		inputs := 3 + int(width%2)
		maxNodes := int(depth % 3)
		options := DauFunctionEnumerationOptions{Inputs: inputs, TargetVariables: inputs, MaxNodes: maxNodes, UseTwo: useTwo, Reduce: reduce, MaxClasses: 10000, MaxCandidates: 1 << 20}
		result, err := EnumerateDauFunctions(options)
		if err != nil {
			t.Fatal(options, err)
		}
		checkDauFunctionEnumeration(t, result)
	})
}

func FuzzDauFunctionWideCanonicalization(f *testing.F) {
	f.Add(uint64(0x96a53c7e), uint8(5), uint8(0x15), true)
	f.Add(uint64(0xd6a59c3e70b4f821), uint8(6), uint8(0x2b), false)
	f.Fuzz(func(t *testing.T, truth uint64, width, phase uint8, output bool) {
		variables := 5 + int(width%2)
		testDauWideCanonicalizer(t, variables, truth, int(phase)&((1<<uint(variables))-1), output)
	})
}
