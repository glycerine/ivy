package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func simpleAndWindow() (*Network, *vec.Vec_Int_t, *vec.Vec_Int_t, *vec.Vec_Int_t, int32) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	pivot := graph.AND(a, b)
	graph.AddOutput(pivot)
	network := NewNetwork(graph)
	mirrors := vec.Vec_IntStartFull(len(graph.Nodes()))
	mapping := vec.Vec_IntStartFull(len(graph.Nodes()))
	window := sbdVector(0, int32(a.NodeID()), int32(b.NodeID()), int32(pivot.NodeID()))
	for i, id := range vec.Vec_IntArray(window) {
		vec.Vec_IntWriteEntry(mapping, int(id), int32(i))
	}
	return network, mirrors, window, mapping, int32(pivot.NodeID())
}
func TestWindowTruthExtractionAndCounterexample(t *testing.T) {
	network, mirrors, window, mapping, pivot := simpleAndWindow()
	empty := sbdVector()
	build := func() *bsat.Solver {
		return Sbd_ManSatSolver(nil, network, mirrors, pivot, window, mapping, empty, empty, 0)
	}
	divVars, values, temp := sbdVector(1, 2), sbdVector(0, 0), sbdVector()
	truth := Sbd_ManSolve(build(), 3, 4, sbdVector(1, 2), divVars, values, temp)
	if truth != 0x8888888888888888 {
		t.Fatalf("AND truth%x", truth)
	}
	result := Sbd_ManSolve(build(), 3, 4, sbdVector(1), divVars, values, temp)
	if result != SBD_SAT_SAT || vec.Vec_IntEntry(values, 0) != 15 || vec.Vec_IntEntry(values, 1) != 14 {
		t.Fatalf("counterexample%x %v", result, vec.Vec_IntArray(values))
	}
	sop := sbdVector()
	if Sbd_ManSolve2(build(), 3, 4, divVars, values, temp, sop) != 1 {
		t.Fatal("SOP extraction failed")
	}
	var inds [64]int32
	support, count := Sbd_ManSolverSupp(sop, inds[:])
	if support != 3 || count != 2 {
		t.Fatalf("SOP support%x count%d", support, count)
	}
}
func TestWindowObservabilityAndQbfReplacement(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	pivot := graph.AND(a, b)
	root := graph.AND(pivot, c)
	graph.AddOutput(root)
	network := NewNetwork(graph)
	mirrors := vec.Vec_IntStartFull(len(graph.Nodes()))
	window := sbdVector(0, int32(a.NodeID()), int32(b.NodeID()), int32(c.NodeID()), int32(pivot.NodeID()), int32(root.NodeID()))
	mapping := vec.Vec_IntStartFull(len(graph.Nodes()))
	for i, id := range vec.Vec_IntArray(window) {
		vec.Vec_IntWriteEntry(mapping, int(id), int32(i))
	}
	tfo, roots := sbdVector(int32(root.NodeID())), sbdVector(int32(root.NodeID()))
	solver := Sbd_ManSatSolver(nil, network, mirrors, int32(pivot.NodeID()), window, mapping, tfo, roots, 0)
	if solver == nil {
		t.Fatal("observable window considered empty")
	}
	if solver.Solve(bsat.Negative(3)) != bsat.Unsatisfiable || solver.Solve(bsat.Positive(3)) != bsat.Satisfiable {
		t.Fatal("root observability miter")
	}
	qbf := Sbd_ManSatSolver(nil, network, mirrors, int32(pivot.NodeID()), window, mapping, tfo, roots, 1)
	for assignment := int32(0); assignment < 16; assignment++ {
		assumptions := make([]bsat.Literal, 4)
		for i, v := range []int32{1, 2, 3, 8} {
			assumptions[i] = bsat.Literal(2*v + 1 - (assignment >> i & 1))
		}
		status := qbf.Solve(assumptions...)
		want := assignment>>2&1 != 0 && ((assignment&1)&(assignment>>1&1)) != (assignment>>3&1)
		if (status == bsat.Satisfiable) != want {
			t.Fatalf("QBF replacement assignment%x", assignment)
		}
	}
}
func TestCareSamplingProducesActualModels(t *testing.T) {
	network, mirrors, window, mapping, pivot := simpleAndWindow()
	solver := Sbd_ManSatSolver(nil, network, mirrors, pivot, window, mapping, sbdVector(), sbdVector(), 0)
	on, off := make([]uint64, 8), make([]uint64, 8)
	if status := Sbd_ManCollectConstantsNew(solver, sbdVector(1, 2), 8, 3, on, off); status != -1 {
		t.Fatalf("nonconstant status%d", status)
	}
	for _, value := range on {
		if value&3 != 3 || value&4 == 0 {
			t.Fatal("onset care point")
		}
	}
	for _, value := range off {
		if value&3 == 3 || value&4 != 0 {
			t.Fatal("offset care point")
		}
	}
	constant := bsat.NewSolver()
	sbdEnsureVariables(constant, 4)
	constant.AddClause(bsat.Negative(3))
	if status := Sbd_ManCollectConstantsNew(constant, sbdVector(1, 2), 1, 3, make([]uint64, 1), make([]uint64, 1)); status != 0 {
		t.Fatalf("constant false status%d", status)
	}
}

func TestCoverColumnsAreRemappedToSatDivisors(t *testing.T) {
	network, mirrors, window, mapping, pivot := simpleAndWindow()
	empty := sbdVector()
	p := &Sbd_Man_t{pPars: &Parameters{LUTSize: 2}, pGia: network, vMirrors: mirrors, vWinObjs: window, vObj2Var: mapping, vTfo: empty, vRoots: empty, vDivSet: sbdVector(), vDivVars: sbdVector(2, 1), vDivValues: sbdVector(0, 0), vLits: sbdVector()}
	p.pSat = Sbd_ManSatSolver(nil, network, mirrors, pivot, window, mapping, empty, empty, 0)
	columns := []uint64{1, 2, 3}
	var truth uint64
	if sbdRefineCover(p, pivot, columns, 2, &truth) != 1 || truth != 0x8888888888888888 {
		t.Fatalf("mapped divisor truth%x", truth)
	}
	if vec.Vec_IntEntry(p.vDivSet, 0) != 2 || vec.Vec_IntEntry(p.vDivSet, 1) != 1 {
		t.Fatal("column positions used as SAT variables")
	}
}
