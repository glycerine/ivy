package fpga

import (
	"fmt"
	"sort"
)

// choiceSignal canonicalizes a secondary node while preserving its signal phase.
func choiceSignal(edge *Fpga_Node_t) *Fpga_Node_t {
	n := Fpga_Regular(edge)
	if n != nil && n.Repr != nil {
		return Fpga_NotCond(n.Repr, Fpga_IsComplement(edge)^boolInt(n.Phase != n.Repr.Phase))
	}
	return edge
}
func dfsNodes(roots []*Fpga_Node_t, equiv bool) *Fpga_NodeVec_t {
	result := Fpga_NodeVecAlloc(0)
	state := map[*Fpga_Node_t]uint8{}
	var visit func(*Fpga_Node_t, bool)
	visit = func(edge *Fpga_Node_t, alternative bool) {
		n := Fpga_Regular(edge)
		if n == nil {
			panic("fpga: unset output")
		}
		if !alternative && n.Repr != nil {
			n = Fpga_Regular(choiceSignal(n))
		}
		if state[n] == 2 {
			return
		}
		if state[n] == 1 {
			panic("fpga: cyclic graph or choices")
		}
		state[n] = 1
		if Fpga_NodeIsAnd(n) != 0 {
			visit(n.One, false)
			visit(n.Two, false)
		}
		if equiv {
			for alt := n.NextE; alt != nil; alt = alt.NextE {
				visit(alt, true)
			}
		}
		state[n] = 2
		result.Nodes = append(result.Nodes, n)
	}
	for _, root := range roots {
		visit(root, false)
	}
	return result
}
func Fpga_MappingDfs(p *Fpga_Man_t, equiv int32) *Fpga_NodeVec_t {
	return dfsNodes(p.Outputs, equiv != 0)
}
func Fpga_MappingDfs_rec(n *Fpga_Node_t, v *Fpga_NodeVec_t, equiv int32) {
	for _, n := range dfsNodes([]*Fpga_Node_t{n}, equiv != 0).Nodes {
		Fpga_NodeVecPushUnique(v, n)
	}
}
func Fpga_MappingDfsNodes(p *Fpga_Man_t, nodes []*Fpga_Node_t, n int32, equiv int32) *Fpga_NodeVec_t {
	return dfsNodes(nodes[:n], equiv != 0)
}
func Fpga_MappingGetAreaFlow(p *Fpga_Man_t) float32 {
	var area float32
	for _, edge := range p.Outputs {
		n := Fpga_Regular(choiceSignal(edge))
		if Fpga_NodeIsAnd(n) != 0 {
			area += n.Best.Flow
		}
	}
	return area
}
func Fpga_MappingArea(p *Fpga_Man_t) float32 {
	seen := map[*Fpga_Node_t]bool{}
	var rec func(*Fpga_Node_t) float32
	rec = func(e *Fpga_Node_t) float32 {
		n := Fpga_Regular(choiceSignal(e))
		if seen[n] || Fpga_NodeIsAnd(n) == 0 {
			return 0
		}
		seen[n] = true
		a := Fpga_CutGetRootArea(p, n.Best)
		for _, leaf := range n.Best.Leaves {
			a += rec(leaf)
		}
		return a
	}
	var area float32
	for _, e := range p.Outputs {
		area += rec(e)
	}
	return area
}
func Fpga_MappingArea_rec(p *Fpga_Man_t, n *Fpga_Node_t, v *Fpga_NodeVec_t) float32 {
	n = Fpga_Regular(n)
	if Fpga_NodeIsAnd(n) == 0 || n.Used {
		return 0
	}
	n.Used = true
	v.Nodes = append(v.Nodes, n)
	a := Fpga_CutGetRootArea(p, n.Best)
	for _, leaf := range n.Best.Leaves {
		a += Fpga_MappingArea_rec(p, leaf, v)
	}
	return a
}
func Fpga_MappingAreaTrav(p *Fpga_Man_t) float32 { return Fpga_MappingArea(p) }
func Fpga_MappingSetRefsAndArea(p *Fpga_Man_t) float32 {
	for _, n := range p.Nodes {
		n.Refs = 0
	}
	p.Const1.Refs = 0
	post := []*Fpga_Node_t{}
	var visit func(*Fpga_Node_t) float32
	visit = func(e *Fpga_Node_t) float32 {
		n := Fpga_Regular(choiceSignal(e))
		n.Refs++
		if n.Refs > 1 || Fpga_NodeIsAnd(n) == 0 {
			return 0
		}
		if n.Best == nil {
			panic("fpga: mapped node without best cut")
		}
		a := Fpga_CutGetRootArea(p, n.Best)
		for _, leaf := range n.Best.Leaves {
			a += visit(leaf)
		}
		post = append(post, n)
		return a
	}
	var area float32
	for _, e := range p.Outputs {
		area += visit(e)
	}
	p.Mapping.Nodes = post
	Fpga_NodeVecReverse(p.Mapping)
	p.AreaGlobal = area
	return area
}
func Fpga_MappingSetRefsAndArea_rec(p *Fpga_Man_t, n *Fpga_Node_t, store *[]*Fpga_Node_t) float32 {
	n = Fpga_Regular(n)
	n.Refs++
	if n.Refs > 1 || Fpga_NodeIsAnd(n) == 0 {
		return 0
	}
	a := Fpga_CutGetRootArea(p, n.Best)
	*store = append(*store, n)
	for _, leaf := range n.Best.Leaves {
		a += Fpga_MappingSetRefsAndArea_rec(p, leaf, store)
	}
	return a
}
func Fpga_MappingCompareOutputDelay(a, b **Fpga_Node_t) int32 {
	return -Fpga_NodeVecCompareArrivals(a, b)
}
func Fpga_MappingFindLatest(p *Fpga_Man_t, indices []int32, max int32) {
	a := make([]int, len(p.Outputs))
	for i := range a {
		a[i] = i
	}
	sort.SliceStable(a, func(i, j int) bool {
		return Fpga_Regular(p.Outputs[a[i]]).Best.Arrival > Fpga_Regular(p.Outputs[a[j]]).Best.Arrival
	})
	for i := 0; i < int(max) && i < len(a) && i < len(indices); i++ {
		indices[i] = int32(a[i])
	}
}
func Fpga_MappingPrintOutputArrivals(p *Fpga_Man_t) {
	for i, e := range p.Outputs {
		fmt.Fprintf(p.Output, "Output %d: %g\n", i, Fpga_Regular(e).Best.Arrival)
	}
}
func Fpga_MappingSetupTruthTables(t [][2]uint32) {
	for i := range t {
		for m := 0; m < 64; m++ {
			if m>>i&1 != 0 {
				t[i][m/32] |= 1 << uint(m%32)
			}
		}
	}
}
func Fpga_MappingSetupMask(mask []uint32, n int32) {
	m := truthMask(int(n))
	if len(mask) > 0 {
		mask[0] = uint32(m)
	}
	if len(mask) > 1 {
		mask[1] = uint32(m >> 32)
	}
}
func Fpga_ManCheckConsistency(p *Fpga_Man_t) int32 {
	for _, n := range p.Nodes {
		if n.manager != p {
			return 0
		}
		if Fpga_NodeIsAnd(n) != 0 && (Fpga_Regular(n.One).manager != p || Fpga_Regular(n.Two).manager != p) {
			return 0
		}
	}
	return 1
}
func Fpga_CompareNodesByLevelDecreasing(a, b **Fpga_Node_t) int32 {
	return -Fpga_NodeVecCompareLevels(a, b)
}
func Fpga_CompareNodesByLevelIncreasing(a, b **Fpga_Node_t) int32 {
	return Fpga_NodeVecCompareLevels(a, b)
}
func Fpga_MappingSortByLevel(p *Fpga_Man_t, v *Fpga_NodeVec_t, increasing int32) {
	Fpga_NodeVecSortByLevel(v)
	if increasing == 0 {
		Fpga_NodeVecReverse(v)
	}
}
func Fpga_DfsLim(p *Fpga_Man_t, n *Fpga_Node_t, levels int32) *Fpga_NodeVec_t {
	v := Fpga_NodeVecAlloc(0)
	Fpga_DfsLim_rec(n, levels, v)
	return v
}
func Fpga_DfsLim_rec(n *Fpga_Node_t, levels int32, v *Fpga_NodeVec_t) {
	seen := map[*Fpga_Node_t]bool{}
	for _, old := range v.Nodes {
		seen[old] = true
	}
	var visit func(*Fpga_Node_t, int32)
	visit = func(edge *Fpga_Node_t, depth int32) {
		n := Fpga_Regular(edge)
		if seen[n] {
			return
		}
		seen[n] = true
		if depth > 1 && Fpga_NodeIsAnd(n) != 0 {
			visit(n.One, depth-1)
			visit(n.Two, depth-1)
		}
		v.Nodes = append(v.Nodes, n)
	}
	visit(n, levels)
}
func Fpga_ManCleanData0(p *Fpga_Man_t) {
	p.Const1.Data0 = nil
	for _, n := range p.Nodes {
		n.Data0 = nil
	}
}
func Fpga_CollectNodeTfo(p *Fpga_Man_t, n *Fpga_Node_t) *Fpga_NodeVec_t {
	v := Fpga_NodeVecAlloc(0)
	seen := map[*Fpga_Node_t]bool{}
	var rec func(*Fpga_Node_t)
	rec = func(n *Fpga_Node_t) {
		if seen[n] {
			return
		}
		seen[n] = true
		for _, f := range n.fanouts {
			rec(f)
		}
		v.Nodes = append(v.Nodes, n)
	}
	rec(Fpga_Regular(n))
	Fpga_NodeVecReverse(v)
	return v
}
func Fpga_CollectNodeTfo_rec(n, pivot *Fpga_Node_t, visited, tfo *Fpga_NodeVec_t) int32 {
	n = Fpga_Regular(n)
	if Fpga_NodeVecPushUnique(visited, n) != 0 {
		return boolInt(n.Used)
	}
	yes := n == Fpga_Regular(pivot)
	if Fpga_NodeIsAnd(n) != 0 {
		a := Fpga_CollectNodeTfo_rec(n.One, pivot, visited, tfo)
		b := Fpga_CollectNodeTfo_rec(n.Two, pivot, visited, tfo)
		yes = yes || a != 0 || b != 0
	}
	n.Used = yes
	if yes {
		tfo.Nodes = append(tfo.Nodes, n)
	}
	return boolInt(yes)
}

