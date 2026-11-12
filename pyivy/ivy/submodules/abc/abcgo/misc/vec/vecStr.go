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

type Vec_Str_t struct {
	nCap, nSize int
	pArray      []byte
}

func Vec_StrAlloc(n int) *Vec_Str_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_StrAllocExact(n)
}
func Vec_StrAllocExact(n int) *Vec_Str_t { return &Vec_Str_t{nCap: n, pArray: make([]byte, n)} }
func Vec_StrStart(n int) *Vec_Str_t      { p := Vec_StrAlloc(n); p.nSize = n; return p }
func Vec_StrAllocArray(a []byte, n int) *Vec_Str_t {
	return &Vec_Str_t{nCap: n, nSize: n, pArray: a[:n]}
}
func Vec_StrAllocArrayCopy(a []byte, n int) *Vec_Str_t {
	p := Vec_StrAllocExact(n)
	p.nSize = n
	copy(p.pArray, a[:n])
	return p
}
func Vec_StrDup(p *Vec_Str_t) *Vec_Str_t {
	q := Vec_StrAllocExact(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray[:p.nSize])
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
func Vec_StrReleaseArray(p *Vec_Str_t) []byte { a := p.pArray; Vec_StrZero(p); return a }
func Vec_StrArray(p *Vec_Str_t) []byte        { return p.pArray[:p.nSize] }
func Vec_StrSize(p *Vec_Str_t) int            { return p.nSize }
func Vec_StrCap(p *Vec_Str_t) int             { return p.nCap }
func Vec_StrMemory(p *Vec_Str_t) float64 {
	if p == nil {
		return 0
	}
	return float64(p.nCap + int(unsafe.Sizeof(*p)))
}
func Vec_StrEntry(p *Vec_Str_t, i int) byte         { return p.pArray[:p.nSize][i] }
func Vec_StrEntryP(p *Vec_Str_t, i int) *byte       { return &p.pArray[:p.nSize][i] }
func Vec_StrWriteEntry(p *Vec_Str_t, i int, x byte) { p.pArray[:p.nSize][i] = x }
func Vec_StrEntryLast(p *Vec_Str_t) byte            { return Vec_StrEntry(p, p.nSize-1) }
func Vec_StrGrow(p *Vec_Str_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]byte, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_StrFill(p *Vec_Str_t, n int, x byte) {
	Vec_StrGrow(p, n)
	for i := 0; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_StrFillExtra(p *Vec_Str_t, n int, x byte) {
	if n <= p.nSize {
		return
	}
	if n > p.nCap {
		Vec_StrGrow(p, max(n, 2*p.nCap))
	}
	for i := p.nSize; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_StrShrink(p *Vec_Str_t, n int) {
	if n < 0 || n > p.nSize {
		panic("invalid vector shrink")
	}
	p.nSize = n
}
func Vec_StrClear(p *Vec_Str_t) { p.nSize = 0 }
func Vec_StrPush(p *Vec_Str_t, x byte) {
	if p.nSize == p.nCap {
		Vec_StrGrow(p, max(16, 2*p.nCap))
	}
	p.pArray[p.nSize] = x
	p.nSize++
}
func Vec_StrPop(p *Vec_Str_t) byte { x := Vec_StrEntryLast(p); p.nSize--; return x }
func Vec_StrEqual(p, q *Vec_Str_t) int {
	if p.nSize != q.nSize {
		return 0
	}
	for i, v := range Vec_StrArray(p) {
		if v != q.pArray[i] {
			return 0
		}
	}
	return 1
}
func Vec_StrSortCompare1(a, b *byte) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_StrSortCompare2(a, b *byte) int { return -Vec_StrSortCompare1(a, b) }
func Vec_StrSort(p *Vec_Str_t, reverse int) {
	sort.Slice(p.pArray[:p.nSize], func(i, j int) bool {
		if reverse != 0 {
			return p.pArray[i] > p.pArray[j]
		}
		return p.pArray[i] < p.pArray[j]
	})
}
func Vec_StrLimit(p *Vec_Str_t) []byte            { return p.pArray[p.nSize:] }
func Vec_StrGetEntry(p *Vec_Str_t, i int) byte    { Vec_StrFillExtra(p, i+1, 0); return p.pArray[i] }
func Vec_StrSetEntry(p *Vec_Str_t, i int, x byte) { Vec_StrFillExtra(p, i+1, 0); p.pArray[i] = x }
func Vec_StrReverseOrder(p *Vec_Str_t) {
	for i, j := 0, p.nSize-1; i < j; i, j = i+1, j-1 {
		p.pArray[i], p.pArray[j] = p.pArray[j], p.pArray[i]
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
		if v > x {
			n++
		}
	}
	return n
}
func Vec_StrCountSmaller(p *Vec_Str_t, x byte) int {
	n := 0
	for _, v := range Vec_StrArray(p) {
		if v < x {
			n++
		}
	}
	return n
}
func Vec_StrSetSize(p *Vec_Str_t, n int) {
	if n < 0 || n > p.nCap {
		panic("string vector size")
	}
	p.nSize = n
}
func Vec_StrPushBuffer(p *Vec_Str_t, a []byte, n int) {
	if p.nSize+n > p.nCap {
		Vec_StrGrow(p, 2*(p.nSize+n))
	}
	copy(p.pArray[p.nSize:], a[:n])
	p.nSize += n
}
func Vec_StrIntPrint(p *Vec_Str_t) {
	fmt.Printf("Vector has %d entries: {", p.nSize)
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
	start := p.nSize
	text := fmt.Sprintf(format, args...)
	Vec_StrPushBuffer(p, []byte(text), len(text))
	Vec_StrGrow(p, p.nSize+1)
	p.pArray[p.nSize] = 0
	return p.pArray[start:p.nSize]
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
	if p.nSize != q.nSize {
		return p.nSize - q.nSize
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
	for ; x >= 0x80; x >>= 7 {
		Vec_StrPush(p, byte(x)|0x80)
	}
	Vec_StrPush(p, byte(x))
}
func Vec_StrGetI(p *Vec_Str_t, pos *int) int32 {
	var value uint32
	for shift := uint(0); ; shift += 7 {
		c := Vec_StrGetC(p, pos)
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
	return string(p.pArray[start : *pos-1])
}
func Vec_StrPutC(p *Vec_Str_t, c byte)        { Vec_StrPush(p, c) }
func Vec_StrGetC(p *Vec_Str_t, pos *int) byte { x := Vec_StrEntry(p, *pos); *pos++; return x }
