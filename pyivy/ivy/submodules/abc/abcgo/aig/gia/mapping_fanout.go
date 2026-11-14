package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManStartMappingFanoutMap(p *Gia_Man_t, counts *Vec_Int_t) *Vec_Int_t {
	result := vec.Vec_IntStart(int(p.nObjs))
	offset := p.nObjs
	for id := int32(0); id < p.nObjs; id++ {
		count := vec.Vec_IntEntry(counts, int(id))
		if count < 0 {
			panic("gia: negative mapping fanout count")
		}
		if count == 0 {
			continue
		}
		vec.Vec_IntWriteEntry(result, int(id), offset)
		offset += count
		vec.Vec_IntFillExtra(result, int(offset), 0)
	}
	return result
}
func Gia_ManStaticMappingFanoutStart(p *Gia_Man_t) {
	if p.vFanoutNums != nil || p.vFanout != nil || p.VMapping == nil {
		panic("gia: invalid mapping fanout initialization")
	}
	counts := vec.Vec_IntStart(int(p.nObjs))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			vec.Vec_IntAddToEntry(counts, int(fanin), 1)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		vec.Vec_IntAddToEntry(counts, int(Gia_ObjFaninId0(obj, obj.id)), 1)
	}
	p.vFanoutNums = counts
	p.vFanout = Gia_ManStartMappingFanoutMap(p, counts)
	used := make([]int32, p.nObjs)
	add := func(id, fanout int32) { Gia_ObjSetFanoutInt(p, Gia_ManObj(p, id), used[id], fanout); used[id]++ }
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			add(fanin, id)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		add(Gia_ObjFaninId0(obj, obj.id), obj.id)
	}
}
func Gia_ManStaticFanoutTest(p *Gia_Man_t) {
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	for _, obj := range p.pObjs {
		fmt.Printf("Obj %d   Fanouts : ", obj.id)
		for i := int32(0); i < Gia_ObjFanoutNum(p, obj); i++ {
			fmt.Printf("%5d ", Gia_ObjFanoutId(p, obj.id, i))
		}
		fmt.Println()
	}
}
