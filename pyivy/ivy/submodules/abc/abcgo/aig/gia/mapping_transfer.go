package gia

import "fmt"

func Gia_ManMappingVerify_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 1
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAndNotBuf(obj) == 0 {
		return 1
	}
	if Gia_ObjIsLut(p, obj.id) == 0 {
		fmt.Printf("Gia_ManMappingVerify: Internal node %d does not have mapping.\n", obj.id)
		return 0
	}
	for _, id := range Gia_ObjLutFanins(p, obj.id) {
		if Gia_ManMappingVerify_rec(p, Gia_ManObj(p, id)) == 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManMappingVerify(p *Gia_Man_t) {
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: mapping is absent")
	}
	Gia_ManIncrementTravId(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsBuf(obj) == 0 && Gia_ObjIsCo(obj) == 0 {
			continue
		}
		fanin := Gia_ObjFanin0(obj)
		if Gia_ObjIsAndNotBuf(fanin) == 0 {
			continue
		}
		if Gia_ObjIsLut(p, fanin.id) == 0 {
			fmt.Printf("Gia_ManMappingVerify: CO driver %d does not have mapping.\n", fanin.id)
			continue
		}
		Gia_ManMappingVerify_rec(p, fanin)
	}
}

// Gia_ManTransferMapping copies LUT records through source Value literals.
func Gia_ManTransferMapping(dst, src *Gia_Man_t) {
	if src.VMapping == nil {
		return
	}
	Gia_ManMappingVerify(src)
	mapping := Vec_IntStart(int(dst.nObjs))
	for id := int32(1); id < src.nObjs; id++ {
		if Gia_ObjIsLut(src, id) == 0 {
			continue
		}
		value := src.pObjs[id].Value
		if value == ^uint32(0) {
			continue
		}
		if value&1 != 0 {
			panic("gia: complemented mapped LUT copy")
		}
		mapped := int32(value >> 1)
		if mapped < 0 || mapped >= dst.nObjs {
			panic("gia: invalid mapped LUT copy")
		}
		Vec_IntWriteEntry(mapping, int(mapped), int32(Vec_IntSize(mapping)))
		place := Vec_IntSize(mapping)
		Vec_IntPush(mapping, Gia_ObjLutSize(src, id))
		for _, fanin := range Gia_ObjLutFanins(src, id) {
			copy := src.pObjs[fanin].Value
			if copy == ^uint32(0) {
				Vec_IntAddToEntry(mapping, place, -1)
			} else {
				Vec_IntPush(mapping, int32(copy>>1))
			}
		}
		mux := Gia_ObjLutMuxId(src, id)
		negative := mux < 0
		if negative {
			mux = -mux
		}
		copy := src.pObjs[mux].Value
		mappedMux := int32(copy >> 1)
		if negative {
			mappedMux = -mappedMux
		}
		Vec_IntPush(mapping, mappedMux)
	}
	dst.VMapping = mapping
	Gia_ManMappingVerify(dst)
}
func Gia_ManTransferPacking(dst, src *Gia_Man_t) {
	if src.VPacking == nil {
		return
	}
	if dst.VPacking != nil {
		panic("gia: destination already has packing")
	}
	data := Vec_IntArray(src.VPacking)
	if len(data) == 0 {
		panic("gia: empty packing")
	}
	result := Vec_IntAlloc(len(data))
	Vec_IntPush(result, data[0])
	entries := int32(0)
	for i := 1; i < len(data); {
		size := int(data[i])
		i++
		if size < 1 || size > 3 || i+size > len(data) {
			panic("gia: invalid packing record")
		}
		Vec_IntPush(result, int32(size))
		for k := 0; k < size; k++ {
			id := data[i+k]
			if Gia_ObjIsLut(src, id) == 0 {
				panic("gia: packed source is not a LUT")
			}
			value := Gia_ManObj(src, id).Value
			if value == ^uint32(0) {
				panic("gia: packed LUT has no copy")
			}
			mapped := int32(value >> 1)
			if Gia_ObjIsLut(dst, mapped) == 0 {
				panic("gia: packed destination is not a LUT")
			}
			Vec_IntPush(result, mapped)
		}
		i += size
		entries++
	}
	if entries != data[0] {
		panic("gia: packing count mismatch")
	}
	dst.VPacking = result
}
