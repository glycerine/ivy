package gia

import "github.com/glycerine/abcgo/misc/vec"

func giaMappedEdge(src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	a, _, _, kind := src.Gate(obj.id)
	if kind != 6 {
		panic("gia: expected mapped output")
	}
	return Gia_ObjLitCopy(src, a)
}
func giaCleanComposed(out *Gia_Man_t) *Gia_Man_t {
	clean := Gia_ManCleanup(out)
	Gia_ManStop(out)
	return clean
}
func Gia_ManDupInvMiter(src, inv *Gia_Man_t) *Gia_Man_t {
	if Gia_ManRegNum(src) <= 0 || Gia_ManRegNum(inv) != 0 || Gia_ManCoNum(inv) != 1 || Gia_ManCiNum(inv) != Gia_ManRegNum(src) {
		panic("gia: invalid inductive invariant interface")
	}
	out := Gia_ManStart(src.nObjs + 2*inv.nObjs)
	out.CopyIdentityFrom(src)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaComposeNodes(out, src, true)
	invariant := func(next bool) int32 {
		Gia_ManFillValue(inv)
		inv.pObjs[0].Value = 0
		for i := int32(0); i < Gia_ManRegNum(src); i++ {
			lit := int32(Gia_ManRo(src, i).Value)
			if next {
				lit = giaMappedEdge(src, Gia_ManRi(src, i))
			}
			Gia_ManCi(inv, i).Value = uint32(lit)
		}
		giaComposeNodes(out, inv, true)
		return giaMappedEdge(inv, Gia_ManCo(inv, 0))
	}
	current, next := invariant(false), invariant(true)
	Gia_ManAppendCo(out, Gia_ManHashAnd(out, current, next^1))
	return giaCleanComposed(out)
}
func Gia_ManDupAppendCones(src *Gia_Man_t, cones []*Gia_Man_t, count, onlyRegs int32) *Gia_Man_t {
	if count < 0 || int(count) > len(cones) {
		panic("gia: invalid cone count")
	}
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaComposeNodes(out, src, true)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for _, cone := range cones[:count] {
		expected := Gia_ManCiNum(src)
		start := int32(0)
		if onlyRegs != 0 {
			expected = Gia_ManRegNum(src)
			start = Gia_ManPiNum(src)
		}
		if cone == nil || Gia_ManPoNum(cone) != 1 || Gia_ManRegNum(cone) != 0 || Gia_ManPiNum(cone) != expected {
			panic("gia: invalid additional cone interface")
		}
		Gia_ManFillValue(cone)
		cone.pObjs[0].Value = 0
		for i := int32(0); i < expected; i++ {
			Gia_ManPi(cone, i).Value = Gia_ManCi(src, start+i).Value
		}
		giaComposeNodes(out, cone, true)
		giaDupMappedOutput(out, cone, Gia_ManPo(cone, 0))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}
func Gia_ManDupSelf(src *Gia_Man_t) *Gia_Man_t {
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	Gia_ManHashAlloc(out)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	control := Gia_ManAppendCi(out)
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	giaComposeNodes(out, src, true)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		ri := Gia_ManRi(src, i)
		lit := Gia_ManHashMux(out, control, giaMappedEdge(src, ri), int32(Gia_ManRo(src, i).Value))
		ri.Value = uint32(Gia_ManAppendCo(out, lit))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}
func Gia_ManDupFlopClass(src *Gia_Man_t, class int32) *Gia_Man_t {
	if src.VFlopClasses == nil || vec.Vec_IntSize(src.VFlopClasses) != int(Gia_ManRegNum(src)) {
		panic("gia: invalid flop classes")
	}
	out := giaDupMappedStart(src)
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		Gia_ManPi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	var selected int32
	for pass := 0; pass < 2; pass++ {
		for i := int32(0); i < Gia_ManRegNum(src); i++ {
			match := vec.Vec_IntEntry(src.VFlopClasses, int(i)) == class
			if match != (pass == 1) {
				continue
			}
			Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
			if match {
				selected++
			}
		}
	}
	giaDupMappedNodes(out, src)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	for pass := 0; pass < 2; pass++ {
		for i := int32(0); i < Gia_ManRegNum(src); i++ {
			match := vec.Vec_IntEntry(src.VFlopClasses, int(i)) == class
			if match == (pass == 1) {
				giaDupMappedOutput(out, src, Gia_ManRi(src, i))
			}
		}
	}
	Gia_ManSetRegNum(out, selected)
	return out
}
func Gia_ManDupTimes(src *Gia_Man_t, times int32) *Gia_Man_t {
	if times <= 0 {
		panic("gia: invalid duplication count")
	}
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	pis, ros, pos, ris := make([]int32, 0), make([]int32, 0), make([]int32, 0), make([]int32, 0)
	for t := int32(0); t < times; t++ {
		Gia_ManFillValue(src)
		src.pObjs[0].Value = 0
		for _, obj := range src.pObjs[1:] {
			_, _, _, kind := src.Gate(obj.id)
			if kind == 1 {
				lit := Gia_ManAppendCi(out)
				obj.Value = uint32(lit)
				if Gia_ObjIsPi(src, obj) != 0 {
					pis = append(pis, lit>>1)
				} else {
					ros = append(ros, lit>>1)
				}
			} else if kind == 6 {
				giaDupMappedOutput(out, src, obj)
				id := int32(obj.Value >> 1)
				if Gia_ObjIsPo(src, obj) != 0 {
					pos = append(pos, id)
				} else {
					ris = append(ris, id)
				}
			} else {
				Gia_ManDupOrderDfs_rec(out, src, obj)
			}
		}
	}
	vec.Vec_IntClear(out.vCis)
	for _, id := range append(pis, ros...) {
		Gia_ObjSetCioId(out.pObjs[id], int32(vec.Vec_IntSize(out.vCis)))
		vec.Vec_IntPush(out.vCis, id)
	}
	vec.Vec_IntClear(out.vCos)
	for _, id := range append(pos, ris...) {
		Gia_ObjSetCioId(out.pObjs[id], int32(vec.Vec_IntSize(out.vCos)))
		vec.Vec_IntPush(out.vCos, id)
	}
	Gia_ManSetRegNum(out, times*Gia_ManRegNum(src))
	return out
}
