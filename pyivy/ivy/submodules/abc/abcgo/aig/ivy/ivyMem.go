// Node memory ownership ported from ABC ivyMem.c (Alan Mishchenko).
package ivy

const IVY_PAGE_SIZE = 12
const IVY_PAGE_MASK = (1 << IVY_PAGE_SIZE) - 1

func Ivy_ManStartMemory(p *Ivy_Man_t) {
	if p == nil || p.memoryStopped {
		panic("Ivy memory cannot be restarted")
	}
	if len(p.pages) != 0 {
		panic("Ivy memory already started")
	}
	p.pages = nil
	p.freeNodes = nil
	p.memoryStopped = false
}
func Ivy_ManStopMemory(p *Ivy_Man_t) {
	if p == nil || p.memoryStopped {
		return
	}
	for _, page := range p.pages {
		for i := range page {
			node := &page[i]
			neg := node.neg
			*node = Ivy_Obj_t{Id: -1, Type: IVY_NONE, neg: neg, recycled: true}
		}
	}
	p.pages = nil
	p.freeNodes = nil
	p.memoryStopped = true
}
func Ivy_ManAddMemory(p *Ivy_Man_t) {
	if p == nil || p.memoryStopped {
		panic("Ivy memory is stopped")
	}
	if len(p.freeNodes) != 0 {
		panic("unused Ivy node slots remain")
	}
	page := make([]Ivy_Obj_t, 1<<IVY_PAGE_SIZE)
	p.pages = append(p.pages, page)
	for i := len(page) - 1; i >= 0; i-- {
		page[i].owner = p
		page[i].recycled = true
		p.freeNodes = append(p.freeNodes, &page[i])
	}
}
func Ivy_ManFetchMemory(p *Ivy_Man_t) *Ivy_Obj_t {
	if p == nil || p.memoryStopped {
		panic("Ivy memory is stopped")
	}
	if len(p.freeNodes) == 0 {
		Ivy_ManAddMemory(p)
	}
	last := len(p.freeNodes) - 1
	node := p.freeNodes[last]
	p.freeNodes[last] = nil
	p.freeNodes = p.freeNodes[:last]
	neg := node.neg
	*node = Ivy_Obj_t{owner: p, neg: neg}
	return node
}
func Ivy_ManRecycleMemory(p *Ivy_Man_t, node *Ivy_Obj_t) {
	if p == nil || node == nil || Ivy_IsComplement(node) != 0 || node.owner != p || node == &p.Ghost || node == p.pConst1 || node.recycled || p.memoryStopped || node.nRefs != 0 {
		panic("invalid Ivy node recycle")
	}
	if node.Id >= 0 && int64(node.Id) < int64(len(p.vObjs)) && p.vObjs[node.Id] == node {
		panic("cannot recycle a published Ivy object")
	}
	neg := node.neg
	*node = Ivy_Obj_t{Type: IVY_NONE, owner: p, neg: neg, recycled: true}
	p.freeNodes = append(p.freeNodes, node)
}
