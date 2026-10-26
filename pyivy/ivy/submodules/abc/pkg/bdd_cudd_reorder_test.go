package abc

// Tests for the standalone CUDD reordering models ported from
// src/bdd/cudd/cuddReorder.c and the related reordering modules.

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"
)

func cuddTestTruth(variables int, bits uint64) []bool {
	table := make([]bool, 1<<variables)
	for i := range table {
		table[i] = bits&(uint64(1)<<i) != 0
	}
	return table
}

// cuddIndependentCost uses symbolic Shannon strings and deliberately shares
// no implementation or node representation with CuddReorderProblem.Cost.
func cuddIndependentCost(table []bool, variables int, order []int) int {
	unique := make(map[string]string)
	var build func(int, int) string
	build = func(depth, assignment int) string {
		if depth == variables {
			if table[assignment] {
				return "1"
			}
			return "0"
		}
		variable := order[depth]
		low := build(depth+1, assignment)
		high := build(depth+1, assignment|1<<variable)
		if low == high {
			return low
		}
		key := fmt.Sprintf("%d(%s,%s)", variable, low, high)
		if value, found := unique[key]; found {
			return value
		}
		unique[key] = key
		return key
	}
	build(0, 0)
	return len(unique)
}

func cuddTestPermutations(n int, visit func([]int)) {
	values := make([]int, n)
	for i := range values {
		values[i] = i
	}
	var generate func(int)
	generate = func(index int) {
		if index == n {
			visit(append([]int(nil), values...))
			return
		}
		for i := index; i < n; i++ {
			values[index], values[i] = values[i], values[index]
			generate(index + 1)
			values[index], values[i] = values[i], values[index]
		}
	}
	generate(0)
}

func TestCuddReorderExhaustiveFourVariableCanonicalCosts(t *testing.T) {
	for variables := 0; variables <= 4; variables++ {
		functions := uint64(1) << (1 << variables)
		for truth := uint64(0); truth < functions; truth++ {
			table := cuddTestTruth(variables, truth)
			problem, err := NewCuddTruthReorderProblem(variables, [][]bool{table})
			if err != nil {
				t.Fatal(variables, truth, err)
			}
			cuddTestPermutations(variables, func(order []int) {
				got, costErr := problem.Cost(order)
				want := cuddIndependentCost(table, variables, order)
				if costErr != nil || got != want {
					t.Fatalf("n=%d truth=%x order=%v got=%d want=%d err=%v", variables, truth, order, got, want, costErr)
				}
			})
		}
	}
}

func TestCuddReorderExactWindowsAndSifting(t *testing.T) {
	table := make([]bool, 16)
	for assignment := range table {
		x0 := assignment&1 != 0
		x1 := assignment&2 != 0
		x2 := assignment&4 != 0
		x3 := assignment&8 != 0
		table[assignment] = (x0 && x3) != (x1 && x2)
	}
	problem, err := NewCuddTruthReorderProblem(4, [][]bool{table})
	if err != nil {
		t.Fatal(err)
	}
	start := []int{0, 1, 2, 3}
	bestCost := math.MaxInt
	var best []int
	cuddTestPermutations(4, func(order []int) {
		cost, costErr := problem.Cost(order)
		if costErr != nil {
			t.Fatal(costErr)
		}
		if cost < bestCost || cost == bestCost && cuddOrderLess(order, best) {
			bestCost, best = cost, append([]int(nil), order...)
		}
	})
	exact, err := problem.Reorder(start, CuddReorderExact, CuddReorderOptions{})
	if err != nil || exact.Cost != bestCost || !slices.Equal(exact.Order, best) {
		t.Fatal(exact, bestCost, best, err)
	}
	for _, method := range []CuddReorderMethod{CuddReorderSift, CuddReorderWindow2, CuddReorderWindow3, CuddReorderWindow4, CuddReorderSymmetry} {
		result, reorderErr := problem.Reorder(start, method, CuddReorderOptions{})
		if reorderErr != nil || result.Cost > result.InitialCost || !cuddValidOrder(result.Order, 4) {
			t.Fatal(method, result, reorderErr)
		}
	}
	partial, err := problem.Reorder(start, CuddReorderExact, CuddReorderOptions{Lower: 1, Upper: 2})
	if err != nil || partial.Order[0] != 0 || partial.Order[3] != 3 {
		t.Fatal(partial, err)
	}
	swapped, swapCost, err := problem.AdjacentSwap(start, 1)
	if err != nil || !slices.Equal(swapped, []int{0, 2, 1, 3}) {
		t.Fatal(swapped, swapCost, err)
	}
	wantSwapCost, _ := problem.Cost(swapped)
	if swapCost != wantSwapCost {
		t.Fatal(swapCost, wantSwapCost)
	}
	grouped, err := problem.ReorderGroups(start, []CuddReorderGroup{{Variables: []int{1, 2}, Hard: true}}, CuddReorderOptions{})
	if err != nil || grouped.Cost > grouped.InitialCost {
		t.Fatal(grouped, err)
	}
	firstPosition, secondPosition := -1, -1
	for position, variable := range grouped.Order {
		if variable == 1 {
			firstPosition = position
		}
		if variable == 2 {
			secondPosition = position
		}
	}
	if secondPosition != firstPosition+1 {
		t.Fatal(grouped.Order)
	}
}

