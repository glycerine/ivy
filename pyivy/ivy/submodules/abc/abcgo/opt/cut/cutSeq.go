package cut

func Cut_NodeShiftCutLeaves(list *Cut_Cut_t, latches int32) {
	for c := list; c != nil; c = c.Next {
		c.USign = 0
		for i, leaf := range c.leaves {
			count := (leaf & CUT_MASK) + latches
			if count < 0 || count > CUT_MASK {
				panic("cut: sequential latch counter overflow")
			}
			c.leaves[i] = (leaf &^ CUT_MASK) | count
			c.USign |= Cut_NodeSign(c.leaves[i])
		}
	}
}
func Cut_NodeComputeCutsSeq(p *Cut_Man_t, node, node0, node1, compl0, compl1, lat0, lat1, trivial, set int32) {
	p.check()
	if p.params.FSeq == 0 {
		panic("cut: sequential computation requires sequential manager")
	}
	oldCount := Cut_CutCountList(p.oldCuts[node])
	if oldCount >= p.params.NKeepMax {
		return
	}
	if oldCount == 0 {
		p.stats.Nodes++
	}
	aOld, aNew := Cut_CutDupList(p, p.oldCuts[node0]), Cut_CutDupList(p, p.newCuts[node0])
	bOld, bNew := Cut_CutDupList(p, p.oldCuts[node1]), Cut_CutDupList(p, p.newCuts[node1])
	defer func() {
		Cut_CutRecycleList(p, aOld)
		Cut_CutRecycleList(p, aNew)
		Cut_CutRecycleList(p, bOld)
		Cut_CutRecycleList(p, bNew)
	}()
	Cut_NodeShiftCutLeaves(aOld, lat0)
	Cut_NodeShiftCutLeaves(aNew, lat0)
	Cut_NodeShiftCutLeaves(bOld, lat1)
	Cut_NodeShiftCutLeaves(bNew, lat1)
	previousBudget := p.budget
	p.budget = p.params.NKeepMax - oldCount
	defer func() { p.budget = previousBudget }()
	var list Cut_List_t
	Cut_ListStart(&list)
	Cut_NodeDoComputeCuts(p, &list, node, compl0, compl1, aOld, bNew, 0, 0)
	Cut_NodeDoComputeCuts(p, &list, node, compl0, compl1, aNew, bOld, 0, 0)
	Cut_NodeDoComputeCuts(p, &list, node, compl0, compl1, aNew, bNew, trivial, 0)
	candidates := Cut_ListFinish(&list)
	var result *Cut_Cut_t
	tail := &result
	for c := candidates; c != nil; {
		next := c.Next
		c.Next = nil
		if Cut_CutFilterOld(p, p.oldCuts[node], c) == 0 {
			if Cut_CutFilterOld(p, p.newCuts[node], c) == 0 {
				*tail = c
				tail = &c.Next
			}
		}
		c = next
	}
	if set >= 0 {
		if p.tempCuts[set] != nil {
			panic("cut: temporary sequential cuts already present")
		}
		p.tempCuts[set] = result
	} else {
		if p.newCuts[node] != nil {
			panic("cut: new sequential cuts already present")
		}
		p.newCuts[node] = result
	}
	if oldCount+Cut_CutCountList(result) >= p.params.NKeepMax {
		p.stats.Limited++
	}
}
func Cut_NodeNewMergeWithOld(p *Cut_Man_t, node int32) {
	list := p.newCuts[node]
	if list == nil {
		return
	}
	delete(p.newCuts, node)
	p.oldCuts[node] = Cut_CutMergeLists(p.oldCuts[node], list)
}
func Cut_NodeTempTransferToNew(p *Cut_Man_t, node, set int32) int32 {
	list := p.tempCuts[set]
	delete(p.tempCuts, set)
	p.newCuts[node] = list
	if list != nil {
		return 1
	}
	return 0
}
func Cut_NodeOldTransferToNew(p *Cut_Man_t, node int32) {
	p.newCuts[node] = p.oldCuts[node]
	delete(p.oldCuts, node)
}
