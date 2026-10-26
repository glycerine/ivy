package abc

// Standalone tests for the Go REO adaptation.

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func reoTestOrders(variables int, visit func([]int)) {
	order := make([]int, variables)
	for i := range order {
		order[i] = i
	}
	var generate func(int)
	generate = func(position int) {
		if position == variables {
			visit(append([]int(nil), order...))
			return
		}
		for i := position; i < variables; i++ {
			order[position], order[i] = order[i], order[position]
			generate(position + 1)
			order[position], order[i] = order[i], order[position]
		}
	}
	generate(0)
}

func reoModelNodeCount(table []bool, order []int) int {
	type modelNode struct{ variable, low, high int }
	terminals := map[bool]int{false: -1, true: -2}
	unique := make(map[modelNode]int)
	next := 1
	var build func(int, int) int
	build = func(depth, assignment int) int {
		if depth == len(order) {
			return terminals[table[assignment]]
		}
		variable := order[depth]
		low := build(depth+1, assignment)
		high := build(depth+1, assignment|1<<variable)
		if low == high {
			return low
		}
		key := modelNode{variable, low, high}
		if found, ok := unique[key]; ok {
			return found
		}
		unique[key] = next
		next++
		return next - 1
	}
	build(0, 0)
	terminalsUsed := 1
	for _, value := range table[1:] {
		if value != table[0] {
			terminalsUsed = 2
			break
		}
	}
	return len(unique) + terminalsUsed
}

func reoBoolTable(bits, variables int) []bool {
	table := make([]bool, 1<<variables)
	for i := range table {
		table[i] = bits&(1<<i) != 0
	}
	return table
}

func TestREOExhaustiveCanonicalNodeModel(t *testing.T) {
	for bits := 0; bits < 256; bits++ {
		table := reoBoolTable(bits, 3)
		problem, err := NewCuddTruthReorderProblem(3, [][]bool{table})
		if err != nil {
			t.Fatal(err)
		}
		reoTestOrders(3, func(order []int) {
			dag, profile, shuffleErr := mustREOManager(t, 3).Shuffle(problem, order, REOOptions{})
			if shuffleErr != nil {
				t.Fatalf("bits=%d order=%v: %v", bits, order, shuffleErr)
			}
			want := reoModelNodeCount(table, order)
			if profile.Nodes != want {
				t.Fatalf("bits=%d order=%v got=%d want=%d", bits, order, profile.Nodes, want)
			}
			for assignment, expected := range table {
				value, valueErr := dag.Value(0, uint64(assignment))
				if valueErr != nil || (value != 0) != expected {
					t.Fatalf("semantic mismatch bits=%d order=%v assignment=%d", bits, order, assignment)
				}
			}
		})
	}
}

