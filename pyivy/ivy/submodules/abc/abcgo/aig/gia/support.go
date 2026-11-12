package gia

import (
	"fmt"
	"time"
)

func Gia_ManIncrementTravId(p *Gia_Man_t) {
	if len(p.pTravIds) < int(p.nObjs) {
		size := 2 * len(p.pTravIds)
		if size < int(p.nObjs)+100 {
			size = int(p.nObjs) + 100
		}
		p.pTravIds = append(p.pTravIds, make([]int32, size-len(p.pTravIds))...)
	}
	p.nTravIdsAlloc = int32(len(p.pTravIds))
	if p.nTravIds == int32(^uint32(0)>>1) {
		clear(p.pTravIds)
		p.nTravIds = 1
	} else {
		p.nTravIds++
	}
}
func Gia_ObjSetTravIdCurrentId(p *Gia_Man_t, id int32)  { p.pTravIds[id] = p.nTravIds }
func Gia_ObjSetTravIdPreviousId(p *Gia_Man_t, id int32) { p.pTravIds[id] = p.nTravIds - 1 }
func Gia_ObjIsTravIdCurrentId(p *Gia_Man_t, id int32) int32 {
	return giaBool(p.pTravIds[id] == p.nTravIds)
}
func Gia_ObjIsTravIdPreviousId(p *Gia_Man_t, id int32) int32 {
	return giaBool(p.pTravIds[id] == p.nTravIds-1)
}
func Gia_ObjSetTravIdCurrent(p *Gia_Man_t, obj *Gia_Obj_t)  { Gia_ObjSetTravIdCurrentId(p, obj.id) }
func Gia_ObjSetTravIdPrevious(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjSetTravIdPreviousId(p, obj.id) }
func Gia_ObjIsTravIdCurrent(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjIsTravIdCurrentId(p, obj.id)
}
func Gia_ObjIsTravIdPrevious(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjIsTravIdPreviousId(p, obj.id)
}

func Gia_ManSuppSize_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	stack := []int32{Gia_ObjId(p, obj)}
	count := int32(0)
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if p.pTravIds[id] == p.nTravIds {
			continue
		}
		p.pTravIds[id] = p.nTravIds
		if id == 0 {
			continue
		}
		node := p.pObjs[id]
		if Gia_ObjIsCi(node) != 0 {
			count++
			continue
		}
		stack = append(stack, giaObjectFanins(p, node)...)
	}
	return count
}
func Gia_ManSuppSizeOne(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	Gia_ManIncrementTravId(p)
	return Gia_ManSuppSize_rec(p, obj)
}
func Gia_ManSuppSize(p *Gia_Man_t, nodes []int32, count int32) int32 {
	Gia_ManIncrementTravId(p)
	total := int32(0)
	for _, id := range nodes[:count] {
		total += Gia_ManSuppSize_rec(p, Gia_ManObj(p, id))
	}
	return total
}
func Gia_ManSuppSizeTest(p *Gia_Man_t) int32 {
	started := time.Now()
	count := int32(0)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ManSuppSizeOne(p, obj) <= 16 {
			count++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\n", count, Gia_ManAndNum(p))
	fmt.Printf("Time = %.2f sec\n", time.Since(started).Seconds())
	return count
}
