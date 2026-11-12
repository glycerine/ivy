package place

import (
	"math"
	"testing"
)

func TestPreplaceCoreAndFixedPads(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind, pad := NewCellType("gate", 2, 1, false), NewCellType("pad", 1, 1, true)
	fixed := NewCell(0, "fixed", kind, 10, 20)
	fixed.m_fixed = true
	AddCell(fixed)
	for i := int32(1); i < 8; i++ {
		AddCell(NewCell(i, "pad", pad, 0, 0))
	}
	if err := Preplace(.8); err != nil {
		t.Fatal(err)
	}
	if g_place_coreBounds.x > 9 || g_place_coreBounds.y > 19.5 || g_place_coreBounds.x+g_place_coreBounds.w < 11 || g_place_coreBounds.y+g_place_coreBounds.h < 20.5 {
		t.Fatal("fixed extent", g_place_coreBounds)
	}
	if fixed.m_x != 10 || fixed.m_y != 20 {
		t.Fatal("moved fixed cell")
	}
	for _, c := range g_place_concreteCells {
		if c != nil && c.m_parent.m_pad && (!qpFinite(c.m_x) || !qpFinite(c.m_y)) {
			t.Fatal("invalid pad position")
		}
	}
	if Preplace(0) == nil {
		t.Fatal("accepted zero utilization")
	}
}
func TestDensityLimitsAndDegenerateCoordinates(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	g_place_coreBounds = Rect{0, 0, 100, 20}
	kind := NewCellType("gate", 1, 1, false)
	for i := int32(0); i < 20; i++ {
		AddCell(NewCell(i, "cell", kind, 10, 10))
	}
	fixed := NewCell(25, "fixed", kind, 30, 10)
	fixed.m_fixed = true
	AddCell(fixed)
	if err := FixDensity(5, 2); err != nil {
		t.Fatal(err)
	}
	for _, c := range g_place_concreteCells {
		if c == nil || c.m_fixed {
			continue
		}
		if math.Abs(float64(c.m_x-10)) > 2.0001 || !qpFinite(c.m_x) {
			t.Fatal("movement", c.m_x)
		}
	}
	if fixed.m_x != 30 {
		t.Fatal("moved fixed")
	}
	if FixDensity(0, 1) == nil || FixDensity(1, -1) == nil {
		t.Fatal("accepted invalid limits")
	}
}
func TestLegalizeObstaclesAndTransactionalFailure(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	g_place_coreBounds = Rect{0, 0, 8, 2}
	g_place_rowHeight = 1
	kind := NewCellType("gate", 2, 1, false)
	obstacle := NewCell(0, "fixed", kind, 1, .5)
	obstacle.m_fixed = true
	AddCell(obstacle)
	for i := int32(1); i < 5; i++ {
		AddCell(NewCell(i, "cell", kind, 1, .5))
	}
	if err := Legalize(); err != nil {
		t.Fatal(err)
	}
	for i, a := range g_place_concreteCells {
		if a == nil {
			continue
		}
		for _, b := range g_place_concreteCells[i+1:] {
			if b != nil && math.Abs(float64(a.m_x-b.m_x)) < 2-1e-6 && math.Abs(float64(a.m_y-b.m_y)) < 1-1e-6 {
				t.Fatal("overlap", a.m_id, b.m_id)
			}
		}
	}
	before := g_place_concreteCells[1].m_x
	AddCell(NewCell(10, "huge", NewCellType("huge", 20, 1, false), 0, 0))
	if Legalize() == nil {
		t.Fatal("accepted oversized cell")
	}
	if g_place_concreteCells[1].m_x != before {
		t.Fatal("partial commit")
	}
}
func TestSanitizeTopBoundaryAndCOGLeaves(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	g_place_coreBounds = Rect{10, 100, 20, 10}
	g_place_rowHeight = 1
	kind := NewCellType("gate", 2, 2, false)
	a := NewCell(0, "a", kind, 15, 1000)
	b := NewCell(1, "b", kind, 25, 105)
	AddCell(a)
	AddCell(b)
	SanitizePlacement()
	if a.m_y < 101 || a.m_y > 109 {
		t.Fatal(a.m_y)
	}
	left := &Partition{m_leaf: true, m_members: []*ConcreteCell{a}, m_bounds: Rect{10, 100, 10, 10}}
	right := &Partition{m_leaf: true, m_members: []*ConcreteCell{b}, m_bounds: Rect{20, 100, 10, 10}}
	g_place_rootPartition = &Partition{m_sub1: left, m_sub2: right}
	g_place_qpProb = nil
	if err := constructQuadraticProblem(); err != nil {
		t.Fatal(err)
	}
	if generateCoGConstraints(nil) != 2 {
		t.Fatal("discarded final COG leaf")
	}
	if err := solveQuadraticProblem(true); err != nil {
		t.Fatal(err)
	}
	if a.m_x != 15 || b.m_x != 25 || a.m_y != 105 || b.m_y != 105 {
		t.Fatal("constraint centers", a, b)
	}
}

func TestQuadraticConnectivitySparseAndIncremental(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	g_place_coreBounds = Rect{0, 0, 20, 10}
	kind := NewCellType("gate", 1, 1, false)
	a, b := NewCell(2, "a", kind, 1, 1), NewCell(7, "b", kind, 19, 9)
	b.m_fixed = true
	AddCell(a)
	AddCell(b)
	AddNet(NewNet(0, []*ConcreteCell{a, b}, 2))
	AddNet(NewNet(1, []*ConcreteCell{a, b}, 3))
	if err := constructQuadraticProblem(); err != nil {
		t.Fatal(err)
	}
	if err := solveQuadraticProblem(false); err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(a.m_x-19)) > 1e-4 || math.Abs(float64(a.m_y-9)) > 1e-4 {
		t.Fatal("anchored clique solve", a)
	}
	if g_place_qpProb.fixed[0] != 1 {
		t.Fatal("sparse hole not fixed")
	}
	g_place_qpProb.x[2] = 13
	g_place_qpProb.y[2] = 4
	c := NewCell(10, "new", kind, 0, 0)
	AddCell(c)
	if err := constructQuadraticProblem(); err != nil {
		t.Fatal(err)
	}
	if g_place_qpProb.x[2] != 13 || g_place_qpProb.y[2] != 4 || g_place_qpProb.x[10] != 10 || g_place_qpProb.y[10] != 5 {
		t.Fatal("incremental initialization")
	}
}
func TestGeometryRejectsNonfiniteInput(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	c := NewCell(0, "bad", NewCellType("gate", 1, 1, false), float32(math.NaN()), 0)
	AddCell(c)
	if Preplace(.8) == nil || FixDensity(1, 1) == nil || Legalize() == nil {
		t.Fatal("accepted nonfinite cell position")
	}
}
