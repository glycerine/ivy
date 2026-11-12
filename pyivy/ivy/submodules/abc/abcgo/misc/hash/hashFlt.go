// Port of src/misc/hash/hashFlt.h. Entries keep stable addresses.
package hash

type Hash_Flt_Entry_t struct {
	key   int32
	data  float32
	pNext *Hash_Flt_Entry_t
}
type Hash_Flt_t struct {
	nSize, nBins int32
	fHash        func(int32, int32) int32
	pArray       []*Hash_Flt_Entry_t
}

func Hash_FltAlloc(nBins int32) *Hash_Flt_t {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	return &Hash_Flt_t{nBins: nBins, fHash: Hash_DefaultHashFunc, pArray: make([]*Hash_Flt_Entry_t, nBins)}
}
func (p *Hash_Flt_t) find(key int32, create bool) *Hash_Flt_Entry_t {
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
	*link = &Hash_Flt_Entry_t{key: key}
	p.nSize++
	return *link
}
func Hash_FltExists(p *Hash_Flt_t, key int32) int32 {
	if p.find(key, false) != nil {
		return 1
	}
	return 0
}
func Hash_FltWriteEntry(p *Hash_Flt_t, key int32, data float32) { p.find(key, true).data = data }
func Hash_FltEntry(p *Hash_Flt_t, key, create int32) float32 {
	if e := p.find(key, create != 0); e != nil {
		return e.data
	}
	return 0
}
func Hash_FltEntryPtr(p *Hash_Flt_t, key int32) *float32 { return &p.find(key, true).data }
func Hash_FltFree(p *Hash_Flt_t) {
	if p != nil {
		p.pArray = nil
		p.nSize = 0
		p.nBins = 0
	}
}
func (p *Hash_Flt_t) Len() int { return int(p.nSize) }

// Range visits in original bin/chain order. Return false to stop.
func (p *Hash_Flt_t) Range(visit func(int32, *float32) bool) {
	for _, e := range p.pArray {
		for ; e != nil; e = e.pNext {
			if !visit(e.key, &e.data) {
				return
			}
		}
	}
}
func Hash_FltRemove(p *Hash_Flt_t, key int32) {
	bin := p.fHash(key, p.nBins)
	link := &p.pArray[bin]
	for *link != nil {
		e := *link
		if e.key == key {
			*link = e.pNext
			p.nSize--
			return
		}
		link = &e.pNext
	}

}
