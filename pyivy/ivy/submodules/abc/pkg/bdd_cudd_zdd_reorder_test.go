package abc

// Stand-alone tests for the ZDD ordering layer.

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func testZDDPermutations(n int) [][]int {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	result := make([][]int, 0)
	_ = cuddPermute(values, func(order []int) error {
		result = append(result, append([]int(nil), order...))
		return nil
	})
	return result
}

type testZDDKey struct {
	variable int
	high     string
	low      string
}

func testZDDModelCost(family []uint64, order []int) int {
	unique := make(map[testZDDKey]string)
	next := 0
	var build func([]uint64, int) string
	build = func(sets []uint64, level int) string {
		if len(sets) == 0 {
			return "0"
		}
		if level == len(order) {
			return "1"
		}
		bit := uint64(1) << order[level]
		low, high := make([]uint64, 0), make([]uint64, 0)
		for _, set := range sets {
			if set&bit == 0 {
				low = append(low, set)
			} else {
				high = append(high, set&^bit)
			}
		}
		h := build(high, level+1)
		l := build(low, level+1)
		if h == "0" {
			return l
		}
		key := testZDDKey{order[level], h, l}
		if id, found := unique[key]; found {
			return id
		}
		next++
		id := string(rune('A' + next))
		unique[key] = id
		return id
	}
	build(family, 0)
	return len(unique)
}

func testZDDUint64Equal(first, second []uint64) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func TestCuddZDDReorderExhaustiveFamilies(t *testing.T) {
	orders := testZDDPermutations(3)
	for mask := 0; mask < 1<<8; mask++ {
		family := make([]uint64, 0)
		for set := 0; set < 8; set++ {
			if mask&(1<<set) != 0 {
				family = append(family, uint64(set))
			}
		}
		problem := &CuddZDDReorderProblem{variables: 3, families: [][]uint64{family}}
		for _, order := range orders {
			cost, err := problem.Cost(order)
			if err != nil {
				t.Fatalf("mask %x order %v: %v", mask, order, err)
			}
			if want := testZDDModelCost(family, order); cost != want {
				t.Fatalf("mask %x order %v cost %d want %d", mask, order, cost, want)
			}
			dag, err := problem.Canonical(order)
			if err != nil {
				t.Fatal(err)
			}
			got, err := dag.Families()
			if err != nil || !testZDDUint64Equal(got[0], family) {
				t.Fatalf("round trip mask %x order %v got %v err %v", mask, order, got, err)
			}
		}
	}
}