// Levelize returns one linked-list head per level, as in the C API.
func Fpga_MappingLevelize(p *Fpga_Man_t, v *Fpga_NodeVec_t) *Fpga_NodeVec_t {
	var highest int32
	for _, n := range v.Nodes {
		n.Level = 0
		if Fpga_NodeIsAnd(n) != 0 {
			n.Level = 1 + max(Fpga_Regular(n.One).Level, Fpga_Regular(n.Two).Level)
		}
		highest = max(highest, n.Level)
	}
	levels := &Fpga_NodeVec_t{Nodes: make([]*Fpga_Node_t, highest+1)}
	for _, n := range v.Nodes {
		n.NextLevel = nil
		if Fpga_NodeIsAnd(n) != 0 {
			n.NextLevel = levels.Nodes[n.Level]
			levels.Nodes[n.Level] = n
		}
	}
	return levels
}
func Fpga_MappingMaxLevel(p *Fpga_Man_t) int32 {
	var level int32
	for _, e := range p.Outputs {
		if Fpga_Regular(e).Level > level {
			level = Fpga_Regular(e).Level
		}
	}
	return level
}
func Fpga_MappingUpdateLevel_rec(p *Fpga_Man_t, n *Fpga_Node_t, maximum int32) int32 {
	memo := map[*Fpga_Node_t]int32{}
	active := map[*Fpga_Node_t]bool{}
	var visit func(*Fpga_Node_t) int32
	visit = func(edge *Fpga_Node_t) int32 {
		n := Fpga_Regular(edge)
		if Fpga_NodeIsAnd(n) == 0 {
			return n.Level
		}
		if value, ok := memo[n]; ok {
			return value
		}
		if active[n] {
			panic("fpga: cyclic choices")
		}
		active[n] = true
		level := 1 + max(visit(n.One), visit(n.Two))
		if n.NextE != nil {
			other := visit(n.NextE)
			if maximum != 0 {
				level = max(level, other)
			} else {
				level = min(level, other)
			}
		}
		delete(active, n)
		n.Level = level
		memo[n] = level
		if n.Repr == nil {
			for alt := n.NextE; alt != nil; alt = alt.NextE {
				alt.Level = level
			}
		}
		return level
	}
	return visit(n)
}
func Fpga_MappingSetChoiceLevels(p *Fpga_Man_t) {
	p.ChoiceNodes = 0
	p.Choices = 0
	for _, edge := range p.Outputs {
		Fpga_MappingUpdateLevel_rec(p, Fpga_Regular(edge), 1)
	}
	for _, n := range p.Ands.Nodes {
		if n.NextE != nil && n.Repr == nil {
			p.ChoiceNodes++
			for a := n.NextE; a != nil; a = a.NextE {
				p.Choices++
			}
		}
	}
}
func Fpga_ManReportChoices(p *Fpga_Man_t) {
	fmt.Fprintf(p.Output, "Choice nodes = %d. Choices = %d.\n", p.ChoiceNodes, p.Choices)
}
func Fpga_MappingOrderCosByLevel(p *Fpga_Man_t) *Fpga_NodeVec_t {
	v := &Fpga_NodeVec_t{Nodes: append([]*Fpga_Node_t(nil), p.Outputs...)}
	Fpga_NodeVecSortByLevel(v)
	Fpga_NodeVecReverse(v)
	return v
}
