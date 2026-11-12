// Port of ABC src/aig/miniaig/miniaig.h (Alan Mishchenko, UC Berkeley).
// Package miniaig implements ABC's compact AND-inverter graph representation.
package miniaig

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const MINI_AIG_NULL = 0x7fffffff
const MINI_AIG_START_SIZE = 255

type Mini_Aig_t struct {
	nCap, nSize, nRegs int
	pArray             []int
}

func Mini_AigGrow(p *Mini_Aig_t, n int) {
	if n <= p.nCap {
		return
	}
	a := make([]int, p.nSize, n)
	copy(a, p.pArray)
	p.pArray = a
	p.nCap = n
}
func Mini_AigPush(p *Mini_Aig_t, a, b int) {
	if p.nSize+2 > p.nCap {
		n := 2 * p.nCap
		if n < MINI_AIG_START_SIZE {
			n = MINI_AIG_START_SIZE
		}
		Mini_AigGrow(p, n)
	}
	p.pArray = append(p.pArray, a, b)
	p.nSize += 2
}
func Mini_AigNodeFanin0(p *Mini_Aig_t, id int) int { return p.pArray[2*id] }
func Mini_AigNodeFanin1(p *Mini_Aig_t, id int) int { return p.pArray[2*id+1] }
func Mini_AigVar2Lit(v, c int) int                 { return 2*v + c }
func Mini_AigLit2Var(l int) int                    { return l >> 1 }
func Mini_AigLitIsCompl(l int) int                 { return l & 1 }
func Mini_AigLitNot(l int) int                     { return l ^ 1 }
func Mini_AigLitNotCond(l, c int) int {
	if c > 0 {
		return l ^ 1
	}
	return l
}
func Mini_AigLitRegular(l int) int                   { return l &^ 1 }
func Mini_AigLitConst0() int                         { return 0 }
func Mini_AigLitConst1() int                         { return 1 }
func Mini_AigLitIsConst0(l int) bool                 { return l == 0 }
func Mini_AigLitIsConst1(l int) bool                 { return l == 1 }
func Mini_AigLitIsConst(l int) bool                  { return l == 0 || l == 1 }
func Mini_AigNodeNum(p *Mini_Aig_t) int              { return p.nSize / 2 }
func Mini_AigNodeIsConst(p *Mini_Aig_t, id int) bool { return id == 0 }
func Mini_AigNodeIsPi(p *Mini_Aig_t, id int) bool {
	return id > 0 && Mini_AigNodeFanin0(p, id) == MINI_AIG_NULL
}
func Mini_AigNodeIsPo(p *Mini_Aig_t, id int) bool {
	return id > 0 && Mini_AigNodeFanin0(p, id) != MINI_AIG_NULL && Mini_AigNodeFanin1(p, id) == MINI_AIG_NULL
}
func Mini_AigNodeIsAnd(p *Mini_Aig_t, id int) bool {
	return id > 0 && Mini_AigNodeFanin0(p, id) != MINI_AIG_NULL && Mini_AigNodeFanin1(p, id) != MINI_AIG_NULL
}
func Mini_AigRegNum(p *Mini_Aig_t) int { return p.nRegs }
func Mini_AigSetRegNum(p *Mini_Aig_t, n int) {
	if n < 0 {
		panic("negative register count")
	}
	p.nRegs = n
}
func Mini_AigStart() *Mini_Aig_t {
	p := &Mini_Aig_t{}
	Mini_AigPush(p, MINI_AIG_NULL, MINI_AIG_NULL)
	return p
}
func Mini_AigStop(p *Mini_Aig_t) {
	if p != nil {
		p.pArray = nil
		p.nSize = 0
		p.nCap = 0
	}
}
func miniCount(p *Mini_Aig_t, pred func(*Mini_Aig_t, int) bool) int {
	n := 0
	for i := 1; i < Mini_AigNodeNum(p); i++ {
		if pred(p, i) {
			n++
		}
	}
	return n
}
func Mini_AigPiNum(p *Mini_Aig_t) int  { return miniCount(p, Mini_AigNodeIsPi) }
func Mini_AigPoNum(p *Mini_Aig_t) int  { return miniCount(p, Mini_AigNodeIsPo) }
func Mini_AigAndNum(p *Mini_Aig_t) int { return miniCount(p, Mini_AigNodeIsAnd) }
func Mini_AigPrintStats(p *Mini_Aig_t) {
	fmt.Printf("PI = %d. PO = %d. Node = %d.\n", Mini_AigPiNum(p), Mini_AigPoNum(p), Mini_AigAndNum(p))
}
func Mini_AigCreatePi(p *Mini_Aig_t) int {
	l := p.nSize
	Mini_AigPush(p, MINI_AIG_NULL, MINI_AIG_NULL)
	return l
}
func Mini_AigCreatePo(p *Mini_Aig_t, l int) int {
	id := p.nSize
	if l < 0 || l >= id {
		panic("invalid output literal")
	}
	Mini_AigPush(p, l, MINI_AIG_NULL)
	return id
}
func Mini_AigAnd(p *Mini_Aig_t, a, b int) int {
	id := p.nSize
	if a < 0 || b < 0 || a >= id || b >= id {
		panic("invalid AND literal")
	}
	Mini_AigPush(p, a, b)
	return id
}
func Mini_AigOr(p *Mini_Aig_t, a, b int) int { return Mini_AigAnd(p, a^1, b^1) ^ 1 }
func Mini_AigMux(p *Mini_Aig_t, c, t, e int) int {
	return Mini_AigOr(p, Mini_AigAnd(p, c, t), Mini_AigAnd(p, c^1, e))
}
func Mini_AigXor(p *Mini_Aig_t, a, b int) int { return Mini_AigMux(p, a, b^1, b) }
func Mini_AigAndProp(p *Mini_Aig_t, a, b int) int {
	if a == 0 || b == 0 || a == b^1 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 || a == b {
		return a
	}
	return Mini_AigAnd(p, a, b)
}
func Mini_AigMuxProp(p *Mini_Aig_t, c, t, e int) int {
	a := Mini_AigAndProp(p, c^1, e)
	b := Mini_AigAndProp(p, c, t)
	return Mini_AigAndProp(p, a^1, b^1) ^ 1
}

