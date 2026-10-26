package abc

// Tests for the numeric ADD port from the seven src/bdd/cudd/cuddAdd*.c units.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func cuddADDTestManager(t *testing.T, variables int) *CuddADDManager {
	t.Helper()
	manager, err := NewCuddADDManager(variables, 1<<17)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func cuddADDTestDense(t *testing.T, manager *CuddADDManager, edge CuddADDEdge, want []float64) {
	t.Helper()
	got, err := manager.Dense(edge)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("dense length=%d want=%d", len(got), len(want))
	}
	for index := range got {
		if math.Float64bits(got[index]) != math.Float64bits(want[index]) {
			t.Fatalf("assignment=%d got=%v want=%v", index, got[index], want[index])
		}
	}
}

func cuddADDTestBinaryModel(operation CuddADDOperation, left, right float64) float64 {
	switch operation {
	case CuddADDPlus:
		return left + right
	case CuddADDTimes:
		return left * right
	case CuddADDThreshold:
		if left >= right {
			return left
		}
		return 0
	case CuddADDSetNZ:
		if right != 0 {
			return right
		}
		return left
	case CuddADDDivide:
		if left == 0 {
			return 0
		}
		return left / right
	case CuddADDMinus:
		return left - right
	case CuddADDMinimum:
		return math.Min(left, right)
	case CuddADDMaximum:
		return math.Max(left, right)
	case CuddADDOneZeroMaximum:
		if left > right {
			return 1
		}
		return 0
	case CuddADDDiff:
		if left == right {
			return math.Inf(1)
		}
		return math.Min(left, right)
	case CuddADDAgreement:
		if left == right {
			return left
		}
		return 0
	case CuddADDOr:
		if left == 1 || right == 1 {
			return 1
		}
		return 0
	case CuddADDNand:
		if left == 1 && right == 1 {
			return 0
		}
		return 1
	case CuddADDNor:
		if left == 0 && right == 0 {
			return 1
		}
		return 0
	case CuddADDXor:
		if left != right {
			return 1
		}
		return 0
	case CuddADDXnor:
		if left == right {
			return 1
		}
		return 0
	default:
		panic("unknown test operation")
	}
}

