/**CFile****************************************************************

  FileName    [vecStr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of characters.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecStr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

package vec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

type Vec_Str_t struct{ data []byte }

func Vec_StrAlloc(n int) *Vec_Str_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_StrAllocExact(n)
}
func Vec_StrAllocExact(n int) *Vec_Str_t { return &Vec_Str_t{data: make([]byte, 0, n)} }
func Vec_StrStart(n int) *Vec_Str_t      { p := Vec_StrAlloc(n); p.resize(n); return p }
func Vec_StrAllocArray(a []byte, n int) *Vec_Str_t {
	return &Vec_Str_t{data: (a[:n])[:n]}
}
func Vec_StrAllocArrayCopy(a []byte, n int) *Vec_Str_t {
	p := Vec_StrAllocExact(n)
	p.resize(n)
	copy(p.storage(), a[:n])
	return p
}
func Vec_StrDup(p *Vec_Str_t) *Vec_Str_t {
	q := Vec_StrAllocExact(cap(p.data))
	q.resize(len(p.data))
	copy(q.storage(), p.data)
	return q
}
func Vec_StrDupArray(p *Vec_Str_t) *Vec_Str_t { q := *p; Vec_StrZero(p); return &q }
func Vec_StrZero(p *Vec_Str_t)                { *p = Vec_Str_t{} }
func Vec_StrErase(p *Vec_Str_t)               { Vec_StrZero(p) }
func Vec_StrFree(p *Vec_Str_t) {
	if p != nil {
		Vec_StrZero(p)
	}
}
func Vec_StrFreeP(p **Vec_Str_t)              { Vec_StrFree(*p); *p = nil }
func Vec_StrReleaseArray(p *Vec_Str_t) []byte { a := p.data; Vec_StrZero(p); return a }
func Vec_StrArray(p *Vec_Str_t) []byte        { return p.data }
func Vec_StrSize(p *Vec_Str_t) int            { return len(p.data) }
func Vec_StrCap(p *Vec_Str_t) int             { return cap(p.data) }
func Vec_StrMemory(p *Vec_Str_t) float64 {
	if p == nil {
		return 0
	}
	return float64(cap(p.data) + int(unsafe.Sizeof(*p)))
}
func Vec_StrEntry(p *Vec_Str_t, i int) byte         { return p.data[i] }
func Vec_StrEntryP(p *Vec_Str_t, i int) *byte       { return &p.data[i] }
func Vec_StrWriteEntry(p *Vec_Str_t, i int, x byte) { p.data[i] = x }
func Vec_StrEntryLast(p *Vec_Str_t) byte            { return Vec_StrEntry(p, len(p.data)-1) }
func Vec_StrGrow(p *Vec_Str_t, n int) {
	if n > cap(p.data) {
		data := make([]byte, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_StrFill(p *Vec_Str_t, n int, x byte) {
	Vec_StrGrow(p, n)
	for i := 0; i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_StrFillExtra(p *Vec_Str_t, n int, x byte) {
	if n <= len(p.data) {
		return
	}
	if n > cap(p.data) {
		Vec_StrGrow(p, max(n, 2*cap(p.data)))
	}
	for i := len(p.data); i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_StrShrink(p *Vec_Str_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.resize(n)
}
func Vec_StrClear(p *Vec_Str_t) { p.resize(0) }
func Vec_StrPush(p *Vec_Str_t, x byte) {
	if len(p.data) == cap(p.data) {
		Vec_StrGrow(p, max(16, 2*cap(p.data)))
	}
	p.storage()[len(p.data)] = x
	p.resize(len(p.data) + 1)
}
func Vec_StrPop(p *Vec_Str_t) byte { x := Vec_StrEntryLast(p); p.resize(len(p.data) - 1); return x }
func Vec_StrEqual(p, q *Vec_Str_t) int {
	if len(p.data) != len(q.data) {
		return 0
	}
	for i, v := range Vec_StrArray(p) {
		if v != q.storage()[i] {
			return 0
		}
	}
	return 1
}
func Vec_StrSortCompare1(a, b *byte) int {
	if int8(*a) < int8(*b) {
		return -1
	}
	if int8(*a) > int8(*b) {
		return 1
	}
	return 0
}
func Vec_StrSortCompare2(a, b *byte) int { return -Vec_StrSortCompare1(a, b) }
func Vec_StrSort(p *Vec_Str_t, reverse int) {
	sort.Slice(p.data, func(i, j int) bool {
		if reverse != 0 {
			return int8(p.storage()[i]) > int8(p.storage()[j])
		}
		return int8(p.storage()[i]) < int8(p.storage()[j])
	})
}
func Vec_StrLimit(p *Vec_Str_t) []byte            { return p.storage()[len(p.data):] }
func Vec_StrGetEntry(p *Vec_Str_t, i int) byte    { Vec_StrFillExtra(p, i+1, 0); return p.storage()[i] }
func Vec_StrSetEntry(p *Vec_Str_t, i int, x byte) { Vec_StrFillExtra(p, i+1, 0); p.storage()[i] = x }
func Vec_StrReverseOrder(p *Vec_Str_t) {
	for i, j := 0, len(p.data)-1; i < j; i, j = i+1, j-1 {
		p.storage()[i], p.storage()[j] = p.storage()[j], p.storage()[i]
	}
}
func Vec_StrCountEntry(p *Vec_Str_t, x byte) int {
	n := 0
	for _, v := range Vec_StrArray(p) {
		if v == x {
			n++
		}
	}
	return n
}
func Vec_StrCountLarger(p *Vec_Str_t, x byte) int {
	n := 0
	for _, v := range Vec_StrArray(p) {
		if int8(v) > int8(x) {
			n++
		}
	}
	return n
}
func Vec_StrCountSmaller(p *Vec_Str_t, x byte) int {
	n := 0
	for _, v := range Vec_StrArray(p) {
		if int8(v) < int8(x) {
			n++
		}
	}
	return n
}
func Vec_StrSetSize(p *Vec_Str_t, n int) {
	if n < 0 || n > cap(p.data) {
		panic("string vector size")
	}
	p.resize(n)
}
func Vec_StrPushBuffer(p *Vec_Str_t, a []byte, n int) {
	if len(p.data)+n > cap(p.data) {
		Vec_StrGrow(p, 2*(len(p.data)+n))
	}
	copy(p.storage()[len(p.data):], a[:n])
	p.resize(len(p.data) + (n))
}
func Vec_StrIntPrint(p *Vec_Str_t) {
	fmt.Printf("Vector has %d entries: {", len(p.data))
	for _, x := range Vec_StrArray(p) {
		fmt.Printf(" %d", int8(x))
	}
	fmt.Println(" }")
}
func Vec_StrPrintNum(p *Vec_Str_t, x int32) { Vec_StrPrintStr(p, strconv.FormatInt(int64(x), 10)) }
func Vec_StrPrintNumStar(p *Vec_Str_t, x int32, digits int) {
	text := strconv.FormatInt(int64(x), 10)
	if x < 0 {
		Vec_StrPush(p, '-')
		text = text[1:]
		digits--
	}
	for i := len(text); i < digits; i++ {
		Vec_StrPush(p, '0')
	}
	Vec_StrPrintStr(p, text)
}
func Vec_StrPrintStr(p *Vec_Str_t, s string) {
	if i := strings.IndexByte(s, 0); i >= 0 {
		s = s[:i]
	}
	Vec_StrPushBuffer(p, []byte(s), len(s))
}
func Vec_StrPrintF(p *Vec_Str_t, format string, args ...any) []byte {
	start := len(p.data)
	text := fmt.Sprintf(format, args...)
	Vec_StrPushBuffer(p, []byte(text), len(text))
	Vec_StrGrow(p, len(p.data)+1)
	p.storage()[len(p.data)] = 0
	return p.storage()[start:len(p.data)]
}
func Vec_StrAppend(p *Vec_Str_t, s string) { Vec_StrPrintStr(p, s) }
func Vec_StrCopy(p *Vec_Str_t, s string)   { Vec_StrClear(p); Vec_StrAppend(p, s); Vec_StrPush(p, 0) }
func Vec_StrSum(p *Vec_Str_t) int32 {
	var sum int32
	for _, x := range Vec_StrArray(p) {
		sum += int32(int8(x))
	}
	return sum
}
func Vec_StrCountEntryLit(p *Vec_Str_t, x byte) int   { return strCountLiteral(p, x, 0) }
func Vec_StrCountLargerLit(p *Vec_Str_t, x byte) int  { return strCountLiteral(p, x, 1) }
func Vec_StrCountSmallerLit(p *Vec_Str_t, x byte) int { return strCountLiteral(p, x, -1) }
func strCountLiteral(p *Vec_Str_t, x byte, order int) int {
	n := 0
	for _, v := range Vec_StrArray(p) {
		value := int32(int8(v)) >> 1
		entry := int32(int8(x))
		if order == 0 && value == entry || order == 1 && value > entry || order == -1 && value < entry {
			n++
		}
	}
	return n
}
func Vec_StrCompareVec(p, q *Vec_Str_t) int {
	if p == nil {
		if q == nil {
			return 0
		}
		return -1
	}
	if q == nil {
		return 1
	}
	if len(p.data) != len(q.data) {
		return len(p.data) - len(q.data)
	}
	return bytes.Compare(Vec_StrArray(p), Vec_StrArray(q))
}
func Vec_StrPutI_ne(p *Vec_Str_t, x int32) {
	var a [4]byte
	binary.BigEndian.PutUint32(a[:], uint32(x))
	Vec_StrPushBuffer(p, a[:], 4)
}
func Vec_StrGetI_ne(p *Vec_Str_t, pos *int) int32 {
	x := int32(binary.BigEndian.Uint32(Vec_StrArray(p)[*pos : *pos+4]))
	*pos += 4
	return x
}
func Vec_StrPutI(p *Vec_Str_t, x int32) {
	if x < 0 {
		panic("negative variable integer")
	}
	for ; x >= 0x80; x >>= 7 {
		Vec_StrPush(p, byte(x)|0x80)
	}
	Vec_StrPush(p, byte(x))
}
func Vec_StrGetI(p *Vec_Str_t, pos *int) int32 {
	var value uint32
	for shift := uint(0); ; shift += 7 {
		c := Vec_StrGetC(p, pos)
		if shift == 28 && c > 0x0f {
			panic("oversized variable integer")
		}
		value |= uint32(c&0x7f) << shift
		if c&0x80 == 0 {
			return int32(value)
		}
		if shift >= 28 {
			panic("oversized variable integer")
		}
	}
}
func Vec_StrPutW(p *Vec_Str_t, x uint64) {
	var a [8]byte
	binary.LittleEndian.PutUint64(a[:], x)
	Vec_StrPushBuffer(p, a[:], 8)
}
func Vec_StrGetW(p *Vec_Str_t, pos *int) uint64 {
	x := binary.LittleEndian.Uint64(Vec_StrArray(p)[*pos : *pos+8])
	*pos += 8
	return x
}
func Vec_StrPutF(p *Vec_Str_t, x float32) {
	var a [4]byte
	binary.LittleEndian.PutUint32(a[:], math.Float32bits(x))
	Vec_StrPushBuffer(p, a[:], 4)
}
func Vec_StrGetF(p *Vec_Str_t, pos *int) float32 {
	x := math.Float32frombits(binary.LittleEndian.Uint32(Vec_StrArray(p)[*pos : *pos+4]))
	*pos += 4
	return x
}
func Vec_StrPutD(p *Vec_Str_t, x float64)        { Vec_StrPutW(p, math.Float64bits(x)) }
func Vec_StrGetD(p *Vec_Str_t, pos *int) float64 { return math.Float64frombits(Vec_StrGetW(p, pos)) }
func Vec_StrPutS(p *Vec_Str_t, s string)         { Vec_StrPrintStr(p, s); Vec_StrPush(p, 0) }
func Vec_StrGetS(p *Vec_Str_t, pos *int) string {
	start := *pos
	for Vec_StrGetC(p, pos) != 0 {
	}
	return string(p.storage()[start : *pos-1])
}
func Vec_StrPutC(p *Vec_Str_t, c byte)        { Vec_StrPush(p, c) }
func Vec_StrGetC(p *Vec_Str_t, pos *int) byte { x := Vec_StrEntry(p, *pos); *pos++; return x }