var s_MiniTruths5 = [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

func Mini_AigTt5HasVar(t uint32, v int) bool {
	return ((t << uint(1<<v)) & s_MiniTruths5[v]) != (t & s_MiniTruths5[v])
}
func Mini_AigTt5Cofactor0(t uint32, v int) uint32 {
	x := t &^ s_MiniTruths5[v]
	return x | x<<uint(1<<v)
}
func Mini_AigTt5Cofactor1(t uint32, v int) uint32 {
	x := t & s_MiniTruths5[v]
	return x | x>>uint(1<<v)
}
func Mini_AigTruth(p *Mini_Aig_t, vars []int, n int, t uint32) int {
	if t == 0 {
		return 0
	}
	if t == ^uint32(0) {
		return 1
	}
	if n <= 0 || n > 5 || len(vars) < n {
		panic("invalid truth support")
	}
	v := n - 1
	for v >= 0 && !Mini_AigTt5HasVar(t, v) {
		v--
	}
	if v < 0 {
		panic("truth outside support")
	}
	a := Mini_AigTruth(p, vars, v, Mini_AigTt5Cofactor0(t, v))
	b := Mini_AigTruth(p, vars, v, Mini_AigTt5Cofactor1(t, v))
	return Mini_AigMuxProp(p, vars[v], b, a)
}
func Mini_AigCheck(p *Mini_Aig_t) bool {
	if p == nil || p.nSize < 2 || p.nSize%2 != 0 || p.nSize != len(p.pArray) {
		return false
	}
	for i := 1; i < Mini_AigNodeNum(p); i++ {
		a, b := p.pArray[2*i], p.pArray[2*i+1]
		if a != MINI_AIG_NULL && (a < 0 || a >= 2*i) {
			return false
		}
		if b != MINI_AIG_NULL && (b < 0 || b >= 2*i) {
			return false
		}
	}
	return true
}

// Dump/Load use the little-endian 32-bit integer format used by ABC on x86.
func Mini_AigDump(p *Mini_Aig_t, name string) {
	f, e := os.Create(name)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	data := make([]int32, 2+p.nSize)
	data[0] = int32(p.nSize)
	data[1] = int32(p.nRegs)
	for i, v := range p.pArray {
		data[i+2] = int32(v)
	}
	if e = binary.Write(f, binary.LittleEndian, data); e != nil {
		panic(e)
	}
}
func Mini_AigLoad(name string) *Mini_Aig_t {
	data, e := os.ReadFile(name)
	if e != nil {
		return nil
	}
	if len(data) < 16 || len(data)%4 != 0 {
		return nil
	}
	size := int(binary.LittleEndian.Uint32(data))
	regs := int(binary.LittleEndian.Uint32(data[4:]))
	if size < 2 || size%2 != 0 || size != (len(data)-8)/4 {
		return nil
	}
	p := &Mini_Aig_t{nCap: size, nSize: size, nRegs: regs, pArray: make([]int, size)}
	for i := range p.pArray {
		p.pArray[i] = int(int32(binary.LittleEndian.Uint32(data[8+4*i:])))
	}
	if !Mini_AigCheck(p) {
		return nil
	}
	return p
}
func Mini_AigDumpVerilog(name, module string, p *Mini_Aig_t) {
	f, e := os.Create(name)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	var ports []string
	for i := 1; i < Mini_AigNodeNum(p); i++ {
		if Mini_AigNodeIsPi(p, i) {
			ports = append(ports, fmt.Sprintf("input wire i%d", i))
		}
		if Mini_AigNodeIsPo(p, i) {
			ports = append(ports, fmt.Sprintf("output wire o%d", i))
		}
	}
	fmt.Fprintf(f, "module %s (%s);\n", module, strings.Join(ports, ", "))
	lit := func(l int) string {
		if l < 2 {
			return fmt.Sprintf("1'b%d", l)
		}
		prefix := ""
		if l&1 != 0 {
			prefix = "~"
		}
		kind := "n"
		if Mini_AigNodeIsPi(p, l>>1) {
			kind = "i"
		}
		return fmt.Sprintf("%s%s%d", prefix, kind, l>>1)
	}
	for i := 1; i < Mini_AigNodeNum(p); i++ {
		if Mini_AigNodeIsAnd(p, i) {
			fmt.Fprintf(f, "wire n%d = %s & %s;\n", i, lit(p.pArray[2*i]), lit(p.pArray[2*i+1]))
		}
		if Mini_AigNodeIsPo(p, i) {
			fmt.Fprintf(f, "assign o%d = %s;\n", i, lit(p.pArray[2*i]))
		}
	}
	fmt.Fprintln(f, "endmodule")
}
