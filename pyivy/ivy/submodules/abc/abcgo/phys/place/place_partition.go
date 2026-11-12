// Recursive placement partitioning, ported from place_partition.c
// (Aaron P. Hurst). Proprietary hMetis is replaced by the native weighted-cut
// implementation in libhmetis.go; see REVIEW.md for its algorithmic differences.
package place

import (
	"fmt"
	"math"
	"sort"
)

var g_place_rootPartition *Partition
var allNetsR2, allNetsL2, allNetsB2, allNetsT2 []*ConcreteNet

func initPartitioning() {
	g_place_numPartitions = 1
	p := &Partition{m_bounds: g_place_coreBounds, m_leaf: true}
	for _, cell := range g_place_concreteCells[:g_place_numCells] {
		if cell != nil && !cell.m_fixed {
			partitionAddMember(p, cell)
		}
	}
	g_place_rootPartition = p
}
func presortNets() {
	nets := g_place_concreteNets[:g_place_numNets]
	allNetsL2 = append(allNetsL2[:0], nets...)
	allNetsR2 = append(allNetsR2[:0], nets...)
	allNetsB2 = append(allNetsB2[:0], nets...)
	allNetsT2 = append(allNetsT2[:0], nets...)
	for _, entry := range []struct {
		nets    []*ConcreteNet
		compare func(*ConcreteNet, *ConcreteNet) int32
	}{{allNetsL2, netSortByL}, {allNetsR2, netSortByR}, {allNetsB2, netSortByB}, {allNetsT2, netSortByT}} {
		sort.SliceStable(entry.nets, func(i, j int) bool { return entry.compare(entry.nets[i], entry.nets[j]) < 0 })
	}
}
func refinePartitions() bool {
	if g_place_rootPartition == nil {
		return true
	}
	return refinePartition(g_place_rootPartition)
}
func reallocPartitions() {
	if g_place_rootPartition != nil {
		reallocPartition(g_place_rootPartition)
	}
}
func partitionAddMember(p *Partition, cell *ConcreteCell) {
	p.m_members = append(p.m_members[:p.m_numMembers], cell)
	p.m_numMembers++
	p.m_area += getCellArea(cell)
}
func resetPartitionMembers(p *Partition) {
	clear(p.m_members)
	p.m_members = p.m_members[:0]
	p.m_numMembers = 0
	p.m_area = 0
}
func partitionNonzeroMembers(p *Partition) int {
	count := 0
	for _, cell := range p.m_members[:p.m_numMembers] {
		if cell != nil && getCellArea(cell) > 0 {
			count++
		}
	}
	return count
}
func partitionChildren(p *Partition) {
	if p == nil || p.m_sub1 == nil || p.m_sub2 == nil {
		panic("partition children are missing")
	}
}
func refinePartition(p *Partition) bool {
	if p == nil {
		panic("nil partition")
	}
	if p.m_done {
		return true
	}
	if !p.m_leaf {
		first := refinePartition(p.m_sub1)
		second := refinePartition(p.m_sub2)
		p.m_done = first && second
		return p.m_done
	}
	g_place_numPartitions++
	p.m_sub1 = &Partition{m_level: p.m_level + 1, m_leaf: true, m_vertical: !p.m_vertical}
	p.m_sub2 = &Partition{m_level: p.m_level + 1, m_leaf: true, m_vertical: !p.m_vertical}
	p.m_leaf = false
	if PARTITION_AREA_ONLY {
		partitionEqualArea(p)
	} else {
		partitionScanlineMincut(p)
	}
	resizePartition(p)
	if p.m_level < REPARTITION_LEVEL_DEPTH {
		if REPARTITION_FM {
			repartitionFM(p)
		} else if REPARTITION_HMETIS {
			if err := repartitionHMetis(p); err != nil {
				fmt.Printf("QPAR-02 : repartitioning: %v\n", err)
			}
		}
	}
	resizePartition(p)
	a, b := partitionNonzeroMembers(p.m_sub1), partitionNonzeroMembers(p.m_sub2)
	p.m_sub1.m_done = a <= LARGEST_FINAL_SIZE
	p.m_sub2.m_done = b <= LARGEST_FINAL_SIZE
	if a == 0 || b == 0 {
		fmt.Println("QPART-35 : WARNING: degenerate partition generated")
		partitionEqualArea(p)
		resizePartition(p)
		p.m_sub1.m_done = true
		p.m_sub2.m_done = true
	}
	p.m_done = p.m_sub1.m_done && p.m_sub2.m_done
	return p.m_done
}
func partitionEqualArea(parent *Partition) {
	partitionChildren(parent)
	members := parent.m_members[:parent.m_numMembers]
	compare := cellSortByY
	if parent.m_vertical {
		compare = cellSortByX
	}
	sort.SliceStable(members, func(i, j int) bool { return compare(members[i], members[j]) < 0 })
	// Recompute area rather than trusting stale accounting after incremental edits.
	parent.m_area = 0
	for _, cell := range members {
		if cell != nil {
			parent.m_area += getCellArea(cell)
		}
	}
	resetPartitionMembers(parent.m_sub1)
	resetPartitionMembers(parent.m_sub2)
	half := parent.m_area * .5
	for _, cell := range members {
		if cell == nil {
			continue
		}
		target := parent.m_sub2
		if parent.m_sub1.m_area < half {
			target = parent.m_sub1
		}
		partitionAddMember(target, cell)
	}
}
func resizePartition(p *Partition) {
	partitionChildren(p)
	a, b := p.m_sub1, p.m_sub2
	fraction := float32(.5)
	total := a.m_area + b.m_area
	p.m_area = total
	if total > 0 {
		fraction = a.m_area / total
	} else if a.m_numMembers+b.m_numMembers > 0 {
		fraction = float32(a.m_numMembers) / float32(a.m_numMembers+b.m_numMembers)
	}
	a.m_bounds = p.m_bounds
	b.m_bounds = p.m_bounds
	if p.m_vertical {
		a.m_bounds.w = p.m_bounds.w * fraction
		b.m_bounds.x = p.m_bounds.x + a.m_bounds.w
		b.m_bounds.w = p.m_bounds.w - a.m_bounds.w
	} else {
		a.m_bounds.h = p.m_bounds.h * fraction
		b.m_bounds.y = p.m_bounds.y + a.m_bounds.h
		b.m_bounds.h = p.m_bounds.h - a.m_bounds.h
	}
}
func reallocPartition(p *Partition) {
	if p == nil || p.m_leaf {
		return
	}
	if PARTITION_AREA_ONLY {
		partitionEqualArea(p)
	} else {
		partitionScanlineMincut(p)
	}
	resizePartition(p)
	if p.m_level < REPARTITION_LEVEL_DEPTH && REPARTITION_HMETIS {
		if err := repartitionHMetis(p); err != nil {
			fmt.Printf("QPAR-02 : repartitioning: %v\n", err)
		}
		resizePartition(p)
	}
	reallocPartition(p.m_sub1)
	reallocPartition(p.m_sub2)
}

