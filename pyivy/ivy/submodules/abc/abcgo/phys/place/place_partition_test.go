package place

import (
	"math"
	"math/rand"
	"testing"
)

func makePartitionFixture(n int) *Partition {
	resetPlacementTest()
	g_place_coreBounds = Rect{w: 100, h: 80}
	kind := NewCellType("unit", 1, 1, false)
	for i := 0; i < n; i++ {
		AddCell(NewCell(int32(i), "", kind, float32(i*7%100), float32(i*11%80)))
	}
	initPartitioning()
	return g_place_rootPartition
}
func TestPartitionRefinementAndArea(t *testing.T) {
	p := makePartitionFixture(100)
	defer resetPlacementTest()
	for iter := 0; !refinePartitions(); iter++ {
		if iter > 15 {
			t.Fatal("did not finish")
		}
	}
	seen := map[*ConcreteCell]bool{}
	var visit func(*Partition)
	visit = func(q *Partition) {
		if q.m_leaf {
			if partitionNonzeroMembers(q) > LARGEST_FINAL_SIZE {
				t.Fatal("oversize leaf")
			}
			for _, c := range q.m_members[:q.m_numMembers] {
				if c != nil {
					if seen[c] {
						t.Fatal("duplicate leaf membership")
					}
					seen[c] = true
				}
			}
			return
		}
		if q.m_sub1.m_area+q.m_sub2.m_area != q.m_area {
			t.Fatal("area mismatch")
		}
		a, b := q.m_sub1.m_bounds, q.m_sub2.m_bounds
		if q.m_vertical {
			if a.w+b.w != q.m_bounds.w || b.x != a.x+a.w {
				t.Fatal("vertical coverage")
			}
		} else if a.h+b.h != q.m_bounds.h || b.y != a.y+a.h {
			t.Fatal("horizontal coverage")
		}
		visit(q.m_sub1)
		visit(q.m_sub2)
	}
	visit(p)
	if len(seen) != 100 {
		t.Fatal("lost cells")
	}
}
func TestIncrementalPartitionSparseAndEmpty(t *testing.T) {
	p := makePartitionFixture(0)
	defer resetPlacementTest()
	kind := NewCellType("unit", 2, 3, false)
	a := NewCell(7, "a", kind, 2, 2)
	AddCell(a)
	incrementalPartition()
	if p.m_numMembers != 1 || p.m_area != 6 {
		t.Fatal("first sparse addition")
	}
	incrementalPartition()
	if p.m_numMembers != 1 {
		t.Fatal("duplicate incremental member")
	}
	fixed := NewCell(19, "fixed", kind, 1, 1)
	fixed.m_fixed = true
	AddCell(fixed)
	incrementalPartition()
	if p.m_numMembers != 1 {
		t.Fatal("fixed cell added")
	}
	if !refinePartition(p) {
		t.Fatal("small partition should finish")
	}
	b := NewCell(30, "b", kind, 90, 70)
	AddCell(b)
	incrementalPartition()
	if p.m_numMembers != 2 || p.m_area != 12 {
		t.Fatal("incremental area")
	}
	for _, q := range []*Partition{p.m_sub1, p.m_sub2} {
		var area float32
		for _, c := range q.m_members[:q.m_numMembers] {
			if c != nil {
				area += getCellArea(c)
			}
		}
		if area != q.m_area {
			t.Fatal("child area")
		}
	}
}
func TestPartitionZeroAreaBounds(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind := NewCellType("zero", 0, 0, false)
	for i := int32(0); i < 4; i++ {
		AddCell(NewCell(i, "", kind, float32(i), float32(i)))
	}
	g_place_coreBounds = Rect{w: 100, h: 80}
	initPartitioning()
	if !refinePartitions() {
		t.Fatal("zero-area termination")
	}
	for _, q := range []*Partition{g_place_rootPartition, g_place_rootPartition.m_sub1, g_place_rootPartition.m_sub2} {
		r := q.m_bounds
		if math.IsNaN(float64(r.w)) || math.IsNaN(float64(r.h)) {
			t.Fatal("NaN bounds")
		}
	}
}
func TestNativeHypergraphKnownCuts(t *testing.T) {
	offsets := []int32{0, 2, 4, 6, 8}
	edges := []int32{0, 1, 0, 1, 2, 3, 2, 3}
	part := make([]int32, 4)
	var cut int32
	if err := HMETIS_PartRecursive(4, 4, nil, offsets, edges, nil, 2, 0, nil, part, &cut); err != nil {
		t.Fatal(err)
	}
	if cut != 0 || part[0] != part[1] || part[2] != part[3] || part[0] == part[2] {
		t.Fatal(part, cut)
	}
	options := make([]int32, 9)
	options[6] = 1
	part = []int32{0, -1, -1, 1}
	if err := HMETIS_PartKway(4, 4, nil, offsets, edges, nil, 2, 0, options, part, &cut); err != nil || cut != 0 || part[0] != 0 || part[3] != 1 {
		t.Fatal(err, part, cut)
	}
}
func TestNativeHypergraphAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(907))
	for trial := 0; trial < 80; trial++ {
		n := 6
		offsets := []int32{0}
		var pins []int32
		for e := 0; e < 7; e++ {
			for v := 0; v < n; v++ {
				if rng.Intn(3) == 0 {
					pins = append(pins, int32(v))
				}
			}
			offsets = append(offsets, int32(len(pins)))
		}
		weights := make([]int32, n)
		for i := range weights {
			weights[i] = int32(rng.Intn(3) + 1)
		}
		part := make([]int32, n)
		g, err := readHypergraph(int32(n), 7, weights, offsets, pins, nil, 2, 30, nil, part)
		if err != nil {
			t.Fatal(err)
		}
		want := int64(1<<63 - 1)
		for assignment := 0; assignment < 1<<n; assignment++ {
			p := make([]int, n)
			for i := range p {
				p[i] = assignment >> i & 1
			}
			if g.feasible(g.loads(p)) {
				want = min(want, g.cost(p))
			}
		}
		var cut int32
		err = HMETIS_PartRecursive(int32(n), 7, weights, offsets, pins, nil, 2, 30, nil, part, &cut)
		if want == 1<<63-1 {
			if err == nil {
				t.Fatal("infeasible accepted")
			}
			continue
		}
		if err != nil || int64(cut) != want {
			t.Fatal(trial, err, cut, want)
		}
	}
}
func TestNativeHypergraphValidation(t *testing.T) {
	if err := HMETIS_PartRecursive(2, 1, nil, []int32{0, 2}, []int32{0, 3}, nil, 2, 0, nil, make([]int32, 2), nil); err == nil {
		t.Fatal("bad vertex")
	}
	part := []int32{0, 0}
	options := make([]int32, 9)
	options[6] = 1
	if err := HMETIS_PartRecursive(2, 0, nil, []int32{0}, nil, nil, 2, 0, options, part, nil); err == nil {
		t.Fatal("infeasible fixed weights")
	}
	if part[0] != 0 || part[1] != 0 {
		t.Fatal("error changed assignments")
	}
	if err := HMETIS_PartRecursive(0, 0, nil, []int32{0}, nil, nil, 2, 0, nil, nil, nil); err != nil {
		t.Fatal("empty graph", err)
	}
}

