package fraig

// fraigNextFanoutPlace returns the link used for this fanin's fanout chain.
func fraigNextFanoutPlace(fanin, fanout *Fraig_Node_t) **Fraig_Node_t {
	if Fraig_Regular(fanout.p1) == fanin {
		return &fanout.pFanFanin1
	}
	assert(Fraig_Regular(fanout.p2) == fanin, "fanout contains fanin")
	return &fanout.pFanFanin2
}

func Fraig_NodeAddFaninFanout(fanin, fanout *Fraig_Node_t) {
	assert(!Fraig_IsComplement(fanin) && !Fraig_IsComplement(fanout), "nodes must be regular")
	assert(Fraig_Regular(fanout.p1) == fanin || Fraig_Regular(fanout.p2) == fanin, "fanout contains fanin")
	if fanin.pFanPivot == nil {
		fanin.pFanPivot = fanout
		return
	}
	pivot := fraigNextFanoutPlace(fanin, fanin.pFanPivot)
	next := fraigNextFanoutPlace(fanin, fanout)
	*next = *pivot
	*pivot = fanout
}

func Fraig_NodeRemoveFaninFanout(fanin, removed *Fraig_Node_t) {
	link := &fanin.pFanPivot
	for current := fanin.pFanPivot; current != nil; {
		next := *fraigNextFanoutPlace(fanin, current)
		if current != removed {
			*link = current
			link = fraigNextFanoutPlace(fanin, current)
		}
		current = next
	}
	*link = nil
}

func Fraig_NodeTransferFanout(from, to *Fraig_Node_t) {
	assert(to.pFanPivot == nil && to.pFanFanin1 == nil && to.pFanFanin2 == nil, "destination has no fanouts")
	for fanout := from.pFanPivot; fanout != nil; {
		next := *fraigNextFanoutPlace(from, fanout)
		if Fraig_Regular(fanout.p1) == from {
			fanout.p1 = Fraig_NotCond(to, Fraig_IsComplement(fanout.p1))
		} else if Fraig_Regular(fanout.p2) == from {
			fanout.p2 = Fraig_NotCond(to, Fraig_IsComplement(fanout.p2))
		}
		fanout = next
	}
	to.pFanPivot, to.pFanFanin1, to.pFanFanin2 = from.pFanPivot, from.pFanFanin1, from.pFanFanin2
	from.pFanPivot, from.pFanFanin1, from.pFanFanin2 = nil, nil, nil
}

func Fraig_NodeGetFanoutNum(node *Fraig_Node_t) int32 {
	var count int32
	for fanout := node.pFanPivot; fanout != nil; fanout = *fraigNextFanoutPlace(node, fanout) {
		count++
	}
	return count
}
