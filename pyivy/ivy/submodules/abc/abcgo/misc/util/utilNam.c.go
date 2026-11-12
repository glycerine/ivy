package util

// String interning from ABC utilNam.c: ID zero is unused; names are retained
// in a NUL-terminated byte arena and chained hash bins preserve insertion IDs.
import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Abc_Nam_t struct {
	store               []byte
	handles, next, bins []int32
	refs                int32
	buffer              []byte
}

func Abc_NamStart(n, average int32) *Abc_Nam_t {
	if n == 0 {
		n = 16
	}
	capacity := ((n*(average+1) + 16) / 4) * 4
	if capacity < 4 {
		capacity = 4
	}
	return &Abc_Nam_t{store: make([]byte, 4, capacity), handles: []int32{-1}, next: []int32{-1}, bins: make([]int32, Abc_PrimeCudd(uint32(n))), refs: 1}
}
func Abc_NamStop(p *Abc_Nam_t) {
	if p != nil {
		*p = Abc_Nam_t{}
	}
}
func Abc_NamRef(p *Abc_Nam_t) *Abc_Nam_t { p.refs++; return p }
func Abc_NamDeref(p *Abc_Nam_t) {
	if p == nil {
		return
	}
	p.refs--
	if p.refs == 0 {
		Abc_NamStop(p)
	}
}
func Abc_NamHandleToStr(p *Abc_Nam_t, h int32) string {
	b := p.store[h:]
	end := 0
	for end < len(b) && b[end] != 0 {
		end++
	}
	return string(b[:end])
}
func Abc_NamIntToHandle(p *Abc_Nam_t, i int32) int32 { return p.handles[i] }
func Abc_NamIntToStr(p *Abc_Nam_t, i int32) string   { return Abc_NamHandleToStr(p, p.handles[i]) }
func Abc_NamIntToNext(p *Abc_Nam_t, i int32) int32   { return p.next[i] }
func Abc_NamIntToNextP(p *Abc_Nam_t, i int32) *int32 { return &p.next[i] }
func Abc_NamPrint(p *Abc_Nam_t, name string) {
	f := os.Stdout
	if name != "" {
		var err error
		f, err = os.Create(name)
		if err != nil {
			Abc_Print(1, "Could not open file %s\n", name)
			return
		}
		defer f.Close()
	}
	for i := 1; i < len(p.handles); i++ {
		fmt.Fprintf(f, "%8d = %s\n", i, Abc_NamIntToStr(p, int32(i)))
	}
}
func Abc_NamObjNumMax(p *Abc_Nam_t) int32 { return int32(len(p.handles)) }
func Abc_NamMemUsed(p *Abc_Nam_t) int32 {
	if p == nil {
		return 0
	}
	return int32(reflect.TypeOf(*p).Size()) + int32(len(p.store)+4*(len(p.handles)+len(p.next)+len(p.bins)))
}
func Abc_NamMemAlloc(p *Abc_Nam_t) int32 {
	if p == nil {
		return 0
	}
	return int32(reflect.TypeOf(*p).Size()) + int32(cap(p.store)+4*(cap(p.handles)+cap(p.next)+cap(p.bins)))
}

var nameHashPrimes = [128]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

