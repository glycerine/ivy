package abc

// Go-only tests for src/misc/extra/extraUtilPerm.c behavior.

import (
	"errors"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"testing"
)

func extraPermutationFamily(t testing.TB, manager *ExtraPermutationZDD, root int) [][]int {
	t.Helper()
	subsets, err := manager.Subsets(root, 1<<16, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	return subsets
}

func extraPermutationBuildFamily(t testing.TB, manager *ExtraPermutationZDD, family [][]int) int {
	t.Helper()
	root := 0
	for _, subset := range family {
		item, err := manager.BuildSet(subset)
		if err != nil {
			t.Fatal(err)
		}
		root, err = manager.Union(root, item)
		if err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func extraFamilyKeys(family [][]int) []string {
	seen := make(map[string]struct{}, len(family))
	result := make([]string, 0, len(family))
	for _, subset := range family {
		key := extraZDDSubsetKey(subset)
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func extraEqualFamily(first, second [][]int) bool {
	return slices.Equal(extraFamilyKeys(first), extraFamilyKeys(second))
}

func TestExtraPermutationZDDConstructionAndCanonicalNodes(t *testing.T) {
	m, err := NewExtraPermutationZDD(5, ExtraPermutationZDDOptions{MaxNodes: 100, CacheEntries: 13})
	if err != nil {
		t.Fatal(err)
	}
	if m.NodeCount() != 7 || m.Stats().CacheEntries != 16 {
		t.Fatalf("initial stats = %+v", m.Stats())
	}
	for variable := 0; variable < 5; variable++ {
		id, _ := m.Variable(variable)
		node, _ := m.Node(id)
		if id != variable+2 || node != (ExtraZDDNode{variable, 1, 0}) {
			t.Fatalf("variable %d = %d %+v", variable, id, node)
		}
	}
	input := []int{4, 1, 3}
	root, err := m.BuildSet(input)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(input, []int{4, 1, 3}) {
		t.Fatal("input mutated")
	}
	again, _ := m.BuildSet([]int{1, 3, 4})
	if root != again {
		t.Fatalf("noncanonical %d %d", root, again)
	}
	if got := extraPermutationFamily(t, m, root); !reflect.DeepEqual(got, [][]int{{1, 3, 4}}) {
		t.Fatalf("family = %v", got)
	}
	if _, err = m.BuildSet([]int{1, 1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate = %v", err)
	}
}

func TestExtraPermutationZDDSetOperations(t *testing.T) {
	m, _ := NewExtraPermutationZDD(6, ExtraPermutationZDDOptions{MaxNodes: 4096})
	a := extraPermutationBuildFamily(t, m, [][]int{{}, {0}, {1, 2}, {0, 3}})
	b := extraPermutationBuildFamily(t, m, [][]int{{}, {0}, {2}, {0, 4}})
	union, err := m.Union(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got := extraPermutationFamily(t, m, union); !extraEqualFamily(got, [][]int{{}, {0}, {1, 2}, {0, 3}, {2}, {0, 4}}) {
		t.Fatalf("union = %v", got)
	}
	difference, err := m.Difference(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got := extraPermutationFamily(t, m, difference); !extraEqualFamily(got, [][]int{{1, 2}, {0, 3}}) {
		t.Fatalf("difference = %v", got)
	}
	// Preserve the source recurrence literally: unlike conventional ZDD
	// intersection, a smaller unmatched top variable retains its true branch.
	zero := extraPermutationBuildFamily(t, m, [][]int{{0}})
	one := extraPermutationBuildFamily(t, m, [][]int{{1}})
	intersection, err := m.Intersection(zero, one)
	if err != nil {
		t.Fatal(err)
	}
	if got := extraPermutationFamily(t, m, intersection); !extraEqualFamily(got, [][]int{{0}}) {
		t.Fatalf("source intersection = %v", got)
	}
	cozero, _ := m.Cofactor0(a, 0)
	if got := extraPermutationFamily(t, m, cozero); !extraEqualFamily(got, [][]int{{}, {1, 2}}) {
		t.Fatalf("co0 = %v", got)
	}
	coone, _ := m.Cofactor1(a, 0)
	if got := extraPermutationFamily(t, m, coone); !extraEqualFamily(got, [][]int{{0}, {0, 3}}) {
		t.Fatalf("co1 = %v", got)
	}
	threshold, _ := m.Threshold(a, 2)
	if got := extraPermutationFamily(t, m, threshold); !extraEqualFamily(got, [][]int{{}, {0}, {1, 2}, {0, 3}}) {
		t.Fatalf("threshold = %v", got)
	}
	paths, _ := m.CountPaths(union)
	if paths.Cmp(big.NewInt(6)) != 0 {
		t.Fatalf("paths = %s", paths)
	}
	nodes, _ := m.CountNodes(union)
	shared, _ := m.CountSharedNodes([]int{a, b})
	if nodes < 1 || shared < nodes {
		t.Fatalf("nodes %d shared %d", nodes, shared)
	}
}

func TestExtraPermutationZDDDotAndMinimalUnion(t *testing.T) {
	m, _ := NewExtraPermutationZDD(7, ExtraPermutationZDDOptions{MaxNodes: 1 << 15})
	a := extraPermutationBuildFamily(t, m, [][]int{{0}, {1, 2}})
	b := extraPermutationBuildFamily(t, m, [][]int{{2}, {3}})
	product, err := m.DotProduct(a, b)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{0, 2}, {0, 3}, {1, 2}, {1, 2, 3}}
	if got := extraPermutationFamily(t, m, product); !extraEqualFamily(got, want) {
		t.Fatalf("dot = %v", got)
	}
	x := extraPermutationBuildFamily(t, m, [][]int{{0}, {0, 1}, {2, 3}, {2, 3, 4}})
	minimal, err := m.MinimalUnion(0, x)
	if err != nil {
		t.Fatal(err)
	}
	// The source assumes each input is already minimal; exercise the actual
	// cross-family reduction with two minimal operands.
	left := extraPermutationBuildFamily(t, m, [][]int{{0}, {2, 3}})
	right := extraPermutationBuildFamily(t, m, [][]int{{0, 1}, {2, 3, 4}, {5}})
	minimal, err = m.MinimalUnion(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if got := extraPermutationFamily(t, m, minimal); !extraEqualFamily(got, [][]int{{0}, {2, 3}, {5}}) {
		t.Fatalf("minimal = %v", got)
	}
	limited, err := m.DotMinimalProduct6(extraPermutationBuildFamily(t, m, [][]int{{0, 1, 2}}), extraPermutationBuildFamily(t, m, [][]int{{3, 4, 5}, {6}}))
	if err != nil {
		t.Fatal(err)
	}
	if got := extraPermutationFamily(t, m, limited); !extraEqualFamily(got, [][]int{{0, 1, 2, 6}, {0, 1, 2, 3, 4, 5}}) {
		t.Fatalf("dot6 = %v", got)
	}
}

func TestExtraPermutationConversionFormattingAndProduct(t *testing.T) {
	fixture := []int{6, 5, 7, 0, 3, 2, 1, 8, 9, 4}
	transpositions, err := ExtraPermutationToTranspositions(fixture)
	if err != nil {
		t.Fatal(err)
	}
	want := []ExtraTransposition{{0, 3}, {1, 6}, {2, 5}, {3, 4}, {4, 9}, {5, 6}, {6, 9}, {7, 9}, {8, 9}}
	if !slices.Equal(transpositions, want) {
		t.Fatalf("transpositions = %v", transpositions)
	}
	round, err := ExtraTranspositionsToPermutation(transpositions, len(fixture))
	if err != nil || !slices.Equal(round, fixture) {
		t.Fatalf("round = %v %v", round, err)
	}
	if FormatExtraPermutation([]int{1, 0}) != "{  1  0 }\n" || FormatExtraTranspositions(nil) != "Empty set\n" || FormatExtraTranspositions([]ExtraTransposition{{1, 3}}) != "(1 3)\n" {
		t.Fatal("format")
	}
	m, _ := NewExtraPermutationZDD(10, ExtraPermutationZDDOptions{MaxNodes: 1 << 16})
	if err = m.ConfigurePermutations(5); err != nil {
		t.Fatal(err)
	}
	p := []int{1, 0, 2, 4, 3}
	q := []int{0, 3, 2, 1, 4}
	pr := extraPermutationBuildFamily(t, m, [][]int{})
	_ = pr
	proot, _ := m.BuildPermutation(p)
	qroot, _ := m.BuildPermutation(q)
	both, err := m.PermutationProduct(proot, qroot)
	if err != nil {
		t.Fatal(err)
	}
	perms, err := m.Permutations(both, 10)
	if err != nil || len(perms) != 1 {
		t.Fatalf("product = %v %v", perms, err)
	}
	// Determine composition independently by applying both index maps.
	wantComposition := make([]int, 5)
	for i := range wantComposition {
		wantComposition[i] = p[q[i]]
	}
	if !slices.Equal(perms[0], wantComposition) {
		t.Fatalf("composition = %v want %v", perms[0], wantComposition)
	}
	variable, _ := m.TranspositionVariable(3, 4)
	moved, err := m.Permute(qroot, variable)
	if err != nil {
		t.Fatal(err)
	}
	movedPerms, _ := m.Permutations(moved, 10)
	swap := []int{0, 1, 2, 4, 3}
	wantMoved := make([]int, 5)
	for i := range wantMoved {
		wantMoved[i] = q[swap[i]]
	}
	if len(movedPerms) != 1 || !slices.Equal(movedPerms[0], wantMoved) {
		t.Fatalf("permute = %v want %v", movedPerms, wantMoved)
	}
}

func TestExtraPermutationZDDGiaAndCubeCheckpoints(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	g.HashStart()
	n, _ := g.HashAnd(a, b)
	g.AppendCO(n)
	analysis, err := AnalyzeExtraZDDGia(g, ExtraPermutationZDDOptions{MaxNodes: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if analysis.PathCount.Cmp(big.NewInt(2)) != 0 || analysis.SharedNodes == 0 || len(analysis.Roots) != 1 {
		t.Fatalf("analysis = %+v", analysis)
	}
	cube, err := EnumerateExtraCubeStatesZDD(ExtraCubeZDDOptions{MaxDepth: 3, ZDD: ExtraPermutationZDDOptions{MaxNodes: 1 << 20, MaxOperations: 1 << 26}})
	if !errors.Is(err, ErrLimit) {
		t.Fatalf("cube error = %v", err)
	}
	if !slices.Equal(cube.LayerTotals, []int{1, 10, 64, 385}) {
		t.Fatalf("cube layers = %v", cube.LayerTotals)
	}
	if !slices.Equal(cube.NodeTotals, []int{0, 63, 443, 2018}) {
		t.Fatalf("cube nodes = %v", cube.NodeTotals)
	}
}

func TestExtraPermutationZDDLimitsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewExtraPermutationZDD(3, ExtraPermutationZDDOptions{MaxNodes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit = %v", err)
	}
	if _, err := NewExtraPermutationZDD(3, ExtraPermutationZDDOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %v", err)
	}
	m, _ := NewExtraPermutationZDD(4, ExtraPermutationZDDOptions{MaxNodes: 8, MaxOperations: 100})
	before := m.clone()
	_, err := m.BuildSet([]int{0, 1, 2, 3})
	if !errors.Is(err, ErrLimit) {
		t.Fatalf("build limit = %v", err)
	}
	if !reflect.DeepEqual(m, before) {
		t.Fatal("failed transaction mutated manager")
	}
	if err = m.ConfigurePermutations(4); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad permutation shape = %v", err)
	}
	var nilManager *ExtraPermutationZDD
	if _, err = nilManager.Union(0, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil = %v", err)
	}
}

func FuzzExtraPermutationZDDSetAlgebra(f *testing.F) {
	for _, seed := range []uint64{0, 1, 0x123456789abcdef, 0xffffffffffffffff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		variables := 1 + int(seed%6)
		seed = seed*6364136223846793005 + 1
		makeFamily := func() [][]int {
			count := 1 + int(seed%8)
			seed = seed*6364136223846793005 + 1
			family := make([][]int, 0, count)
			seen := map[string]struct{}{}
			for i := 0; i < count; i++ {
				subset := []int{}
				for v := 0; v < variables; v++ {
					if seed>>uint((i*7+v)%64)&1 != 0 {
						subset = append(subset, v)
					}
				}
				key := extraZDDSubsetKey(subset)
				if _, ok := seen[key]; !ok {
					seen[key] = struct{}{}
					family = append(family, subset)
				}
			}
			return family
		}
		aFamily, bFamily := makeFamily(), makeFamily()
		m, err := NewExtraPermutationZDD(variables, ExtraPermutationZDDOptions{MaxNodes: 1 << 15, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		a := extraPermutationBuildFamily(t, m, aFamily)
		b := extraPermutationBuildFamily(t, m, bFamily)
		union, err := m.Union(a, b)
		if err != nil {
			t.Fatal(err)
		}
		wantUnion := append(append([][]int{}, aFamily...), bFamily...)
		if got := extraPermutationFamily(t, m, union); !extraEqualFamily(got, wantUnion) {
			t.Fatalf("union %v want %v", got, wantUnion)
		}
		difference, err := m.Difference(a, b)
		if err != nil {
			t.Fatal(err)
		}
		bKeys := map[string]struct{}{}
		for _, s := range bFamily {
			bKeys[extraZDDSubsetKey(s)] = struct{}{}
		}
		wantDiff := [][]int{}
		for _, s := range aFamily {
			if _, ok := bKeys[extraZDDSubsetKey(s)]; !ok {
				wantDiff = append(wantDiff, s)
			}
		}
		if got := extraPermutationFamily(t, m, difference); !extraEqualFamily(got, wantDiff) {
			t.Fatalf("difference %v want %v", got, wantDiff)
		}
	})
}
