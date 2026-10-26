package abc

// Stand-alone tests for src/bdd/cudd/cuddMatMult.c.

import (
	"errors"
	"math"
	"testing"
)

func cuddMatrixInputs(manager *CuddADDManager, left, right []float64, bits int) (CuddADDEdge, CuddADDEdge, error) {
	variables := bits * 3
	size := 1 << uint(variables)
	order := 1 << uint(bits)
	leftDense := make([]float64, size)
	rightDense := make([]float64, size)
	mask := order - 1
	for assignment := range size {
		x := assignment & mask
		z := (assignment >> uint(bits)) & mask
		y := (assignment >> uint(2*bits)) & mask
		leftDense[assignment] = left[x*order+z]
		rightDense[assignment] = right[z*order+y]
	}
	leftEdge, err := manager.FromDense(leftDense, ExtraBDDOptions{})
	if err != nil {
		return 0, 0, err
	}
	rightEdge, err := manager.FromDense(rightDense, ExtraBDDOptions{})
	return leftEdge, rightEdge, err
}

func cuddDenseArithmeticProduct(left, right []float64, order int) []float64 {
	result := make([]float64, order*order)
	for x := range order {
		for y := range order {
			for z := range order {
				result[x*order+y] += left[x*order+z] * right[z*order+y]
			}
		}
	}
	return result
}

func cuddDenseTriangle(left, right []float64, order int) []float64 {
	result := make([]float64, order*order)
	for x := range order {
		for y := range order {
			best := math.Inf(1)
			for z := range order {
				best = math.Min(best, left[x*order+z]+right[z*order+y])
			}
			result[x*order+y] = best
		}
	}
	return result
}

func cuddCheckMatrixResult(t *testing.T, manager *CuddADDManager, edge CuddADDEdge, want []float64, bits int) {
	t.Helper()
	order := 1 << uint(bits)
	for assignment := range 1 << uint(bits*3) {
		x := assignment & (order - 1)
		y := (assignment >> uint(2*bits)) & (order - 1)
		got, err := manager.Evaluate(edge, uint64(assignment))
		if err != nil || got != want[x*order+y] {
			t.Fatalf("assignment %d: got %g, %v want %g", assignment, got, err, want[x*order+y])
		}
	}
}

