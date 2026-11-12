// Intrusive fanout lists ported from ABC ivyFanout.c (Alan Mishchenko).
package ivy

func Ivy_ObjNextFanoutPlace(node, fanout *Ivy_Obj_t) **Ivy_Obj_t {
	if Ivy_ObjFanin0(fanout) == node {
		return &fanout.pNextFan0
	}
	if Ivy_ObjFanin1(fanout) == node {
		return &fanout.pNextFan1
	}
	panic("fanout does not read this object")
}
func Ivy_ObjPrevFanoutPlace(node, fanout *Ivy_Obj_t) **Ivy_Obj_t {
	if Ivy_ObjFanin0(fanout) == node {
		return &fanout.pPrevFan0
	}
	if Ivy_ObjFanin1(fanout) == node {
		return &fanout.pPrevFan1
	}
	panic("fanout does not read this object")
}
func Ivy_ObjNextFanout(node, fanout *Ivy_Obj_t) *Ivy_Obj_t {
	if fanout == nil {
		return nil
	}
	return *Ivy_ObjNextFanoutPlace(node, fanout)
}
func Ivy_ObjPrevFanout(node, fanout *Ivy_Obj_t) *Ivy_Obj_t {
	if fanout == nil {
		return nil
	}
	return *Ivy_ObjPrevFanoutPlace(node, fanout)
}
func Ivy_ObjPrevNextFanoutPlace(node, fanout *Ivy_Obj_t) **Ivy_Obj_t {
	previous := Ivy_ObjPrevFanout(node, fanout)
	if previous == nil {
		return &node.pFanout
	}
	return Ivy_ObjNextFanoutPlace(node, previous)
}
func Ivy_ObjNextPrevFanoutPlace(node, fanout *Ivy_Obj_t) **Ivy_Obj_t {
	next := Ivy_ObjNextFanout(node, fanout)
	if next == nil {
		return nil
	}
	return Ivy_ObjPrevFanoutPlace(node, next)
}
func Ivy_ManStartFanout(p *Ivy_Man_t) {
	if p.fFanout != 0 {
		panic("fanouts already enabled")
	}
	p.fFanout = 1
	for _, o := range p.vObjs {
		if o == nil {
			continue
		}
		if a := Ivy_ObjFanin0(o); a != nil {
			Ivy_ObjAddFanout(p, a, o)
		}
		if b := Ivy_ObjFanin1(o); b != nil {
			Ivy_ObjAddFanout(p, b, o)
		}
	}
}
func Ivy_ManStopFanout(p *Ivy_Man_t) {
	if p.fFanout == 0 {
		panic("fanouts not enabled")
	}
	p.fFanout = 0
	for _, o := range p.vObjs {
		if o != nil {
			o.pFanout = nil
			o.pNextFan0 = nil
			o.pNextFan1 = nil
			o.pPrevFan0 = nil
			o.pPrevFan1 = nil
		}
	}
}
func Ivy_ObjAddFanout(p *Ivy_Man_t, node, fanout *Ivy_Obj_t) {
	if p.fFanout == 0 || node.owner != p || fanout.owner != p {
		panic("invalid fanout ownership/mode")
	}
	next, prev := Ivy_ObjNextFanoutPlace(node, fanout), Ivy_ObjPrevFanoutPlace(node, fanout)
	if *next != nil || *prev != nil || node.pFanout == fanout {
		panic("fanout already linked")
	}
	*next = node.pFanout
	if node.pFanout != nil {
		*Ivy_ObjPrevFanoutPlace(node, node.pFanout) = fanout
	}
	node.pFanout = fanout
}
func Ivy_ObjDeleteFanout(p *Ivy_Man_t, node, fanout *Ivy_Obj_t) {
	if p.fFanout == 0 || node.owner != p || fanout.owner != p {
		panic("invalid fanout ownership/mode")
	}
	place := Ivy_ObjPrevNextFanoutPlace(node, fanout)
	if *place != fanout {
		panic("fanout absent")
	}
	next, prev := Ivy_ObjNextFanoutPlace(node, fanout), Ivy_ObjPrevFanoutPlace(node, fanout)
	*place = *next
	if following := Ivy_ObjNextPrevFanoutPlace(node, fanout); following != nil {
		if *following != fanout {
			panic("fanout back-link mismatch")
		}
		*following = *prev
	}
	*next = nil
	*prev = nil
}
func Ivy_ObjPatchFanout(p *Ivy_Man_t, node, old, next *Ivy_Obj_t) {
	if p.fFanout == 0 {
		panic("fanouts not enabled")
	}
	place := Ivy_ObjPrevNextFanoutPlace(node, old)
	if *place != old {
		panic("fanout absent")
	}
	*place = next
	if following := Ivy_ObjNextPrevFanoutPlace(node, old); following != nil {
		if *following != old {
			panic("fanout back-link mismatch")
		}
		*following = next
	}
}
func Ivy_ObjCollectFanouts(p *Ivy_Man_t, node *Ivy_Obj_t, result *[]*Ivy_Obj_t) {
	if p.fFanout == 0 || Ivy_IsComplement(node) != 0 {
		panic("regular object with fanouts required")
	}
	*result = (*result)[:0]
	seen := map[*Ivy_Obj_t]bool{}
	for fan := node.pFanout; fan != nil; fan = Ivy_ObjNextFanout(node, fan) {
		if seen[fan] {
			panic("fanout cycle")
		}
		seen[fan] = true
		*result = append(*result, fan)
	}
}
func Ivy_ObjReadFirstFanout(p *Ivy_Man_t, node *Ivy_Obj_t) *Ivy_Obj_t { return node.pFanout }
func Ivy_ObjFanoutNum(p *Ivy_Man_t, node *Ivy_Obj_t) int32 {
	var list []*Ivy_Obj_t
	Ivy_ObjCollectFanouts(p, node, &list)
	return int32(len(list))
}
