// Basic traversal/buffer helpers ported from ABC ivyUtil.c.
package ivy

func Ivy_ManCleanTravId(p *Ivy_Man_t) {
	p.nTravIds = 1
	for _, node := range p.vObjs {
		if node != nil {
			node.TravId = 0
		}
	}
}
func Ivy_ManIncrementTravId(p *Ivy_Man_t) {
	if p.nTravIds >= (1<<30)-1-1000 {
		Ivy_ManCleanTravId(p)
	}
	p.nTravIds++
}
func Ivy_ObjReal(edge *Ivy_Obj_t) *Ivy_Obj_t {
	seen := map[*Ivy_Obj_t]bool{}
	for edge != nil && Ivy_ObjIsBuf(edge) != 0 {
		node := Ivy_Regular(edge)
		if seen[node] {
			panic("buffer cycle")
		}
		seen[node] = true
		edge = Ivy_NotCond(node.pFanin0, Ivy_IsComplement(edge))
	}
	return edge
}
