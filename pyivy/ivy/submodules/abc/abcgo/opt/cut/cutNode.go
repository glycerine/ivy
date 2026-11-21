package cut

import (
	"time"
)

func Cut_CutCheckDominance(dom, cut *Cut_Cut_t) int32 {
	if cutContains(cut.leaves, dom.leaves) {
		return 1
	}
	return 0
}
func cutListDominates(list *Cut_List_t, leaves []int32, sign uint32) bool {
	for size := 1; size <= len(leaves); size++ {
		for old := list.head[size]; old != nil; old = old.Next {
			if old.USign&^sign == 0 && cutContains(leaves, old.leaves) {
				return true
			}
		}
	}
	return false
}

func Cut_CutFilterOne(p *Cut_Man_t, list *Cut_List_t, c *Cut_Cut_t) int32 {
	if cutListDominates(list, c.leaves, c.USign) {
		p.stats.Filtered++
		Cut_CutRecycle(p, c)
		return 1
	}
	cutRemoveSupersets(p, list, c)
	return 0
}

func cutRemoveSupersets(p *Cut_Man_t, list *Cut_List_t, c *Cut_Cut_t) {
	for size := len(c.leaves) + 1; size <= int(p.params.NVarsMax); size++ {
		slot := &list.head[size]
		var last *Cut_Cut_t
		for *slot != nil {
			old := *slot
			if c.USign&^old.USign == 0 && cutContains(old.leaves, c.leaves) {
				*slot = old.Next
				list.count--
				p.stats.Filtered++
				Cut_CutRecycle(p, old)
			} else {
				last = old
				slot = &old.Next
			}
		}
		list.tail[size] = last
	}
}
func Cut_CutFilterOneEqual(p *Cut_Man_t, list *Cut_List_t, c *Cut_Cut_t) int32 {
	for old := list.head[len(c.leaves)]; old != nil; old = old.Next {
		if Cut_CutCompare(c, old) == 0 {
			p.stats.Filtered++
			Cut_CutRecycle(p, c)
			return 1
		}
	}
	return 0
}
func cutFrontierDominates(list, c *Cut_Cut_t) bool {
	for old := list; old != nil; old = old.Next {
		if old.USign&^c.USign == 0 && cutContains(c.leaves, old.leaves) {
			return true
		}
	}
	return false
}

// Remove strict supersets and update the caller's head when the first cut goes.
func cutPruneFrontier(p *Cut_Man_t, slot **Cut_Cut_t, c *Cut_Cut_t) int32 {
	var removed int32
	for *slot != nil {
		old := *slot
		if len(c.leaves) < len(old.leaves) && c.USign&^old.USign == 0 && cutContains(old.leaves, c.leaves) {
			*slot = old.Next
			p.stats.Filtered++
			Cut_CutRecycle(p, old)
			removed++
		} else {
			slot = &old.Next
		}
	}
	return removed
}

func Cut_CutFilterOld(p *Cut_Man_t, list, c *Cut_Cut_t) int32 {
	if cutFrontierDominates(list, c) {
		p.stats.Filtered++
		Cut_CutRecycle(p, c)
		return 1
	}
	// This entry point, like the C routine, receives the trivial head by
	// value. The internal frontier API can replace a nontrivial head as well.
	if list != nil {
		cutPruneFrontier(p, &list.Next, c)
	}
	return 0
}

type cutSeqFrontiers struct {
	old, current *Cut_Cut_t
	limit        int32
}

func (f *cutSeqFrontiers) filter(p *Cut_Man_t, c *Cut_Cut_t) bool {
	for _, list := range [2]**Cut_Cut_t{&f.old, &f.current} {
		if cutFrontierDominates(*list, c) {
			return true
		}
		f.limit += cutPruneFrontier(p, list, c)
	}
	return false
}

func (p *Cut_Man_t) globalCutEligible(leaves []int32) bool {
	if len(leaves) == 1 {
		return true
	}
	for _, leaf := range leaves {
		node := leaf
		if p.params.FSeq != 0 {
			node >>= CUT_SHIFT
		}
		if int(node) < len(p.attrs) && p.attrs[node] != 0 {
			return true
		}
	}
	return false
}

func Cut_CutFilterGlobal(p *Cut_Man_t, c *Cut_Cut_t) int32 {
	if p.globalCutEligible(c.leaves) {
		return 0
	}
	p.stats.Filtered++
	Cut_CutRecycle(p, c)
	return 1
}

// Cut lists are ordered by size. The stop pointer also lets a tree
// boundary expose just its trivial cut without changing the fanin list.
func cutListBounds(head *Cut_Cut_t, limit int32, tree bool) (full, stop *Cut_Cut_t) {
	if tree {
		if len(head.leaves) != 1 {
			panic("cut: tree boundary lacks trivial cut")
		}
		return head.Next, head.Next
	}
	for head != nil && int32(len(head.leaves)) < limit {
		head = head.Next
	}
	return head, nil
}

