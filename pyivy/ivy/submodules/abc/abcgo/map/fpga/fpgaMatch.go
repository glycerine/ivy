package fpga

func interrupted(p *Fpga_Man_t) bool {
	if p.Context != nil {
		select {
		case <-p.Context.Done():
			p.LastError = p.Context.Err()
			return true
		default:
		}
	}
	return false
}
func Fpga_MappingMatches(p *Fpga_Man_t, delay int32) int32 {
	for _, n := range p.Ands.Nodes {
		if interrupted(p) {
			return 0
		}
		if Fpga_NodeIsAnd(n) != 0 && n.Repr == nil {
			if Fpga_MatchNode(p, n, delay) == 0 {
				return 0
			}
		}
	}
	return 1
}
func Fpga_MatchNode(p *Fpga_Man_t, n *Fpga_Node_t, delay int32) int32 {
	if n.EstimatedFanouts < 0 {
		n.EstimatedFanouts = float32(n.Refs)
	} else {
		n.EstimatedFanouts = (2*n.EstimatedFanouts + float32(n.Refs)) / 3
	}
	var best *Fpga_Cut_t
	for c := n.Cuts.Next; c != nil; c = c.Next {
		Fpga_CutGetParameters(p, c)
		if delay == 0 && c.Arrival > n.Required+p.Epsilon {
			continue
		}
		if best == nil || delay != 0 && (c.Arrival < best.Arrival-p.Epsilon || abs(c.Arrival-best.Arrival) <= p.Epsilon && c.Flow < best.Flow-p.Epsilon) || delay == 0 && (c.Flow < best.Flow-p.Epsilon || abs(c.Flow-best.Flow) <= p.Epsilon && c.Arrival < best.Arrival-p.Epsilon) {
			best = c
		}
	}
	if best == nil {
		if n.Best != nil {
			return 1
		}
		return 0
	}
	n.Old = n.Best
	n.Best = best
	return 1
}
func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
func Fpga_MappingMatchesArea(p *Fpga_Man_t) int32 { return Fpga_MappingMatchesAreaArray(p, p.Ands) }
func Fpga_MappingMatchesAreaArray(p *Fpga_Man_t, v *Fpga_NodeVec_t) int32 {
	for _, n := range v.Nodes {
		if interrupted(p) {
			return 0
		}
		if Fpga_NodeIsAnd(n) != 0 && n.Repr == nil {
			Fpga_MatchNodeArea(p, n)
		}
	}
	return 1
}
func recoverNode(p *Fpga_Man_t, n *Fpga_Node_t, switching bool) int32 {
	old := n.Best
	referenced := n.Refs > 0
	if referenced && switching {
		Fpga_CutDerefSwitch(p, n, old, 0)
	} else if referenced {
		Fpga_CutDeref(p, n, old, 0)
	}
	best := old
	bestCost := float32(FPGA_FLOAT_LARGE)
	for c := n.Cuts.Next; c != nil; c = c.Next {
		Fpga_TimeCutComputeArrival(p, c)
		if c.Arrival > n.Required+p.Epsilon {
			continue
		}
		var cost float32
		if switching {
			cost = Fpga_CutGetSwitchDerefed(p, n, c)
		} else {
			cost = Fpga_CutGetAreaDerefed(p, c)
		}
		c.Flow = cost
		if cost < bestCost-p.Epsilon || abs(cost-bestCost) <= p.Epsilon && c.Arrival < best.Arrival-p.Epsilon {
			best = c
			bestCost = cost
		}
	}
	n.Old = old
	n.Best = best
	if referenced && switching {
		best.Flow = Fpga_CutRefSwitch(p, n, best, 0)
	} else if referenced {
		best.Flow = Fpga_CutRef(p, n, best, 0)
	}
	return 1
}
func Fpga_MatchNodeArea(p *Fpga_Man_t, n *Fpga_Node_t) int32 { return recoverNode(p, n, false) }
func Fpga_MappingMatchesSwitch(p *Fpga_Man_t) int32 {
	for _, n := range p.Ands.Nodes {
		if interrupted(p) {
			return 0
		}
		if Fpga_NodeIsAnd(n) != 0 && n.Repr == nil {
			Fpga_MatchNodeSwitch(p, n)
		}
	}
	return 1
}
func Fpga_MatchNodeSwitch(p *Fpga_Man_t, n *Fpga_Node_t) int32 { return recoverNode(p, n, true) }
func Fpga_FindBestNode(p *Fpga_Man_t, v *Fpga_NodeVec_t, node **Fpga_Node_t, cut **Fpga_Cut_t) float32 {
	best := float32(-FPGA_FLOAT_LARGE)
	*node = nil
	*cut = nil
	for _, n := range v.Nodes {
		if Fpga_NodeIsAnd(n) == 0 || n.Refs == 0 {
			continue
		}
		old := n.Best
		a := Fpga_CutGetAreaRefed(p, old)
		Fpga_CutDeref(p, n, old, 0)
		for c := n.Cuts.Next; c != nil; c = c.Next {
			Fpga_TimeCutComputeArrival(p, c)
			if c.Arrival > n.Required+p.Epsilon {
				continue
			}
			gain := a - Fpga_CutGetAreaDerefed(p, c)
			if gain > best {
				best = gain
				*node = n
				*cut = c
			}
		}
		Fpga_CutRef(p, n, old, 0)
	}
	return best
}
