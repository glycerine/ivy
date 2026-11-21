package cut

func checkCutShift(list *Cut_Cut_t, latches int32) {
	for c := list; c != nil; c = c.Next {
		for _, leaf := range c.leaves {
			count := leaf & CUT_MASK
			if latches < -count || latches > CUT_MASK-count {
				panic("cut: sequential latch counter overflow")
			}
		}
	}
}

func shiftCutLeaves(list *Cut_Cut_t, latches int32) {
	for c := list; c != nil; c = c.Next {
		c.USign = 0
		for i, leaf := range c.leaves {
			c.leaves[i] = leaf + latches
			c.USign |= Cut_NodeSign(c.leaves[i])
		}
	}
}

func Cut_NodeShiftCutLeaves(list *Cut_Cut_t, latches int32) {
	checkCutShift(list, latches)
	shiftCutLeaves(list, latches)
}

type cutSeqInputs struct {
	old, current *Cut_Cut_t
	latches      int32
	copied       bool
}

func cutListsOverlap(lists ...*Cut_Cut_t) bool {
	// Two acyclic lists share a suffix exactly when they end at the same cut.
	for i, tail := range lists {
		if tail == nil {
			continue
		}
		for tail.Next != nil {
			tail = tail.Next
		}
		for _, previous := range lists[:i] {
			if previous == tail {
				return true
			}
		}
		lists[i] = tail
	}
	return false
}

func prepareSeqInputs(p *Cut_Man_t, old, current *Cut_Cut_t, latches int32, aliases bool) cutSeqInputs {
	if latches != 0 {
		checkCutShift(old, latches)
		checkCutShift(current, latches)
	}
	if aliases {
		old, current = Cut_CutDupList(p, old), Cut_CutDupList(p, current)
	}
	if latches != 0 {
		shiftCutLeaves(old, latches)
		shiftCutLeaves(current, latches)
	}
	return cutSeqInputs{old, current, latches, aliases}
}

func (inputs cutSeqInputs) restore(p *Cut_Man_t) {
	if inputs.copied {
		Cut_CutRecycleList(p, inputs.old)
		Cut_CutRecycleList(p, inputs.current)
	} else if inputs.latches != 0 {
		shiftCutLeaves(inputs.old, -inputs.latches)
		shiftCutLeaves(inputs.current, -inputs.latches)
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
	if set >= 0 && p.tempCuts[set] != nil {
		panic("cut: temporary sequential cuts already present")
	}
	if set < 0 && p.newCuts[node] != nil {
		panic("cut: new sequential cuts already present")
	}
	if oldCount == 0 {
		p.stats.Nodes++
	}
	aliases := node == node0 || node == node1 || node0 == node1
	old0, current0 := p.oldCuts[node0], p.newCuts[node0]
	old1, current1 := p.oldCuts[node1], p.newCuts[node1]
	if !aliases && (lat0 != 0 || lat1 != 0) {
		aliases = cutListsOverlap(old0, current0, old1, current1, p.oldCuts[node], p.newCuts[node])
	}
	a := prepareSeqInputs(p, old0, current0, lat0, aliases)
	defer a.restore(p)
	b := prepareSeqInputs(p, old1, current1, lat1, aliases)
	defer b.restore(p)
	frontiers := cutSeqFrontiers{old: p.oldCuts[node], current: p.newCuts[node], limit: p.params.NKeepMax - oldCount}
	var list Cut_List_t
	Cut_ListStart(&list)
	cutNodeDoComputeCuts(p, &list, node, compl0, compl1, a.old, b.current, 0, 0, &frontiers)
	cutNodeDoComputeCuts(p, &list, node, compl0, compl1, a.current, b.old, 0, 0, &frontiers)
	cutNodeDoComputeCuts(p, &list, node, compl0, compl1, a.current, b.current, trivial, 0, &frontiers)
	result := Cut_ListFinish(&list)
	p.oldCuts[node], p.newCuts[node] = frontiers.old, frontiers.current
	if set >= 0 {
		p.tempCuts[set] = result
	} else {
		p.newCuts[node] = result
	}
	if list.count >= frontiers.limit {
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
