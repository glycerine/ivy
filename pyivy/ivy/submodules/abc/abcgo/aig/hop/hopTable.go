// Structural hashing, ported from ABC hopTable.c (Alan Mishchenko).
package hop

import "fmt"

func hopEdgeKey(o *Hop_Obj_t) uint64 {
	if o == nil {
		return 0
	}
	return uint64(uint32(Hop_Regular(o).Id))<<1 | uint64(Hop_IsComplement(o))
}
func Hop_Hash(o *Hop_Obj_t, size int32) uint64 {
	return (hopEdgeKey(o.pFanin0)*7937 + hopEdgeKey(o.pFanin1)*2971 + uint64(o.Type)*911) % uint64(size)
}
func Hop_TableFind(p *Hop_Man_t, o *Hop_Obj_t) **Hop_Obj_t {
	slot := &p.pTable[Hop_Hash(o, p.nTableSize)]
	for *slot != nil {
		n := *slot
		if n.Type == o.Type && n.pFanin0 == o.pFanin0 && n.pFanin1 == o.pFanin1 {
			break
		}
		slot = &n.pNext
	}
	return slot
}
func Hop_TableLookup(p *Hop_Man_t, o *Hop_Obj_t) *Hop_Obj_t { return *Hop_TableFind(p, o) }
func Hop_TableInsert(p *Hop_Man_t, o *Hop_Obj_t) {
	slot := Hop_TableFind(p, o)
	assert(*slot == nil, "unique structural node")
	*slot = o
	o.pNext = nil
	if Hop_ManNodeNum(p) > 2*p.nTableSize {
		Hop_TableResize(p)
	}
}
func Hop_TableDelete(p *Hop_Man_t, o *Hop_Obj_t) {
	slot := Hop_TableFind(p, o)
	assert(*slot == o, "node belongs to hash table")
	*slot = o.pNext
	o.pNext = nil
}
func Hop_TableCountEntries(p *Hop_Man_t) int32 {
	var n int32
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			n++
		}
	}
	return n
}
func Hop_TableResize(p *Hop_Man_t) {
	old := p.pTable
	p.nTableSize = p.nTableSize*2 + 1
	p.pTable = make([]*Hop_Obj_t, p.nTableSize)
	for _, head := range old {
		for o := head; o != nil; {
			next := o.pNext
			i := Hop_Hash(o, p.nTableSize)
			o.pNext = p.pTable[i]
			p.pTable[i] = o
			o = next
		}
	}
}
func Hop_TableProfile(p *Hop_Man_t) {
	for i, head := range p.pTable {
		n := 0
		for o := head; o != nil; o = o.pNext {
			n++
		}
		fmt.Printf("%d: %d\n", i, n)
	}
}
