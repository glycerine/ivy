package place

import (
	"math"
	"testing"
)

func closeQP(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 2e-3*math.Max(1, math.Abs(float64(want))) {
		t.Fatalf("got %g want %g", got, want)
	}
}
func TestQPChainAndIsolatedCells(t *testing.T) {
	p, err := NewQPProblem([]float32{0, 9, 10, 7}, []float32{0, 1, 8, 9}, []bool{true, false, true, false}, []QPEdge{{0, 1, 1}, {1, 2, 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Solve(); err != nil {
		t.Fatal(err)
	}
	x, y := p.Coordinates()
	closeQP(t, x[1], 5)
	closeQP(t, y[1], 4)
	closeQP(t, x[3], 7)
	closeQP(t, y[3], 9)
	closeQP(t, p.Objective(), 82)
	if err = p.Solve(); err != nil {
		t.Fatal("repeat", err)
	}
}
func TestQPWeightedCOG(t *testing.T) {
	p, _ := NewQPProblem([]float32{1, 9}, []float32{2, 4}, []bool{false, false}, []QPEdge{{0, 1, 1}})
	if err := p.SetAreas([]float32{1, 3}); err != nil {
		t.Fatal(err)
	}
	if err := p.AddCenterOfGravity([]int32{0, 1}, 4, 7); err != nil {
		t.Fatal(err)
	}
	if err := p.Solve(); err != nil {
		t.Fatal(err)
	}
	x, y := p.Coordinates()
	for i := range x {
		closeQP(t, x[i], 4)
		closeQP(t, y[i], 7)
	}
	closeQP(t, p.Objective(), 0)
}
func TestQPAllFixedAndDependent(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 3}, []float32{0, 4}, []bool{true, true}, []QPEdge{{0, 1, 2}})
	if err := p.Solve(); err != nil {
		t.Fatal(err)
	}
	closeQP(t, p.Objective(), 50)
	p, _ = NewQPProblem([]float32{0, 10}, []float32{0, 0}, []bool{false, true}, []QPEdge{{0, 1, 1}})
	p.SetAreas([]float32{1, 2})
	p.AddCenterOfGravity([]int32{0, 1}, 8, 0)
	if err := p.Solve(); err != nil {
		t.Fatal(err)
	}
	x, _ := p.Coordinates()
	closeQP(t, x[0], 4)
}
func TestQPActiveBounds(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 8, 0}, []float32{0, 0, 0}, []bool{false, false, true}, []QPEdge{{0, 2, 100}, {1, 2, 1}})
	p.SetAreas([]float32{1, 1, 1})
	p.AddCenterOfGravity([]int32{0, 1}, 4, 0)
	p.SetBounds(5, 5)
	if err := p.Solve(); err != nil {
		t.Fatal(err, p.x, p.loop_fail)
	}
	x, _ := p.Coordinates()
	closeQP(t, x[0], 3)
	closeQP(t, x[1], 5)
}
func TestQPLoopConstraint(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 10, 0, 10}, []float32{0, 0, 0, 0}, []bool{false, false, true, true}, []QPEdge{{0, 2, 1}, {1, 3, 1}})
	p.AddLoop([]int32{0, 1}, 8, 0)
	if err := p.Solve(); err != nil {
		t.Fatal(err, p.x, p.loop_penalty)
	}
	x, _ := p.Coordinates()
	closeQP(t, x[0], 4)
	closeQP(t, x[1], 6)
}
func TestQPGradientFiniteDifference(t *testing.T) {
	p, _ := NewQPProblem([]float32{2, 8, 0}, []float32{3, 5, 0}, []bool{false, false, true}, []QPEdge{{0, 2, 2}, {1, 2, 3}, {0, 1, 1}})
	p.SetAreas([]float32{1, 2, 1})
	p.AddCenterOfGravity([]int32{0, 1}, 4, 3)
	p.AddLoop([]int32{0, 1, 2}, 100, 2)
	qps_init(p)
	qpPrepareVariables(p)
	qps_settp(p)
	gradient := make([]float32, p.priv_n)
	qps_dfunc(p, gradient)
	for i := range gradient {
		old := p.priv_cp[i]
		const step float32 = .01
		p.priv_cp[i] = old + step
		qps_settp(p)
		hi := qps_func(p)
		p.priv_cp[i] = old - step
		qps_settp(p)
		lo := qps_func(p)
		p.priv_cp[i] = old
		got := (hi - lo) / (2 * step)
		if math.Abs(float64(got-gradient[i])) > .1 {
			t.Fatal(i, got, gradient[i])
		}
	}
	qps_clean(p)
}

func TestQPEstimatePreservesCurrentResiduals(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 3, 10}, []float32{0, 0, 0}, []bool{false, false, true}, []QPEdge{{0, 2, 1}, {1, 2, 1}})
	p.AddLoop([]int32{0, 1}, 2, 0)
	qps_init(p)
	p.max_enable = 1
	p.max_x, p.max_y = 10, 10
	qpPrepareVariables(p)
	qps_settp(p)
	p.priv_f = qps_func(p)
	residual, objective := p.priv_lt[0], p.f
	qps_estopt(p)
	if p.priv_lt[0] != residual || p.f != objective {
		t.Fatal("estimate changed current values", p.priv_lt[0], residual, p.f, objective)
	}
	qps_clean(p)
}
func TestQPMultipleLoopConstraints(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 10, 20, 0, 10, 20}, make([]float32, 6), []bool{false, false, false, true, true, true}, []QPEdge{{0, 3, 1}, {1, 4, 1}, {2, 5, 1}})
	p.AddLoop([]int32{0, 1}, 8, 0)
	p.AddLoop([]int32{1, 2}, 8, 0)
	if err := p.Solve(); err != nil {
		t.Fatal(err, p.x, p.loop_penalty)
	}
	x, _ := p.Coordinates()
	closeQP(t, x[0], 8)
	closeQP(t, x[1], 10)
	closeQP(t, x[2], 12)
}
func TestQPInvalidAndInfeasibleInputs(t *testing.T) {
	p, _ := NewQPProblem([]float32{0, 10}, []float32{0, 0}, []bool{true, true}, []QPEdge{{0, 1, 1}})
	p.AddLoop([]int32{0, 1}, 8, 0)
	if err := p.Solve(); err == nil {
		t.Fatal("infeasible fixed loop accepted")
	}
	x, _ := p.Coordinates()
	if x[0] != 0 || x[1] != 10 {
		t.Fatal("fixed cells moved")
	}
	p, _ = NewQPProblem([]float32{0, 1}, []float32{0, 1}, []bool{false, false}, nil)
	p.SetAreas([]float32{1, 1})
	p.AddCenterOfGravity([]int32{0, 1}, 1, 1)
	p.AddCenterOfGravity([]int32{0}, 1, 1)
	if err := p.Solve(); err == nil {
		t.Fatal("overlapping COGs accepted")
	}
	empty, err := NewQPProblem(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = empty.Solve(); err != nil || empty.Objective() != 0 {
		t.Fatal("empty problem", err)
	}
}

func TestQPBoundsWithTranslationNullspace(t *testing.T) {
	p, _ := NewQPProblem([]float32{20, 30}, []float32{-10, -20}, []bool{false, false}, []QPEdge{{0, 1, 1}})
	p.SetBounds(5, 5)
	if err := p.Solve(); err != nil {
		t.Fatal(err)
	}
	x, y := p.Coordinates()
	for i := range x {
		if x[i] < 0 || x[i] > 5 || y[i] < 0 || y[i] > 5 {
			t.Fatal("outside bounds", x, y)
		}
	}
	closeQP(t, p.Objective(), 0)
}