func TestNativeHypergraphKwayAndHeuristic(t *testing.T) {
	for _, n := range []int{6, 40} {
		parts := int32(2)
		if n == 6 {
			parts = 3
		}
		cluster := n / int(parts)
		offsets := []int32{0}
		var pins []int32
		for base := 0; base < n; base += cluster {
			for v := base; v < base+cluster-1; v++ {
				pins = append(pins, int32(v), int32(v+1))
				offsets = append(offsets, int32(len(pins)))
			}
		}
		assignment := make([]int32, n)
		var cuts int32
		if err := HMETIS_PartKway(int32(n), int32(len(offsets)-1), nil, offsets, pins, nil, parts, 0, nil, assignment, &cuts); err != nil || cuts != 0 {
			t.Fatal(n, err, assignment, cuts)
		}
		loads := make([]int, parts)
		for _, p := range assignment {
			if p < 0 || p >= parts {
				t.Fatal("partition range")
			}
			loads[p]++
		}
		for _, load := range loads {
			if load != cluster {
				t.Fatal("balance", loads)
			}
		}
	}
}

func TestPlacementHypergraphRefinementQuality(t *testing.T) {
	resetPlacementTest()
	defer resetPlacementTest()
	kind := NewCellType("unit", 1, 1, false)
	var cells []*ConcreteCell
	for i, x := range []float32{48, 49, 51, 52} {
		c := NewCell(int32(i), "", kind, x, 0)
		AddCell(c)
		cells = append(cells, c)
	}
	for i, pair := range [][2]int{{0, 2}, {1, 3}} {
		AddNet(NewNet(int32(i), []*ConcreteCell{cells[pair[0]], cells[pair[1]]}, 1))
	}
	parent := &Partition{m_members: cells, m_numMembers: 4, m_vertical: true, m_bounds: Rect{w: 100, h: 80}, m_sub1: &Partition{m_leaf: true}, m_sub2: &Partition{m_leaf: true}}
	partitionEqualArea(parent)
	resizePartition(parent)
	if err := repartitionHMetis(parent); err != nil {
		t.Fatal(err)
	}
	side := map[*ConcreteCell]int{}
	for i, p := range []*Partition{parent.m_sub1, parent.m_sub2} {
		for _, c := range p.m_members[:p.m_numMembers] {
			side[c] = i
		}
	}
	for _, n := range g_place_concreteNets[:g_place_numNets] {
		if side[n.m_terms[0]] != side[n.m_terms[1]] {
			t.Fatal("failed to remove geometric cross cuts")
		}
	}
	if len(side) != 4 || parent.m_sub1.m_area != 2 || parent.m_sub2.m_area != 2 {
		t.Fatal("refinement balance/membership")
	}
}
func TestIncrementalReopensCompletedLeaf(t *testing.T) {
	p := makePartitionFixture(4)
	defer resetPlacementTest()
	refinePartition(p)
	if !p.m_done {
		t.Fatal("initial completion")
	}
	kind := NewCellType("unit", 1, 1, false)
	for i := int32(4); i < 29; i++ {
		AddCell(NewCell(i, "", kind, 0, 0))
	}
	incrementalPartition()
	if p.m_done || p.m_area != 29 {
		t.Fatal("completed state was not reopened")
	}
	for i := 0; !refinePartitions(); i++ {
		if i > 10 {
			t.Fatal("incremental refinement did not finish")
		}
	}
}
func TestNativeHypergraphBadIntermediateOffsets(t *testing.T) {
	err := HMETIS_PartRecursive(2, 2, nil, []int32{0, 100, 2}, []int32{0, 1}, nil, 2, 0, nil, make([]int32, 2), nil)
	if err == nil {
		t.Fatal("bad intermediate offset accepted")
	}
}

func TestNativeHypergraphWeightedEdges(t *testing.T) {
	part := make([]int32, 4)
	offsets := []int32{0, 2, 4, 6}
	pins := []int32{0, 1, 1, 2, 2, 3}
	var cost int32
	if err := HMETIS_PartRecursive(4, 3, nil, offsets, pins, []int32{1, 100, 1}, 2, 0, nil, part, &cost); err != nil || cost != 2 || part[1] != part[2] {
		t.Fatal(err, part, cost)
	}
	if err := HMETIS_PartRecursive(1, 0, nil, []int32{0}, nil, nil, 1, 0, nil, []int32{2}, nil); err == nil {
		t.Fatal("invalid seed accepted")
	}
}
