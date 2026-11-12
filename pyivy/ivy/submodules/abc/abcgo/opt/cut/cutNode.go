package cut

import (
	"sort"
	"time"
)

func Cut_CutCheckDominance(dom, cut *Cut_Cut_t) int32 {
	if cutContains(cut.leaves, dom.leaves) {
		return 1
	}
	return 0
}
func Cut_CutFilterOne(p *Cut_Man_t, list *Cut_List_t, c *Cut_Cut_t) int32 {
	for size := 1; size <= len(c.leaves); size++ {
		for old := list.head[size]; old != nil; old = old.Next {
			if cutContains(c.leaves, old.leaves) {
				p.stats.Filtered++
				Cut_CutRecycle(p, c)
				return 1
			}
		}
	}
	for size := len(c.leaves) + 1; size <= int(p.params.NVarsMax); size++ {
		slot := &list.head[size]
		for *slot != nil {
			old := *slot
			if cutContains(old.leaves, c.leaves) {
				*slot = old.Next
				p.stats.Filtered++
				Cut_CutRecycle(p, old)
			} else {
				slot = &old.Next
			}
		}
		list.tail[size] = slot
	}
	return 0
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
func Cut_CutFilterOld(p *Cut_Man_t, list, c *Cut_Cut_t) int32 {
	for old := list; old != nil; old = old.Next {
		if cutContains(c.leaves, old.leaves) {
			p.stats.Filtered++
			Cut_CutRecycle(p, c)
			return 1
		}
	}
	return 0
}
func Cut_CutFilterGlobal(p *Cut_Man_t, c *Cut_Cut_t) int32 {
	if len(c.leaves) == 1 {
		return 0
	}
	for _, leaf := range c.leaves {
		node := leaf
		if p.params.FSeq != 0 {
			node >>= CUT_SHIFT
		}
		if int(node) < len(p.attrs) && p.attrs[node] != 0 {
			return 0
		}
	}
	p.stats.Filtered++
	Cut_CutRecycle(p, c)
	return 1
}
func cutListCount(list *Cut_List_t) int32 {
	var count int32
	for _, head := range list.head {
		count += Cut_CutCountList(head)
	}
	return count
}
func cutListParts(list *Cut_Cut_t, max int32) (small, large []*Cut_Cut_t) {
	for c := list; c != nil; c = c.Next {
		if int32(len(c.leaves)) == max {
			large = append(large, c)
		} else {
			small = append(small, c)
		}
	}
	return
}
func Cut_NodeDoComputeCuts(p *Cut_Man_t, super *Cut_List_t, node, compl0, compl1 int32, a, b *Cut_Cut_t, trivial, treeCode int32) {
	limit := p.params.NKeepMax
	if p.budget >= 0 {
		limit = p.budget
	}
	if cutListCount(super) >= limit {
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
	if a == nil || b == nil || p.params.FLocal != 0 && treeCode != 0 {
		return
	}
	small0, large0 := cutListParts(a, p.params.NVarsMax)
	small1, large1 := cutListParts(b, p.params.NVarsMax)
	if treeCode&1 != 0 {
		if len(a.leaves) != 1 {
			panic("cut: tree boundary lacks trivial cut")
		}
		small0 = []*Cut_Cut_t{a}
		large0 = nil
	}
	if treeCode&2 != 0 {
		if len(b.leaves) != 1 {
			panic("cut: tree boundary lacks trivial cut")
		}
		small1 = []*Cut_Cut_t{b}
		large1 = nil
	}
	for _, pair := range []struct{ a, b []*Cut_Cut_t }{{small0, small1}, {small0, large1}, {large0, small1}, {large0, large1}} {
		for _, left := range pair.a {
			for _, right := range pair.b {
				limit := p.params.NKeepMax
				if p.budget >= 0 {
					limit = p.budget
				}
				if cutListCount(super) >= limit {
					return
				}
				c := Cut_CutMergeTwo(p, left, right)
				if c == nil {
					continue
				}
				c.Num0, c.Num1 = left.Num0&2047, right.Num0&2047
				if p.params.FGlobal != 0 && Cut_CutFilterGlobal(p, c) != 0 {
					continue
				}
				if p.params.FFilter != 0 && Cut_CutFilterOne(p, super, c) != 0 {
					continue
				}
				if p.params.FTruth != 0 {
					Cut_TruthCompute(p, c, left, right, compl0, compl1)
				}
				Cut_ListAdd(super, c)
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
		best.leaves = []int32{min(node0, node1), max(node0, node1)}
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
func cutUnionLists(p *Cut_Man_t, lists []*Cut_Cut_t, keepTrivial bool, old, new *Cut_Cut_t) *Cut_Cut_t {
	var super Cut_List_t
	Cut_ListStart(&super)
	var candidates []*Cut_Cut_t
	var top *Cut_Cut_t
	for i, list := range lists {
		if list == nil {
			continue
		}
		start := list
		if keepTrivial {
			start = list.Next
			list.Next = nil
			if i == 0 {
				top = list
				Cut_ListAdd(&super, list)
			} else {
				Cut_CutRecycle(p, list)
			}
		}
		for c := start; c != nil; {
			next := c.Next
			c.Next = nil
			candidates = append(candidates, c)
			c = next
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return len(candidates[i].leaves) < len(candidates[j].leaves) })
	for _, c := range candidates {
		if cutListCount(&super) >= p.params.NKeepMax {
			Cut_CutRecycle(p, c)
			continue
		}
		if p.params.FFilter != 0 {
			if Cut_CutFilterOne(p, &super, c) != 0 {
				continue
			}
			if Cut_CutFilterOld(p, old, c) != 0 {
				continue
			}
			if Cut_CutFilterOld(p, new, c) != 0 {
				continue
			}
		}
		if top != nil {
			c.FCompl = top.FSimul != c.FSimul
		}
		Cut_ListAdd(&super, c)
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
	result := cutUnionLists(p, lists, first != 0, p.oldCuts[nodes[0]], nil)
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
		if cutListCount(&super) >= p.params.NKeepMax {
			break
		}
		copy := Cut_CutDupList(p, &Cut_Cut_t{NVarsMax: c.NVarsMax, leaves: c.leaves, truth: c.truth, FSimul: c.FSimul, FCompl: c.FCompl, USign: c.USign})
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
