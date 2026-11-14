package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

func Mf_ManDeriveMappingGia(p *Mf_Man_t) *Gia_Man_t {
	if p.pPars.FCutMin == 0 {
		panic("gia: truth-minimized mapping required")
	}
	src := p.pGia
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	copies := make([]int32, src.nObjs)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for _, id := range vec.Vec_IntArray(src.vCis) {
		copies[id] = Gia_ManAppendCi(dst)
	}
	records := []int32{0}
	offsets := make(map[int32]int32)
	var scratch []uint32
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		size, fun := Mf_CutSize(cut), Mf_CutFunc(cut)
		if size == 0 {
			if fun>>1 != 0 {
				panic("gia: invalid constant truth")
			}
			copies[obj.id] = fun
			continue
		}
		if size == 1 {
			if fun>>1 != 1 {
				panic("gia: invalid unit truth")
			}
			copies[obj.id] = copies[cut[1]] ^ (fun & 1)
			continue
		}
		leaves := make([]int32, size)
		for i, id := range cut[1 : 1+size] {
			leaves[i] = copies[id]
			if leaves[i] < 0 {
				panic("gia: missing mapped leaf")
			}
		}
		t := vec.Vec_MemReadEntry(p.vTtMem, int(uint32(fun)>>1))
		truth := make([]uint32, 2*len(t))
		for i, w := range t {
			truth[2*i], truth[2*i+1] = uint32(w), uint32(w>>32)
		}
		lit := kit.Kit_TruthToGia(dst, truth, size, &scratch, leaves, 0)
		copies[obj.id] = lit ^ (fun & 1)
		offsets[lit>>1] = int32(len(records))
		records = append(records, size)
		for _, leaf := range leaves {
			records = append(records, leaf>>1)
		}
		records = append(records, copies[obj.id]>>1)
	}
	for _, id := range vec.Vec_IntArray(src.vCos) {
		obj := Gia_ManObj(src, id)
		lit := copies[Gia_ObjFanin0(obj).id]
		if lit < 0 {
			panic("gia: missing output driver")
		}
		Gia_ManAppendCo(dst, lit^Gia_ObjFaninC0(obj))
	}
	mapping := vec.Vec_IntStart(int(dst.nObjs))
	for id, offset := range offsets {
		vec.Vec_IntWriteEntry(mapping, int(id), dst.nObjs+offset)
	}
	for _, v := range records {
		vec.Vec_IntPush(mapping, v)
	}
	dst.vMapping = mapping
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}

func Mf_ManDeriveMappingCoarse(p *Mf_Man_t) *Gia_Man_t {
	src := p.pGia
	if p.pPars.FCutMin != 0 || src.pMuxes == nil {
		panic("gia: coarse mapping requires coarse gates without cut minimization")
	}
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	for _, id := range vec.Vec_IntArray(src.vCis) {
		Gia_ManObj(src, id).Value = uint32(Gia_ManAppendCi(dst))
	}
	records := []int32{0}
	offsets := make(map[int32]int32)
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		var lit int32
		if Gia_ObjIsMuxId(src, obj.id) != 0 {
			lit = Gia_ManAppendMux(dst, Gia_ObjFanin2Copy(src, obj), Gia_ObjFanin1Copy(obj), Gia_ObjFanin0Copy(obj))
		} else if Gia_ObjIsXor(obj) != 0 {
			lit = Gia_ManAppendXor(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		} else {
			lit = Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		}
		obj.Value = uint32(lit)
		if Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		size := Mf_CutSize(cut)
		offsets[lit>>1] = int32(len(records))
		records = append(records, size)
		for _, id := range cut[1 : 1+size] {
			records = append(records, int32(Gia_ManObj(src, id).Value)>>1)
		}
		records = append(records, lit>>1)
	}
	for _, id := range vec.Vec_IntArray(src.vCos) {
		obj := Gia_ManObj(src, id)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	mapping := vec.Vec_IntStart(int(dst.nObjs))
	for id, offset := range offsets {
		vec.Vec_IntWriteEntry(mapping, int(id), dst.nObjs+offset)
	}
	for _, v := range records {
		vec.Vec_IntPush(mapping, v)
	}
	dst.vMapping = mapping
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
