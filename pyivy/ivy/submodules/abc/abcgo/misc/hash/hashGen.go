// Port of src/misc/hash/hashGen.h. Keys and values are Go-owned; the C
// fFreeKey ownership switch has no effect on garbage-collected storage.
package hash

import "strings"

type Hash_GenHashFunction_t func(any, int32) int32
type Hash_GenCompFunction_t func(any, any) int32
type Hash_Gen_Entry_t struct {
	key   any
	data  any
	pNext *Hash_Gen_Entry_t
}
type Hash_Gen_t struct {
	nSize, nBins int32
	fHash        Hash_GenHashFunction_t
	fComp        Hash_GenCompFunction_t
	fFreeKey     int32
	pArray       []*Hash_Gen_Entry_t
}

func Hash_DefaultHashFuncStr(key any, nBins int32) int32 {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	var h uint32
	for _, b := range []byte(key.(string)) {
		if b == 0 {
			break
		}
		h = h*6 + uint32(b)
	}
	return int32(h % uint32(nBins))
}

// C string keys end at the first NUL, including during comparison.
func hashCString(s string) string {
	if i := strings.IndexByte(s, 0); i >= 0 {
		return s[:i]
	}
	return s
}
func Hash_DefaultCmpFuncStr(a, b any) int32 {
	return int32(strings.Compare(hashCString(a.(string)), hashCString(b.(string))))
}
func hashInteger(key any) int64 {
	switch v := key.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case uintptr:
		return int64(v)
	default:
		panic("hash: integer key required")
	}
}
func Hash_DefaultHashFuncInt(key any, nBins int32) int32 {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	h := hashInteger(key) % int64(nBins)
	if h < 0 {
		h += int64(nBins)
	}
	return int32(h)
}
func Hash_DefaultCmpFuncInt(a, b any) int32 {
	x, y := hashInteger(a), hashInteger(b)
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
func Hash_GenAlloc(nBins int32, hash Hash_GenHashFunction_t, compare Hash_GenCompFunction_t, freeKey int32) *Hash_Gen_t {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	if hash == nil {
		hash = Hash_DefaultHashFuncStr
	}
	if compare == nil {
		compare = Hash_DefaultCmpFuncStr
	}
	return &Hash_Gen_t{nBins: nBins, fHash: hash, fComp: compare, fFreeKey: freeKey, pArray: make([]*Hash_Gen_Entry_t, nBins)}
}
func Hash_GenEntry(p *Hash_Gen_t, key any, create int32) *Hash_Gen_Entry_t {
	bin := p.fHash(key, p.nBins)
	link := &p.pArray[bin]
	for *link != nil {
		if p.fComp((*link).key, key) == 0 {
			return *link
		}
		link = &(*link).pNext
	}
	if create == 0 {
		return nil
	}
	*link = &Hash_Gen_Entry_t{key: key}
	p.nSize++
	return *link
}
func Hash_GenExists(p *Hash_Gen_t, key any) int32 {
	if Hash_GenEntry(p, key, 0) != nil {
		return 1
	}
	return 0
}
func Hash_GenWriteEntry(p *Hash_Gen_t, key, data any) { Hash_GenEntry(p, key, 1).data = data }
func Hash_GenRemove(p *Hash_Gen_t, key any) any {
	bin := p.fHash(key, p.nBins)
	link := &p.pArray[bin]
	for *link != nil {
		e := *link
		if p.fComp(e.key, key) == 0 {
			*link = e.pNext
			p.nSize--
			return e.data
		}
		link = &e.pNext
	}
	return nil
}
func Hash_GenFree(p *Hash_Gen_t) {
	if p != nil {
		p.pArray = nil
		p.nSize = 0
		p.nBins = 0
	}
}
func (p *Hash_Gen_t) Len() int                 { return int(p.nSize) }
func (e *Hash_Gen_Entry_t) Key() any           { return e.key }
func (e *Hash_Gen_Entry_t) Value() any         { return e.data }
func (e *Hash_Gen_Entry_t) SetValue(value any) { e.data = value }
func (p *Hash_Gen_t) Range(visit func(*Hash_Gen_Entry_t) bool) {
	for _, e := range p.pArray {
		for ; e != nil; e = e.pNext {
			if !visit(e) {
				return
			}
		}
	}
}
