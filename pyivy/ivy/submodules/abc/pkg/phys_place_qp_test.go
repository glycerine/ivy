package abc

// Tests for src/phys/place/place_qpsolver.h and place_qpsolver.c.

import (
	"errors"
	"math"
	"math/rand"
	"testing"
)

func qpsClose(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

func qpsWirelength(problem QPSProblem, positions []QPSPoint) float64 {
	value := 0.0
	for _, edge := range problem.Edges {
		value += edge.Weight * qpsDistanceSquared(positions[edge.CellA], positions[edge.CellB])
	}
	return value
}

func TestQPSWeightedPlacementAndLifecycle(t *testing.T) {
	problem := QPSProblem{
		Positions: []QPSPoint{{X: 0, Y: 2}, {X: 10, Y: 6}, {X: -40, Y: 100}},
		Fixed:     []bool{true, true, false},
		Edges:     []QPSEdge{{CellA: 2, CellB: 0, Weight: 1}, {CellA: 2, CellB: 1, Weight: 3}},
	}
	initial := qpsWirelength(problem, problem.Positions)
	solver, err := NewQPSSolver(problem, QPSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve()
	if err != nil {
		t.Fatal(err)
	}
	if !qpsClose(result.Positions[2].X, 7.5, 1e-8) || !qpsClose(result.Positions[2].Y, 5, 1e-8) {
		t.Fatalf("weighted optimum = %+v", result.Positions[2])
	}
	if result.Wirelength > initial || !result.BoundsSatisfied || !result.LoopsSatisfied {
		t.Fatalf("bad result %+v initial=%g", result, initial)
	}
	if problem.Positions[2].X != -40 || problem.Positions[2].Y != 100 {
		t.Fatal("solver mutated caller-owned input")
	}
	second, err := solver.Solve()
	if err != nil || second.Positions[2] != result.Positions[2] {
		t.Fatalf("repeat solve = %+v, %v", second, err)
	}
	if err := solver.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := solver.Solve(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("solve after close error = %v", err)
	}
	if err := solver.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("second close error = %v", err)
	}
}

func TestQPSCenterConstraint(t *testing.T) {
	problem := QPSProblem{
		Positions: []QPSPoint{{}, {X: 10}, {X: 2}, {X: 3}},
		Fixed:     []bool{true, true, false, false},
		Edges:     []QPSEdge{{CellA: 0, CellB: 2, Weight: 1}, {CellA: 1, CellB: 3, Weight: 1}},
		Areas:     []float64{1, 1, 1, 3},
		Centers:   []QPSCenterConstraint{{Cells: []int{2, 3}, Center: QPSPoint{X: 4, Y: 2}}},
	}
	result, err := SolveQPS(problem, QPSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !qpsClose(result.Positions[2].X, -1.4, 1e-7) || !qpsClose(result.Positions[3].X, 5.8, 1e-7) {
		t.Fatalf("center optimum = %+v", result.Positions)
	}
	x := result.Positions[2].X + 3*result.Positions[3].X
	y := result.Positions[2].Y + 3*result.Positions[3].Y
	if !qpsClose(x/4, 4, 1e-10) || !qpsClose(y/4, 2, 1e-10) {
		t.Fatalf("center not preserved: %g %g", x/4, y/4)
	}
}

func TestQPSBoundsAndLoopConstraint(t *testing.T) {
	bounded := QPSProblem{
		Positions: []QPSPoint{{X: 2, Y: 0.5}, {X: 10, Y: 2}},
		Fixed:     []bool{true, false},
		Edges:     []QPSEdge{{CellA: 0, CellB: 1, Weight: 1}},
		MaxX:      3,
		MaxY:      1,
		UseBounds: true,
	}
	result, err := SolveQPS(bounded, QPSOptions{MaxOuterIterations: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if !result.BoundsSatisfied || result.Positions[1].X < -2e-3 || result.Positions[1].X > 3.002 || result.Positions[1].Y < -2e-3 || result.Positions[1].Y > 1.002 {
		t.Fatalf("bounds result = %+v", result)
	}
	if !qpsClose(result.Positions[1].X, 2, 3e-3) || !qpsClose(result.Positions[1].Y, 0.5, 3e-3) {
		t.Fatalf("bounded optimum = %+v", result.Positions[1])
	}

	looped := QPSProblem{
		Positions: []QPSPoint{{X: 0}, {X: 10}, {X: 5}},
		Fixed:     []bool{true, true, false},
		Edges:     []QPSEdge{{CellA: 1, CellB: 2, Weight: 1}},
		Loops:     []QPSLoopConstraint{{Cells: []int{0, 2}, MaxSquaredLength: 8}},
	}
	result, err = SolveQPS(looped, QPSOptions{MaxOuterIterations: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if !result.LoopsSatisfied || result.LoopResiduals[0] > 2e-2 || !qpsClose(result.Positions[2].X, 2, 5e-3) {
		t.Fatalf("loop result = %+v", result)
	}
}

func qpsDenseReference(problem QPSProblem) []QPSPoint {
	index := make([]int, len(problem.Positions))
	for i := range index {
		index[i] = -1
	}
	floating := 0
	for i := range index {
		if !problem.Fixed[i] {
			index[i] = floating
			floating++
		}
	}
	matrix := make([][]float64, floating)
	rightX := make([]float64, floating)
	rightY := make([]float64, floating)
	for i := range matrix {
		matrix[i] = make([]float64, floating)
	}
	for _, edge := range problem.Edges {
		a, b, weight := edge.CellA, edge.CellB, edge.Weight
		if index[a] >= 0 {
			matrix[index[a]][index[a]] += weight
			if index[b] >= 0 {
				matrix[index[a]][index[b]] -= weight
			} else {
				rightX[index[a]] += weight * problem.Positions[b].X
				rightY[index[a]] += weight * problem.Positions[b].Y
			}
		}
		if index[b] >= 0 {
			matrix[index[b]][index[b]] += weight
			if index[a] >= 0 {
				matrix[index[b]][index[a]] -= weight
			} else {
				rightX[index[b]] += weight * problem.Positions[a].X
				rightY[index[b]] += weight * problem.Positions[a].Y
			}
		}
	}
	solve := func(right []float64) []float64 {
		a := make([][]float64, floating)
		b := append([]float64(nil), right...)
		for i := range a {
			a[i] = append([]float64(nil), matrix[i]...)
		}
		for col := 0; col < floating; col++ {
			pivot := col
			for row := col + 1; row < floating; row++ {
				if math.Abs(a[row][col]) > math.Abs(a[pivot][col]) {
					pivot = row
				}
			}
			a[col], a[pivot] = a[pivot], a[col]
			b[col], b[pivot] = b[pivot], b[col]
			for row := col + 1; row < floating; row++ {
				factor := a[row][col] / a[col][col]
				for k := col; k < floating; k++ {
					a[row][k] -= factor * a[col][k]
				}
				b[row] -= factor * b[col]
			}
		}
		result := make([]float64, floating)
		for row := floating - 1; row >= 0; row-- {
			value := b[row]
			for k := row + 1; k < floating; k++ {
				value -= a[row][k] * result[k]
			}
			result[row] = value / a[row][row]
		}
		return result
	}
	x, y := solve(rightX), solve(rightY)
	result := append([]QPSPoint(nil), problem.Positions...)
	for cell, variable := range index {
		if variable >= 0 {
			result[cell] = QPSPoint{X: x[variable], Y: y[variable]}
		}
	}
	return result
}

func TestQPSIndependentDenseModel5000(t *testing.T) {
	random := rand.New(rand.NewSource(0x515053))
	for trial := 0; trial < 5000; trial++ {
		floating := 1 + random.Intn(4)
		problem := QPSProblem{Positions: make([]QPSPoint, floating+2), Fixed: make([]bool, floating+2)}
		problem.Fixed[0], problem.Fixed[1] = true, true
		problem.Positions[0] = QPSPoint{X: float64(random.Intn(21) - 10), Y: float64(random.Intn(21) - 10)}
		problem.Positions[1] = QPSPoint{X: float64(random.Intn(21) - 10), Y: float64(random.Intn(21) - 10)}
		for cell := 2; cell < floating+2; cell++ {
			problem.Positions[cell] = QPSPoint{X: float64(random.Intn(21) - 10), Y: float64(random.Intn(21) - 10)}
			problem.Edges = append(problem.Edges,
				QPSEdge{CellA: 0, CellB: cell, Weight: float64(1 + random.Intn(5))},
				QPSEdge{CellA: 1, CellB: cell, Weight: float64(1 + random.Intn(5))})
			if cell > 2 {
				problem.Edges = append(problem.Edges, QPSEdge{CellA: cell - 1, CellB: cell, Weight: float64(1 + random.Intn(5))})
			}
		}
		want := qpsDenseReference(problem)
		got, err := SolveQPS(problem, QPSOptions{Tolerance: 1e-8})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		for cell := 2; cell < floating+2; cell++ {
			if !qpsClose(got.Positions[cell].X, want[cell].X, 2e-4) || !qpsClose(got.Positions[cell].Y, want[cell].Y, 2e-4) {
				t.Fatalf("trial %d cell %d: got %+v want %+v", trial, cell, got.Positions[cell], want[cell])
			}
		}
	}
}

func TestQPSErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	valid := QPSProblem{Positions: []QPSPoint{{}, {X: 1}}, Fixed: []bool{true, false}, Edges: []QPSEdge{{CellA: 0, CellB: 1, Weight: 1}}}
	tests := []QPSProblem{
		{},
		{Positions: []QPSPoint{{}}, Fixed: nil},
		{Positions: []QPSPoint{{}, {}}, Fixed: []bool{true, false}, Edges: []QPSEdge{{CellA: 0, CellB: 0, Weight: 1}}},
		{Positions: []QPSPoint{{}, {}}, Fixed: []bool{true, false}, Edges: []QPSEdge{{CellA: 0, CellB: 1, Weight: -1}}},
		{Positions: []QPSPoint{{}, {}}, Fixed: []bool{true, false}, Areas: []float64{1, 1}, Centers: []QPSCenterConstraint{{Cells: []int{0}, Center: QPSPoint{}}}},
		{Positions: []QPSPoint{{}, {}}, Fixed: []bool{true, false}, Loops: []QPSLoopConstraint{{Cells: []int{0, 0}, MaxSquaredLength: 1}}},
	}
	for i, problem := range tests {
		if _, err := NewQPSSolver(problem, QPSOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("case %d error = %v", i, err)
		}
	}
	if _, err := NewQPSSolver(valid, QPSOptions{MaxCells: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("cell limit error = %v", err)
	}
	if _, err := NewQPSSolver(valid, QPSOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel error = %v", err)
	}
	original := append([]QPSPoint(nil), valid.Positions...)
	if _, err := SolveQPS(valid, QPSOptions{MaxInnerIterations: 1, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("solve cancellation = %v", err)
	}
	for i := range original {
		if valid.Positions[i] != original[i] {
			t.Fatal("failed solve mutated input")
		}
	}
}

func FuzzQPSDenseModel(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{255, 0, 17, 42, 91})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		floating := 1 + int(data[0]%4)
		problem := QPSProblem{Positions: make([]QPSPoint, floating+2), Fixed: make([]bool, floating+2)}
		problem.Fixed[0], problem.Fixed[1] = true, true
		at := 1
		next := func() byte {
			if at >= len(data) {
				return byte(at*37 + 11)
			}
			value := data[at]
			at++
			return value
		}
		for i := range problem.Positions {
			problem.Positions[i] = QPSPoint{X: float64(int(next()%41) - 20), Y: float64(int(next()%41) - 20)}
		}
		for cell := 2; cell < len(problem.Positions); cell++ {
			problem.Edges = append(problem.Edges,
				QPSEdge{CellA: 0, CellB: cell, Weight: float64(1 + next()%9)},
				QPSEdge{CellA: 1, CellB: cell, Weight: float64(1 + next()%9)})
			if cell > 2 && next()&1 != 0 {
				problem.Edges = append(problem.Edges, QPSEdge{CellA: cell - 1, CellB: cell, Weight: float64(1 + next()%9)})
			}
		}
		want := qpsDenseReference(problem)
		got, err := SolveQPS(problem, QPSOptions{Tolerance: 1e-8})
		if err != nil {
			t.Fatal(err)
		}
		if got.Wirelength > qpsWirelength(problem, problem.Positions)+1e-7 {
			t.Fatal("objective increased")
		}
		for cell := 2; cell < len(problem.Positions); cell++ {
			if !qpsClose(got.Positions[cell].X, want[cell].X, 3e-4) || !qpsClose(got.Positions[cell].Y, want[cell].Y, 3e-4) {
				t.Fatalf("cell %d: got %+v want %+v", cell, got.Positions[cell], want[cell])
			}
		}
	})
}