func Cut_NodeDoComputeCuts(p *Cut_Man_t, super *Cut_List_t, node, compl0, compl1 int32, a, b *Cut_Cut_t, trivial, treeCode int32) {
	cutNodeDoComputeCuts(p, super, node, compl0, compl1, a, b, trivial, treeCode, nil)
}

func cutNodeDoComputeCuts(p *Cut_Man_t, super *Cut_List_t, node, compl0, compl1 int32, a, b *Cut_Cut_t, trivial, treeCode int32, frontiers *cutSeqFrontiers) {
	limit := p.params.NKeepMax
	if frontiers != nil {
		limit = frontiers.limit
	}
	if super.count >= limit {
		return
	}
	if a != nil && b != nil {
		x, y := ((a.FSimul != a.FCompl) != (compl0 != 0)), ((b.FSimul != b.FCompl) != (compl1 != 0))
		p.simulation = x && y
		if p.xor {
			p.simulation = x != y
		}
	}
	if trivial != 0 {
		Cut_ListAdd(super, Cut_CutCreateTriv(p, node))
	}
	if a == nil || b == nil || super.count >= limit || p.params.FLocal != 0 && treeCode != 0 {
		return
	}
	var scratch [CUT_SIZE_MAX]int32
	process := func(left, right *Cut_Cut_t) bool {
		leaves := mergeLeaves(left.leaves, right.leaves, p.params.NVarsMax, scratch[:0])
		if leaves == nil {
			return false
		}
		sign := left.USign | right.USign
		// Count the candidate before pruning to preserve ABC's peak count.
		// Allocate its owned leaf and truth storage only when it survives.
		p.countAllocation()
		rejected := p.params.FGlobal != 0 && !p.globalCutEligible(leaves)
		if !rejected && p.params.FFilter != 0 {
			rejected = cutListDominates(super, leaves, sign)
			if !rejected {
				candidate := Cut_Cut_t{leaves: leaves, USign: sign}
				cutRemoveSupersets(p, super, &candidate)
				if frontiers != nil {
					rejected = frontiers.filter(p, &candidate)
					limit = frontiers.limit
				}
			}
		}
		if rejected {
			p.stats.Filtered++
			p.stats.Recycled++
			p.stats.Current--
			return false
		}
		c := newMergedCut(p, leaves, sign)
		c.Num0, c.Num1 = left.Num0&2047, right.Num0&2047
		if p.params.FTruth != 0 {
			Cut_TruthCompute(p, c, left, right, compl0, compl1)
		}
		Cut_ListAdd(super, c)
		return super.count >= limit
	}
	full0, stop0 := cutListBounds(a, p.params.NVarsMax, treeCode&1 != 0)
	full1, stop1 := cutListBounds(b, p.params.NVarsMax, treeCode&2 != 0)
	for left := a; left != full0; left = left.Next {
		for right := b; right != full1; right = right.Next {
			if process(left, right) {
				return
			}
		}
	}
	for left := a; left != full0; left = left.Next {
		for right := full1; right != stop1; right = right.Next {
			if left.USign&^right.USign == 0 && process(left, right) {
				return
			}
		}
	}
	for right := b; right != full1; right = right.Next {
		for left := full0; left != stop0; left = left.Next {
			if right.USign&^left.USign == 0 && process(left, right) {
				return
			}
		}
	}
	for left := full0; left != stop0; left = left.Next {
		for right := full1; right != stop1; right = right.Next {
			if left.USign == right.USign && process(left, right) {
				return
			}
		}
	}
}
func Cut_NodeComputeCuts(p *Cut_Man_t, node, node0, node1, compl0, compl1, trivial, treeCode int32) *Cut_Cut_t {
	p.check()
	if p.newCuts[node] != nil {
		panic("cut: node cuts already computed")
	}
	started := time.Now()
	p.stats.Nodes++
	if p.params.FRecord != 0 {
		Cut_CutNumberList(p.newCuts[node0])
		Cut_CutNumberList(p.newCuts[node1])
	}
	var list Cut_List_t
	Cut_ListStart(&list)
	Cut_NodeDoComputeCuts(p, &list, node, compl0, compl1, p.newCuts[node0], p.newCuts[node1], trivial, treeCode)
	cuts := Cut_ListFinish(&list)
	p.newCuts[node] = cuts
	if p.params.FRecordAig != 0 {
		for c := cuts; c != nil; c = c.Next {
			if len(c.leaves) > 4 {
				p.recorded = append(p.recorded, Cut_TruthRecord{Leaves: append([]int32(nil), c.leaves...), Truth: append([]uint32(nil), c.truth...)})
			}
		}
	}
	p.stats.MergeTime += time.Since(started)
	if cuts == nil {
		p.stats.NoCuts++
	}
	if Cut_CutCountList(cuts) >= p.params.NKeepMax {
		p.stats.Limited++
	}
	if p.params.FRecord != 0 {
		for c := cuts; c != nil; c = c.Next {
			if !(trivial != 0 && len(c.leaves) == 1 && c.leaves[0] == node) {
				kind := uint8(1)
				if p.xor {
					kind = 2
				}
				c.proof = &Cut_Proof{Node: node, Sources: [2]int32{node0, node1}, Phases: [2]bool{compl0 != 0, compl1 != 0}, Kind: kind}
			}
		}
		p.records[node] = cutRecord(cuts)
	}
	if p.params.FMap != 0 && p.params.FSeq == 0 {
		Cut_NodeMapping(p, cuts, node, node0, node1)
	}
	return cuts
}
func Cut_CutListVerify(list *Cut_Cut_t) int32 {
	seen := map[*Cut_Cut_t]bool{}
	var previous *Cut_Cut_t
	for c := list; c != nil; c = c.Next {
		if seen[c] || len(c.leaves) < 1 || int32(len(c.leaves)) > c.NVarsMax {
			return 0
		}
		seen[c] = true
		for i := 1; i < len(c.leaves); i++ {
			if c.leaves[i-1] >= c.leaves[i] {
				return 0
			}
		}
		if previous != nil && len(previous.leaves) > len(c.leaves) {
			return 0
		}
		for dom := list; dom != c; dom = dom.Next {
			if cutContains(c.leaves, dom.leaves) {
				return 0
			}
		}
		previous = c
	}
	return 1
}
func Cut_NodeMapping2(p *Cut_Man_t, cuts *Cut_Cut_t, node, node0, node1 int32) int32 {
	bestDelay := int32(1 << 30)
	var best *Cut_Cut_t
	for c := cuts; c != nil; c = c.Next {
		if len(c.leaves) == 1 {
			continue
		}
		var delay int32
		for _, leaf := range c.leaves {
			delay = max(delay, p.delays[leaf])
		}
		delay++
		if delay < bestDelay {
			bestDelay, best = delay, c
		}
	}
	if best == nil {
		p.stats.MappingDelay = -1
		return -1
	}
	p.delays[node] = bestDelay
	p.best[node] = best
	p.stats.MappingDelay = max(p.stats.MappingDelay, int64(bestDelay))
	return bestDelay
}
func Cut_NodeMapping(p *Cut_Man_t, cuts *Cut_Cut_t, node, node0, node1 int32) int32 {
	d0, d1 := p.delays[node0], p.delays[node1]
	a, b := p.best[node0], p.best[node1]
	if d0 == 0 {
		a = p.newCuts[node0]
	}
	if d1 == 0 {
		b = p.newCuts[node1]
	}
	delay := max(d0, d1)
	if d0 == d1 {
		if d0 == 0 {
			delay = 1
		}
	} else if d0 > d1 {
		b = p.newCuts[node1]
	} else {
		a = p.newCuts[node0]
	}
	if a == nil || b == nil {
		panic("cut: mapping fanin cuts unavailable")
	}
	best := Cut_CutMergeTwo(p, a, b)
	if best == nil {
		delay++
		best = Cut_CutAlloc(p)
		best.setLeaves([]int32{min(node0, node1), max(node0, node1)})
		best.USign = Cut_NodeSign(node0) | Cut_NodeSign(node1)
	}
	p.delays[node] = delay
	p.best[node] = best
	p.stats.MappingDelay = max(p.stats.MappingDelay, int64(delay))
	return delay
}
func Cut_ManMappingArea_rec(p *Cut_Man_t, node int32) int32 {
	cut := p.best[node]
	if cut == nil || len(cut.leaves) == 1 {
		return 0
	}
	delete(p.best, node)
	count := int32(1)
	for _, leaf := range cut.leaves {
		count += Cut_ManMappingArea_rec(p, leaf)
	}
	return count
}
func cutUnionLists(p *Cut_Man_t, lists []*Cut_Cut_t, keepTrivial bool, top *Cut_Cut_t, frontiers *cutSeqFrontiers) *Cut_Cut_t {
	var super Cut_List_t
	Cut_ListStart(&super)
	limit := p.params.NKeepMax
	if frontiers != nil {
		limit = frontiers.limit
	}
	consider := func(c *Cut_Cut_t) {
		if super.count >= limit {
			Cut_CutRecycle(p, c)
			return
		}
		if p.params.FFilter != 0 {
			if Cut_CutFilterOne(p, &super, c) != 0 {
				return
			}
			if frontiers != nil {
				rejected := frontiers.filter(p, c)
				limit = frontiers.limit
				if rejected {
					p.stats.Filtered++
					Cut_CutRecycle(p, c)
					return
				}
			}
		}
		if top != nil {
			c.FCompl = top.FSimul != c.FSimul
		}
		Cut_ListAdd(&super, c)
	}
	// Native choice union considers all sublimit cuts in source-list order
	// before it visits any full-width cuts. Reuse the list of input heads
	// to remember where each full-width suffix starts.
	for i, list := range lists {
		if list != nil && keepTrivial {
			trivial := list
			list = list.Next
			trivial.Next = nil
			if i == 0 {
				if top == nil {
					top = trivial
				}
				Cut_ListAdd(&super, trivial)
			} else {
				Cut_CutRecycle(p, trivial)
			}
		}
		for list != nil && int32(len(list.leaves)) < p.params.NVarsMax {
			next := list.Next
			consider(list)
			list = next
		}
		lists[i] = list
	}
	for _, list := range lists {
		for list != nil {
			next := list.Next
			consider(list)
			list = next
		}
	}
	return Cut_ListFinish(&super)
}
func Cut_NodeUnionCuts(p *Cut_Man_t, nodes []int32) *Cut_Cut_t {
	if len(nodes) == 0 {
		return nil
	}
	lists := make([]*Cut_Cut_t, len(nodes))
	for i, node := range nodes {
		lists[i] = p.newCuts[node]
		delete(p.newCuts, node)
	}
	result := cutUnionLists(p, lists, true, nil, nil)
	p.newCuts[nodes[0]] = result
	if p.params.FRecord != 0 {
		p.records[nodes[0]] = cutRecord(result)
	}
	p.stats.Nodes -= int64(len(nodes) - 1)
	return result
}
func Cut_NodeUnionCutsSeq(p *Cut_Man_t, nodes []int32, set, first int32) *Cut_Cut_t {
	if len(nodes) == 0 {
		return nil
	}
	root := nodes[0]
	frontiers := cutSeqFrontiers{old: p.oldCuts[root], limit: p.params.NKeepMax}
	top := frontiers.old
	if top == nil {
		top = p.newCuts[root]
	}
	if set >= 0 {
		frontiers.current = p.newCuts[root]
	}
	if first == 0 {
		frontiers.limit--
	} // ABC reserves the representative's cut.
	lists := make([]*Cut_Cut_t, len(nodes))
	for i, node := range nodes {
		if i == 0 && set >= 0 {
			lists[i] = p.tempCuts[set]
			delete(p.tempCuts, set)
		} else {
			lists[i] = p.newCuts[node]
			delete(p.newCuts, node)
		}
	}
	result := cutUnionLists(p, lists, first != 0, top, &frontiers)
	p.oldCuts[root] = frontiers.old
	if set >= 0 {
		p.newCuts[root] = frontiers.current
	}
	if set >= 0 {
		p.tempCuts[set] = result
	} else {
		p.newCuts[nodes[0]] = result
	}
	return result
}

