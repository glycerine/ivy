// Node memory ownership ported from ABC ivyMem.c (Alan Mishchenko).
package ivy

const IVY_PAGE_SIZE = 12
const IVY_PAGE_MASK = (1 << IVY_PAGE_SIZE) - 1
const ivyBlockBits = 7
const ivyBlockSize = 1 << ivyBlockBits

type ivyMemoryPage [1 << (IVY_PAGE_SIZE - ivyBlockBits)]*[ivyBlockSize]Ivy_Obj_t

// Physical blocks are issued on demand within each native 4,096-slot page.
// Only previously issued slots need entries in the recycle stack.
func (p *Ivy_Man_t) ownsSlot(n *Ivy_Obj_t) bool {
	slot := n.poolSlot
	if uint64(slot) >= p.nextSlot {
		return false
	}
	page := &p.pages[slot>>IVY_PAGE_SIZE]
	block := page[(slot&IVY_PAGE_MASK)>>ivyBlockBits]
	return &block[slot&(ivyBlockSize-1)] == n
}

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
	for i := range p.pages {
		for _, block := range p.pages[i] {
			if block == nil {
				break
			}
			for j := range block {
				node := &block[j]
				*node = Ivy_Obj_t{Id: -1, neg: node.neg, recycled: true}
			}
		}
	}
	p.pages = nil
	p.freeNodes = nil
	p.nextSlot = 0
	p.memoryStopped = true
}
func Ivy_ManAddMemory(p *Ivy_Man_t) {
	if p == nil || p.memoryStopped {
		panic("Ivy memory is stopped")
	}
	if len(p.freeNodes) != 0 || int(p.nextSlot>>IVY_PAGE_SIZE) != len(p.pages) {
		panic("unused Ivy node slots remain")
	}
	if len(p.pages) == 1<<(32-IVY_PAGE_SIZE) {
		panic("Ivy slot index exhausted")
	}
	p.pages = append(p.pages, ivyMemoryPage{})
}
func Ivy_ManFetchMemory(p *Ivy_Man_t) *Ivy_Obj_t {
	if p == nil || p.memoryStopped {
		panic("Ivy memory is stopped")
	}
	if last := len(p.freeNodes) - 1; last >= 0 {
		node := p.freeNodes[last]
		p.freeNodes[last] = nil
		p.freeNodes = p.freeNodes[:last]
		*node = Ivy_Obj_t{owner: p, neg: node.neg, poolSlot: node.poolSlot}
		return node
	}
	slot := p.nextSlot
	page := slot >> IVY_PAGE_SIZE
	if int(page) == len(p.pages) {
		Ivy_ManAddMemory(p)
	}
	block := &p.pages[page][(slot&IVY_PAGE_MASK)>>ivyBlockBits]
	if *block == nil {
		*block = new([ivyBlockSize]Ivy_Obj_t)
	}
	node := &(*block)[slot&(ivyBlockSize-1)]
	node.owner, node.poolSlot = p, uint32(slot)
	p.nextSlot++
	return node
}
func Ivy_ManRecycleMemory(p *Ivy_Man_t, node *Ivy_Obj_t) {
	if p == nil || node == nil || Ivy_IsComplement(node) != 0 || node.owner != p || node == &p.Ghost || node == p.pConst1 || node.recycled || p.memoryStopped || node.nRefs != 0 {
		panic("invalid Ivy node recycle")
	}
	if !p.ownsSlot(node) {
		panic("ivy: object is not a slot in this manager")
	}
	if node.Id >= 0 && int64(node.Id) < int64(len(p.vObjs)) && p.vObjs[node.Id] == node {
		panic("cannot recycle a published Ivy object")
	}
	*node = Ivy_Obj_t{owner: p, neg: node.neg, poolSlot: node.poolSlot, recycled: true}
	p.freeNodes = append(p.freeNodes, node)
}