func mustREOManager(t testing.TB, variables int) *REOManager {
	t.Helper()
	manager, err := NewREOManager(variables, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func reoExplicitAPL(dag CuddOrderedDAG) float64 {
	total := 0
	for root := range dag.Roots {
		for assignment := 0; assignment < 1<<len(dag.Order); assignment++ {
			edge := dag.Roots[root]
			for edge > 0 {
				total++
				node := dag.Nodes[edge-1]
				if assignment&(1<<node.Variable) == 0 {
					edge = node.Low
				} else {
					edge = node.High
				}
			}
		}
	}
	return float64(total) / float64(uint64(1)<<uint(len(dag.Order)))
}

func TestREOMultiOutputProfilesAndModes(t *testing.T) {
	variables := 4
	tables := [][]bool{
		func() []bool {
			result := make([]bool, 16)
			for x := range result {
				result[x] = x&1 != 0 && x&8 != 0 || x&2 != 0
			}
			return result
		}(),
		func() []bool {
			result := make([]bool, 16)
			for x := range result {
				result[x] = (x&1 != 0) != (x&4 != 0)
			}
			return result
		}(),
	}
	problem, err := NewCuddTruthReorderProblem(variables, tables)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []REOMinimization{REOMinimizeNodes, REOMinimizeWidth, REOMinimizeAPL} {
		manager := mustREOManager(t, variables)
		if err = manager.SetMinimization(mode); err != nil {
			t.Fatal(err)
		}
		if err = manager.SetIterations(2); err != nil {
			t.Fatal(err)
		}
		if err = manager.SetVerification(true); err != nil {
			t.Fatal(err)
		}
		result, reorderErr := manager.Reorder(problem, []int{3, 2, 1, 0}, REOOptions{})
		if reorderErr != nil {
			t.Fatalf("mode %d: %v", mode, reorderErr)
		}
		if reoMetric(result.Final, mode) > reoMetric(result.Initial, mode)+1e-12 || result.Evaluations == 0 || result.Passes == 0 {
			t.Fatalf("bad improvement: %+v", result)
		}
		if math.Abs(result.Final.APL-reoExplicitAPL(result.DAG)) > 1e-12 {
			t.Fatalf("APL mismatch got %g want %g", result.Final.APL, reoExplicitAPL(result.DAG))
		}
		totalNodes := 0
		for _, count := range result.Final.NodesByLevel {
			totalNodes += count
		}
		if result.Final.Nodes != totalNodes {
			t.Fatalf("node profile mismatch nodes=%d levels=%v", result.Final.Nodes, result.Final.NodesByLevel)
		}
		if len(result.Permutation) != variables || !sort.IntsAreSorted(sortedCopy(result.Permutation)) {
			t.Fatalf("bad permutation %v", result.Permutation)
		}
	}
}

func sortedCopy(values []int) []int {
	result := append([]int(nil), values...)
	sort.Ints(result)
	return result
}

func TestREOShuffleSwapUnitsOwnershipAndText(t *testing.T) {
	problem, err := NewCuddTruthReorderProblem(3, [][]bool{reoBoolTable(0x96, 3), reoBoolTable(0xe8, 3)})
	if err != nil {
		t.Fatal(err)
	}
	manager := mustREOManager(t, 3)
	dag, profile, err := manager.Shuffle(problem, []int{2, 0, 1}, REOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	order, swapped, err := manager.SwapAdjacent(problem, []int{2, 0, 1}, 1, REOOptions{})
	if err != nil || !reflect.DeepEqual(order, []int{2, 1, 0}) || swapped.Nodes < 1 {
		t.Fatalf("swap %v %+v %v", order, swapped, err)
	}
	units, err := REOTransferToUnits(dag)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := units.TransferToDAG()
	if err != nil || !reflect.DeepEqual(restored, dag) {
		t.Fatalf("unit roundtrip: %v", err)
	}
	units.Order[0] = 99
	if dag.Order[0] == 99 {
		t.Fatal("unit result aliases input")
	}
	if text := profile.Text(); !strings.Contains(text, "level 0 nodes") || strings.Count(text, "\n") != 4 {
		t.Fatalf("unexpected profile text %q", text)
	}
	bad := restored
	bad.Nodes[len(bad.Nodes)-1].Low = len(bad.Nodes)
	if _, err = REOTransferToUnits(bad); !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed DAG accepted: %v", err)
	}
	bad = restored
	bad.Terminals = append(bad.Terminals, bad.Terminals[0])
	if _, err = REOTransferToUnits(bad); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate/unreachable terminal accepted: %v", err)
	}
	oversized := REOUnitDAG{Order: make([]int, 21), Roots: []int{-1}, Terminals: []uint64{0}}
	for i := range oversized.Order {
		oversized.Order[i] = i
	}
	if _, err = oversized.TransferToDAG(); !errors.Is(err, ErrLimit) {
		t.Fatalf("oversized reverse transfer accepted: %v", err)
	}
}

func TestREOLimitsCancellationLifecycleAndAtomicResults(t *testing.T) {
	problem, err := NewCuddTruthReorderProblem(4, [][]bool{reoBoolTable(0x6996, 4)})
	if err != nil {
		t.Fatal(err)
	}
	manager := mustREOManager(t, 4)
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{MaxEvaluations: 1}); result.Order != nil || !errors.Is(reorderErr, ErrLimit) {
		t.Fatalf("evaluation limit result=%+v err=%v", result, reorderErr)
	}
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{MaxOperations: 1}); result.Order != nil || !errors.Is(reorderErr, ErrLimit) {
		t.Fatalf("operation limit result=%+v err=%v", result, reorderErr)
	}
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{MaxBytes: 8}); result.Order != nil || !errors.Is(reorderErr, ErrLimit) {
		t.Fatalf("byte limit result=%+v err=%v", result, reorderErr)
	}
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{MaxOperations: 60}); result.Order != nil || !errors.Is(reorderErr, ErrLimit) {
		t.Fatalf("profile operation limit result=%+v err=%v", result, reorderErr)
	}
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{MaxBytes: 512}); result.Order != nil || !errors.Is(reorderErr, ErrLimit) {
		t.Fatalf("profile memory limit result=%+v err=%v", result, reorderErr)
	}
	if result, reorderErr := manager.Reorder(problem, nil, REOOptions{Stop: func() bool { return true }}); result.Order != nil || !errors.Is(reorderErr, ErrCancelled) {
		t.Fatalf("cancel result=%+v err=%v", result, reorderErr)
	}
	if err = manager.SetRemapping(true); err != nil {
		t.Fatal(err)
	}
	if _, reorderErr := manager.Reorder(problem, nil, REOOptions{}); !errors.Is(reorderErr, ErrUnsupported) {
		t.Fatalf("remapping should be explicit unsupported behavior: %v", reorderErr)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = manager.Shuffle(problem, []int{0, 1, 2, 3}, REOOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager accepted: %v", err)
	}
}