func TestCuddADDDenseCanonicalAndApply(t *testing.T) {
	manager := cuddADDTestManager(t, 4)
	leftValues := make([]float64, 16)
	rightValues := make([]float64, 16)
	for assignment := range leftValues {
		leftValues[assignment] = float64((assignment & 3) + 1)
		rightValues[assignment] = float64(((assignment >> 2) & 3) + 2)
	}
	left, err := manager.FromDense(leftValues, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	leftAgain, err := manager.FromDense(slices.Clone(leftValues), ExtraBDDOptions{})
	if err != nil || leftAgain != left {
		t.Fatalf("canonical edge=%d again=%d err=%v", left, leftAgain, err)
	}
	right, err := manager.FromDense(rightValues, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	operations := []CuddADDOperation{CuddADDPlus, CuddADDTimes, CuddADDThreshold, CuddADDSetNZ, CuddADDDivide, CuddADDMinus, CuddADDMinimum, CuddADDMaximum, CuddADDOneZeroMaximum, CuddADDDiff, CuddADDAgreement}
	for _, operation := range operations {
		result, applyErr := manager.Apply(operation, left, right, ExtraBDDOptions{})
		if applyErr != nil {
			t.Fatalf("operation=%d: %v", operation, applyErr)
		}
		want := make([]float64, len(leftValues))
		for index := range want {
			want[index] = cuddADDTestBinaryModel(operation, leftValues[index], rightValues[index])
		}
		cuddADDTestDense(t, manager, result, want)
	}
	booleanLeft := make([]float64, 16)
	booleanRight := make([]float64, 16)
	for assignment := range booleanLeft {
		booleanLeft[assignment] = float64(assignment & 1)
		booleanRight[assignment] = float64((assignment >> 1) & 1)
	}
	a, _ := manager.FromDense(booleanLeft, ExtraBDDOptions{})
	b, _ := manager.FromDense(booleanRight, ExtraBDDOptions{})
	for _, operation := range []CuddADDOperation{CuddADDOr, CuddADDNand, CuddADDNor, CuddADDXor, CuddADDXnor} {
		result, applyErr := manager.Apply(operation, a, b, ExtraBDDOptions{})
		if applyErr != nil {
			t.Fatal(applyErr)
		}
		want := make([]float64, 16)
		for index := range want {
			want[index] = cuddADDTestBinaryModel(operation, booleanLeft[index], booleanRight[index])
		}
		cuddADDTestDense(t, manager, result, want)
	}
}

func TestCuddADDIteUnaryFindAndBits(t *testing.T) {
	manager := cuddADDTestManager(t, 3)
	condition, _ := manager.Variable(1)
	thenEdge, _ := manager.FromDense([]float64{1.2, 2.4, 3.1, 4.6, -1.2, -2.4, -3.1, -4.6}, ExtraBDDOptions{})
	elseEdge, _ := manager.Constant(7)
	result, err := manager.Ite(condition, thenEdge, elseEdge, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{7, 7, 3.1, 4.6, 7, 7, -3.1, -4.6}
	cuddADDTestDense(t, manager, result, want)
	if _, err = manager.Ite(thenEdge, thenEdge, elseEdge, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("non-Boolean ITE err=%v", err)
	}
	constant, ok, err := manager.IteConstant(condition, elseEdge, elseEdge, ExtraBDDOptions{})
	if err != nil || !ok || constant != 7 {
		t.Fatalf("ite constant=%v ok=%v err=%v", constant, ok, err)
	}
	evaluated, ok, err := manager.EvalConst(condition, result, -9, ExtraBDDOptions{})
	if err != nil || ok || evaluated != 0 {
		t.Fatalf("eval constant=%v ok=%v err=%v", evaluated, ok, err)
	}
	complement, _ := manager.Complement(thenEdge, ExtraBDDOptions{})
	for assignment, value := range []float64{1.2, 2.4, 3.1, 4.6, -1.2, -2.4, -3.1, -4.6} {
		got, _ := manager.Evaluate(complement, uint64(assignment))
		if got != 0 || value == 0 {
			t.Fatalf("complement assignment=%d got=%v", assignment, got)
		}
	}
	negated, _ := manager.Negate(thenEdge, ExtraBDDOptions{})
	negValues, _ := manager.Dense(negated)
	if negValues[3] != -4.6 || negValues[7] != 4.6 {
		t.Fatalf("negated=%v", negValues)
	}
	rounded, _ := manager.RoundOff(thenEdge, 0, ExtraBDDOptions{})
	cuddADDTestDense(t, manager, rounded, []float64{2, 3, 4, 5, -1, -2, -3, -4})
	maximum, err := manager.FindMax(thenEdge, ExtraBDDOptions{})
	if err != nil || maximum != 4.6 {
		t.Fatalf("max=%v err=%v", maximum, err)
	}
	minimum, _ := manager.FindMin(thenEdge, ExtraBDDOptions{})
	if minimum != -4.6 {
		t.Fatalf("min=%v", minimum)
	}
	integers, _ := manager.FromDense([]float64{0, 1, 2, 3, 4, 5, 6, 7}, ExtraBDDOptions{})
	bit, _ := manager.IthBit(integers, 1, ExtraBDDOptions{})
	cuddADDTestDense(t, manager, bit, []float64{0, 0, 1, 1, 0, 0, 1, 1})
	inverse, err := manager.ScalarInverse(integers, 0.5, ExtraBDDOptions{})
	if !errors.Is(err, ErrDimension) || inverse != 0 {
		t.Fatalf("inverse=%d err=%v", inverse, err)
	}
	positive, _ := manager.FromDense([]float64{1, 2, 4, 8, 1, 2, 4, 8}, ExtraBDDOptions{})
	inverse, err = manager.ScalarInverse(positive, 0.5, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddADDTestDense(t, manager, inverse, []float64{1, .5, .25, .125, 1, .5, .25, .125})
	leq, err := manager.Leq(inverse, positive, ExtraBDDOptions{})
	if err != nil || !leq {
		t.Fatalf("leq=%v err=%v", leq, err)
	}
}

func TestCuddADDAbstractionWalshAndResidue(t *testing.T) {
	manager := cuddADDTestManager(t, 4)
	values := make([]float64, 16)
	for assignment := range values {
		values[assignment] = float64((assignment & 1) + 2*((assignment>>1)&1) + 1)
	}
	edge, _ := manager.FromDense(values, ExtraBDDOptions{})
	exist, err := manager.ExistAbstract(edge, []int{0}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantExist := make([]float64, 16)
	for assignment := range wantExist {
		base := assignment &^ 1
		wantExist[assignment] = values[base] + values[base|1]
	}
	cuddADDTestDense(t, manager, exist, wantExist)
	constant, _ := manager.Constant(3)
	absentExist, err := manager.ExistAbstract(constant, []int{0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddADDTestDense(t, manager, absentExist, slices.Repeat([]float64{12}, 16))
	absentUniversal, err := manager.UnivAbstract(constant, []int{0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddADDTestDense(t, manager, absentUniversal, slices.Repeat([]float64{81}, 16))
	universal, _ := manager.UnivAbstract(edge, []int{1}, ExtraBDDOptions{})
	wantUniversal := make([]float64, 16)
	for assignment := range wantUniversal {
		base := assignment &^ 2
		wantUniversal[assignment] = values[base] * values[base|2]
	}
	cuddADDTestDense(t, manager, universal, wantUniversal)
	boolean, _ := manager.FromDense([]float64{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1}, ExtraBDDOptions{})
	orAbstract, _ := manager.OrAbstract(boolean, []int{0}, ExtraBDDOptions{})
	cuddADDTestDense(t, manager, orAbstract, []float64{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1})
	walsh, err := manager.Walsh([]int{0, 1}, []int{2, 3}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		parity := ((assignment>>0)&1)&((assignment>>2)&1) ^ ((assignment>>1)&1)&((assignment>>3)&1)
		want := 1.0
		if parity != 0 {
			want = -1
		}
		got, _ := manager.Evaluate(walsh, uint64(assignment))
		if got != want {
			t.Fatalf("walsh assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	// The rows of a 4-by-4 Walsh matrix are pairwise orthogonal.
	for rowA := 0; rowA < 4; rowA++ {
		for rowB := 0; rowB < 4; rowB++ {
			dot := 0.0
			for column := 0; column < 4; column++ {
				a, _ := manager.Evaluate(walsh, uint64(rowA|(column<<2)))
				b, _ := manager.Evaluate(walsh, uint64(rowB|(column<<2)))
				dot += a * b
			}
			wantDot := 0.0
			if rowA == rowB {
				wantDot = 4
			}
			if dot != wantDot {
				t.Fatalf("Walsh rows %d,%d dot=%v want=%v", rowA, rowB, dot, wantDot)
			}
		}
	}
	residue, err := manager.Residue(3, 3, 0, 0, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		got, _ := manager.Evaluate(residue, uint64(assignment))
		if want := float64((assignment & 7) % 3); got != want {
			t.Fatalf("residue assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	twos, _ := manager.Residue(3, 5, CuddADDResidueTwosComplement|CuddADDResidueMSB, 0, ExtraBDDOptions{})
	for assignment := 0; assignment < 8; assignment++ {
		number := ((assignment >> 2) & 1) + 2*((assignment>>1)&1) + 4*(assignment&1)
		if assignment&1 != 0 {
			number -= 8
		}
		want := number % 5
		if want < 0 {
			want += 5
		}
		got, _ := manager.Evaluate(twos, uint64(assignment))
		if got != float64(want) {
			t.Fatalf("twos assignment=%d got=%v want=%d", assignment, got, want)
		}
	}
}

func TestCuddADDLimitsCancellationAndLifecycle(t *testing.T) {
	manager := cuddADDTestManager(t, 3)
	left, _ := manager.FromDense([]float64{1, 2, 3, 4, 5, 6, 7, 8}, ExtraBDDOptions{})
	right, _ := manager.FromDense([]float64{8, 7, 6, 5, 4, 3, 2, 1}, ExtraBDDOptions{})
	before := manager.NodeCount()
	if _, err := manager.Apply(CuddADDTimes, left, right, ExtraBDDOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) || manager.NodeCount() != before {
		t.Fatalf("limit err=%v nodes=%d before=%d", err, manager.NodeCount(), before)
	}
	calls := 0
	if _, err := manager.Apply(CuddADDPlus, left, right, ExtraBDDOptions{Stop: func() bool { calls++; return calls > 2 }}); !errors.Is(err, ErrCancelled) || manager.NodeCount() != before {
		t.Fatalf("cancel err=%v nodes=%d before=%d", err, manager.NodeCount(), before)
	}
	if _, err := manager.Constant(math.NaN()); !errors.Is(err, ErrDimension) {
		t.Fatalf("NaN err=%v", err)
	}
	zero, _ := manager.Constant(0)
	quotient, err := manager.Apply(CuddADDDivide, zero, zero, ExtraBDDOptions{})
	if err != nil || quotient != zero {
		t.Fatalf("zero divide zero=%d err=%v", quotient, err)
	}
	if _, err = manager.Apply(CuddADDOr, left, right, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("non-Boolean OR err=%v", err)
	}
	infinity, _ := manager.Constant(math.Inf(1))
	if _, err = manager.IthBit(infinity, 0, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("infinite bit extraction err=%v", err)
	}
	negative, _ := manager.Constant(-1)
	if _, err = manager.Log(negative, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative logarithm err=%v", err)
	}
	if _, err := manager.FromDense([]float64{1}, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension err=%v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Constant(1); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed err=%v", err)
	}
}

func TestCuddADDQueriesAndGeneratorsHonorBudgetsBeforeAllocation(t *testing.T) {
	queries, err := NewCuddADDManager(20, 128)
	if err != nil {
		t.Fatal(err)
	}
	condition, _ := queries.Variable(0)
	value, _ := queries.Variable(1)
	cancelled := ExtraBDDOptions{Stop: func() bool { return true }}
	if _, _, err = queries.EvalConst(condition, value, -1, cancelled); !errors.Is(err, ErrCancelled) {
		t.Fatalf("EvalConst cancellation err=%v", err)
	}
	if _, err = queries.Leq(condition, value, cancelled); !errors.Is(err, ErrCancelled) {
		t.Fatalf("Leq cancellation err=%v", err)
	}
	if _, err = queries.FindMax(value, cancelled); !errors.Is(err, ErrCancelled) {
		t.Fatalf("FindMax cancellation err=%v", err)
	}
	limited := ExtraBDDOptions{MaxOperations: 1}
	if _, _, err = queries.EvalConst(condition, value, -1, limited); !errors.Is(err, ErrLimit) {
		t.Fatalf("EvalConst operation limit err=%v", err)
	}
	if _, err = queries.Leq(condition, value, limited); !errors.Is(err, ErrLimit) {
		t.Fatalf("Leq operation limit err=%v", err)
	}
	if _, err = queries.FindMin(value, limited); !errors.Is(err, ErrLimit) {
		t.Fatalf("FindMin operation limit err=%v", err)
	}

	generators := cuddADDTestManager(t, 4)
	before := generators.NodeCount()
	if _, err = generators.Walsh([]int{0}, []int{1}, ExtraBDDOptions{MaxOperations: 16}); !errors.Is(err, ErrLimit) || generators.NodeCount() != before {
		t.Fatalf("Walsh cumulative limit err=%v nodes=%d want=%d", err, generators.NodeCount(), before)
	}
	calls := 0
	if _, err = generators.Residue(3, 5, 0, 0, ExtraBDDOptions{Stop: func() bool { calls++; return calls > 20 }}); !errors.Is(err, ErrCancelled) || generators.NodeCount() != before {
		t.Fatalf("Residue cancellation err=%v nodes=%d want=%d", err, generators.NodeCount(), before)
	}

	large, err := NewCuddADDManager(10, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	largeBefore := large.NodeCount()
	bytesLimited := ExtraBDDOptions{MaxBytes: 10_000}
	if _, err = large.Walsh([]int{0}, []int{1}, bytesLimited); !errors.Is(err, ErrLimit) || large.NodeCount() != largeBefore {
		t.Fatalf("Walsh byte limit err=%v nodes=%d want=%d", err, large.NodeCount(), largeBefore)
	}
	if _, err = large.Residue(4, 7, 0, 0, bytesLimited); !errors.Is(err, ErrLimit) || large.NodeCount() != largeBefore {
		t.Fatalf("Residue byte limit err=%v nodes=%d want=%d", err, large.NodeCount(), largeBefore)
	}
}

func TestCuddADDIndependentModel(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	operations := []CuddADDOperation{CuddADDPlus, CuddADDTimes, CuddADDMinus, CuddADDMinimum, CuddADDMaximum, CuddADDThreshold, CuddADDOneZeroMaximum}
	for trial := 0; trial < 5000; trial++ {
		manager := cuddADDTestManager(t, 3)
		leftValues := make([]float64, 8)
		rightValues := make([]float64, 8)
		for index := range leftValues {
			leftValues[index] = float64(int(next()%9) - 4)
			rightValues[index] = float64(int(next()%9) - 4)
		}
		left, err := manager.FromDense(leftValues, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		right, err := manager.FromDense(rightValues, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		operation := operations[next()%uint64(len(operations))]
		result, err := manager.Apply(operation, left, right, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := range leftValues {
			want := cuddADDTestBinaryModel(operation, leftValues[assignment], rightValues[assignment])
			got, evalErr := manager.Evaluate(result, uint64(assignment))
			if evalErr != nil || got != want {
				t.Fatalf("trial=%d operation=%d assignment=%d got=%v want=%v err=%v", trial, operation, assignment, got, want, evalErr)
			}
		}
	}
}

func TestCuddADDAlgebraicLawsExhaustiveFourVariables(t *testing.T) {
	manager := cuddADDTestManager(t, 4)
	leftValues := make([]float64, 16)
	rightValues := make([]float64, 16)
	for assignment := 0; assignment < 16; assignment++ {
		leftValues[assignment] = float64((assignment%7)-3) / 2
		rightValues[assignment] = float64(((assignment*5)%9)-4) / 2
	}
	left, _ := manager.FromDense(leftValues, ExtraBDDOptions{})
	right, _ := manager.FromDense(rightValues, ExtraBDDOptions{})
	zero, _ := manager.Constant(0)
	one, _ := manager.Constant(1)
	for _, operation := range []CuddADDOperation{CuddADDPlus, CuddADDTimes, CuddADDMinimum, CuddADDMaximum} {
		forward, err := manager.Apply(operation, left, right, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		reverse, err := manager.Apply(operation, right, left, ExtraBDDOptions{})
		if err != nil || forward != reverse {
			t.Fatalf("operation %d is not canonical-commutative: %d %d err=%v", operation, forward, reverse, err)
		}
	}
	plusZero, _ := manager.Apply(CuddADDPlus, left, zero, ExtraBDDOptions{})
	timesOne, _ := manager.Apply(CuddADDTimes, left, one, ExtraBDDOptions{})
	minusSelf, _ := manager.Apply(CuddADDMinus, left, left, ExtraBDDOptions{})
	if plusZero != left || timesOne != left || minusSelf != zero {
		t.Fatalf("identities plus=%d times=%d minus=%d left=%d zero=%d", plusZero, timesOne, minusSelf, left, zero)
	}
	negated, _ := manager.Negate(left, ExtraBDDOptions{})
	restored, _ := manager.Negate(negated, ExtraBDDOptions{})
	if restored != left {
		t.Fatalf("double negation=%d want=%d", restored, left)
	}
	conditionValues := make([]float64, 16)
	for assignment := range conditionValues {
		conditionValues[assignment] = float64(assignment & 1)
	}
	condition, _ := manager.FromDense(conditionValues, ExtraBDDOptions{})
	selected, err := manager.Ite(condition, left, right, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		want := rightValues[assignment]
		if assignment&1 != 0 {
			want = leftValues[assignment]
		}
		got, _ := manager.Evaluate(selected, uint64(assignment))
		if got != want {
			t.Fatalf("ITE assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	nodes := manager.NodeCount()
	if _, _, err := manager.IteConstant(condition, left, right, ExtraBDDOptions{}); err != nil || manager.NodeCount() != nodes {
		t.Fatalf("ITE constant err=%v nodes=%d want=%d", err, manager.NodeCount(), nodes)
	}
}

func FuzzCuddADDOperations(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4}, []byte{4, 3, 2, 1}, uint8(CuddADDPlus))
	f.Add([]byte{0, 1, 0xff}, []byte{2, 0xfe, 3}, uint8(CuddADDMaximum))
	f.Fuzz(func(t *testing.T, leftBytes, rightBytes []byte, operationByte uint8) {
		if len(leftBytes) == 0 || len(rightBytes) == 0 {
			return
		}
		manager, err := NewCuddADDManager(3, 1<<12)
		if err != nil {
			t.Fatal(err)
		}
		leftValues := make([]float64, 8)
		rightValues := make([]float64, 8)
		for index := range leftValues {
			leftValues[index] = float64(int8(leftBytes[index%len(leftBytes)]))
			rightValues[index] = float64(int8(rightBytes[index%len(rightBytes)]))
		}
		left, err := manager.FromDense(leftValues, ExtraBDDOptions{MaxOperations: 1 << 15})
		if err != nil {
			t.Fatal(err)
		}
		right, err := manager.FromDense(rightValues, ExtraBDDOptions{MaxOperations: 1 << 15})
		if err != nil {
			t.Fatal(err)
		}
		operations := []CuddADDOperation{CuddADDPlus, CuddADDTimes, CuddADDMinus, CuddADDMinimum, CuddADDMaximum, CuddADDThreshold, CuddADDOneZeroMaximum}
		operation := operations[int(operationByte)%len(operations)]
		result, err := manager.Apply(operation, left, right, ExtraBDDOptions{MaxOperations: 1 << 15})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := range leftValues {
			want := cuddADDTestBinaryModel(operation, leftValues[assignment], rightValues[assignment])
			got, evalErr := manager.Evaluate(result, uint64(assignment))
			if evalErr != nil || got != want {
				t.Fatalf("operation=%d assignment=%d got=%v want=%v err=%v", operation, assignment, got, want, evalErr)
			}
		}
	})
}