func TestCuddReorderInteractionSymmetryAndSharedCost(t *testing.T) {
	first := make([]bool, 8)
	second := make([]bool, 8)
	for assignment := range first {
		first[assignment] = (assignment&1 != 0) != (assignment&2 != 0)
		second[assignment] = assignment&4 != 0
	}
	problem, err := NewCuddTruthReorderProblem(3, [][]bool{first, second})
	if err != nil {
		t.Fatal(err)
	}
	matrix, err := problem.InteractionMatrix()
	if err != nil || !matrix.Interacts(0, 1) || matrix.Interacts(0, 2) || matrix.Interacts(1, 2) {
		t.Fatal(matrix, err)
	}
	symmetric, err := problem.Symmetric(0, 1)
	if err != nil || !symmetric {
		t.Fatal(symmetric, err)
	}
	symmetric, err = problem.Symmetric(0, 2)
	if err != nil || symmetric {
		t.Fatal(symmetric, err)
	}
	groups, err := problem.SymmetryGroups([]int{2, 0, 1})
	if err != nil || len(groups) != 1 || !slices.Equal(groups[0].Variables, []int{0, 1}) {
		t.Fatal(groups, err)
	}
	one, _ := NewCuddTruthReorderProblem(3, [][]bool{first})
	oneCost, _ := one.Cost([]int{0, 1, 2})
	sharedCost, _ := problem.Cost([]int{0, 1, 2})
	if sharedCost <= oneCost {
		t.Fatal(oneCost, sharedCost)
	}
	dag, err := problem.Canonical([]int{2, 0, 1})
	dagCost, _ := problem.Cost([]int{2, 0, 1})
	if err != nil || len(dag.Roots) != 2 || len(dag.Nodes) != dagCost {
		t.Fatal(dag, dagCost, err)
	}
	for root, table := range [][]bool{first, second} {
		for assignment, want := range table {
			value, valueErr := dag.Value(root, uint64(assignment))
			if valueErr != nil || (value != 0) != want {
				t.Fatal(root, assignment, value, want, valueErr)
			}
		}
	}
}

func TestCuddReorderCoverageInventory(t *testing.T) {
	coverage := CuddReorderingCoverage()
	if len(coverage.Complete) != 1 || len(coverage.Partial) < 9 || len(coverage.NativeOnly) < 3 {
		t.Fatal(coverage)
	}
	coverage.Complete[0] = "changed"
	if CuddReorderingCoverage().Complete[0] == "changed" {
		t.Fatal("coverage inventory aliases caller storage")
	}
}

