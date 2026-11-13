package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

const TAB_UNUSED = 0x7FFF

type Tab_Obj_t struct {
	Table, Next                  int32
	Cost, LitA, LitB, LitC, Func uint32
}
type Tab_Tab_t struct {
	SizeMask, nBins int32
	pBins           []Tab_Obj_t
}

func Tab_TabAlloc(logSize int32) *Tab_Tab_t {
	if logSize < 4 || logSize > 30 {
		panic("bmc: invalid divisor table size")
	}
	n := int32(1) << logSize
	return &Tab_Tab_t{SizeMask: n - 1, nBins: 1, pBins: make([]Tab_Obj_t, n)}
}
func Tab_TabFree(p *Tab_Tab_t) {
	if p != nil {
		p.pBins = nil
		p.nBins = 0
	}
}
func Tab_Hash(a, b, c, fn, mask int32) int32 {
	return (a*50331653 + b*100663319 + c*201326611 + fn*402653189) & mask
}
func Tab_TabRehash(p *Tab_Tab_t) {
	if p.nBins != p.SizeMask+1 {
		panic("bmc: divisor table not full")
	}
	old := p.nBins
	p.pBins = append(p.pBins, make([]Tab_Obj_t, old)...)
	p.SizeMask = 2*p.SizeMask + 1
	for i := int32(0); i < old; i++ {
		p.pBins[i].Table = 0
		p.pBins[i].Next = 0
	}
	for i := int32(1); i < old; i++ {
		e := &p.pBins[i]
		h := Tab_Hash(int32(e.LitA), int32(e.LitB), int32(e.LitC), int32(e.Func), p.SizeMask)
		e.Next = p.pBins[h].Table
		p.pBins[h].Table = i
	}
}
func Tab_TabEntry(p *Tab_Tab_t, i int32) *Tab_Obj_t {
	if i == 0 {
		return nil
	}
	return &p.pBins[i]
}
func Tab_TabHashAdd(p *Tab_Tab_t, lits []int32, fn, cost int32) int32 {
	if len(lits) < 3 || fn < 0 || fn > 3 || cost < 0 {
		panic("bmc: invalid divisor")
	}
	for _, v := range lits[:3] {
		if v < 0 || v > TAB_UNUSED {
			panic("bmc: divisor literal out of range")
		}
	}
	if p.nBins == p.SizeMask+1 {
		Tab_TabRehash(p)
	}
	h := Tab_Hash(lits[0], lits[1], lits[2], fn, p.SizeMask)
	for i := p.pBins[h].Table; i != 0; i = p.pBins[i].Next {
		e := &p.pBins[i]
		if e.LitA == uint32(lits[0]) && e.LitB == uint32(lits[1]) && e.LitC == uint32(lits[2]) && e.Func == uint32(fn) {
			e.Cost = (e.Cost + uint32(cost)) & 0x1ffff
			return 1
		}
	}
	i := p.nBins
	p.nBins++
	p.pBins[i] = Tab_Obj_t{Table: p.pBins[i].Table, Next: p.pBins[h].Table, LitA: uint32(lits[0]), LitB: uint32(lits[1]), LitC: uint32(lits[2]), Func: uint32(fn), Cost: uint32(cost) & 0x1ffff}
	p.pBins[h].Table = i
	return 0
}
func Tab_TabFindBest(p *Tab_Tab_t, nDivs int32) *vec.Vec_Int_t {
	order := make([]int, p.nBins-1)
	for i := range order {
		order[i] = i + 1
	}
	sort.SliceStable(order, func(i, j int) bool { return p.pBins[order[i]].Cost > p.pBins[order[j]].Cost })
	result := vec.Vec_IntAlloc(4 * int(nDivs))
	names := []string{"const1", "and", "xor", "mux"}
	for i, index := range order {
		e := p.pBins[index]
		if int32(i) >= nDivs || e.Cost <= 1 {
			break
		}
		fmt.Printf("Lit0 = %5d.  Lit1 = %5d.  Lit2 = %5d.  Func = %s.  Cost = %3d.\n", e.LitA, e.LitB, e.LitC, names[e.Func], e.Cost)
		vec.Vec_IntPushFour(result, int32(e.Func), int32(e.LitA), int32(e.LitB), int32(e.LitC))
	}
	return result
}