func TestREONodeObjectiveIncludesReachableTerminals(t *testing.T) {
	constant, err := NewCuddTruthReorderProblem(2, [][]bool{{true, true, true, true}})
	if err != nil {
		t.Fatal(err)
	}
	_, profile, err := mustREOManager(t, 2).Shuffle(constant, []int{0, 1}, REOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if profile.Nodes != 1 || !reflect.DeepEqual(profile.NodesByLevel, []int{1}) {
		t.Fatalf("constant unit missing from node objective: %+v", profile)
	}

	nonconstant, err := NewCuddTruthReorderProblem(1, [][]bool{{false, true}})
	if err != nil {
		t.Fatal(err)
	}
	_, profile, err = mustREOManager(t, 1).Shuffle(nonconstant, []int{0}, REOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if profile.Nodes != 3 || !reflect.DeepEqual(profile.NodesByLevel, []int{1, 2}) {
		t.Fatalf("terminal units missing from node objective: %+v", profile)
	}
	limited, err := NewREOManager(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = limited.Shuffle(nonconstant, []int{0}, REOOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node ceiling ignored terminal units: %v", err)
	}
}

func TestREOPreservesUnsupportedVariableSlots(t *testing.T) {
	table := make([]bool, 8)
	for assignment := range table {
		table[assignment] = assignment&1 != 0 && assignment&4 != 0
	}
	problem, err := NewCuddTruthReorderProblem(3, [][]bool{table})
	if err != nil {
		t.Fatal(err)
	}
	result, err := mustREOManager(t, 3).Reorder(problem, []int{0, 1, 2}, REOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Order[1] != 1 {
		t.Fatalf("unsupported variable moved from its manager-order slot: %v", result.Order)
	}
	compactTable := make([]bool, 4)
	for assignment := range compactTable {
		compactTable[assignment] = assignment&1 != 0 && assignment&2 != 0
	}
	compact, err := NewCuddTruthReorderProblem(2, [][]bool{compactTable})
	if err != nil {
		t.Fatal(err)
	}
	_, compactProfile, err := mustREOManager(t, 2).Shuffle(compact, []int{0, 1}, REOOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Final.Nodes != compactProfile.Nodes || result.Final.Width != compactProfile.Width || result.Final.APL != compactProfile.APL || len(result.Final.WidthByLevel) != 3 {
		t.Fatalf("unsupported slot changed compact REO profile: full=%+v compact=%+v", result.Final, compactProfile)
	}
}

func TestREODeterministicModelCampaign5000(t *testing.T) {
	state := uint64(0x6a09e667f3bcc909)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	manager := mustREOManager(t, 3)
	for trial := 0; trial < 5000; trial++ {
		bits := int(next() & 255)
		table := reoBoolTable(bits, 3)
		problem, err := NewCuddTruthReorderProblem(3, [][]bool{table})
		if err != nil {
			t.Fatal(err)
		}
		order := []int{0, 1, 2}
		for i := 2; i > 0; i-- {
			j := int(next() % uint64(i+1))
			order[i], order[j] = order[j], order[i]
		}
		_, profile, err := manager.Shuffle(problem, order, REOOptions{})
		if err != nil || profile.Nodes != reoModelNodeCount(table, order) {
			t.Fatalf("trial %d bits=%x order=%v profile=%+v err=%v", trial, bits, order, profile, err)
		}
	}
}

func TestREOCoverageAndBuildInventory(t *testing.T) {
	coverage := REOSourceCoverage()
	want := []string{"reoApi.c", "reoCore.c", "reoProfile.c", "reoShuffle.c", "reoSift.c", "reoSwap.c", "reoTransfer.c", "reoUnits.c"}
	if !reflect.DeepEqual(coverage.Build, want) || len(coverage.Complete) == 0 || len(coverage.Partial) == 0 || len(coverage.NativeOnly) == 0 {
		t.Fatalf("coverage mismatch %+v", coverage)
	}
	inventory := REODefinitionInventory()
	if len(inventory) != 39 {
		t.Fatalf("definition inventory has %d entries", len(inventory))
	}
	seen := make(map[string]struct{}, len(inventory))
	for _, name := range inventory {
		if _, found := seen[name]; found {
			t.Fatalf("duplicate definition %q", name)
		}
		seen[name] = struct{}{}
	}
}

func FuzzREOCanonicalAndProfiles(f *testing.F) {
	for _, seed := range []uint64{0, 1, 0x96, 0xe8, 0xff, 0x123456789abcdef0} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		variables := int(seed%4) + 1
		size := 1 << variables
		table := make([]bool, size)
		state := seed | 1
		for i := range table {
			state = state*6364136223846793005 + 1442695040888963407
			table[i] = state>>63 != 0
		}
		problem, err := NewCuddTruthReorderProblem(variables, [][]bool{table})
		if err != nil {
			t.Fatal(err)
		}
		order := make([]int, variables)
		for i := range order {
			order[i] = i
		}
		for i := variables - 1; i > 0; i-- {
			j := int((seed >> uint(i*5)) % uint64(i+1))
			order[i], order[j] = order[j], order[i]
		}
		dag, profile, err := mustREOManager(t, variables).Shuffle(problem, order, REOOptions{})
		if err != nil || profile.Nodes != reoModelNodeCount(table, order) || math.Abs(profile.APL-reoExplicitAPL(dag)) > 1e-12 {
			t.Fatalf("seed=%x order=%v profile=%+v err=%v", seed, order, profile, err)
		}
	})
}