func Cut_NodeComputeCutsXor(p *Cut_Man_t, node, node0, node1, compl0, compl1, trivial, tree int32) *Cut_Cut_t {
	old := p.xor
	p.xor = true
	defer func() { p.xor = old }()
	return Cut_NodeComputeCuts(p, node, node0, node1, compl0, compl1, trivial, tree)
}
func Cut_NodeComputeBuffer(p *Cut_Man_t, node, fanin, compl int32) *Cut_Cut_t {
	if p.newCuts[node] != nil {
		panic("cut: buffer cuts already computed")
	}
	if p.newCuts[fanin] == nil {
		panic("cut: buffer fanin cuts unavailable")
	}
	p.simulation = (p.newCuts[fanin].FSimul != p.newCuts[fanin].FCompl) != (compl != 0)
	var super Cut_List_t
	Cut_ListStart(&super)
	Cut_ListAdd(&super, Cut_CutCreateTriv(p, node))
	source := cutSlice(p.newCuts[fanin])
	for i, c := range source {
		if super.count >= p.params.NKeepMax {
			break
		}
		copy := p.copyBufferCut(c)
		if compl != 0 {
			for j := range copy.truth {
				copy.truth[j] = ^copy.truth[j]
			}
			copy.FSimul = !copy.FSimul
		}
		Cut_ListAdd(&super, copy)
		if p.params.FRecord != 0 {
			copy.proof = &Cut_Proof{Node: node, Sources: [2]int32{fanin, 0}, Phases: [2]bool{compl != 0, false}, Kind: 3}
		}
		copy.Num0 = uint16(i) & 2047
	}
	result := Cut_ListFinish(&super)
	p.newCuts[node] = result
	p.stats.Nodes++
	if p.params.FRecord != 0 {
		p.records[node] = cutRecord(result)
	}
	p.delays[node] = p.delays[fanin]
	p.best[node] = p.best[fanin]
	return result
}