func TestCuddMatrixMultiplyAndTriangle(t *testing.T) {
	manager, err := NewCuddADDManager(3, 4096)
	if err != nil {
		t.Fatal(err)
	}
	left := []float64{1, 2, 3, 4}
	right := []float64{5, 6, 7, 8}
	a, b, err := cuddMatrixInputs(manager, left, right, 1)
	if err != nil {
		t.Fatal(err)
	}
	product, err := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	alternative, err := manager.TimesPlus(a, b, []int{1, 1}, ExtraBDDOptions{})
	if err != nil || alternative != product {
		t.Fatalf("TimesPlus = %d, %v; MatrixMultiply = %d", alternative, err, product)
	}
	cuddCheckMatrixResult(t, manager, product, cuddDenseArithmeticProduct(left, right, 2), 1)

	triangle, err := manager.Triangle(a, b, []int{1}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddCheckMatrixResult(t, manager, triangle, cuddDenseTriangle(left, right, 2), 1)
}

func TestCuddMatrixMissingVariableScaling(t *testing.T) {
	manager, _ := NewCuddADDManager(3, 256)
	two, _ := manager.Constant(2)
	three, _ := manager.Constant(3)
	product, err := manager.MatrixMultiply(two, three, []int{0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	value, err := manager.Evaluate(product, 0)
	if err != nil || value != 24 {
		t.Fatalf("missing-variable arithmetic product = %g, %v; want 24", value, err)
	}
	triangle, err := manager.Triangle(two, three, []int{0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	value, err = manager.Evaluate(triangle, 7)
	if err != nil || value != 5 {
		t.Fatalf("missing-variable triangle = %g, %v; want 5", value, err)
	}
}

func TestCuddMatrixOuterSum(t *testing.T) {
	manager, _ := NewCuddADDManager(2, 1024)
	matrix, err := manager.FromDense([]float64{9, 8, 7, 6}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	row, err := manager.FromDense([]float64{2, 4, 2, 4}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	column, err := manager.FromDense([]float64{3, 3, 5, 5}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := manager.OuterSum(matrix, row, column, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{5, 7, 7, 6}
	for assignment, expected := range want {
		got, evalErr := manager.Evaluate(result, uint64(assignment))
		if evalErr != nil || got != expected {
			t.Fatalf("OuterSum[%d] = %g, %v; want %g", assignment, got, evalErr, expected)
		}
	}

	inf, _ := manager.Constant(math.Inf(1))
	unchanged, err := manager.OuterSum(matrix, inf, column, ExtraBDDOptions{})
	if err != nil || unchanged != matrix {
		t.Fatalf("infinite outer sum = %d, %v; want matrix %d", unchanged, err, matrix)
	}
}

func TestCuddMatrixExhaustiveBinaryTwoByTwo(t *testing.T) {
	for leftBits := 0; leftBits < 16; leftBits++ {
		for rightBits := 0; rightBits < 16; rightBits++ {
			manager, err := NewCuddADDManager(3, 512)
			if err != nil {
				t.Fatal(err)
			}
			left := make([]float64, 4)
			right := make([]float64, 4)
			for index := range 4 {
				left[index] = float64(leftBits >> uint(index) & 1)
				right[index] = float64(rightBits >> uint(index) & 1)
			}
			a, b, err := cuddMatrixInputs(manager, left, right, 1)
			if err != nil {
				t.Fatal(err)
			}
			product, err := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{})
			if err != nil {
				t.Fatalf("matrices %d/%d: %v", leftBits, rightBits, err)
			}
			cuddCheckMatrixResult(t, manager, product, cuddDenseArithmeticProduct(left, right, 2), 1)
			triangle, err := manager.Triangle(a, b, []int{1}, ExtraBDDOptions{})
			if err != nil {
				t.Fatalf("triangle %d/%d: %v", leftBits, rightBits, err)
			}
			cuddCheckMatrixResult(t, manager, triangle, cuddDenseTriangle(left, right, 2), 1)
		}
	}
}

func TestCuddMatrixFourByFourModel(t *testing.T) {
	manager, _ := NewCuddADDManager(6, 1<<16)
	left := []float64{0, 2, -1, 3, 4, 1, 2, 0, -2, 3, 1, 5, 2, 0, 4, -1}
	right := []float64{3, 1, 0, 2, -1, 4, 2, 1, 5, 0, 3, -2, 1, 2, 4, 0}
	a, b, err := cuddMatrixInputs(manager, left, right, 2)
	if err != nil {
		t.Fatal(err)
	}
	product, err := manager.MatrixMultiply(a, b, []int{2, 3}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddCheckMatrixResult(t, manager, product, cuddDenseArithmeticProduct(left, right, 4), 2)
	triangle, err := manager.Triangle(a, b, []int{2, 3}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddCheckMatrixResult(t, manager, triangle, cuddDenseTriangle(left, right, 4), 2)
}

func TestCuddMatrixNoncontiguousAbstraction(t *testing.T) {
	manager, err := NewCuddADDManager(4, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	leftValues := make([]float64, 16)
	rightValues := make([]float64, 16)
	for assignment := range 16 {
		leftValues[assignment] = float64((assignment*5+3)%11 - 5)
		rightValues[assignment] = float64((assignment*7+1)%13 - 6)
	}
	left, err := manager.FromDense(leftValues, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	right, err := manager.FromDense(rightValues, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	product, err := manager.MatrixMultiply(left, right, []int{2, 0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	triangle, err := manager.Triangle(left, right, []int{0, 2}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := range 16 {
		base := assignment &^ (1 | 4)
		wantProduct := 0.0
		wantTriangle := math.Inf(1)
		for selected := range 4 {
			candidate := base | (selected & 1) | ((selected & 2) << 1)
			term := leftValues[candidate] * rightValues[candidate]
			wantProduct += term
			wantTriangle = math.Min(wantTriangle, leftValues[candidate]+rightValues[candidate])
		}
		gotProduct, evalErr := manager.Evaluate(product, uint64(assignment))
		if evalErr != nil || gotProduct != wantProduct {
			t.Fatalf("product[%d] = %g, %v; want %g", assignment, gotProduct, evalErr, wantProduct)
		}
		gotTriangle, evalErr := manager.Evaluate(triangle, uint64(assignment))
		if evalErr != nil || gotTriangle != wantTriangle {
			t.Fatalf("triangle[%d] = %g, %v; want %g", assignment, gotTriangle, evalErr, wantTriangle)
		}
	}
}

func TestCuddMatrixUndefinedInfinityCombinationRollsBack(t *testing.T) {
	manager, err := NewCuddADDManager(1, 128)
	if err != nil {
		t.Fatal(err)
	}
	left, err := manager.FromDense([]float64{math.Inf(1), 1}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	right, err := manager.FromDense([]float64{1, math.Inf(-1)}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	start := manager.NodeCount()
	if _, err := manager.MatrixMultiply(left, right, []int{0}, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("opposite-infinity sum error = %v; want %v", err, ErrDimension)
	}
	if manager.NodeCount() != start {
		t.Fatalf("opposite-infinity failure changed node count from %d to %d", start, manager.NodeCount())
	}
}

func TestCuddMatrixFailuresAreAtomic(t *testing.T) {
	manager, _ := NewCuddADDManager(3, 1024)
	a, b, err := cuddMatrixInputs(manager, []float64{1, 2, 3, 4}, []float64{5, 6, 7, 8}, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := manager.NodeCount()
	checks := []struct {
		name string
		run  func() error
		want error
	}{
		{"bad variable", func() error { _, e := manager.MatrixMultiply(a, b, []int{3}, ExtraBDDOptions{}); return e }, ErrDimension},
		{"bad edge", func() error {
			_, e := manager.MatrixMultiply(CuddADDEdge(1<<30), b, []int{1}, ExtraBDDOptions{})
			return e
		}, ErrInvalidLiteral},
		{"bad options", func() error { _, e := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{MaxBytes: -1}); return e }, ErrDimension},
		{"operation limit", func() error {
			_, e := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{MaxOperations: 1})
			return e
		}, ErrLimit},
		{"byte limit", func() error {
			_, e := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{MaxBytes: start*32 - 1})
			return e
		}, ErrLimit},
		{"temporary byte limit", func() error {
			_, e := manager.MatrixMultiply(a, b, nil, ExtraBDDOptions{MaxBytes: start*32 + manager.Variables()})
			return e
		}, ErrLimit},
		{"cancel", func() error {
			_, e := manager.Triangle(a, b, []int{1}, ExtraBDDOptions{Stop: func() bool { return true }})
			return e
		}, ErrCancelled},
	}
	for _, check := range checks {
		if err := check.run(); !errors.Is(err, check.want) {
			t.Fatalf("%s: got %v want %v", check.name, err, check.want)
		}
		if manager.NodeCount() != start {
			t.Fatalf("%s changed node count from %d to %d", check.name, start, manager.NodeCount())
		}
	}
	zero, _ := manager.Constant(0)
	inf, _ := manager.Constant(math.Inf(1))
	start = manager.NodeCount()
	product, err := manager.MatrixMultiply(zero, inf, nil, ExtraBDDOptions{})
	if err != nil || product != zero {
		t.Fatalf("0*Inf product = %d, %v; want zero", product, err)
	}
	if manager.NodeCount() != start {
		t.Fatal("absorbing numeric result created nodes")
	}
	negativeInf, _ := manager.Constant(math.Inf(-1))
	triangle, err := manager.Triangle(inf, negativeInf, nil, ExtraBDDOptions{})
	if err != nil || triangle != inf {
		t.Fatalf("Inf + -Inf distance = %d, %v; want positive infinity", triangle, err)
	}
}

func TestCuddMatrixClosedManager(t *testing.T) {
	manager, _ := NewCuddADDManager(0, 8)
	one, _ := manager.Constant(1)
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.MatrixMultiply(one, one, nil, ExtraBDDOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager error = %v", err)
	}
}

func TestCuddMatrixDeterministicModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state = state*6364136223846793005 + 1442695040888963407
		return state
	}
	for trial := 0; trial < 5000; trial++ {
		left := make([]float64, 4)
		right := make([]float64, 4)
		for index := range 4 {
			left[index] = float64(int(next()%9) - 4)
			right[index] = float64(int(next()%9) - 4)
		}
		manager, _ := NewCuddADDManager(3, 1024)
		a, b, err := cuddMatrixInputs(manager, left, right, 1)
		if err != nil {
			t.Fatalf("trial %d build: %v", trial, err)
		}
		product, err := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{})
		if err != nil {
			t.Fatalf("trial %d product: %v", trial, err)
		}
		cuddCheckMatrixResult(t, manager, product, cuddDenseArithmeticProduct(left, right, 2), 1)
		triangle, err := manager.Triangle(a, b, []int{1}, ExtraBDDOptions{})
		if err != nil {
			t.Fatalf("trial %d triangle: %v", trial, err)
		}
		cuddCheckMatrixResult(t, manager, triangle, cuddDenseTriangle(left, right, 2), 1)
	}
}

func FuzzCuddMatrixSemirings(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0, 0, 0, 0, 1, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 8 {
			t.Skip()
		}
		left := make([]float64, 4)
		right := make([]float64, 4)
		for index := range 4 {
			left[index] = float64(int8(data[index]) % 17)
			right[index] = float64(int8(data[index+4]) % 17)
		}
		manager, err := NewCuddADDManager(3, 1024)
		if err != nil {
			t.Fatal(err)
		}
		a, b, err := cuddMatrixInputs(manager, left, right, 1)
		if err != nil {
			t.Fatal(err)
		}
		product, err := manager.MatrixMultiply(a, b, []int{1}, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		cuddCheckMatrixResult(t, manager, product, cuddDenseArithmeticProduct(left, right, 2), 1)
		triangle, err := manager.Triangle(a, b, []int{1}, ExtraBDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		cuddCheckMatrixResult(t, manager, triangle, cuddDenseTriangle(left, right, 2), 1)
	})
}