func TestCuddReorderBDDAndADDExtraction(t *testing.T) {
	bdd, err := NewExtraBDDManager(3, 128)
	if err != nil {
		t.Fatal(err)
	}
	x0, _ := bdd.Variable(0)
	x1, _ := bdd.Variable(1)
	x2, _ := bdd.Variable(2)
	and, err := bdd.And(x0, x2, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	root, err := bdd.Or(and, x1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	problem, err := NewCuddBDDReorderProblem(bdd, []ExtraBDDLit{root}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		want, evalErr := bdd.Evaluate(root, []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if evalErr != nil || (problem.tables[0][assignment] != 0) != want {
			t.Fatal(assignment, want, evalErr)
		}
	}
	add, err := NewCuddADDManager(2, 64)
	if err != nil {
		t.Fatal(err)
	}
	a0, _ := add.Variable(0)
	a1, _ := add.Variable(1)
	sum, err := add.Apply(CuddADDPlus, a0, a1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	addProblem, err := NewCuddADDReorderProblem(add, []CuddADDEdge{sum}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment, want := range []float64{0, 1, 1, 2} {
		if got := math.Float64frombits(addProblem.tables[0][assignment]); got != want {
			t.Fatal(assignment, got, want)
		}
	}
}

func TestCuddReorderDeterministicAnnealAndGenetic(t *testing.T) {
	problem, err := NewCuddTruthReorderProblem(5, [][]bool{cuddTestTruth(5, 0x9e3779b9)})
	if err != nil {
		t.Fatal(err)
	}
	start := []int{4, 3, 2, 1, 0}
	options := CuddReorderOptions{Seed: 17, MaxPasses: 5, Generations: 6, Population: 8, MaxEvaluations: 2000}
	for _, method := range []CuddReorderMethod{CuddReorderAnneal, CuddReorderGenetic} {
		first, firstErr := problem.Reorder(start, method, options)
		second, secondErr := problem.Reorder(start, method, options)
		if firstErr != nil || secondErr != nil || !slices.Equal(first.Order, second.Order) || first.Cost != second.Cost || first.Cost > first.InitialCost {
			t.Fatal(method, first, second, firstErr, secondErr)
		}
	}
}

func TestCuddReorderLimitsCancellationAndNoMutation(t *testing.T) {
	problem, err := NewCuddTruthReorderProblem(4, [][]bool{cuddTestTruth(4, 0x6996)})
	if err != nil {
		t.Fatal(err)
	}
	order := []int{0, 1, 2, 3}
	before := append([]int(nil), order...)
	if result, reorderErr := problem.Reorder(order, CuddReorderExact, CuddReorderOptions{MaxEvaluations: 1}); !errors.Is(reorderErr, ErrLimit) || result.Order != nil || !slices.Equal(order, before) {
		t.Fatal(result, reorderErr, order)
	}
	if result, reorderErr := problem.Reorder(order, CuddReorderSift, CuddReorderOptions{Stop: func() bool { return true }}); !errors.Is(reorderErr, ErrCancelled) || result.Order != nil || !slices.Equal(order, before) {
		t.Fatal(result, reorderErr, order)
	}
	stopCalls := 0
	if result, reorderErr := problem.Reorder(order, CuddReorderSift, CuddReorderOptions{Stop: func() bool {
		stopCalls++
		return stopCalls == 2
	}}); !errors.Is(reorderErr, ErrCancelled) || result.Order != nil {
		t.Fatal(result, reorderErr, stopCalls)
	}
	large, _ := NewCuddTruthReorderProblem(11, [][]bool{make([]bool, 1<<11)})
	largeOrder := make([]int, 11)
	for i := range largeOrder {
		largeOrder[i] = i
	}
	if _, reorderErr := large.Reorder(largeOrder, CuddReorderExact, CuddReorderOptions{}); !errors.Is(reorderErr, ErrLimit) {
		t.Fatal(reorderErr)
	}
	if _, costErr := problem.Cost([]int{0, 0, 2, 3}); !errors.Is(costErr, ErrDimension) {
		t.Fatal(costErr)
	}
	oversizedTables := make([][]bool, cuddReorderMaxDenseCells/(1<<20)+1)
	if _, constructErr := NewCuddTruthReorderProblem(20, oversizedTables); !errors.Is(constructErr, ErrLimit) {
		t.Fatal(constructErr)
	}
	singleton, singletonErr := problem.Reorder(order, CuddReorderExact, CuddReorderOptions{Lower: 0, Upper: 0, RangeSet: true})
	if singletonErr != nil || !slices.Equal(singleton.Order, order) || singleton.Passes != 1 {
		t.Fatal(singleton, singletonErr)
	}
	if _, groupErr := problem.ReorderGroups(order, nil, CuddReorderOptions{Lower: 0, Upper: 0, RangeSet: true}); !errors.Is(groupErr, ErrDimension) {
		t.Fatal(groupErr)
	}
	dag, dagErr := problem.Canonical(order)
	if dagErr != nil || len(dag.Nodes) == 0 {
		t.Fatal(dagErr)
	}
	dag.Roots[0] = 1
	dag.Nodes[0].Low, dag.Nodes[0].High = 1, 1
	if _, valueErr := dag.Value(0, 0); !errors.Is(valueErr, ErrInvariant) {
		t.Fatal(valueErr)
	}
}

func TestCuddLevelQueueAndLinearMap(t *testing.T) {
	queue, err := NewCuddLevelQueue(4)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		key   uint64
		level int
	}{{10, 2}, {11, 0}, {12, 2}, {13, 1}} {
		added, addErr := queue.Enqueue(row.key, row.level)
		if addErr != nil || !added {
			t.Fatal(row, added, addErr)
		}
	}
	if added, addErr := queue.Enqueue(10, 3); addErr != nil || added {
		t.Fatal(added, addErr)
	}
	for _, want := range []struct {
		key   uint64
		level int
	}{{11, 0}, {13, 1}, {10, 2}, {12, 2}} {
		key, level, ok := queue.Dequeue()
		if !ok || key != want.key || level != want.level {
			t.Fatal(key, level, ok, want)
		}
	}
	mapping, err := NewCuddLinearMap(4)
	if err != nil {
		t.Fatal(err)
	}
	if err = mapping.Combine(0, 2); err != nil {
		t.Fatal(err)
	}
	for transformed := uint64(0); transformed < 16; transformed++ {
		original, mapErr := mapping.Assignment(transformed)
		if mapErr != nil || original&1 != ((transformed&1)^((transformed>>2)&1)) {
			t.Fatal(transformed, original, mapErr)
		}
	}
	rows := mapping.Rows()
	rows[0] = 0
	if mapping.Rows()[0] == 0 {
		t.Fatal("linear rows alias caller storage")
	}
	copyMapping := mapping
	if err := copyMapping.Combine(1, 3); err != nil {
		t.Fatal(err)
	}
	if slices.Equal(copyMapping.Rows(), mapping.Rows()) || mapping.Rows()[1] != uint64(1)<<1 {
		t.Fatal("linear map copies alias storage")
	}
	var nilMapping *CuddLinearMap
	if err := nilMapping.Combine(0, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestCuddReorderFiveThousandDeterministicCases(t *testing.T) {
	state := uint64(1)
	for trial := 0; trial < 5000; trial++ {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		variables := 1 + int(state%4)
		mask := uint64(1)<<(1<<variables) - 1
		problem, err := NewCuddTruthReorderProblem(variables, [][]bool{cuddTestTruth(variables, state&mask)})
		if err != nil {
			t.Fatal(trial, err)
		}
		order := make([]int, variables)
		for i := range order {
			order[i] = variables - 1 - i
		}
		result, reorderErr := problem.Reorder(order, CuddReorderSift, CuddReorderOptions{MaxEvaluations: 256, MaxPasses: 2})
		if reorderErr != nil || result.Cost > result.InitialCost || !cuddValidOrder(result.Order, variables) {
			t.Fatal(trial, result, reorderErr)
		}
	}
}

func FuzzCuddReorderCanonicalAndExact(f *testing.F) {
	f.Add(byte(3), uint64(0x96), uint64(7))
	f.Add(byte(4), uint64(0x6996), uint64(19))
	f.Fuzz(func(t *testing.T, variableByte byte, truth, seed uint64) {
		variables := 1 + int(variableByte%5)
		table := cuddTestTruth(variables, truth)
		problem, err := NewCuddTruthReorderProblem(variables, [][]bool{table})
		if err != nil {
			t.Fatal(err)
		}
		order := make([]int, variables)
		for i := range order {
			order[i] = variables - 1 - i
		}
		result, err := problem.Reorder(order, CuddReorderExact, CuddReorderOptions{Seed: seed, MaxEvaluations: 1000})
		if err != nil {
			t.Fatal(err)
		}
		if result.Cost > result.InitialCost || !cuddValidOrder(result.Order, variables) {
			t.Fatal(result)
		}
		got, err := problem.Cost(result.Order)
		if err != nil || got != cuddIndependentCost(table, variables, result.Order) {
			t.Fatal(got, result, err)
		}
	})
}
