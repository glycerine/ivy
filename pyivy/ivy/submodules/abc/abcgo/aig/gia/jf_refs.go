package gia

import "github.com/glycerine/abcgo/misc/vec"

func Jf_ObjIsUnit(p *Gia_Obj_t) bool { return p.fMark0 == 0 }
func Jf_ObjCleanUnit(p *Gia_Obj_t) {
	if !Jf_ObjIsUnit(p) {
		panic("JF object is already non-unit")
	}
	p.fMark0 = 1
}
func Jf_ObjSetUnit(p *Gia_Obj_t) { p.fMark0 = 0 }

// jfInitRefs discounts duplicated MUX controls and XOR data inputs. Coarsened
// MUX children are marked non-unit so mapping can treat them as a single node.
func jfInitRefs(p *Gia_Man_t, coarsen bool) ([]float32, int32) {
	if p.Refs != nil {
		panic("gia: reference counts already initialized")
	}
	p.Refs = make([]int32, p.nObjs)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		Gia_ObjRefFanin0Inc(p, obj)
		if Gia_ObjIsBuf(obj) != 0 {
			continue
		}
		Gia_ObjRefFanin1Inc(p, obj)
		if Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var data1, data0 *Gia_Obj_t
		control := Gia_ObjRecognizeMux(obj, &data1, &data0)
		Gia_ObjRefDec(p, Gia_Regular(control))
		if Gia_Regular(data1) == Gia_Regular(data0) {
			Gia_ObjRefDec(p, Gia_Regular(data1))
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ObjRefFanin0Inc(p, Gia_ManObj(p, id))
	}
	var coarse int32
	if coarsen {
		for _, obj := range p.pObjs {
			if Gia_ObjIsMuxType(obj) == 0 {
				continue
			}
			for _, child := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
				if Gia_ObjRefNum(p, child) != 1 {
					continue
				}
				Jf_ObjSetUnit(Gia_ObjFanin0(child))
				Jf_ObjSetUnit(Gia_ObjFanin1(child))
				Jf_ObjCleanUnit(child)
				coarse++
			}
		}
	}
	result := make([]float32, len(p.Refs))
	for i, refs := range p.Refs {
		if refs < 1 {
			refs = 1
		}
		result[i] = float32(refs)
	}
	return result, coarse
}
