// Port of ABC src/aig/miniaig/minilut.h (Alan Mishchenko, UC Berkeley).
package miniaig

import (
	"encoding/binary"
	"fmt"
	"os"
)

const MINI_LUT_NULL = 0x7fffffff
const MINI_LUT_NULL2 = 0x7ffffffe
const MINI_LUT_START_SIZE = 255

type Mini_Lut_t struct {
	nRegs, LutSize int
	pArray         []int
	pTruths        []uint32
}

func Mini_LutWordNum(n int) int {
	if n > 5 {
		return 1 << uint(n-5)
	}
	return 1
}
func Mini_LutGrow(p *Mini_Lut_t, n int) {
	if n <= (cap(p.pArray) / p.LutSize) {
		return
	}
	a := make([]int, len(p.pArray), n*p.LutSize)
	copy(a, p.pArray)
	p.pArray = a
	t := make([]uint32, len(p.pTruths), n*Mini_LutWordNum(p.LutSize))
	copy(t, p.pTruths)
	p.pTruths = t
}
func Mini_LutPush(p *Mini_Lut_t, n int, vars []int, truth []uint32) {
	if n < 0 || n > p.LutSize || len(vars) < n {
		panic("invalid LUT fanins")
	}
	if Mini_LutNodeNum(p) == (cap(p.pArray) / p.LutSize) {
		cap := (cap(p.pArray) / p.LutSize) * 2
		if cap < 255 {
			cap = 255
		}
		Mini_LutGrow(p, cap)
	}
	for _, v := range vars[:n] {
		if v < 0 || v >= Mini_LutNodeNum(p) {
			panic("non-topological LUT")
		}
	}
	for i := 0; i < p.LutSize; i++ {
		v := MINI_LUT_NULL
		if i < n {
			v = vars[i]
		}
		p.pArray = append(p.pArray, v)
	}
	words := Mini_LutWordNum(p.LutSize)
	if truth != nil && len(truth) < words {
		panic("short LUT truth")
	}
	for i := 0; i < words; i++ {
		var v uint32
		if truth != nil {
			v = truth[i]
		}
		p.pTruths = append(p.pTruths, v)
	}
}
func Mini_LutNodeFanin(p *Mini_Lut_t, id, k int) int { return p.pArray[p.LutSize*id+k] }
func Mini_LutNodeTruth(p *Mini_Lut_t, id int) []uint32 {
	w := Mini_LutWordNum(p.LutSize)
	return p.pTruths[id*w : (id+1)*w]
}
func Mini_LutNodeConst0() int                        { return 0 }
func Mini_LutNodeConst1() int                        { return 1 }
func Mini_LutNodeNum(p *Mini_Lut_t) int              { return len(p.pArray) / p.LutSize }
func Mini_LutNodeIsConst(p *Mini_Lut_t, id int) bool { return id == 0 || id == 1 }
func Mini_LutNodeIsPi(p *Mini_Lut_t, id int) bool {
	return id > 1 && Mini_LutNodeFanin(p, id, 0) == MINI_LUT_NULL
}
func Mini_LutNodeIsPo(p *Mini_Lut_t, id int) bool {
	return id > 1 && Mini_LutNodeFanin(p, id, 0) != MINI_LUT_NULL && Mini_LutNodeFanin(p, id, 1) == MINI_LUT_NULL2
}
func Mini_LutNodeIsNode(p *Mini_Lut_t, id int) bool {
	return id > 1 && !Mini_LutNodeIsPi(p, id) && !Mini_LutNodeIsPo(p, id)
}
func Mini_LutSize(p *Mini_Lut_t) int         { return p.LutSize }
func Mini_LutRegNum(p *Mini_Lut_t) int       { return p.nRegs }
func Mini_LutSetRegNum(p *Mini_Lut_t, n int) { p.nRegs = n }
func Mini_LutStart(n int) *Mini_Lut_t {
	if n < 2 || n > 16 {
		panic("LUT size outside 2..16")
	}
	p := &Mini_Lut_t{LutSize: n}
	Mini_LutPush(p, 0, nil, nil)
	Mini_LutPush(p, 0, nil, nil)
	for i := range Mini_LutNodeTruth(p, 1) {
		Mini_LutNodeTruth(p, 1)[i] = ^uint32(0)
	}
	return p
}
func Mini_LutStop(p *Mini_Lut_t) {
	if p != nil {
		p.pArray = nil
		p.pTruths = nil
	}
}
func Mini_LutCreatePi(p *Mini_Lut_t) int { Mini_LutPush(p, 0, nil, nil); return Mini_LutNodeNum(p) - 1 }
func Mini_LutCreatePo(p *Mini_Lut_t, v int) int {
	Mini_LutPush(p, 1, []int{v}, nil)
	p.pArray[p.LutSize*(Mini_LutNodeNum(p)-1)+1] = MINI_LUT_NULL2
	return Mini_LutNodeNum(p) - 1
}
func Mini_LutCreateNode(p *Mini_Lut_t, n int, v []int, t []uint32) int {
	Mini_LutPush(p, n, v, t)
	return Mini_LutNodeNum(p) - 1
}
func Mini_LutCheck(p *Mini_Lut_t) bool {
	if p == nil || p.LutSize < 2 || p.LutSize > 16 || Mini_LutNodeNum(p) < 2 || len(p.pArray) != Mini_LutNodeNum(p)*p.LutSize || len(p.pTruths) != Mini_LutNodeNum(p)*Mini_LutWordNum(p.LutSize) {
		return false
	}
	for id := 0; id < 2; id++ {
		for k := 0; k < p.LutSize; k++ {
			if Mini_LutNodeFanin(p, id, k) != MINI_LUT_NULL {
				return false
			}
		}
		expected := uint32(0)
		if id == 1 {
			expected = ^uint32(0)
		}
		for _, truth := range Mini_LutNodeTruth(p, id) {
			if truth != expected {
				return false
			}
		}
	}
	for i := 2; i < Mini_LutNodeNum(p); i++ {
		ended := false
		for k := 0; k < p.LutSize; k++ {
			v := Mini_LutNodeFanin(p, i, k)
			if v == MINI_LUT_NULL {
				ended = true
				continue
			}
			if v == MINI_LUT_NULL2 && k == 1 && !ended {
				ended = true
				continue
			}
			if ended || v < 0 || v >= i {
				return false
			}
		}
	}
	return true
}
func Mini_LutPrintStats(p *Mini_Lut_t) {
	a, b, c := 0, 0, 0
	for i := 2; i < Mini_LutNodeNum(p); i++ {
		switch {
		case Mini_LutNodeIsPi(p, i):
			a++
		case Mini_LutNodeIsPo(p, i):
			b++
		default:
			c++
		}
	}
	fmt.Printf("PI = %d. PO = %d. LUT = %d.\n", a, b, c)
}
func Mini_LutDump(p *Mini_Lut_t, name string) {
	f, e := os.Create(name)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	d := []int32{int32(Mini_LutNodeNum(p)), int32(p.nRegs), int32(p.LutSize)}
	for _, v := range p.pArray {
		d = append(d, int32(v))
	}
	for _, v := range p.pTruths {
		d = append(d, int32(v))
	}
	if e = binary.Write(f, binary.LittleEndian, d); e != nil {
		panic(e)
	}
}
func Mini_LutLoad(name string) *Mini_Lut_t {
	d, e := os.ReadFile(name)
	if e != nil || len(d) < 12 || len(d)%4 != 0 {
		return nil
	}
	word := func(i int) int { return int(int32(binary.LittleEndian.Uint32(d[4*i:]))) }
	n, r, k := word(0), word(1), word(2)
	if k < 2 || k > 16 || n < 2 || n > (len(d)-12)/4 {
		return nil
	}
	w := Mini_LutWordNum(k)
	if len(d)/4 != 3+n*(k+w) {
		return nil
	}
	p := &Mini_Lut_t{nRegs: r, LutSize: k, pArray: make([]int, n*k), pTruths: make([]uint32, n*w)}
	for i := range p.pArray {
		p.pArray[i] = word(3 + i)
	}
	for i := range p.pTruths {
		p.pTruths[i] = uint32(word(3 + n*k + i))
	}
	if !Mini_LutCheck(p) {
		return nil
	}
	return p
}
