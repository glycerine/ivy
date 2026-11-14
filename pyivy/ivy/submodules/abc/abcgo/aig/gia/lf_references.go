package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lf_ManSetFlowRefInc(g *Gia_Man_t, refs *Vec_Flt_t, offsets *Vec_Int_t, id int32) {
	if Gia_ObjIsAndNotBuf(Gia_ManObj(g, id)) != 0 {
		vec.Vec_FltAddToEntry(refs, int(vec.Vec_IntEntry(offsets, int(id))), 1)
	}
}
func Lf_ManSetFlowRefs(g *Gia_Man_t, refs *Vec_Flt_t, offsets *Vec_Int_t) {
	vec.Vec_FltFill(refs, int(Gia_ManAndNotBufNum(g)), 0)
	change := func(obj *Gia_Obj_t, delta float32) {
		if Gia_ObjIsAndNotBuf(obj) != 0 {
			vec.Vec_FltAddToEntry(refs, int(vec.Vec_IntEntry(offsets, int(obj.id))), delta)
		}
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		change(Gia_ObjFanin0(obj), 1)
		if Gia_ObjIsBuf(obj) != 0 {
			continue
		}
		change(Gia_ObjFanin1(obj), 1)
		if g.PMuxes != nil {
			if Gia_ObjIsMuxId(g, obj.id) != 0 {
				change(Gia_ObjFanin2(g, obj), 1)
			}
		} else if Gia_ObjIsMuxType(obj) != 0 {
			var d1, d0 *Gia_Obj_t
			ctrl := Gia_Regular(Gia_ObjRecognizeMux(obj, &d1, &d0))
			d0 = Gia_Regular(d0)
			d1 = Gia_Regular(d1)
			change(ctrl, -1)
			if d0 == d1 {
				change(d0, -1)
			}
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		change(Gia_ObjFanin0(Gia_ManObj(g, id)), 1)
	}
	for i := range vec.Vec_FltArray(refs) {
		vec.Vec_FltUpdateEntry(refs, i, 1)
	}
}
func Lf_ManSetCutRefs(p *Lf_Man_t) {
	used := vec.Vec_PtrSize(&p.vMemSets)*(1<<LF_LOG_PAGE) - vec.Vec_IntSize(&p.vFreeSets)
	if used != 0 {
		fmt.Printf("The number of used cutsets = %d.\n", used)
	}
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if obj.Value != 0 {
			panic("LF cut references not cleared")
		}
		if Gia_ObjIsBuf(obj) != 0 {
			continue
		}
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
			if Gia_ObjIsAndNotBuf(fanin) != 0 {
				fanin.Value++
			}
		}
		if Gia_ObjIsMuxId(p.pGia, obj.id) != 0 {
			fanin := Gia_ObjFanin2(p.pGia, obj)
			if Gia_ObjIsAndNotBuf(fanin) != 0 {
				fanin.Value++
			}
		}
		if Gia_ObjSibl(p.pGia, obj.id) != 0 {
			sibling := Gia_ObjSiblObj(p.pGia, obj.id)
			if Gia_ObjIsAndNotBuf(sibling) != 0 {
				sibling.Value++
			}
		}
	}
}
