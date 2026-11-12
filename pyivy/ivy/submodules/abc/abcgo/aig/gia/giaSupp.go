package gia

// Gia_Min2AddClausesMux encodes f = if i then t else e, including the
// redundant consensus clauses used by the original implementation.
func Gia_Min2AddClausesMux(p *Gia_Man_t, node *Gia_Obj_t, solver *satoko_t) {
	if Gia_IsComplement(node) != 0 || node.fMark0 == 0 {
		panic("gia: invalid marked MUX")
	}
	var then, otherwise *Gia_Obj_t
	control := Gia_ObjRecognizeMux(node, &then, &otherwise)
	f := int(Gia_Min2ObjSatId(p, node)) * 2
	i := int(Gia_Min2ObjSatId(p, control)) * 2
	t := int(Gia_Min2ObjSatId(p, Gia_Regular(then)))*2 + int(Gia_IsComplement(then))
	e := int(Gia_Min2ObjSatId(p, Gia_Regular(otherwise)))*2 + int(Gia_IsComplement(otherwise))
	clauses := [][]int{{i ^ 1, t ^ 1, f}, {i ^ 1, t, f ^ 1}, {i, e ^ 1, f}, {i, e, f ^ 1}}
	if t>>1 != e>>1 {
		clauses = append(clauses, []int{t, e, f ^ 1}, []int{t ^ 1, e ^ 1, f})
	}
	for _, clause := range clauses {
		if !solver.AddClause(clause) {
			panic("gia: inconsistent MUX clause")
		}
	}
}

func Gia_Min2AddClausesSuper(p *Gia_Man_t, node *Gia_Obj_t, super *Vec_Ptr_t, solver *satoko_t) {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsAnd(node) == 0 {
		panic("gia: invalid supergate")
	}
	f := int(Gia_Min2ObjSatId(p, node)) * 2
	clause := make([]int, 0, Vec_PtrSize(super)+1)
	for _, entry := range Vec_PtrArray(super) {
		fanin := entry.(*Gia_Obj_t)
		lit := int(Gia_Min2ObjSatId(p, Gia_Regular(fanin)))*2 + int(Gia_IsComplement(fanin))
		if !solver.AddClause([]int{lit, f ^ 1}) {
			panic("gia: inconsistent supergate clause")
		}
		clause = append(clause, lit^1)
	}
	clause = append(clause, f)
	if !solver.AddClause(clause) {
		panic("gia: inconsistent supergate clause")
	}
}

func Gia_Min2CollectSuper_rec(obj *Gia_Obj_t, super *Vec_Ptr_t, first, useMuxes int32) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsCi(obj) != 0 || first == 0 && Gia_ObjValue(obj) > 1 || useMuxes != 0 && obj.fMark0 != 0 {
		Vec_PtrPushUnique(super, obj)
		return
	}
	Gia_Min2CollectSuper_rec(Gia_ObjChild0(obj), super, 0, useMuxes)
	Gia_Min2CollectSuper_rec(Gia_ObjChild1(obj), super, 0, useMuxes)
}
func Gia_Min2CollectSuper(obj *Gia_Obj_t, useMuxes int32, super *Vec_Ptr_t) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsCi(obj) != 0 {
		panic("gia: invalid supergate root")
	}
	Vec_PtrClear(super)
	Gia_Min2CollectSuper_rec(obj, super, 1, useMuxes)
}
func Gia_Min2ObjAddToFrontier(p *Gia_Man_t, obj *Gia_Obj_t, frontier *Vec_Ptr_t, solver *satoko_t, vars *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsConst0(obj) != 0 {
		panic("gia: invalid SAT frontier node")
	}
	if Gia_Min2ObjSatId(p, obj) >= 0 {
		return
	}
	Vec_IntPush(vars, Gia_ObjId(p, obj))
	Gia_Min2ObjSetSatId(p, obj, int32(solver.AddVariable(false)))
	if Gia_ObjIsAnd(obj) != 0 {
		Vec_PtrPush(frontier, obj)
	}
}
func Gia_Min2ObjGetCnfVar(p *Gia_Man2Min_t, id int32) int32 {
	obj := Gia_ManObj(p.pGia, id)
	if v := Gia_Min2ObjSatId(p.pGia, obj); v >= 0 {
		return v
	}
	if Gia_ObjIsCi(obj) != 0 {
		Vec_IntPush(p.vSatVars, id)
		return Gia_Min2ObjSetSatId(p.pGia, obj, int32(p.pSat.AddVariable(false)))
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid SAT root")
	}
	Vec_PtrClear(p.vFrontier)
	Gia_Min2ObjAddToFrontier(p.pGia, obj, p.vFrontier, p.pSat, p.vSatVars)
	for i := 0; i < Vec_PtrSize(p.vFrontier); i++ {
		node := Vec_PtrEntry(p.vFrontier, i).(*Gia_Obj_t)
		if node.fMark0 != 0 {
			Vec_PtrClear(p.vFanins)
			for _, child := range []*Gia_Obj_t{Gia_ObjFanin0(node), Gia_ObjFanin1(node)} {
				Vec_PtrPushUnique(p.vFanins, Gia_ObjFanin0(child))
				Vec_PtrPushUnique(p.vFanins, Gia_ObjFanin1(child))
			}
		} else {
			Gia_Min2CollectSuper(node, 1, p.vFanins)
		}
		for _, entry := range Vec_PtrArray(p.vFanins) {
			Gia_Min2ObjAddToFrontier(p.pGia, Gia_Regular(entry.(*Gia_Obj_t)), p.vFrontier, p.pSat, p.vSatVars)
		}
		if node.fMark0 != 0 {
			Gia_Min2AddClausesMux(p.pGia, node, p.pSat)
		} else {
			Gia_Min2AddClausesSuper(p.pGia, node, p.vFanins, p.pSat)
		}
		if Vec_PtrSize(p.vFanins) <= 1 {
			panic("gia: invalid SAT fanins")
		}
	}
	return Gia_Min2ObjSatId(p.pGia, obj)
}
