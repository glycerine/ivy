// Port of src/misc/hash/hashPtr.h. Entries keep stable addresses.
package hash

type Hash_Ptr_Entry_t struct {
	key   int32
	data  any
	pNext *Hash_Ptr_Entry_t
}
type Hash_Ptr_t struct {
	nSize, nBins int32
	fHash        func(int32, int32) int32
	pArray       []*Hash_Ptr_Entry_t
}

func Hash_PtrAlloc(nBins int32) *Hash_Ptr_t {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	return &Hash_Ptr_t{nBins: nBins, fHash: Hash_DefaultHashFunc, pArray: make([]*Hash_Ptr_Entry_t, nBins)}
}
func (p *Hash_Ptr_t) find(key int32, create bool) *Hash_Ptr_Entry_t {
	bin := p.fHash(key, p.nBins)
	link := &p.pArray[bin]
	for *link != nil {
		if (*link).key == key {
			return *link
		}
		link = &(*link).pNext
	}
	if !create {
		return nil
	}
	*link = &Hash_Ptr_Entry_t{key: key}
	p.nSize++
	return *link
}
func Hash_PtrExists(p *Hash_Ptr_t, key int32) int32 {
	if p.find(key, false) != nil {
		return 1
	}
	return 0
}
func Hash_PtrWriteEntry(p *Hash_Ptr_t, key int32, data any) { p.find(key, true).data = data }
func Hash_PtrEntry(p *Hash_Ptr_t, key, create int32) any {
	if e := p.find(key, create != 0); e != nil {
		return e.data
	}
	return nil
}
func Hash_PtrEntryPtr(p *Hash_Ptr_t, key int32) *any { return &p.find(key, true).data }
func Hash_PtrFree(p *Hash_Ptr_t) {
	if p != nil {
		p.pArray = nil
		p.nSize = 0
		p.nBins = 0
	}
}
func (p *Hash_Ptr_t) Len() int { return int(p.nSize) }

// Range visits in original bin/chain order. Return false to stop.
func (p *Hash_Ptr_t) Range(visit func(int32, *any) bool) {
	for _, e := range p.pArray {
		for ; e != nil; e = e.pNext {
			if !visit(e.key, &e.data) {
				return
			}
		}
	}
}
func Hash_PtrRemove(p *Hash_Ptr_t, key int32) any {
	bin := p.fHash(key, p.nBins)
	link := &p.pArray[bin]
	for *link != nil {
		e := *link
		if e.key == key {
			*link = e.pNext
			p.nSize--
			data := e.data
			return data
		}
		link = &e.pNext
	}
	return nil
}