func repartitionHMetis(parent *Partition) error { return partitionRefineCut(parent, true) }
func repartitionFM(parent *Partition) {
	if err := partitionRefineCut(parent, false); err != nil {
		fmt.Printf("QPAR-02 : gain refinement: %v\n", err)
	}
}
func partitionRefineCut(parent *Partition, nearCut bool) error {
	partitionChildren(parent)
	assignment, weights := make([]int32, g_place_numCells), make([]int32, g_place_numCells)
	cut, width := parent.m_sub2.m_bounds.y, parent.m_bounds.h
	if parent.m_vertical {
		cut, width = parent.m_sub2.m_bounds.x, parent.m_bounds.w
	}
	for id, cell := range g_place_concreteCells[:g_place_numCells] {
		if cell != nil {
			coordinate := cell.m_y
			if parent.m_vertical {
				coordinate = cell.m_x
			}
			if coordinate >= cut {
				assignment[id] = 1
			}
		}
	}
	targets := 0
	for _, cell := range parent.m_members[:parent.m_numMembers] {
		if cell == nil {
			continue
		}
		if cell.m_id < 0 || cell.m_id >= g_place_numCells || g_place_concreteCells[cell.m_id] != cell {
			return fmt.Errorf("partition member is not in database")
		}
		area := float64(getCellArea(cell))
		if area < 0 || math.IsNaN(area) || math.IsInf(area, 0) || area*1000 > math.MaxInt32 {
			return fmt.Errorf("invalid partition cell area")
		}
		weight := int32(math.Round(area * 1000))
		if area > 0 && weight == 0 {
			weight = 1
		}
		weights[cell.m_id] = weight
		coordinate := cell.m_y
		if parent.m_vertical {
			coordinate = cell.m_x
		}
		if !cell.m_fixed && (!nearCut || math.Abs(float64(coordinate-cut)) < float64(width*REPARTITION_TARGET_FRACTION)) {
			assignment[cell.m_id] = -1
			targets++
		}
	}
	if targets == 0 {
		return nil
	}
	offsets := []int32{0}
	var edges []int32
	for _, net := range g_place_concreteNets[:g_place_numNets] {
		if net == nil || net.m_numTerms <= 1 {
			continue
		}
		for _, cell := range net.m_terms[:net.m_numTerms] {
			if cell == nil {
				return fmt.Errorf("nil hyperedge terminal")
			}
			edges = append(edges, cell.m_id)
		}
		offsets = append(offsets, int32(len(edges)))
	}
	options := make([]int32, 9)
	options[0] = 1
	options[6] = 1
	options[7] = 12261980
	var afterCuts int32
	if err := HMETIS_PartRecursive(g_place_numCells, int32(len(offsets)-1), weights, offsets, edges, nil, 2, int32(100*MAX_PARTITION_NONSYMMETRY), options, assignment, &afterCuts); err != nil {
		return err
	}
	resetPartitionMembers(parent.m_sub1)
	resetPartitionMembers(parent.m_sub2)
	for _, cell := range parent.m_members[:parent.m_numMembers] {
		if cell == nil {
			continue
		}
		target := parent.m_sub1
		if assignment[cell.m_id] != 0 {
			target = parent.m_sub2
		}
		partitionAddMember(target, cell)
	}
	return nil
}

