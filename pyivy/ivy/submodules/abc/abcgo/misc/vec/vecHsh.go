// Integer and vector interning, ported from ABC vecHsh.h (Alan Mishchenko).
package vec

import (
	"encoding/binary"
	"fmt"
	"slices"
)

type Hsh_Int1Man_t struct{ vData, vNext, vTable *Vec_Int_t }
type Hsh_IntObj_t struct{ iData, iNext int32 }
type Hsh_IntMan_t struct {
	nSize         int
	vData, vTable *Vec_Int_t
	objs          []*Hsh_IntObj_t
}
type Hsh_VecObj_t struct {
	nSize  int
	iNext  int32
	pArray []int32
}
type Hsh_VecMan_t struct {
	vTable                *Vec_Int_t
	objs                  []*Hsh_VecObj_t
	vTemp, vTemp1, vTemp2 Vec_Int_t
}

func hshPrime(n int) int {
	if n < 2 {
		return 2
	}
	for {
		prime := true
		for d := 2; d <= n/d; d++ {
			if n%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			return n
		}
		n++
	}
}
func hshJenkinsBytes(data []int32) uint32 {
	var hash uint32
	var bytes [4]byte
	for _, word := range data {
		binary.NativeEndian.PutUint32(bytes[:], uint32(word))
		for _, b := range bytes {
			hash += uint32(b)
			hash += hash << 10
			hash ^= hash >> 6
		}
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15
	return hash
}
func Hsh_IntData(p *Hsh_IntMan_t, iData int) []int32 {
	if iData < 0 || p.nSize <= 0 || iData >= Vec_IntSize(p.vData)/p.nSize {
		panic("vec: hash record out of bounds")
	}
	return Vec_IntArray(p.vData)[p.nSize*iData : p.nSize*(iData+1)]
}
func Hsh_IntObj(p *Hsh_IntMan_t, iObj int32) *Hsh_IntObj_t {
	if iObj == -1 {
		return nil
	}
	return p.objs[iObj]
}

// Explicit byte conversion replaces the C union without unsafe pointer punning.
func Hsh_IntWord(data, next int32) uint64 {
	var bytes [8]byte
	binary.NativeEndian.PutUint32(bytes[:4], uint32(data))
	binary.NativeEndian.PutUint32(bytes[4:], uint32(next))
	return binary.NativeEndian.Uint64(bytes[:])
}
func Hsh_VecObj(p *Hsh_VecMan_t, i int32) *Hsh_VecObj_t {
	if i == -1 {
		return nil
	}
	return p.objs[i]
}

func Hsh_Int1ManStart(entries int) *Hsh_Int1Man_t {
	if entries < 0 {
		panic("vec: negative hash capacity")
	}
	return &Hsh_Int1Man_t{Vec_IntAlloc(entries), Vec_IntAlloc(entries), Vec_IntStartFull(hshPrime(entries))}
}
func Hsh_Int1ManStop(p *Hsh_Int1Man_t) {
	if p != nil {
		Vec_IntFree(p.vData)
		Vec_IntFree(p.vNext)
		Vec_IntFree(p.vTable)
		*p = Hsh_Int1Man_t{}
	}
}
func Hsh_Int1ManEntryNum(p *Hsh_Int1Man_t) int { return Vec_IntSize(p.vData) }
func Hsh_Int1ManHash(data int32, tableSize int) int {
	if tableSize <= 0 {
		panic("vec: invalid hash table size")
	}
	return int(hshJenkinsBytes([]int32{data}) % uint32(tableSize))
}
func Hsh_Int1ManLookupInt(p *Hsh_Int1Man_t, data int32) *int32 {
	place := Vec_IntEntryP(p.vTable, Hsh_Int1ManHash(data, Vec_IntSize(p.vTable)))
	for *place >= 0 {
		item := int(*place)
		if Vec_IntEntry(p.vData, item) == data {
			return place
		}
		place = Vec_IntEntryP(p.vNext, item)
	}
	return place
}
func Hsh_Int1ManLookup(p *Hsh_Int1Man_t, data int32) int32 { return *Hsh_Int1ManLookupInt(p, data) }
func Hsh_Int1ManAdd(p *Hsh_Int1Man_t, data int32) int32 {
	if Vec_IntSize(p.vData) > Vec_IntSize(p.vTable) {
		Vec_IntFill(p.vTable, hshPrime(2*Vec_IntSize(p.vTable)), -1)
		for i, value := range Vec_IntArray(p.vData) {
			place := Vec_IntEntryP(p.vTable, Hsh_Int1ManHash(value, Vec_IntSize(p.vTable)))
			Vec_IntWriteEntry(p.vNext, i, *place)
			*place = int32(i)
		}
	}
	place := Hsh_Int1ManLookupInt(p, data)
	if *place >= 0 {
		return *place
	}
	id := int32(Vec_IntSize(p.vData))
	*place = id
	Vec_IntPush(p.vData, data)
	Vec_IntPush(p.vNext, -1)
	return id
}

func Hsh_IntManStart(data *Vec_Int_t, size, entries int) *Hsh_IntMan_t {
	if data == nil || size <= 0 || entries < 0 {
		panic("vec: invalid fixed-record hash dimensions")
	}
	return &Hsh_IntMan_t{nSize: size, vData: data, vTable: Vec_IntStartFull(hshPrime(entries)), objs: make([]*Hsh_IntObj_t, 0, entries)}
}
func Hsh_IntManStop(p *Hsh_IntMan_t) {
	if p != nil {
		Vec_IntFree(p.vTable)
		*p = Hsh_IntMan_t{}
	}
}
func Hsh_IntManEntryNum(p *Hsh_IntMan_t) int { return len(p.objs) }
func Hsh_IntManHash(data []int32, size, tableSize int) int {
	if size < 0 || size > len(data) || tableSize <= 0 {
		panic("vec: invalid hash dimensions")
	}
	return int(hshJenkinsBytes(data[:size]) % uint32(tableSize))
}
func Hsh_IntManHash2(data []int32, size, tableSize int) int {
	if size < 0 || size > len(data) || tableSize <= 0 {
		panic("vec: invalid hash dimensions")
	}
	primes := [7]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var key uint32
	var bytes [4]byte
	for i, word := range data[:size] {
		binary.NativeEndian.PutUint32(bytes[:], uint32(word))
		for j, b := range bytes {
			key += uint32(b) * primes[(4*i+j)%7]
		}
	}
	return int(key % uint32(tableSize))
}
func Hsh_IntManLookup(p *Hsh_IntMan_t, data []int32) *int32 {
	if len(data) < p.nSize {
		panic("vec: incomplete fixed hash record")
	}
	place := Vec_IntEntryP(p.vTable, Hsh_IntManHash(data, p.nSize, Vec_IntSize(p.vTable)))
	for *place >= 0 {
		o := p.objs[*place]
		if slices.Equal(data[:p.nSize], Hsh_IntData(p, int(o.iData))) {
			return place
		}
		place = &o.iNext
	}
	return place
}
func Hsh_IntManAdd(p *Hsh_IntMan_t, iData int) int32 {
	if len(p.objs) > Vec_IntSize(p.vTable) {
		Vec_IntFill(p.vTable, hshPrime(2*Vec_IntSize(p.vTable)), -1)
		for i, o := range p.objs {
			place := Vec_IntEntryP(p.vTable, Hsh_IntManHash(Hsh_IntData(p, int(o.iData)), p.nSize, Vec_IntSize(p.vTable)))
			o.iNext = *place
			*place = int32(i)
		}
	}
	data := Hsh_IntData(p, iData)
	place := Hsh_IntManLookup(p, data)
	if *place >= 0 {
		return *place
	}
	id := int32(len(p.objs))
	*place = id
	p.objs = append(p.objs, &Hsh_IntObj_t{iData: int32(iData), iNext: -1})
	return id
}
func Hsh_IntManHashArray(data *Vec_Int_t, size int) *Vec_Int_t {
	if size <= 0 || Vec_IntSize(data)%size != 0 {
		panic("vec: incomplete fixed hash records")
	}
	entries := Vec_IntSize(data) / size
	p := Hsh_IntManStart(data, size, entries)
	result := Vec_IntAlloc(entries)
	for i := 0; i < entries; i++ {
		Vec_IntPush(result, Hsh_IntManAdd(p, i))
	}
	Hsh_IntManStop(p)
	return result
}
func hshWordsAsInts(data *Vec_Wrd_t) *Vec_Int_t {
	result := Vec_IntAlloc(2 * Vec_WrdSize(data))
	var bytes [8]byte
	for _, word := range Vec_WrdArray(data) {
		binary.NativeEndian.PutUint64(bytes[:], word)
		Vec_IntPush(result, int32(binary.NativeEndian.Uint32(bytes[:4])))
		Vec_IntPush(result, int32(binary.NativeEndian.Uint32(bytes[4:])))
	}
	return result
}
func Hsh_WrdManHashArray(data *Vec_Wrd_t, size int) *Vec_Int_t {
	if size <= 0 || Vec_WrdSize(data)%size != 0 {
		panic("vec: incomplete word hash records")
	}
	return Hsh_IntManHashArray(hshWordsAsInts(data), 2*size)
}
func Hsh_WrdManHashArrayStart(data *Vec_Wrd_t, size int) *Hsh_IntMan_t {
	if size <= 0 || Vec_WrdSize(data)%size != 0 {
		panic("vec: incomplete word hash records")
	}
	entries := Vec_WrdSize(data) / size
	p := Hsh_IntManStart(hshWordsAsInts(data), 2*size, entries)
	for i := 0; i < entries; i++ {
		Hsh_IntManAdd(p, i)
	}
	if len(p.objs) != entries {
		panic("vec: hash-array start requires unique records")
	}
	return p
}

func Hsh_VecManStart(entries int) *Hsh_VecMan_t {
	if entries < 0 {
		panic("vec: negative hash capacity")
	}
	return &Hsh_VecMan_t{vTable: Vec_IntStartFull(hshPrime(entries)), objs: make([]*Hsh_VecObj_t, 0, entries)}
}
func Hsh_VecManStop(p *Hsh_VecMan_t) {
	if p != nil {
		Vec_IntFree(p.vTable)
		*p = Hsh_VecMan_t{}
	}
}
func Hsh_VecReadArray(p *Hsh_VecMan_t, i int32) []int32 { return p.objs[i].pArray }
func hshVecView(o *Hsh_VecObj_t, v *Vec_Int_t) *Vec_Int_t {
	*v = Vec_Int_t{nSize: o.nSize, nCap: o.nSize, pArray: o.pArray}
	return v
}
func Hsh_VecReadEntry(p *Hsh_VecMan_t, i int32) *Vec_Int_t  { return hshVecView(p.objs[i], &p.vTemp) }
func Hsh_VecReadEntry1(p *Hsh_VecMan_t, i int32) *Vec_Int_t { return hshVecView(p.objs[i], &p.vTemp1) }
func Hsh_VecReadEntry2(p *Hsh_VecMan_t, i int32) *Vec_Int_t { return hshVecView(p.objs[i], &p.vTemp2) }
func Hsh_VecSize(p *Hsh_VecMan_t) int                       { return len(p.objs) }
func Hsh_VecManHash(v *Vec_Int_t, tableSize int) int {
	if tableSize <= 0 {
		panic("vec: invalid hash table size")
	}
	primes := [7]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var key uint32
	for i, value := range Vec_IntArray(v) {
		key += uint32(value) * primes[i%7]
	}
	return int(key % uint32(tableSize))
}
func Hsh_VecManAdd(p *Hsh_VecMan_t, v *Vec_Int_t) int32 {
	if len(p.objs) > Vec_IntSize(p.vTable) {
		Vec_IntFill(p.vTable, hshPrime(2*Vec_IntSize(p.vTable)), -1)
		for i, o := range p.objs {
			view := Vec_Int_t{nSize: o.nSize, nCap: o.nSize, pArray: o.pArray}
			place := Vec_IntEntryP(p.vTable, Hsh_VecManHash(&view, Vec_IntSize(p.vTable)))
			o.iNext = *place
			*place = int32(i)
		}
	}
	place := Vec_IntEntryP(p.vTable, Hsh_VecManHash(v, Vec_IntSize(p.vTable)))
	data := Vec_IntArray(v)
	for *place >= 0 {
		o := p.objs[*place]
		if slices.Equal(o.pArray, data) {
			return *place
		}
		place = &o.iNext
	}
	id := int32(len(p.objs))
	*place = id
	p.objs = append(p.objs, &Hsh_VecObj_t{nSize: len(data), iNext: -1, pArray: append([]int32(nil), data...)})
	return id
}

func Hsh_Int1ManHashArrayTest() {
	p := Hsh_Int1ManStart(10)
	defer Hsh_Int1ManStop(p)
	for _, value := range []int32{5, 7, 7, 7, 8, 5, 3, 4} {
		fmt.Printf("Entry = %d. Hashed = %d.\n", value, Hsh_Int1ManAdd(p, value))
	}
	fmt.Printf("Size = %d.\n", Hsh_Int1ManEntryNum(p))
}
func Hsh_IntManHashArrayTest() {
	data := Vec_IntAllocArrayCopy([]int32{12, 17, 13, 12, 15, 3, 16, 16, 12, 17, 12, 12}, 12)
	defer Vec_IntFree(data)
	result := Hsh_IntManHashArray(data, 2)
	defer Vec_IntFree(result)
	Vec_IntPrint(data)
	Vec_IntPrint(result)
}
func Hsh_VecManHashTest() {
	p := Hsh_VecManStart(5)
	defer Hsh_VecManStop(p)
	result := Vec_IntAlloc(40)
	defer Vec_IntFree(result)
	for i := 0; i < 20; i++ {
		v := Vec_IntStartNatural(i)
		Vec_IntPush(result, Hsh_VecManAdd(p, v))
		Vec_IntFree(v)
	}
	for i := 20; i > 0; i-- {
		v := Vec_IntStartNatural(i)
		Vec_IntPush(result, Hsh_VecManAdd(p, v))
		Vec_IntFree(v)
	}
	Vec_IntPrint(result)
}
