package abc

// Stand-alone semantic tests for src/bdd/cudd/cuddApprox.c, cuddClip.c,
// cuddSubsetHB.c, cuddSubsetSP.c, cuddPriority.c, and cuddSign.c.

import (
	"errors"
	"math"
	"math/bits"
	"testing"
)

func cuddApproxManager(t testing.TB, variables int) *ExtraBDDManager {
	t.Helper()
	manager, err := NewExtraBDDManager(variables, 1<<17)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func cuddApproxFromTable(t testing.TB, manager *ExtraBDDManager, table uint64) ExtraBDDLit {
	t.Helper()
	values := make([]bool, 1<<uint(manager.variables))
	for assignment := range values {
		values[assignment] = table>>uint(assignment)&1 != 0
	}
	options, _ := normalizeExtraBDDOptions(ExtraBDDOptions{})
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: options, transaction: &transaction}
	root, err := manager.analysisBuild(values, &work)
	if err != nil {
		manager.rollback(&transaction)
		t.Fatal(err)
	}
	return root
}

func cuddApproxTable(t testing.TB, manager *ExtraBDDManager, root ExtraBDDLit) uint64 {
	t.Helper()
	var table uint64
	assignment := make([]bool, manager.variables)
	for row := 0; row < 1<<uint(manager.variables); row++ {
		for variable := range assignment {
			assignment[variable] = row>>uint(variable)&1 != 0
		}
		value, err := manager.Evaluate(root, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if value {
			table |= uint64(1) << uint(row)
		}
	}
	return table
}

func TestCuddApproxExhaustiveFourVariableContracts(t *testing.T) {
	manager := cuddApproxManager(t, 4)
	algorithms := []func(ExtraBDDLit, CuddApproxOptions) (ExtraBDDLit, error){
		manager.UnderApprox,
		manager.RemapUnderApprox,
		manager.SubsetHeavyBranch,
		manager.SubsetShortPaths,
	}
	for table := uint64(0); table < 1<<16; table++ {
		root := cuddApproxFromTable(t, manager, table)
		for index, algorithm := range algorithms {
			result, err := algorithm(root, CuddApproxOptions{Threshold: 3, Quality: 1})
			if err != nil {
				t.Fatalf("table %#x algorithm %d: %v", table, index, err)
			}
			got := cuddApproxTable(t, manager, result)
			if got&^table != 0 {
				t.Fatalf("table %#x algorithm %d produced non-subset %#x", table, index, got)
			}
			nodes, err := manager.CountNodes(result, ExtraBDDOptions{})
			if err != nil || nodes > 3 {
				t.Fatalf("table %#x algorithm %d nodes=%d err=%v", table, index, nodes, err)
			}
		}
		over, err := manager.OverApprox(root, CuddApproxOptions{Threshold: 3})
		if err != nil {
			t.Fatalf("table %#x over: %v", table, err)
		}
		if got := cuddApproxTable(t, manager, over); table&^got != 0 {
			t.Fatalf("table %#x is not contained by %#x", table, got)
		}
	}
}

func TestCuddClippingExhaustiveIndependentModel(t *testing.T) {
	manager := cuddApproxManager(t, 3)
	mask := uint64(0xff)
	for left := uint64(0); left <= mask; left++ {
		leftRoot := cuddApproxFromTable(t, manager, left)
		for right := uint64(0); right <= mask; right++ {
			rightRoot := cuddApproxFromTable(t, manager, right)
			exact := left & right
			under, err := manager.ClippingAnd(leftRoot, rightRoot, 1, CuddClipUnder, ExtraBDDOptions{})
			if err != nil {
				t.Fatal(err)
			}
			over, err := manager.ClippingAnd(leftRoot, rightRoot, 1, CuddClipOver, ExtraBDDOptions{})
			if err != nil {
				t.Fatal(err)
			}
			underTable, overTable := cuddApproxTable(t, manager, under), cuddApproxTable(t, manager, over)
			if underTable&^exact != 0 || exact&^overTable != 0 {
				t.Fatalf("%#x & %#x exact=%#x under=%#x over=%#x", left, right, exact, underTable, overTable)
			}
		}
	}
}

func TestCuddPriorityPredicatesSelectionAndSignatures(t *testing.T) {
	manager := cuddApproxManager(t, 4)
	x, y := []int{0, 1}, []int{2, 3}
	greater, err := manager.CuddXGreaterY(x, y, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	equal, err := manager.CuddXEqualY(x, y, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	difference, err := manager.CuddDifferenceAtLeast(x, y, 1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		xv := (assignment&1)<<1 | (assignment>>1)&1
		yv := ((assignment>>2)&1)<<1 | (assignment>>3)&1
		row := make([]bool, 4)
		for variable := range row {
			row[variable] = assignment>>uint(variable)&1 != 0
		}
		gotGreater, _ := manager.Evaluate(greater, row)
		gotEqual, _ := manager.Evaluate(equal, row)
		gotDifference, _ := manager.Evaluate(difference, row)
		if gotGreater != (xv > yv) || gotEqual != (xv == yv) || gotDifference != (xv-yv >= 1) {
			t.Fatalf("assignment %04b x=%d y=%d got=(%v,%v,%v)", assignment, xv, yv, gotGreater, gotEqual, gotDifference)
		}
	}
	relation := cuddApproxFromTable(t, manager, 0xffff)
	selected, err := manager.PrioritySelect(relation, x, y, CuddPrioritySmallest, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	selectedTable := cuddApproxTable(t, manager, selected)
	for assignment := 0; assignment < 16; assignment++ {
		yv := ((assignment>>2)&1)<<1 | (assignment>>3)&1
		want := yv == 0
		if (selectedTable>>uint(assignment)&1 != 0) != want {
			t.Fatalf("priority row %04b", assignment)
		}
	}
	projection, err := manager.CompatibleProjection(relation, y, ExtraBDDOptions{})
	if err != nil || cuddApproxTable(t, manager, projection) != selectedTable {
		t.Fatalf("compatible projection table=%#x want %#x err=%v", cuddApproxTable(t, manager, projection), selectedTable, err)
	}
	closest, err := manager.ClosestCube(greater, equal, ExtraBDDOptions{})
	if err != nil || !closest.Exists || closest.Minterms != 1 {
		t.Fatalf("closest=%+v err=%v", closest, err)
	}
	closestTable := cuddApproxTable(t, manager, closest.Root)
	truth := cuddApproxTable(t, manager, greater)
	equalTable := cuddApproxTable(t, manager, equal)
	if closestTable&truth == 0 {
		t.Fatalf("closest cube %#x does not imply first operand %#x", closestTable, truth)
	}
	distance, wantDistance := manager.variables+1, manager.variables+1
	for firstRow := 0; firstRow < 16; firstRow++ {
		for secondRow := 0; secondRow < 16; secondRow++ {
			if closestTable>>uint(firstRow)&1 != 0 && equalTable>>uint(secondRow)&1 != 0 {
				distance = min(distance, bits.OnesCount(uint(firstRow^secondRow)))
			}
			if truth>>uint(firstRow)&1 != 0 && equalTable>>uint(secondRow)&1 != 0 {
				wantDistance = min(wantDistance, bits.OnesCount(uint(firstRow^secondRow)))
			}
		}
	}
	if distance != wantDistance {
		t.Fatalf("closest distance=%d want %d", distance, wantDistance)
	}
	signatures, err := manager.CofactorMintermSignatures(greater, ExtraBDDOptions{})
	if err != nil || len(signatures) != 5 {
		t.Fatalf("signatures=%v err=%v", signatures, err)
	}
	for variable := 0; variable < 4; variable++ {
		count := 0
		for assignment := 0; assignment < 16; assignment++ {
			if assignment>>uint(variable)&1 != 0 && truth>>uint(assignment)&1 != 0 {
				count++
			}
		}
		if math.Abs(signatures[variable]-float64(count)/8) > 1e-12 {
			t.Fatalf("signature %d=%g want %g", variable, signatures[variable], float64(count)/8)
		}
	}
	if math.Abs(signatures[4]-float64(bits.OnesCount64(truth))/16) > 1e-12 {
		t.Fatalf("on-set signature %g", signatures[4])
	}
}

func TestCuddPriorityBDDSelectionHammingModel(t *testing.T) {
	manager := cuddApproxManager(t, 6)
	x, y, z := []int{0, 1}, []int{2, 3}, []int{4, 5}
	priority, err := manager.CuddDistanceXYGreaterXZ(x, y, z, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	selected, err := manager.PrioritySelectByBDD(ExtraBDDTrue, priority, x, y, z, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assignment := make([]bool, 6)
	for row := 0; row < 64; row++ {
		for variable := range assignment {
			assignment[variable] = row>>uint(variable)&1 != 0
		}
		value, evalErr := manager.Evaluate(selected, assignment)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		xv := (row&1)<<1 | row>>1&1
		yv := (row>>2&1)<<1 | row>>3&1
		if value != (xv == yv) {
			t.Fatalf("row %06b x=%d y=%d selected=%v", row, xv, yv, value)
		}
	}
}

func TestCuddCompatibleProjectionSelectsNonzeroRepresentatives(t *testing.T) {
	manager := cuddApproxManager(t, 3)
	// For retained variable 0, the allowed projected values (variables 1,2)
	// are {2,3} when it is false and {1,3} when it is true. Neither selected
	// representative is the all-zero value.
	relation := cuddApproxFromTable(t, manager, 1<<2|1<<6|1<<5|1<<7)
	projection, err := manager.CompatibleProjection(relation, []int{1, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cuddApproxTable(t, manager, projection), uint64(1<<2|1<<5); got != want {
		t.Fatalf("projection=%#x want %#x", got, want)
	}
}

func TestCuddBiasedApproximationKeepsInputWithinThreshold(t *testing.T) {
	manager := cuddApproxManager(t, 4)
	root := cuddApproxFromTable(t, manager, 0x6996)
	nodes, err := manager.CountNodes(root, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	under, err := manager.BiasedUnderApprox(root, ExtraBDDFalse, CuddApproxOptions{Threshold: nodes})
	if err != nil || under != root {
		t.Fatalf("under=%d want %d err=%v", under, root, err)
	}
	over, err := manager.BiasedOverApprox(root, ExtraBDDFalse, CuddApproxOptions{Threshold: nodes})
	if err != nil || over != root {
		t.Fatalf("over=%d want %d err=%v", over, root, err)
	}
}

func TestCuddApproxLimitsCancellationAndRollback(t *testing.T) {
	manager := cuddApproxManager(t, 4)
	root := cuddApproxFromTable(t, manager, 0x6996)
	before := len(manager.nodes)
	if result, err := manager.UnderApprox(root, CuddApproxOptions{Threshold: 3, ExtraBDDOptions: ExtraBDDOptions{Stop: func() bool { return true }}}); result != 0 || !errors.Is(err, ErrCancelled) || len(manager.nodes) != before {
		t.Fatalf("cancel result=%d err=%v nodes=%d/%d", result, err, len(manager.nodes), before)
	}
	if result, err := manager.ClippingAnd(root, root.Not(), 5, CuddClipDirection(9), ExtraBDDOptions{}); result != 0 || !errors.Is(err, ErrDimension) {
		t.Fatalf("bad direction result=%d err=%v", result, err)
	}
	if _, err := manager.UnderApprox(root, CuddApproxOptions{Threshold: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative threshold: %v", err)
	}
	if _, err := manager.CuddXGreaterY([]int{0}, []int{0}, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("overlapping variables: %v", err)
	}
	small := cuddApproxManager(t, 1)
	smallBefore := len(small.nodes)
	if result, err := small.UnderApprox(2, CuddApproxOptions{Threshold: 1, ExtraBDDOptions: ExtraBDDOptions{MaxOperations: 3}}); result != 0 || !errors.Is(err, ErrLimit) || len(small.nodes) != smallBefore {
		t.Fatalf("cumulative limit result=%d err=%v nodes=%d/%d", result, err, len(small.nodes), smallBefore)
	}
	if signatures, err := manager.CofactorMintermSignatures(ExtraBDDTrue, ExtraBDDOptions{Stop: func() bool { return true }}); signatures != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("constant signature cancellation signatures=%v err=%v", signatures, err)
	}
	if result, err := manager.CompatibleProjection(ExtraBDDFalse, []int{0}, ExtraBDDOptions{Stop: func() bool { return true }}); result != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatalf("constant projection cancellation result=%d err=%v", result, err)
	}
	large := cuddApproxManager(t, 26)
	x, y := make([]int, 13), make([]int, 13)
	for index := range x {
		x[index], y[index] = index, index+13
	}
	if result, err := large.CuddXGreaterY(x, y, ExtraBDDOptions{}); result != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("predicate preflight result=%d err=%v", result, err)
	}
}

func TestCuddApproxDeterministicCampaign5000(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	manager := cuddApproxManager(t, 4)
	for iteration := 0; iteration < 5000; iteration++ {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		table := state & 0xffff
		root := cuddApproxFromTable(t, manager, table)
		threshold := 1 + int(state>>20%4)
		under, err := manager.SubsetHeavyBranch(root, CuddApproxOptions{Threshold: threshold})
		if err != nil {
			t.Fatalf("iteration %d: %v", iteration, err)
		}
		got := cuddApproxTable(t, manager, under)
		if got&^table != 0 {
			t.Fatalf("iteration %d: %#x not subset %#x", iteration, got, table)
		}
		nodes, _ := manager.CountNodes(under, ExtraBDDOptions{})
		if nodes > threshold {
			t.Fatalf("iteration %d: %d > %d", iteration, nodes, threshold)
		}
	}
}

func FuzzCuddApproxContracts(f *testing.F) {
	for _, seed := range []uint16{0, 1, 0xffff, 0x6996, 0x9669, 0xa55a} {
		f.Add(seed, uint8(2))
	}
	f.Fuzz(func(t *testing.T, table uint16, thresholdByte uint8) {
		manager := cuddApproxManager(t, 4)
		root := cuddApproxFromTable(t, manager, uint64(table))
		threshold := 1 + int(thresholdByte%5)
		under, err := manager.SubsetShortPaths(root, CuddApproxOptions{Threshold: threshold})
		if err != nil {
			t.Fatal(err)
		}
		got := cuddApproxTable(t, manager, under)
		if got&^uint64(table) != 0 {
			t.Fatalf("%#x is not a subset of %#x", got, table)
		}
		nodes, _ := manager.CountNodes(under, ExtraBDDOptions{})
		if nodes > threshold {
			t.Fatalf("nodes=%d threshold=%d", nodes, threshold)
		}
	})
}
