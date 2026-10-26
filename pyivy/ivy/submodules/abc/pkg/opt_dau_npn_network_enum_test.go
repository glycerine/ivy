package abc

// Go-only exact-class and frontier tests for DAU network enumeration.

import (
	"errors"
	"slices"
	"testing"
)

func checkDauNetworkEnumeration(t *testing.T, result *DauNetworkEnumeration) {
	t.Helper()
	classes := result.Classes()
	tables := make([]*DauNPNClasses, result.variables+1)
	for variables := range tables {
		tables[variables], _ = EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
	}
	seen := make(map[uint64]bool, len(classes))
	for index, class := range classes {
		if seen[class.Truth] {
			t.Fatal("duplicate", index, class)
		}
		seen[class.Truth] = true
		canonical, support, err := canonicalizeDauFunctionTruth(class.Truth, result.variables, tables)
		if err != nil || canonical != class.Truth || support != class.Support {
			t.Fatal("canonical", index, class, canonical, support, err)
		}
		if index > 0 && class.Nodes < classes[index-1].Nodes {
			t.Fatal("node order", index, classes[index-1], class)
		}
	}
	for _, level := range result.Levels() {
		if level.Start < 0 || level.Start > level.End || level.End > len(classes) {
			t.Fatal("range", level)
		}
		proper := 0
		for _, class := range classes[level.Start:level.End] {
			if class.Nodes != level.Nodes {
				t.Fatal("cost", level, class)
			}
			if class.Support < result.variables {
				proper++
			}
		}
		if proper != level.NewProper {
			t.Fatal("proper support count", level, proper)
		}
	}
}

func TestEnumerateDauNetworksFirstLevel(t *testing.T) {
	result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 4, MaxNodes: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.Variables() != 4 || len(result.Classes()) != 4 || len(result.Levels()) != 1 {
		t.Fatal(result.Variables(), result.Classes(), result.Levels())
	}
	if result.Levels()[0] != (DauNetworkLevel{Nodes: 1, Start: 2, End: 4, NewProper: 2, Candidates: 3}) {
		t.Fatal(result.Levels()[0])
	}
	checkDauNetworkEnumeration(t, result)
	classes, levels := result.Classes(), result.Levels()
	classes[0].Truth = ^uint64(0)
	levels[0].End = 0
	if result.Classes()[0].Truth != 0 || result.Levels()[0].End != 4 {
		t.Fatal("accessors alias internal storage")
	}
}

func TestEnumerateDauNetworksAllFourInputClasses(t *testing.T) {
	result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 4, MaxNodes: 31})
	if err != nil {
		t.Fatal(err)
	}
	checkDauNetworkEnumeration(t, result)
	if len(result.Classes()) != 222 {
		t.Fatal("incomplete four-input NPN domain", len(result.Classes()), result.Levels())
	}
	if levels := result.Levels(); len(levels) == 0 || levels[len(levels)-1].Start != levels[len(levels)-1].End {
		t.Fatal("enumeration did not reach an empty fixed point", levels)
	}
	exact, err := EnumerateDauNPNClasses(4, DauNPNClassOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]uint64, 0, len(result.Classes()))
	for _, class := range result.Classes() {
		canonical, err := exact.Canonical(class.Truth & 0xffff)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, canonical)
	}
	slices.Sort(got)
	if !slices.Equal(got, exact.Representatives()) {
		t.Fatal("frontier classes differ from independently exhaustive NPN orbits")
	}
}

func TestEnumerateDauNetworksOptions(t *testing.T) {
	for _, options := range []DauNetworkEnumerationOptions{
		{Variables: 0, MaxNodes: 1},
		{Variables: 5, MaxNodes: 1},
		{Variables: 3, MaxNodes: -1},
		{Variables: 3, MaxNodes: 32},
	} {
		if result, err := EnumerateDauNetworks(options); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 1, MaxClasses: 3}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 1, MaxCandidates: 2}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 1, Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	withTwo, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 3, UseTwo: true})
	if err != nil {
		t.Fatal(err)
	}
	checkDauNetworkEnumeration(t, withTwo)
	again, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 3, UseTwo: true})
	if err != nil || !slices.Equal(withTwo.Classes(), again.Classes()) || !slices.Equal(withTwo.Levels(), again.Levels()) {
		t.Fatal("nondeterministic", err)
	}
	withoutTwo, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: 3, MaxNodes: 3})
	if err != nil || withTwo.Levels()[2].Candidates <= withoutTwo.Levels()[2].Candidates {
		t.Fatal("two-node source branch was not explored", err, withTwo.Levels(), withoutTwo.Levels())
	}
}

func TestDauNetworkTwoCandidateAssignmentModel(t *testing.T) {
	truth := uint64(0x96)
	for _, reverse := range []bool{false, true} {
		got := dauNetworkTwoCandidate(truth, 3, 0, 1, reverse)
		for assignment := 0; assignment < 8; assignment++ {
			value := func(first, second bool) bool {
				source := assignment &^ 3
				if first {
					source |= 1
				}
				if second {
					source |= 2
				}
				return truth&(1<<uint(source)) != 0
			}
			first := assignment&1 != 0
			second := assignment&2 != 0
			inner := first != second
			cofactor0 := value(false, !inner)
			cofactor1 := value(true, !inner)
			outer := first && second
			if reverse {
				outer = !outer
			}
			want := cofactor0
			if outer {
				want = cofactor1
			}
			if (got&(1<<uint(assignment)) != 0) != want {
				t.Fatal(reverse, assignment, got, want)
			}
		}
	}
}

func FuzzEnumerateDauNetworks(f *testing.F) {
	f.Add(uint8(3), uint8(2), false)
	f.Add(uint8(4), uint8(1), true)
	f.Fuzz(func(t *testing.T, width, depth uint8, useTwo bool) {
		variables := 1 + int(width%4)
		maxNodes := int(depth % 4)
		result, err := EnumerateDauNetworks(DauNetworkEnumerationOptions{Variables: variables, MaxNodes: maxNodes, UseTwo: useTwo, MaxClasses: 1000, MaxCandidates: 1 << 20})
		if err != nil {
			t.Fatal(variables, maxNodes, useTwo, err)
		}
		checkDauNetworkEnumeration(t, result)
	})
}