// partitionScanlineMincut evaluates balanced geometric cuts against the actual
// net connectivity. The old source had only a disabled C++ experiment here.
func partitionScanlineMincut(parent *Partition) {
	partitionEqualArea(parent)
	members := make([]*ConcreteCell, 0, parent.m_numMembers)
	for _, cell := range parent.m_members[:parent.m_numMembers] {
		if cell != nil {
			members = append(members, cell)
		}
	}
	if len(members) < 2 {
		return
	}
	positions := make(map[*ConcreteCell]int, len(members))
	for i, cell := range members {
		positions[cell] = i
	}
	best, bestCuts := -1, int(^uint(0)>>1)
	var area float32
	lower, upper := parent.m_area*(.5-MAX_PARTITION_NONSYMMETRY), parent.m_area*(.5+MAX_PARTITION_NONSYMMETRY)
	for split := 1; split < len(members); split++ {
		area += getCellArea(members[split-1])
		if area < lower || area > upper {
			continue
		}
		cuts := 0
		for _, net := range g_place_concreteNets[:g_place_numNets] {
			if net == nil {
				continue
			}
			left, right := false, false
			for _, cell := range net.m_terms[:net.m_numTerms] {
				if pos, ok := positions[cell]; ok {
					left = left || pos < split
					right = right || pos >= split
				}
			}
			if left && right {
				cuts++
			}
		}
		if cuts < bestCuts {
			best, bestCuts = split, cuts
		}
	}
	if best < 0 {
		return
	}
	resetPartitionMembers(parent.m_sub1)
	resetPartitionMembers(parent.m_sub2)
	for i, cell := range members {
		target := parent.m_sub2
		if i < best {
			target = parent.m_sub1
		}
		partitionAddMember(target, cell)
	}
}
func incrementalSubpartition(p *Partition, cells []*ConcreteCell, n int32) {
	if p == nil {
		panic("nil incremental partition")
	}
	for _, cell := range cells[:n] {
		if cell == nil {
			panic("nil incremental cell")
		}
		partitionAddMember(p, cell)
	}
	if p.m_leaf {
		if partitionNonzeroMembers(p) > LARGEST_FINAL_SIZE {
			p.m_done = false
		}
		return
	}
	partitionChildren(p)
	cut := p.m_sub2.m_bounds.y
	if p.m_vertical {
		cut = p.m_sub2.m_bounds.x
	}
	var first, second []*ConcreteCell
	for _, cell := range cells[:n] {
		coordinate := cell.m_y
		if p.m_vertical {
			coordinate = cell.m_x
		}
		if coordinate < cut {
			first = append(first, cell)
		} else {
			second = append(second, cell)
		}
	}
	if len(first) > 0 {
		incrementalSubpartition(p.m_sub1, first, int32(len(first)))
	}
	if len(second) > 0 {
		incrementalSubpartition(p.m_sub2, second, int32(len(second)))
	}
	p.m_done = p.m_sub1.m_done && p.m_sub2.m_done
}
func incrementalPartition() {
	if g_place_rootPartition == nil {
		panic("partitioning not initialized")
	}
	existing := make(map[*ConcreteCell]bool, g_place_rootPartition.m_numMembers)
	for _, cell := range g_place_rootPartition.m_members[:g_place_rootPartition.m_numMembers] {
		if cell != nil {
			existing[cell] = true
		}
	}
	var added []*ConcreteCell
	for _, cell := range g_place_concreteCells[:g_place_numCells] {
		if cell != nil && !cell.m_fixed && !existing[cell] {
			added = append(added, cell)
		}
	}
	fmt.Printf("QPRT-50 : incremental partitioning with %d new cells\n", len(added))
	if len(added) > 0 {
		incrementalSubpartition(g_place_rootPartition, added, int32(len(added)))
	}
}
func InitPartitioning()         { initPartitioning() }
func RefinePartitions() bool    { return refinePartitions() }
func ReallocatePartitions()     { reallocPartitions() }
func IncrementalPartitioning()  { incrementalPartition() }
func RootPartition() *Partition { return g_place_rootPartition }
func (p *Partition) Members() []*ConcreteCell {
	return append([]*ConcreteCell(nil), p.m_members[:p.m_numMembers]...)
}
func (p *Partition) Bounds() Rect { return p.m_bounds }
func (p *Partition) IsLeaf() bool { return p.m_leaf }
