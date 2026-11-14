package gia

import "github.com/glycerine/abcgo/misc/vec"

func strSuperRec(p *Gia_Man_t, lit int32, xor bool) {
	obj := Gia_ManObj(p, lit>>1)
	stop := lit&1 != 0 || Gia_ObjRefNum(p, obj) > 1 || vec.Vec_IntSize(p.vSuper) > STR_SUPER
	if xor {
		stop = stop || Gia_ObjIsXor(obj) == 0
	} else {
		stop = stop || Gia_ObjIsAndReal(p, obj) == 0
	}
	if stop {
		vec.Vec_IntPush(p.vSuper, lit)
		return
	}
	a, b := 2*Gia_ObjFaninId0(obj, obj.id)+Gia_ObjFaninC0(obj), 2*Gia_ObjFaninId1(obj, obj.id)+Gia_ObjFaninC1(obj)
	if xor && (a&1 != 0 || b&1 != 0) {
		panic("gia: complemented XOR supergate input")
	}
	strSuperRec(p, a, xor)
	strSuperRec(p, b, xor)
}
func strSuperCollect(p *Gia_Man_t, obj *Gia_Obj_t) {
	if p.vSuper == nil {
		p.vSuper = vec.Vec_IntAlloc(STR_SUPER)
	} else {
		vec.Vec_IntClear(p.vSuper)
	}
	xor := Gia_ObjIsXor(obj) != 0
	if !xor && Gia_ObjIsAndReal(p, obj) == 0 {
		panic("gia: invalid supergate")
	}
	a, b := 2*Gia_ObjFaninId0(obj, obj.id)+Gia_ObjFaninC0(obj), 2*Gia_ObjFaninId1(obj, obj.id)+Gia_ObjFaninC1(obj)
	strSuperRec(p, a, xor)
	strSuperRec(p, b, xor)
	vec.Vec_IntSort(p.vSuper, 0)
	if xor {
		Gia_ManSimplifyXor(p.vSuper)
	} else {
		Gia_ManSimplifyAnd(p.vSuper)
	}
}
func Str_ManNormalize_rec(ntk *Str_Ntk_t, p *Gia_Man_t, obj *Gia_Obj_t, groups *Vec_Wec_t, roots *Vec_Int_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	obj.Value = 0
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: normalization expects AND")
	}
	makeMux := func(mux *Gia_Obj_t) {
		inputs := []int32{Gia_ObjFanin0Copy(mux), Gia_ObjFanin1Copy(mux), Gia_ObjFanin2Copy(p, mux)}
		if inputs[2]&1 != 0 {
			inputs[2] ^= 1
			inputs[0], inputs[1] = inputs[1], inputs[0]
		}
		mux.Value = uint32(Str_ObjCreate(ntk, STR_MUX, inputs))
	}
	if Gia_ObjIsMux(p, obj) != 0 {
		groupID := vec.Vec_IntEntry(roots, int(obj.id))
		if groupID < 0 {
			Str_ManNormalize_rec(ntk, p, Gia_ObjFanin0(obj), groups, roots)
			Str_ManNormalize_rec(ntk, p, Gia_ObjFanin1(obj), groups, roots)
			Str_ManNormalize_rec(ntk, p, Gia_ObjFanin2(p, obj), groups, roots)
			makeMux(obj)
			return
		}
		group := vec.Vec_WecEntry(groups, int(groupID))
		for _, id := range vec.Vec_IntArray(group) {
			Str_MuxInputsCollect(p, Gia_ManObj(p, id), p.vSuper)
			ids := append([]int32(nil), vec.Vec_IntArray(p.vSuper)...)
			for _, leaf := range ids {
				Str_ManNormalize_rec(ntk, p, Gia_ManObj(p, leaf), groups, roots)
			}
		}
		for _, id := range vec.Vec_IntArray(group) {
			root := Gia_ManObj(p, id)
			Str_MuxStructCollect(p, root, p.vSuper)
			for _, mux := range vec.Vec_IntArray(p.vSuper) {
				makeMux(Gia_ManObj(p, mux))
			}
			for _, mux := range vec.Vec_IntArray(p.vSuper) {
				Str_NtkObj(ntk, int32(Gia_ManObj(p, mux).Value>>1)).iTop = int32(root.Value >> 1)
			}
			ntk.nTrees++
		}
		ids := vec.Vec_IntArray(group)
		end := Gia_ManObj(p, ids[len(ids)-1])
		for _, id := range ids {
			Str_NtkObj(ntk, int32(Gia_ManObj(p, id).Value>>1)).iTop = int32(end.Value >> 1)
		}
		ntk.nGroups++
		return
	}
	strSuperCollect(p, obj)
	lits := append([]int32(nil), vec.Vec_IntArray(p.vSuper)...)
	for i, lit := range lits {
		fan := Gia_ManObj(p, lit>>1)
		Str_ManNormalize_rec(ntk, p, fan, groups, roots)
		lits[i] = int32(fan.Value) ^ (lit & 1)
	}
	kind := int32(STR_AND)
	if Gia_ObjIsXor(obj) != 0 {
		kind = STR_XOR
	}
	obj.Value = uint32(Str_ObjCreate(ntk, kind, lits))
}
func Str_ManNormalizeInt(p *Gia_Man_t, groups *Vec_Wec_t, roots *Vec_Int_t) *Str_Ntk_t {
	if p.PMuxes == nil {
		panic("gia: normalization needs real mux metadata")
	}
	if p.vSuper == nil {
		p.vSuper = vec.Vec_IntAlloc(STR_SUPER)
	}
	if p.vStore == nil {
		p.vStore = vec.Vec_IntAlloc(STR_SUPER)
	}
	Gia_ManFillValue(p)
	ntk := Str_NtkCreate(p.nObjs, 1+Gia_ManCoNum(p)+2*Gia_ManAndNum(p)+Gia_ManMuxNum(p))
	Gia_ManConst0(p).Value = 0
	for _, obj := range p.pObjs[1:] {
		if Gia_ObjIsCi(obj) != 0 {
			obj.Value = uint32(Str_ObjCreate(ntk, STR_PI, nil))
		} else if Gia_ObjIsCo(obj) != 0 {
			Str_ManNormalize_rec(ntk, p, Gia_ObjFanin0(obj), groups, roots)
			obj.Value = uint32(Str_ObjCreate(ntk, STR_PO, []int32{Gia_ObjFanin0Copy(obj)}))
		}
	}
	return ntk
}
func Str_ManNormalize(p *Gia_Man_t) *Str_Ntk_t {
	muxes := Gia_ManDupMuxes(p, 5)
	groups := Str_ManDeriveTrees(muxes)
	Str_ManCheckOverlap(muxes, groups)
	roots := Str_ManCreateRoots(groups, muxes.nObjs)
	ntk := Str_ManNormalizeInt(muxes, groups, roots)
	Gia_ManCleanMark0(muxes)
	Gia_ManStop(muxes)
	return ntk
}
