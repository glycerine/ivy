package gia

/**CFile****************************************************************

  FileName    [giaAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"time"
)

const (
	AGI_PI uint64 = 0xFFFFFFFF00000000
	AGI_RO uint64 = 0xFFFFFFFE00000000
	AGI_PO uint64 = 0xFFFFFFFD00000000
	AGI_RI uint64 = 0xFFFFFFFC00000000
	AGI_C0 uint64 = 0xFFFFFFFBFFFFFFFA
	AGI_M0 uint64 = 0x00000000FFFFFFFF
	AGI_M1 uint64 = 0xFFFFFFFF00000000
)

type Agi_Man_t struct {
	pName, pSpec                           string
	nNodes, nRegs                          int32
	nTravIds                               uint32
	vCis, vCos                             []int32
	pObjs                                  []uint64
	pThird, pTravIds, pNext, pTable, pCopy []uint32
}

func Agi_ManObjNum(p *Agi_Man_t) int32          { return int32(len(p.pObjs)) }
func Agi_ManCiNum(p *Agi_Man_t) int32           { return int32(len(p.vCis)) }
func Agi_ManCoNum(p *Agi_Man_t) int32           { return int32(len(p.vCos)) }
func Agi_ManNodeNum(p *Agi_Man_t) int32         { return p.nNodes }
func Agi_ObjLit0(p *Agi_Man_t, id int32) uint32 { return uint32(p.pObjs[id]) }
func Agi_ObjLit1(p *Agi_Man_t, id int32) uint32 { return uint32(p.pObjs[id] >> 32) }
func Agi_ObjLit2(p *Agi_Man_t, id int32) uint32 {
	if p.pThird == nil {
		return ^uint32(0)
	}
	return p.pThird[id]
}
func Agi_ObjVar0(p *Agi_Man_t, id int32) int32 { return int32(Agi_ObjLit0(p, id) >> 1) }
func Agi_ObjVar1(p *Agi_Man_t, id int32) int32 { return int32(Agi_ObjLit1(p, id) >> 1) }
func Agi_ObjVar2(p *Agi_Man_t, id int32) int32 { return int32(Agi_ObjLit2(p, id) >> 1) }
func Agi_ObjSetLit0(p *Agi_Man_t, id int32, lit uint32) {
	p.pObjs[id] = p.pObjs[id]&AGI_M1 | uint64(lit)
}
func Agi_ObjSetLit1(p *Agi_Man_t, id int32, lit uint32) {
	p.pObjs[id] = p.pObjs[id]&AGI_M0 | uint64(lit)<<32
}
func Agi_ObjSetLit2(p *Agi_Man_t, id int32, lit uint32) {
	if p.pThird == nil {
		p.pThird = make([]uint32, len(p.pObjs))
		for i := range p.pThird {
			p.pThird[i] = ^uint32(0)
		}
	}
	p.pThird[id] = lit
}
func Agi_ObjIsC0(p *Agi_Man_t, id int32) int32 { return giaBool(id == 0) }
func Agi_ObjIsPi(p *Agi_Man_t, id int32) int32 { return giaBool(p.pObjs[id]&AGI_PI == AGI_PI) }
func Agi_ObjIsRo(p *Agi_Man_t, id int32) int32 { return giaBool(p.pObjs[id]&AGI_PI == AGI_RO) }
func Agi_ObjIsPo(p *Agi_Man_t, id int32) int32 { return giaBool(p.pObjs[id]&AGI_PI == AGI_PO) }
func Agi_ObjIsRi(p *Agi_Man_t, id int32) int32 { return giaBool(p.pObjs[id]&AGI_PI == AGI_RI) }
func Agi_ObjIsCi(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsPi(p, id) != 0 || Agi_ObjIsRo(p, id) != 0)
}
func Agi_ObjIsCo(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsPo(p, id) != 0 || Agi_ObjIsRi(p, id) != 0)
}
func Agi_ObjIsNode(p *Agi_Man_t, id int32) int32 { return giaBool(p.pObjs[id] < AGI_C0) }
func Agi_ObjIsBuf(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjLit0(p, id) == Agi_ObjLit1(p, id))
}
func Agi_ObjIsAnd(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsNode(p, id) != 0 && Agi_ObjLit0(p, id) < Agi_ObjLit1(p, id))
}
func Agi_ObjIsXor(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsNode(p, id) != 0 && Agi_ObjLit0(p, id) > Agi_ObjLit1(p, id))
}
func Agi_ObjIsMux(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsAnd(p, id) != 0 && Agi_ObjLit2(p, id) != ^uint32(0))
}
func Agi_ObjIsMaj(p *Agi_Man_t, id int32) int32 {
	return giaBool(Agi_ObjIsXor(p, id) != 0 && Agi_ObjLit2(p, id) != ^uint32(0))
}
func Agi_ManAppendObj(p *Agi_Man_t) int32 {
	id := Agi_ManObjNum(p)
	if id >= int32(objectNoFanin) {
		panic("gia: AGI object limit exceeded")
	}
	p.pObjs = append(p.pObjs, 0)
	p.pTravIds = append(p.pTravIds, 0)
	if p.pThird != nil {
		p.pThird = append(p.pThird, ^uint32(0))
	}
	return id
}
func Agi_ManAppendCi(p *Agi_Man_t) int32 {
	id := Agi_ManAppendObj(p)
	p.pObjs[id] = AGI_PI | uint64(len(p.vCis))
	p.vCis = append(p.vCis, id)
	return id << 1
}
func Agi_ManAppendCo(p *Agi_Man_t, lit int32) int32 {
	if lit < 0 || lit>>1 >= Agi_ManObjNum(p) {
		panic("gia: invalid AGI output literal")
	}
	id := Agi_ManAppendObj(p)
	p.pObjs[id] = AGI_PO | uint64(lit)
	p.vCos = append(p.vCos, id)
	return id << 1
}
func agiAppendNode(p *Agi_Man_t, a, b int32) int32 {
	if a < 0 || b < 0 || a>>1 >= Agi_ManObjNum(p) || b>>1 >= Agi_ManObjNum(p) {
		panic("gia: invalid AGI node fanins")
	}
	id := Agi_ManAppendObj(p)
	p.pObjs[id] = uint64(uint32(b))<<32 | uint64(uint32(a))
	p.nNodes++
	return id << 1
}
func Agi_ManAppendAnd(p *Agi_Man_t, a, b int32) int32 {
	if a >= b {
		panic("gia: AGI AND fanins must be ordered")
	}
	return agiAppendNode(p, a, b)
}
func Agi_ManAlloc(capacity int32) *Agi_Man_t {
	if capacity < 16 {
		capacity = 16
	}
	return &Agi_Man_t{pObjs: append(make([]uint64, 0, int(capacity)), AGI_C0), pTravIds: make([]uint32, 1, int(capacity))}
}
func Agi_ManFree(p *Agi_Man_t) {
	if p != nil {
		*p = Agi_Man_t{}
	}
}
func Agi_ManFromGia(p *Gia_Man_t) *Agi_Man_t {
	out := Agi_ManAlloc(p.nObjs)
	out.pName = p.pName
	out.pSpec = p.PSpec
	out.nRegs = p.nRegs
	p.pObjs[0].Value = 0
	for _, obj := range p.pObjs[1:] {
		var lit int32
		switch {
		case Gia_ObjIsCi(obj) != 0:
			lit = Agi_ManAppendCi(out)
			if Gia_ObjCioId(obj) >= Gia_ManPiNum(p) {
				out.pObjs[lit>>1] = AGI_RO | uint64(Gia_ObjCioId(obj))
			}
		case Gia_ObjIsCo(obj) != 0:
			lit = Agi_ManAppendCo(out, Gia_ObjFanin0Copy(obj))
			if Gia_ObjCioId(obj) >= Gia_ManPoNum(p) {
				out.pObjs[lit>>1] = AGI_RI | uint64(uint32(Gia_ObjFanin0Copy(obj)))
			}
		case Gia_ObjIsAnd(obj) != 0:
			lit = agiAppendNode(out, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
			if int(obj.id) < len(p.PMuxes) && p.PMuxes[obj.id] != 0 {
				control := p.PMuxes[obj.id]
				copy := p.pObjs[control>>1].Value ^ (control & 1)
				Agi_ObjSetLit2(out, lit>>1, copy)
			}
		default:
			panic("gia: unsupported AGI source object")
		}
		obj.Value = uint32(lit)
	}
	return out
}
func Agi_ManSuppSize_rec(p *Agi_Man_t, id int32) int32 {
	stack := []int32{id}
	count := int32(0)
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if p.pTravIds[node] == p.nTravIds {
			continue
		}
		p.pTravIds[node] = p.nTravIds
		if node == 0 {
			continue
		}
		if Agi_ObjIsCi(p, node) != 0 {
			count++
			continue
		}
		stack = append(stack, Agi_ObjVar0(p, node))
		if Agi_ObjIsNode(p, node) != 0 {
			stack = append(stack, Agi_ObjVar1(p, node))
			if third := Agi_ObjLit2(p, node); third != ^uint32(0) {
				stack = append(stack, int32(third>>1))
			}
		}
	}
	return count
}
func Agi_ManSuppSizeOne(p *Agi_Man_t, id int32) int32 {
	p.nTravIds++
	if p.nTravIds == 0 {
		clear(p.pTravIds)
		p.nTravIds = 1
	}
	return Agi_ManSuppSize_rec(p, id)
}
func Agi_ManSuppSizeTest(p *Agi_Man_t) int32 {
	started := time.Now()
	count := int32(0)
	for id := int32(1); id < Agi_ManObjNum(p); id++ {
		if Agi_ObjIsNode(p, id) != 0 && Agi_ManSuppSizeOne(p, id) <= 16 {
			count++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\n", count, p.nNodes)
	fmt.Printf("Time = %.2f sec\n", time.Since(started).Seconds())
	return count
}
func Agi_ManTest(p *Gia_Man_t) {
	Gia_ManSuppSizeTest(p)
	copy := Agi_ManFromGia(p)
	Agi_ManSuppSizeTest(copy)
	Agi_ManFree(copy)
}
