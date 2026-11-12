package mapper

// The C build enables explicit fanout links only under MAP_ALLOCATE_FANOUT.
// Go keeps the optional relation as a compact adjacency list.
func Map_NodeAddFaninFanout(fanin, fanout *Map_Node_t) {
	fanin, fanout = Map_Regular(fanin), Map_Regular(fanout)
	if fanin == nil || fanout == nil {
		return
	}
	for _, existing := range fanin.fanouts {
		if existing == fanout {
			return
		}
	}
	fanin.fanouts = append(fanin.fanouts, fanout)
}

func Map_NodeRemoveFaninFanout(fanin, remove *Map_Node_t) {
	fanin, remove = Map_Regular(fanin), Map_Regular(remove)
	if fanin == nil || remove == nil {
		return
	}
	for i, fanout := range fanin.fanouts {
		if fanout == remove {
			copy(fanin.fanouts[i:], fanin.fanouts[i+1:])
			fanin.fanouts[len(fanin.fanouts)-1] = nil
			fanin.fanouts = fanin.fanouts[:len(fanin.fanouts)-1]
			return
		}
	}
}

func Map_NodeGetFanoutNum(node *Map_Node_t) int32 {
	if node == nil {
		return 0
	}
	return int32(len(Map_Regular(node).fanouts))
}
