package gia

import "fmt"

func Gia_ManDupMuxesNoHash(p *Gia_Man_t) *Gia_Man_t {
	if p.PMuxes != nil {
		panic("gia: expected expanded muxes")
	}
	p.Refs = nil
	Gia_ManCreateRefs(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsMuxType(obj) != 0 {
			Gia_ObjRefDec(p, Gia_ObjFanin0(obj))
			Gia_ObjRefDec(p, Gia_ObjFanin1(obj))
		}
	}
	dst := Gia_ManStart(p.nObjs)
	dst.pName, dst.PSpec = p.pName, p.PSpec
	dst.EnableMuxes()
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	copyLit := func(obj *Gia_Obj_t) int32 { return Gia_ObjLitCopy(p, Gia_ObjToLit(p, obj)) }
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjRefNumId(p, obj.id) == 0 {
			continue
		}
		var value int32
		if Gia_ObjIsMuxType(obj) == 0 {
			value = Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		} else {
			var a, b *Gia_Obj_t
			if Gia_ObjRecognizeExor(obj, &a, &b) != 0 {
				x, y := copyLit(a), copyLit(b)
				value = (x & 1) ^ (y & 1) ^ Gia_ManAppendXorReal(dst, x&^1, y&^1)
			} else {
				control := Gia_ObjRecognizeMux(obj, &a, &b)
				x, y := copyLit(b), copyLit(a)
				if x == y {
					value = x
				} else if x>>1 == y>>1 {
					y = copyLit(control)
					value = (x & 1) ^ (y & 1) ^ Gia_ManAppendXorReal(dst, x&^1, y&^1)
				} else {
					value = Gia_ManAppendMuxReal(dst, copyLit(control), copyLit(a), copyLit(b))
				}
			}
		}
		obj.Value = uint32(value)
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(p))
	if Gia_ManHasDangling(dst) != 0 {
		panic("gia: mux duplication produced dangling nodes")
	}
	return dst
}
func Str_PrintState(cost, super []int32, matrix []uint64, size int32) {
	for i := int32(0); i < size; i++ {
		fmt.Printf("%6d : %6d : %3d   %3d\n", i, super[i]>>1, cost[i]>>4, cost[i]&15)
	}
	fmt.Println()
}
