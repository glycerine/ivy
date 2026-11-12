package place

import "testing"

func resetPlacementTest() {
	g_place_numCells = 0
	g_place_numNets = 0
	g_place_concreteCells = nil
	g_place_concreteNets = nil
	g_place_concreteCellsSize = 0
	g_place_concreteNetsSize = 0
	g_place_rootPartition = nil
}
func TestPlacementGeometryAndEstimate(t *testing.T) {
	resetPlacementTest()
	kind := NewCellType("gate", 2, 3, false)
	a, b := NewCell(0, "a", kind, 10, 20), NewCell(1, "b", kind, 15, 27)
	net := NewNet(0, []*ConcreteCell{a, b}, 1)
	box := NetBounds(net)
	if box != (Rect{x: 10, y: 20, w: 5, h: 7}) || NetWirelength(net) != 12 || CellArea(a) != 6 {
		t.Fatal(box)
	}
	probe := NewCell(2, "probe", kind, 8, 17)
	if got := EstimateWirelengthIncrease(probe, []*ConcreteNet{net}); got != 5 {
		t.Fatal("x/y boundary distances", got)
	}
	probe.SetPosition(12, 22)
	if EstimateWirelengthIncrease(probe, []*ConcreteNet{net}) != 0 {
		t.Fatal("inside bounding box")
	}
	empty := NewNet(3, nil, 1)
	if NetWirelength(empty) != 0 || EstimateWirelengthIncrease(probe, []*ConcreteNet{empty}) != 0 {
		t.Fatal("empty net geometry")
	}
}
func TestPlacementSparseDatabaseDeletion(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind := NewCellType("gate", 1, 1, false)
	a, b := NewCell(3, "a", kind, 0, 0), NewCell(100, "b", kind, 4, 7)
	AddCell(a)
	AddCell(b)
	n := NewNet(20, []*ConcreteCell{a, b}, 1)
	AddNet(n)
	if g_place_numCells != 101 || g_place_numNets != 21 || TotalWirelength() != 11 {
		t.Fatal("sparse insertion")
	}
	DeleteNet(n)
	DeleteCell(b)
	if g_place_numCells != 4 || g_place_numNets != 0 {
		t.Fatal("tail trimming")
	}
	DeleteCell(a)
	if g_place_numCells != 0 {
		t.Fatal("last cell deletion")
	}
}
func TestPlacementPartitionDeletion(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind := NewCellType("gate", 2, 3, false)
	a, b := NewCell(0, "a", kind, 0, 0), NewCell(1, "b", kind, 0, 0)
	left := &Partition{m_leaf: true, m_members: []*ConcreteCell{a}, m_numMembers: 1, m_area: 6}
	right := &Partition{m_leaf: true, m_members: []*ConcreteCell{b}, m_numMembers: 1, m_area: 6}
	root := &Partition{m_members: []*ConcreteCell{a, b}, m_numMembers: 2, m_area: 12, m_sub1: left, m_sub2: right}
	g_place_rootPartition = root
	AddCell(a)
	AddCell(b)
	DeleteCell(a)
	if root.m_area != 6 || left.m_area != 0 || right.m_area != 6 || root.m_members[0] != nil {
		t.Fatal("partition ownership")
	}
}
func TestPlacementComparators(t *testing.T) {
	kind := NewCellType("t", 1, 1, false)
	a, b := NewCell(1, "a", kind, 2, 9), NewCell(2, "b", kind, 3, 5)
	x, y := NewNet(1, []*ConcreteCell{a}, 1), NewNet(2, []*ConcreteCell{b}, 1)
	for _, compare := range []func(*ConcreteNet, *ConcreteNet) int32{netSortByL, netSortByR, netSortByB, netSortByT, netSortByID} {
		if compare(nil, nil) != 0 || compare(nil, x) != -1 || compare(x, nil) != 1 {
			t.Fatal("net nil order")
		}
	}
	if cellSortByX(a, b) != -1 || cellSortByY(a, b) != 1 || cellSortByID(a, b) != -1 || netSortByB(x, y) != 1 {
		t.Fatal("coordinate ordering")
	}
	if sqHashId(3, 7) != 4 {
		t.Fatal("hash")
	}
}

func TestPlacementInputOwnershipAndStaleDelete(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind := NewCellType("type", 1, 1, false)
	a, b := NewCell(0, "a", kind, 0, 0), NewCell(0, "replacement", kind, 1, 1)
	terms := []*ConcreteCell{a}
	net := NewNet(0, terms, 1)
	terms[0] = b
	if net.m_terms[0] != a {
		t.Fatal("net aliases caller terminal list")
	}
	AddCell(a)
	AddCell(b)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("stale delete accepted")
			}
		}()
		DeleteCell(a)
	}()
	if g_place_concreteCells[0] != b {
		t.Fatal("stale delete removed replacement")
	}
}
