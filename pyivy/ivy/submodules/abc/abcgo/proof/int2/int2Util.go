package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Int2_ManComputeCoPres returns CO indices needed by the SOP. A nil SOP
// denotes the initial-state cube, which uses the first register output.
func Int2_ManComputeCoPres(vSop *Vec_Int_t, nRegs int32) *Vec_Int_t {
	if nRegs < 0 {
		panic("int2: negative register count")
	}
	out := vec.Vec_IntAlloc(100)
	if vSop == nil {
		if nRegs > 0 {
			vec.Vec_IntPush(out, 0)
		}
		return out
	}
	used := make([]bool, nRegs)
	for pos := 1; pos < vec.Vec_IntSize(vSop); {
		count := int(vec.Vec_IntEntry(vSop, pos))
		pos++
		if count < 0 || pos+count > vec.Vec_IntSize(vSop) {
			vec.Vec_IntFree(out)
			panic("int2: malformed SOP cube")
		}
		for k := 0; k < count; k++ {
			lit := vec.Vec_IntEntry(vSop, pos+k)
			reg := lit >> 1
			if reg < 0 || reg >= nRegs {
				vec.Vec_IntFree(out)
				panic("int2: SOP literal is outside register range")
			}
			used[reg] = true
		}
		pos += count
	}
	for reg, yes := range used {
		if yes {
			vec.Vec_IntPush(out, int32(reg+1))
		}
	}
	return out
}

func Int2_ManCollectInternal_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t, seen map[int32]bool) {
	id := gia.Gia_ObjId(p, obj)
	if seen[id] {
		return
	}
	seen[id] = true
	if gia.Gia_ObjIsCi(obj) != 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("int2: expected a CI or AND in support cone")
	}
	Int2_ManCollectInternal_rec(p, gia.Gia_ObjFanin0(obj), nodes, seen)
	Int2_ManCollectInternal_rec(p, gia.Gia_ObjFanin1(obj), nodes, seen)
	vec.Vec_IntPush(nodes, id)
}

func Int2_ManCollectInternal(p *Gia_Man_t, vCoPres *Vec_Int_t) *Vec_Int_t {
	if p == nil || vCoPres == nil {
		panic("int2: nil graph or CO set")
	}
	nodes := vec.Vec_IntAlloc(1000)
	seen := make(map[int32]bool, gia.Gia_ManObjNum(p))
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		seen[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = true
	}
	seen[gia.Gia_ObjId(p, gia.Gia_ManConst0(p))] = true
	for i := 0; i < vec.Vec_IntSize(vCoPres); i++ {
		coID := vec.Vec_IntEntry(vCoPres, i)
		if coID < 0 || coID >= gia.Gia_ManCoNum(p) {
			vec.Vec_IntFree(nodes)
			panic("int2: CO index outside graph")
		}
		co := gia.Gia_ManCo(p, coID)
		Int2_ManCollectInternal_rec(p, gia.Gia_ObjFanin0(co), nodes, seen)
	}
	return nodes
}

// Int2_ManProbToGia builds a single-output graph for a prioritized cube SOP.
func Int2_ManProbToGia(p *Gia_Man_t, vSop *Vec_Int_t) *Gia_Man_t {
	if p == nil || gia.Gia_ManPoNum(p) != 1 {
		panic("int2: expected a non-nil single-output graph")
	}
	nRegs := gia.Gia_ManRegNum(p)
	coPres := Int2_ManComputeCoPres(vSop, nRegs)
	nodes := Int2_ManCollectInternal(p, coPres)
	defer vec.Vec_IntFree(coPres)
	defer vec.Vec_IntFree(nodes)

	out := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	name, spec := p.Identity()
	out.SetIdentity(name, spec)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(p, i), gia.Gia_ManAppendCi(out))
	}
	gia.Gia_ManHashAlloc(out)
	for i := 0; i < vec.Vec_IntSize(nodes); i++ {
		obj := gia.Gia_ManObj(p, vec.Vec_IntEntry(nodes, i))
		lit := gia.Gia_ManHashAnd(out, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
		gia.Gia_ObjSetValue(obj, lit)
	}
	for i := 0; i < vec.Vec_IntSize(coPres); i++ {
		co := gia.Gia_ManCo(p, vec.Vec_IntEntry(coPres, i))
		gia.Gia_ObjSetValue(co, gia.Gia_ObjFanin0Copy(co))
	}
	sop := int32(0)
	if vSop == nil {
		// The initial suffix is the property's output. Subsequent suffixes
		// use the SOP over next-state outputs below.
		sop = gia.Gia_ObjValue(gia.Gia_ManPo(p, 0))
	} else {
		for pos := 1; pos < vec.Vec_IntSize(vSop); {
			count := int(vec.Vec_IntEntry(vSop, pos))
			pos++
			if count < 0 || pos+count > vec.Vec_IntSize(vSop) {
				gia.Gia_ManHashStop(out)
				gia.Gia_ManStop(out)
				panic("int2: malformed SOP cube")
			}
			cube := int32(1)
			for k := 0; k < count; k++ {
				lit := vec.Vec_IntEntry(vSop, pos+k)
				reg := lit >> 1
				if reg < 0 || reg >= nRegs {
					gia.Gia_ManHashStop(out)
					gia.Gia_ManStop(out)
					panic("int2: SOP literal is outside register range")
				}
				ro := gia.Gia_ManRi(p, reg)
				mapped := gia.Gia_ObjFanin0Copy(ro) ^ (lit & 1)
				cube = gia.Gia_ManHashAnd(out, cube, mapped)
			}
			pos += count
			sop = gia.Gia_ManHashOr(out, sop, cube)
		}
	}
	gia.Gia_ManAppendCo(out, sop)
	gia.Gia_ManHashStop(out)
	clean := gia.Gia_ManCleanup(out)
	gia.Gia_ManStop(out)
	return clean
}
