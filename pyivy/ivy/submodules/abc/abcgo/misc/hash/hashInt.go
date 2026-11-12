// Port of src/misc/hash/hashInt.h. Entries keep stable addresses.
package hash

type Hash_Int_Entry_t struct {
	key   int32
	data  int32
	pNext *Hash_Int_Entry_t
}
type Hash_Int_t struct {
	nSize, nBins int32
	fHash        func(int32, int32) int32
	pArray       []*Hash_Int_Entry_t
}

func Hash_IntAlloc(nBins int32) *Hash_Int_t {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	return &Hash_Int_t{nBins: nBins, fHash: Hash_DefaultHashFunc, pArray: make([]*Hash_Int_Entry_t, nBins)}
}
func (p *Hash_Int_t) find(key int32, create bool) *Hash_Int_Entry_t {
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
	*link = &Hash_Int_Entry_t{key: key}
	p.nSize++
	return *link
}
func Hash_IntExists(p *Hash_Int_t, key int32) int32 {
	if p.find(key, false) != nil {
		return 1
	}
	return 0
}
func Hash_IntWriteEntry(p *Hash_Int_t, key int32, data int32) { p.find(key, true).data = data }
func Hash_IntEntry(p *Hash_Int_t, key, create int32) int32 {
	if e := p.find(key, create != 0); e != nil {
		return e.data
	}
	return 0
}
func Hash_IntEntryPtr(p *Hash_Int_t, key int32) *int32 { return &p.find(key, true).data }
func Hash_IntFree(p *Hash_Int_t) {
	if p != nil {
		p.pArray = nil
		p.nSize = 0
		p.nBins = 0
	}
}
func (p *Hash_Int_t) Len() int { return int(p.nSize) }

// Range visits in original bin/chain order. Return false to stop.
func (p *Hash_Int_t) Range(visit func(int32, *int32) bool) {
	for _, e := range p.pArray {
		for ; e != nil; e = e.pNext {
			if !visit(e.key, &e.data) {
				return
			}
		}
	}
}