func TestCuddZDDReorderManagerSnapshot(t *testing.T) {
	manager, err := NewCuddZDDManager(4, 1024)
	if err != nil {
		t.Fatal(err)
	}
	first, err := manager.FromBitsets([]uint64{0, 3, 5, 12}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.FromBitsets([]uint64{1, 2, 4, 8}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	problem, err := NewCuddZDDReorderProblem(manager, []CuddZDDEdge{first, second}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if problem.Variables() != 4 {
		t.Fatalf("variables %d", problem.Variables())
	}
	dag, err := problem.Canonical([]int{3, 2, 1, 0})
	if err != nil {
		t.Fatal(err)
	}
	families, err := dag.Families()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(families[0], []uint64{0, 3, 5, 12}) || !reflect.DeepEqual(families[1], []uint64{1, 2, 4, 8}) {
		t.Fatalf("families %v", families)
	}
	if _, err := NewCuddZDDReorderProblem(manager, []CuddZDDEdge{9999}, CuddZDDOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid root: %v", err)
	}
	if _, err := NewCuddZDDReorderProblem(manager, []CuddZDDEdge{first}, CuddZDDOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("operation limit: %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := problem.Cost([]int{0, 1, 2, 3}); err != nil {
		t.Fatalf("snapshot depends on manager: %v", err)
	}
}

func TestCuddZDDOrderedDAGRejectsCycles(t *testing.T) {
	dag := CuddZDDOrderedDAG{
		Order: []int{0}, Roots: []int{1},
		Nodes: []CuddZDDOrderedNode{{Variable: 0, High: 1, Low: 0}},
	}
	if _, err := dag.Families(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("cycle: %v", err)
	}
	dag.Nodes[0] = CuddZDDOrderedNode{Variable: 2, High: -1}
	if _, err := dag.Families(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("variable: %v", err)
	}
	dag = CuddZDDOrderedDAG{
		Order: []int{0, 1}, Roots: []int{2},
		Nodes: []CuddZDDOrderedNode{
			{Variable: 0, High: -1, Low: 0},
			{Variable: 1, High: 1, Low: 0},
		},
	}
	if _, err := dag.Families(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("out-of-order child: %v", err)
	}
	dag.Nodes = []CuddZDDOrderedNode{{Variable: 0, High: 0, Low: -1}}
	dag.Roots[0] = 1
	if _, err := dag.Families(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("unreduced node: %v", err)
	}
}

func TestCuddZDDTraversalRejectsMalformedManager(t *testing.T) {
	manager, err := NewCuddZDDManager(3, 64)
	if err != nil {
		t.Fatal(err)
	}
	root, err := manager.FromBitsets([]uint64{1, 3}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	node := manager.nodes[root]
	delete(manager.unique, cuddZDDKey{variable: node.variable, high: node.high, low: node.low})
	manager.nodes[root].low = CuddZDDEdge(len(manager.nodes) + 7)
	manager.unique[cuddZDDKey{variable: manager.nodes[root].variable, high: manager.nodes[root].high, low: manager.nodes[root].low}] = root
	if _, err := manager.Paths(root, CuddZDDOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("paths malformed edge: %v", err)
	}
	if _, err := NewCuddZDDReorderProblem(manager, []CuddZDDEdge{root}, CuddZDDOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("snapshot malformed edge: %v", err)
	}
}

func TestCuddZDDReorderExactAndBounds(t *testing.T) {
	problem := &CuddZDDReorderProblem{variables: 4, families: [][]uint64{{1, 3, 5, 7, 8}}}
	start := []int{3, 2, 1, 0}
	result, err := problem.Reorder(start, CuddReorderExact, CuddReorderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantCost := int(^uint(0) >> 1)
	var want []int
	for _, order := range testZDDPermutations(4) {
		cost, costErr := problem.Cost(order)
		if costErr != nil {
			t.Fatal(costErr)
		}
		if cost < wantCost || cost == wantCost && (want == nil || cuddOrderLess(order, want)) {
			wantCost, want = cost, append([]int(nil), order...)
		}
	}
	if result.Cost != wantCost || !reflect.DeepEqual(result.Order, want) || result.Passes != 1 {
		t.Fatalf("result %+v want %v/%d", result, want, wantCost)
	}
	if _, err := problem.Reorder(start, CuddReorderExact, CuddReorderOptions{MaxEvaluations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit %v", err)
	}
	stops := 0
	if _, err := problem.Reorder(start, CuddReorderSift, CuddReorderOptions{Stop: func() bool { stops++; return stops > 1 }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	if !reflect.DeepEqual(start, []int{3, 2, 1, 0}) {
		t.Fatalf("input mutated: %v", start)
	}
}

func TestCuddZDDInteractionsSymmetryTransform(t *testing.T) {
	problem := &CuddZDDReorderProblem{variables: 4, families: [][]uint64{{1, 2, 4, 3, 5, 6}, {8}}}
	sort.Slice(problem.families[0], func(i, j int) bool { return problem.families[0][i] < problem.families[0][j] })
	matrix, err := problem.InteractionMatrix()
	if err != nil {
		t.Fatal(err)
	}
	if !matrix.Interacts(0, 2) || matrix.Interacts(0, 3) {
		t.Fatalf("unexpected interaction")
	}
	symmetric, err := problem.Symmetric(0, 1)
	if err != nil || !symmetric {
		t.Fatalf("symmetry %v %v", symmetric, err)
	}
	groups, err := problem.SymmetryGroups([]int{0, 1, 2, 3})
	if err != nil || len(groups) != 1 || !reflect.DeepEqual(groups[0].Variables, []int{0, 1, 2}) {
		t.Fatalf("groups %v err %v", groups, err)
	}
	mapping, err := NewCuddLinearMap(4)
	if err != nil {
		t.Fatal(err)
	}
	if err := mapping.Combine(0, 1); err != nil {
		t.Fatal(err)
	}
	transformed, err := problem.Transform(mapping)
	if err != nil {
		t.Fatal(err)
	}
	if len(transformed.families[0]) != len(problem.families[0]) {
		t.Fatalf("transform lost entries")
	}
	for i, original := range problem.families[0] {
		want := original
		if (original>>1)&1 != 0 {
			want ^= 1
		}
		if transformed.families[0][i] != want && !containsUint64(transformed.families[0], want) {
			t.Fatalf("transform missing %d from %d: %v", want, original, transformed.families[0])
		}
	}
	identity, _ := NewCuddLinearMap(4)
	copyProblem, err := problem.Transform(identity)
	if err != nil || !reflect.DeepEqual(copyProblem.families, problem.families) {
		t.Fatalf("identity %v %v", copyProblem, err)
	}
}

func containsUint64(values []uint64, target uint64) bool {
	index := sort.Search(len(values), func(i int) bool { return values[i] >= target })
	return index < len(values) && values[index] == target
}

func TestCuddZDDReorderGroups(t *testing.T) {
	problem := &CuddZDDReorderProblem{variables: 4, families: [][]uint64{{1, 3, 5, 8, 12}}}
	groups := []CuddReorderGroup{{Variables: []int{1, 2}, Hard: true}}
	result, err := problem.ReorderGroups([]int{0, 1, 2, 3}, groups, CuddReorderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	first, second := -1, -1
	for i, variable := range result.Order {
		if variable == 1 {
			first = i
		}
		if variable == 2 {
			second = i
		}
	}
	if second != first+1 {
		t.Fatalf("group split: %v", result.Order)
	}
	if _, err := problem.ReorderGroups([]int{0, 1, 2, 3}, []CuddReorderGroup{{Variables: []int{0, 2}}}, CuddReorderOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("noncontiguous group: %v", err)
	}
}

func TestCuddZDDPathsAndText(t *testing.T) {
	manager, _ := NewCuddZDDManager(4, 256)
	root, err := manager.FromBitsets([]uint64{0, 1, 6}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	paths, err := manager.Paths(root, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sets := make([]uint64, 0, len(paths))
	for _, path := range paths {
		var set uint64
		for variable, value := range path {
			if value == 1 {
				set |= uint64(1) << variable
			}
		}
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i] < sets[j] })
	if !reflect.DeepEqual(sets, []uint64{0, 1, 6}) {
		t.Fatalf("sets %v paths %v", sets, paths)
	}
	text, err := manager.DebugText(root, CuddZDDOptions{})
	if err != nil || text == "" {
		t.Fatalf("text %q %v", text, err)
	}
	for _, tc := range []struct {
		path []uint8
		want string
	}{
		{[]uint8{1, 2, 2, 1}, "10"},
		{[]uint8{0, 0, 2, 2}, "--"},
		{[]uint8{1, 1}, "?"},
		{[]uint8{3, 0}, "?"},
		{[]uint8{0, 9}, "0"},
		{[]uint8{64, 0}, "?"},
	} {
		got, stringErr := CuddZDDCoverPathString(tc.path)
		if stringErr != nil || got != tc.want {
			t.Fatalf("path %v got %q err %v", tc.path, got, stringErr)
		}
	}
	if _, err := CuddZDDCoverPathString([]uint8{1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("odd path: %v", err)
	}
	if _, err := manager.Paths(root, CuddZDDOptions{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("path bytes: %v", err)
	}
}

func TestCuddZDDReorderModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5dd0))
	for trial := 0; trial < 5000; trial++ {
		variables := 1 + rng.Intn(6)
		count := rng.Intn(1 << min(variables, 4))
		setMap := make(map[uint64]struct{})
		for len(setMap) < count {
			setMap[uint64(rng.Intn(1<<variables))] = struct{}{}
		}
		family := make([]uint64, 0, len(setMap))
		for set := range setMap {
			family = append(family, set)
		}
		sort.Slice(family, func(i, j int) bool { return family[i] < family[j] })
		order := rng.Perm(variables)
		problem := &CuddZDDReorderProblem{variables: variables, families: [][]uint64{family}}
		got, err := problem.Cost(order)
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if want := testZDDModelCost(family, order); got != want {
			t.Fatalf("trial %d got %d want %d", trial, got, want)
		}
	}
}

func TestCuddZDDReorderCoverageInventory(t *testing.T) {
	coverage := CuddZDDReorderingCoverage()
	if len(coverage.Complete) != 1 || len(coverage.Partial) != 7 || len(coverage.NativeOnly) != 3 {
		t.Fatalf("coverage %+v", coverage)
	}
}

func FuzzCuddZDDReorderCanonical(f *testing.F) {
	f.Add(uint8(3), uint64(0x96), uint64(5))
	f.Add(uint8(4), uint64(0xf00f), uint64(19))
	f.Fuzz(func(t *testing.T, rawVariables uint8, mask, seed uint64) {
		variables := 1 + int(rawVariables%6)
		family := make([]uint64, 0)
		limit := 1 << variables
		for set := 0; set < limit && set < 64; set++ {
			if mask&(uint64(1)<<set) != 0 {
				family = append(family, uint64(set))
			}
		}
		order := rand.New(rand.NewSource(int64(seed))).Perm(variables)
		problem := &CuddZDDReorderProblem{variables: variables, families: [][]uint64{family}}
		cost, err := problem.Cost(order)
		if err != nil {
			t.Fatal(err)
		}
		if cost != testZDDModelCost(family, order) {
			t.Fatalf("cost mismatch")
		}
		dag, err := problem.Canonical(order)
		if err != nil {
			t.Fatal(err)
		}
		got, err := dag.Families()
		if err != nil || !testZDDUint64Equal(got[0], family) {
			t.Fatalf("round trip got %v err %v", got, err)
		}
	})
}
