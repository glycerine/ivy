package gia

import "fmt"

func Gia_ObjTempRepr(p *Gia_Man_t, id, level int32) *Gia_Obj_t {
	repr := Gia_ObjRepr(p, id)
	if Gia_ObjProved(p, id) == 0 || Gia_ObjFailed(p, id) != 0 || repr == GIA_VOID {
		return nil
	}
	if repr == 0 {
		return Gia_ManConst0(p)
	}
	for member := repr; member != 0; member = Gia_ObjNext(p, member) {
		if Gia_ObjFailed(p, member) != 0 || Gia_ObjProved(p, member) == 0 {
			continue
		}
		if member >= id {
			return nil
		}
		if Gia_ObjLevelId(p, member) < level {
			return Gia_ManObj(p, member)
		}
	}
	panic("gia: no temporary representative")
}
func Gia_GenerateReducedLevel(p *Gia_Man_t, level int32, roots **Vec_Ptr_t) *Gia_Man_t {
	*roots = Vec_PtrAlloc(100)
	result := Gia_ManStart(p.nObjs)
	result.pName, result.PSpec = p.pName, p.PSpec
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(result))
	}
	Gia_ManHashAlloc(result)
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjLevelId(p, id) > level {
			continue
		}
		if Gia_ObjLevelId(p, id) == level {
			Vec_PtrPush(*roots, obj)
		}
		if Gia_ObjLevelId(p, id) < level {
			if repr := Gia_ObjTempRepr(p, id, level); repr != nil {
				obj.Value = repr.Value ^ repr.fPhase ^ obj.fPhase
				continue
			}
		}
		obj.Value = uint32(Gia_ManHashAnd(result, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
	}
	Gia_ManCreateRefs(result)
	Gia_ManFillValue(result)
	Gia_ManIncrementTravId(result)
	return result
}
func Gia_CollectRelatedClasses(p *Gia_Man_t, roots *Vec_Ptr_t) *Vec_Ptr_t {
	classes := Vec_PtrAlloc(100)
	seen := make(map[int32]bool)
	seen[0] = true
	for i := 0; i < Vec_PtrSize(roots); i++ {
		root := Vec_PtrEntry(roots, i).(*Gia_Obj_t)
		repr := Gia_ObjRepr(p, Gia_ObjId(p, root))
		if repr == GIA_VOID || seen[repr] {
			continue
		}
		seen[repr] = true
		Vec_PtrPush(classes, Gia_ManObj(p, repr))
	}
	return classes
}
func Gia_CollectClassMembers(p *Gia_Man_t, repr *Gia_Obj_t, members *Vec_Ptr_t, level int32) *Gia_Obj_t {
	Vec_PtrClear(members)
	var temp *Gia_Obj_t
	for id := Gia_ObjId(p, repr); id != 0; id = Gia_ObjNext(p, id) {
		current := Gia_ObjLevelId(p, id)
		if current == level {
			Vec_PtrPush(members, Gia_ManObj(p, id))
		}
		if temp == nil && current < level {
			temp = Gia_ManObj(p, id)
		}
	}
	return temp
}
func Gia_GiarfStorePatternTry(info, pres *Vec_Ptr_t, bit int32, lits []int32, nLits int32) int32 {
	if bit < 0 || bit >= 32 || nLits < 0 || int(nLits) > len(lits) {
		panic("gia: invalid simulation pattern")
	}
	mask := uint32(1) << uint(bit)
	for _, lit := range lits[:nLits] {
		data := Vec_PtrEntry(info, int(lit>>1)).([]uint32)
		present := Vec_PtrEntry(pres, int(lit>>1)).([]uint32)
		if present[0]&mask != 0 && (data[0]&mask != 0) == (lit&1 != 0) {
			return 0
		}
	}
	for _, lit := range lits[:nLits] {
		data := Vec_PtrEntry(info, int(lit>>1)).([]uint32)
		present := Vec_PtrEntry(pres, int(lit>>1)).([]uint32)
		present[0] |= mask
		if (data[0]&mask != 0) == (lit&1 != 0) {
			data[0] ^= mask
		}
	}
	return 1
}
func Gia_GiarfStorePattern(info, pres *Vec_Ptr_t, cex *Vec_Int_t) int32 {
	for bit := int32(1); bit < 32; bit++ {
		if Gia_GiarfStorePatternTry(info, pres, bit, Vec_IntArray(cex), int32(Vec_IntSize(cex))) != 0 {
			return 1
		}
	}
	return 0
}
func Gia_GiarfInsertPattern(p *Hcd_Man_t, cex *Vec_Int_t, bit int32) {
	if bit < 0 || bit >= 32 {
		panic("gia: invalid pattern bit")
	}
	mask := uint32(1) << uint(bit)
	for _, lit := range Vec_IntArray(cex) {
		data := Hcd_ObjSimP(p, Gia_ManCi(p.pGia, lit>>1).id)
		if (data[0]&mask != 0) == (lit&1 != 0) {
			data[0] ^= mask
		}
	}
}
func Gia_GiarfPrintClasses(p *Gia_Man_t) {
	var fails, proves, both, total int32
	for id := int32(0); id < p.nObjs; id++ {
		f, q := Gia_ObjFailed(p, id), Gia_ObjProved(p, id)
		fails += f
		proves += q
		if f != 0 && q != 0 {
			both++
		}
		if Gia_ObjReprObj(p, id) != nil {
			total++
		}
	}
	fmt.Printf("nFails = %7d.  nProves = %7d.  nBoth = %7d.  nTotal = %7d.\n", fails, proves, both, total)
}
