// Placement database and geometry, ported from place_base.c (Aaron P. Hurst).
package place

var g_place_numCells, g_place_numNets int32
var g_place_rowHeight float32 = 1
var g_place_coreBounds, g_place_padBounds Rect
var g_place_concreteCells []*ConcreteCell
var g_place_concreteCellsSize int32
var g_place_concreteNets []*ConcreteNet
var g_place_concreteNetsSize int32

func getNetBBox(net *ConcreteNet) Rect {
	if net == nil {
		panic("nil net")
	}
	if net.m_numTerms == 0 {
		return Rect{}
	}
	terms := net.m_terms[:net.m_numTerms]
	first := terms[0]
	if first == nil {
		panic("nil net terminal")
	}
	left, right, bottom, top := first.m_x, first.m_x, first.m_y, first.m_y
	for _, cell := range terms[1:] {
		if cell == nil {
			panic("nil net terminal")
		}
		left = min(left, cell.m_x)
		right = max(right, cell.m_x)
		bottom = min(bottom, cell.m_y)
		top = max(top, cell.m_y)
	}
	return Rect{x: left, y: bottom, w: right - left, h: top - bottom}
}
func getNetWirelength(net *ConcreteNet) float32 { r := getNetBBox(net); return r.w + r.h }
func getTotalWirelength() float32 {
	var total float32
	for _, net := range g_place_concreteNets[:g_place_numNets] {
		if net != nil {
			total += getNetWirelength(net)
		}
	}
	return total
}
func getCellArea(cell *ConcreteCell) float32 {
	if cell == nil || cell.m_parent == nil {
		panic("cell has no type")
	}
	return cell.m_parent.m_width * cell.m_parent.m_height
}
func growPlacementSlots[T any](slots []*T, id int32) []*T {
	if id < 0 {
		panic("negative placement id")
	}
	if int(id) < len(slots) {
		return slots
	}
	size := int64(id)*3/2 + 20
	if size > 1<<31-1 || size > int64(int(^uint(0)>>1)) {
		panic("placement id too large")
	}
	return append(slots, make([]*T, int(size)-len(slots))...)
}
func addConcreteNet(net *ConcreteNet) {
	if net == nil {
		panic("nil net")
	}
	g_place_concreteNets = growPlacementSlots(g_place_concreteNets, net.m_id)
	g_place_concreteNetsSize = int32(len(g_place_concreteNets))
	g_place_numNets = max(g_place_numNets, net.m_id+1)
	g_place_concreteNets[net.m_id] = net
}
func delConcreteNet(net *ConcreteNet) {
	if net == nil || net.m_id < 0 || net.m_id >= g_place_numNets {
		panic("invalid net deletion")
	}
	if g_place_concreteNets[net.m_id] != net {
		panic("net is not registered")
	}
	g_place_concreteNets[net.m_id] = nil
	for g_place_numNets > 0 && g_place_concreteNets[g_place_numNets-1] == nil {
		g_place_numNets--
	}
}
func addConcreteCell(cell *ConcreteCell) {
	if cell == nil {
		panic("nil cell")
	}
	g_place_concreteCells = growPlacementSlots(g_place_concreteCells, cell.m_id)
	g_place_concreteCellsSize = int32(len(g_place_concreteCells))
	g_place_numCells = max(g_place_numCells, cell.m_id+1)
	g_place_concreteCells[cell.m_id] = cell
}
func delCellFromPartition(cell *ConcreteCell, p *Partition) {
	if cell == nil || p == nil {
		panic("nil partition deletion")
	}
	for i, member := range p.m_members[:p.m_numMembers] {
		if member != cell {
			continue
		}
		p.m_members[i] = nil
		p.m_area -= getCellArea(cell)
		if !p.m_leaf {
			delCellFromPartition(cell, p.m_sub1)
			delCellFromPartition(cell, p.m_sub2)
		}
		return
	}
}
func delConcreteCell(cell *ConcreteCell) {
	if cell == nil || cell.m_id < 0 || cell.m_id >= g_place_numCells {
		panic("invalid cell deletion")
	}
	if g_place_concreteCells[cell.m_id] != cell {
		panic("cell is not registered")
	}
	g_place_concreteCells[cell.m_id] = nil
	for g_place_numCells > 0 && g_place_concreteCells[g_place_numCells-1] == nil {
		g_place_numCells--
	}
	if g_place_rootPartition != nil {
		delCellFromPartition(cell, g_place_rootPartition)
	}
}
func comparePlacementFloat(a, b float32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func comparePlacementID(a, b int32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func netPositionCompare(a, b *ConcreteNet, axis int32) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	ra, rb := getNetBBox(a), getNetBBox(b)
	switch axis {
	case 0:
		return comparePlacementFloat(ra.x, rb.x)
	case 1:
		return comparePlacementFloat(ra.x+ra.w, rb.x+rb.w)
	case 2:
		return comparePlacementFloat(ra.y+ra.h, rb.y+rb.h)
	default:
		return comparePlacementFloat(ra.y, rb.y)
	}
}
func netSortByL(a, b *ConcreteNet) int32 { return netPositionCompare(a, b, 0) }
func netSortByR(a, b *ConcreteNet) int32 { return netPositionCompare(a, b, 1) }
func netSortByB(a, b *ConcreteNet) int32 { return netPositionCompare(a, b, 2) }
func netSortByT(a, b *ConcreteNet) int32 { return netPositionCompare(a, b, 3) }
func netSortByID(a, b *ConcreteNet) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	return comparePlacementID(a.m_id, b.m_id)
}
func cellSortByX(a, b *ConcreteCell) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	return comparePlacementFloat(a.m_x, b.m_x)
}
func cellSortByY(a, b *ConcreteCell) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	return comparePlacementFloat(a.m_y, b.m_y)
}
func cellSortByID(a, b *ConcreteCell) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	return comparePlacementID(a.m_id, b.m_id)
}

// Exported entry points provide typed Go access to the placement database.
func AddCell(cell *ConcreteCell)             { addConcreteCell(cell) }
func DeleteCell(cell *ConcreteCell)          { delConcreteCell(cell) }
func AddNet(net *ConcreteNet)                { addConcreteNet(net) }
func DeleteNet(net *ConcreteNet)             { delConcreteNet(net) }
func NetBounds(net *ConcreteNet) Rect        { return getNetBBox(net) }
func NetWirelength(net *ConcreteNet) float32 { return getNetWirelength(net) }
func TotalWirelength() float32               { return getTotalWirelength() }
func CellArea(cell *ConcreteCell) float32    { return getCellArea(cell) }
