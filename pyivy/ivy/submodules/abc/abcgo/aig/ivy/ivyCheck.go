// Graph consistency helpers ported from ABC ivyCheck.c.
package ivy

// Ivy_ManCheckFanoutNums counts dangling logic nodes, as in the C API.
func Ivy_ManCheckFanoutNums(p *Ivy_Man_t) int32 {
	var count int32
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsNode(n) != 0 && n.nRefs == 0 {
			count++
		}
	}
	return count
}

// Ivy_ManCheckFanouts checks both directions and each intrusive previous link.
// Malformed/cyclic lists return zero instead of looping or panicking.
func Ivy_ManCheckFanouts(p *Ivy_Man_t) int32 {
	if p.fFanout == 0 {
		return 1
	}
	expected := map[*Ivy_Obj_t]map[*Ivy_Obj_t]bool{}
	for _, node := range p.vObjs {
		if node == nil {
			continue
		}
		for _, edge := range []*Ivy_Obj_t{node.pFanin0, node.pFanin1} {
			if edge == nil {
				continue
			}
			fanin := Ivy_Regular(edge)
			if expected[fanin] == nil {
				expected[fanin] = map[*Ivy_Obj_t]bool{}
			}
			expected[fanin][node] = true
		}
	}
	for _, fanin := range p.vObjs {
		if fanin == nil {
			continue
		}
		seen := map[*Ivy_Obj_t]bool{}
		var prev *Ivy_Obj_t
		for node := fanin.pFanout; node != nil; {
			if seen[node] || node.owner != p || node.recycled || !expected[fanin][node] {
				return 0
			}
			seen[node] = true
			var next, back *Ivy_Obj_t
			if Ivy_Regular(node.pFanin0) == fanin {
				next, back = node.pNextFan0, node.pPrevFan0
			} else if Ivy_Regular(node.pFanin1) == fanin {
				next, back = node.pNextFan1, node.pPrevFan1
			} else {
				return 0
			}
			if back != prev {
				return 0
			}
			prev, node = node, next
		}
		if len(seen) != len(expected[fanin]) || int32(len(seen)) != fanin.nRefs {
			return 0
		}
	}
	return 1
}

func Ivy_ManCheckChoices(p *Ivy_Man_t) int32 {
	if p == nil || p.pHaig == nil {
		return 1
	}
	history := p.pHaig
	for _, root := range history.vObjs {
		if root == nil || root.nRefs == 0 {
			continue
		}
		if Ivy_IsComplement(root.pEquiv) != 0 {
			return 0
		}
		seen := map[*Ivy_Obj_t]bool{root: true}
		for n := root.pEquiv; n != nil && n != root; n = Ivy_Regular(n.pEquiv) {
			if Ivy_IsComplement(n) != 0 || n.owner != history || n.recycled || seen[n] || n.nRefs > 1 {
				return 0
			}
			seen[n] = true
		}
	}
	return 1
}