// A negative limit selects the whole C string; nonnegative limits are byte
// indices into s, replacing C's pointer-to-end arguments.
func namLimited(s string, limit int) string {
	if limit >= 0 {
		s = s[:limit]
	}
	if end := strings.IndexByte(s, 0); end >= 0 {
		s = s[:end]
	}
	return s
}
func Abc_NamStrHash(s string, limit int, size int32) int32 {
	s = namLimited(s, limit)
	var h uint32
	for i := range len(s) {
		term := uint32(int32(int8(s[i]))) * nameHashPrimes[i&127]
		if i&1 != 0 {
			h *= term
		} else {
			h ^= term
		}
	}
	return int32(h % uint32(size))
}
func Abc_NamStrHash2(s string, limit int, size int32) int32 {
	s = namLimited(s, limit)
	var h uint32
	for i := range len(s) {
		h += uint32(int32(int8(s[i])))
		h += h << 10
		h ^= h >> 6
	}
	h += h << 3
	h ^= h >> 11
	h += h << 15
	return int32(h % uint32(size))
}
func Abc_NamStrcmp(s string, limit int, other string) int32 {
	return int32(strings.Compare(namLimited(s, limit), namLimited(other, -1)))
}
func Abc_NamStrHashFind(p *Abc_Nam_t, s string, limit int) *int32 {
	s = namLimited(s, limit)
	if s == "" {
		panic("name must be nonempty")
	}
	place := &p.bins[Abc_NamStrHash(s, -1, int32(len(p.bins)))]
	for *place != 0 {
		if Abc_NamIntToStr(p, *place) == s {
			break
		}
		place = &p.next[*place]
	}
	return place
}
func Abc_NamStrHashResize(p *Abc_Nam_t) {
	p.bins = make([]int32, Abc_PrimeCudd(uint32(3*len(p.bins))))
	clear(p.next)
	for i := 1; i < len(p.handles); i++ {
		place := Abc_NamStrHashFind(p, Abc_NamIntToStr(p, int32(i)), -1)
		*place = int32(i)
	}
}
func Abc_NamStrFind(p *Abc_Nam_t, s string) int32 { return *Abc_NamStrHashFind(p, s, -1) }
func Abc_NamStrFindLim(p *Abc_Nam_t, s string, limit int) int32 {
	return *Abc_NamStrHashFind(p, s, limit)
}
func Abc_NamStrFindOrAdd(p *Abc_Nam_t, s string, found *int32) int32 {
	return Abc_NamStrFindOrAddLim(p, s, -1, found)
}
func Abc_NamStrFindOrAddLim(p *Abc_Nam_t, s string, limit int, found *int32) int32 {
	s = namLimited(s, limit)
	place := Abc_NamStrHashFind(p, s, -1)
	if *place != 0 {
		if found != nil {
			*found = 1
		}
		return *place
	}
	if found != nil {
		*found = 0
	}
	id := int32(len(p.handles))
	*place = id
	handle := int32(len(p.store))
	p.store = append(p.store, []byte(s)...)
	p.store = append(p.store, 0)
	p.handles = append(p.handles, handle)
	p.next = append(p.next, 0)
	if len(p.handles) > 2*len(p.bins) {
		Abc_NamStrHashResize(p)
	}
	return id
}
func Abc_NamStrFindOrAddF(p *Abc_Nam_t, format string, args ...any) int32 {
	return Abc_NamStrFindOrAdd(p, fmt.Sprintf(format, args...), nil)
}
func Abc_NamStr(p *Abc_Nam_t, id int32) string {
	if id <= 0 {
		return ""
	}
	return Abc_NamIntToStr(p, id)
}
func Abc_NamBuffer(p *Abc_Nam_t) *[]byte { p.buffer = p.buffer[:0]; return &p.buffer }
func Abc_NamComputeIdMap(a, b *Abc_Nam_t) []int32 {
	result := make([]int32, len(a.handles))
	for i := 1; i < len(result); i++ {
		result[i] = Abc_NamStrFind(b, Abc_NamIntToStr(a, int32(i)))
	}
	return result
}
func Abc_NamReportCommon(ids []int32, a, b *Abc_Nam_t) int32 {
	var count int32
	for _, id := range ids {
		if Abc_NamStrFind(b, Abc_NamStr(a, id)) > 0 {
			count++
		}
	}
	return count
}
func Abc_NamReportUnique(ids []int32, a, b *Abc_Nam_t) string {
	for _, id := range ids {
		s := Abc_NamStr(a, id)
		if Abc_NamStrFind(b, s) == 0 {
			return s
		}
	}
	return ""
}
