package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecHash.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Hashing integer pairs/triples into an integer.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecHash.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// Hash_IntObj_t holds one stable object. Entry zero is the absent sentinel.
// Pair tables use iData2 as mutable payload; triple tables include it in the key.
type Hash_IntObj_t struct{ iData0, iData1, iData2, iNext int32 }
type Hash_IntMan_t struct {
	vTable  *Vec_Int_t
	objects []*Hash_IntObj_t
	nRefs   int
}

func Hash_IntObj(p *Hash_IntMan_t, i int32) *Hash_IntObj_t {
	if i == 0 {
		return nil
	}
	return p.objects[int(i)]
}
func Hash_IntObjData0(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData0 }
func Hash_IntObjData1(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData1 }
func Hash_IntObjData2(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData2 }

// Inc returns the previous payload; Dec returns the new payload, matching C.
func Hash_Int2ObjInc(p *Hash_IntMan_t, i int32) int32 {
	o := Hash_IntObj(p, i)
	v := o.iData2
	o.iData2++
	return v
}
func Hash_Int2ObjDec(p *Hash_IntMan_t, i int32) int32 {
	o := Hash_IntObj(p, i)
	o.iData2--
	return o.iData2
}
func Hash_Int2ObjSetData2(p *Hash_IntMan_t, i, d int32) { Hash_IntObj(p, i).iData2 = d }
func Hash_IntManStart(n int) *Hash_IntMan_t {
	if n < 0 || n > int(^uint32(0)>>1)-100 {
		panic("invalid integer hash capacity")
	}
	n += 100
	return &Hash_IntMan_t{vTable: Vec_IntStart(hashIntPrime(n)), objects: make([]*Hash_IntObj_t, 1, n), nRefs: 1}
}
func Hash_IntManStop(p *Hash_IntMan_t) {
	if p == nil {
		return
	}
	Vec_IntFree(p.vTable)
	p.vTable = nil
	p.objects = nil
	p.nRefs = 0
}
func Hash_IntManRef(p *Hash_IntMan_t) *Hash_IntMan_t {
	if p == nil || p.nRefs <= 0 {
		panic("reference to released integer hash")
	}
	p.nRefs++
	return p
}
func Hash_IntManDeref(p *Hash_IntMan_t) {
	if p == nil {
		return
	}
	if p.nRefs <= 0 {
		panic("unbalanced integer hash dereference")
	}
	p.nRefs--
	if p.nRefs == 0 {
		Hash_IntManStop(p)
	}
}
func Hash_IntManEntryNum(p *Hash_IntMan_t) int {
	if len(p.objects) == 0 {
		return 0
	}
	return len(p.objects) - 1
}
func Hash_IntManBucketLengths(p *Hash_IntMan_t) []int {
	result := make([]int, Vec_IntSize(p.vTable))
	for i, head := range Vec_IntArray(p.vTable) {
		for id := head; id != 0; id = p.objects[int(id)].iNext {
			result[i]++
		}
	}
	return result
}
func Hash_IntManProfile(p *Hash_IntMan_t) {
	for _, n := range Hash_IntManBucketLengths(p) {
		fmt.Printf("%d ", n)
	}
	fmt.Println()
}

// The explicit uint32 operations preserve the original unsigned C wraparound.
func Hash_Int2ManHash(a, b int32, size int) int {
	if size <= 0 {
		panic("invalid hash table size")
	}
	return int((uint32(4177)*uint32(a) + uint32(7873)*uint32(b)) % uint32(size))
}
func Hsh_Int3ManHash(a, b, c int32, size int) int {
	if size <= 0 {
		panic("invalid hash table size")
	}
	return int((uint32(4177)*uint32(a) + uint32(7873)*uint32(b) + uint32(1699)*uint32(c)) % uint32(size))
}
func hashIntLive(p *Hash_IntMan_t) {
	if p == nil || p.nRefs <= 0 {
		panic("use of released integer hash")
	}
}

// Lookup returns the actual bucket/chain link. Zero denotes absence; the link
// remains valid until table resize or manager release (as in the C API).
func Hash_Int2ManLookup(p *Hash_IntMan_t, a, b int32) *int32 {
	hashIntLive(p)
	place := Vec_IntEntryP(p.vTable, Hash_Int2ManHash(a, b, Vec_IntSize(p.vTable)))
	for *place != 0 {
		o := Hash_IntObj(p, *place)
		if o.iData0 == a && o.iData1 == b {
			return place
		}
		place = &o.iNext
	}
	return place
}
func Hsh_Int3ManLookup(p *Hash_IntMan_t, a, b, c int32) *int32 {
	hashIntLive(p)
	place := Vec_IntEntryP(p.vTable, Hsh_Int3ManHash(a, b, c, Vec_IntSize(p.vTable)))
	for *place != 0 {
		o := Hash_IntObj(p, *place)
		if o.iData0 == a && o.iData1 == b && o.iData2 == c {
			return place
		}
		place = &o.iNext
	}
	return place
}
func hashIntResize(p *Hash_IntMan_t, mode int) {
	if len(p.objects) <= Vec_IntSize(p.vTable) {
		return
	}
	Vec_IntFill(p.vTable, hashIntPrime(2*Vec_IntSize(p.vTable)), 0)
	for i := 1; i < len(p.objects); i++ {
		o := p.objects[i]
		o.iNext = 0
		var place *int32
		if mode == 2 {
			place = Hash_Int2ManLookup(p, o.iData0, o.iData1)
		} else {
			place = Hsh_Int3ManLookup(p, o.iData0, o.iData1, o.iData2)
		}
		if *place != 0 {
			panic("duplicate key during integer hash resize")
		}
		*place = int32(i)
	}
}
func Hash_Int2ManInsert(p *Hash_IntMan_t, a, b, data int32) int32 {
	hashIntLive(p)
	hashIntResize(p, 2)
	place := Hash_Int2ManLookup(p, a, b)
	if *place != 0 {
		return *place
	}
	if len(p.objects) > int(^uint32(0)>>1) {
		panic("integer hash id overflow")
	}
	id := int32(len(p.objects))
	p.objects = append(p.objects, &Hash_IntObj_t{iData0: a, iData1: b, iData2: data})
	*place = id
	return id
}
func Hsh_Int3ManInsert(p *Hash_IntMan_t, a, b, c int32) int32 {
	hashIntLive(p)
	hashIntResize(p, 3)
	place := Hsh_Int3ManLookup(p, a, b, c)
	if *place != 0 {
		return *place
	}
	if len(p.objects) > int(^uint32(0)>>1) {
		panic("integer hash id overflow")
	}
	id := int32(len(p.objects))
	p.objects = append(p.objects, &Hash_IntObj_t{iData0: a, iData1: b, iData2: c})
	*place = id
	return id
}
func Hash_IntManHashArrayTest() {
	p := Hash_IntManStart(10)
	defer Hash_IntManStop(p)
	a := Hash_Int2ManInsert(p, 10, 11, 12)
	b := Hash_Int2ManInsert(p, 20, 21, 22)
	if a == 0 || b == 0 || Hash_Int2ManInsert(p, 10, 11, 12) != a {
		panic("integer hash self-test failed")
	}
}
func hashIntPrime(n int) int {
	if n < 3 {
		return 3
	}
	if n%2 == 0 {
		n++
	}
	for {
		prime := true
		for d := 3; d <= n/d; d += 2 {
			if n%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			return n
		}
		n += 2
	}
}
