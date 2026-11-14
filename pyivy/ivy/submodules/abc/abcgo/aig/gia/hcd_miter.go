package gia

// Hcd_ManChoiceMiter_rec copies a cone and its CO into the shared manager.
func Hcd_ManChoiceMiter_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	// Both C recursive copiers are identical. Share the implementation and
	// its real-gate normalization rather than duplicating the AND-only bug.
	return Gia_ManChoiceMiter_rec(dst, src, obj)
}
func Hcd_ManChoiceMiter(networks *Vec_Ptr_t) *Gia_Man_t {
	if Vec_PtrSize(networks) == 0 {
		panic("gia: empty choice miter")
	}
	first := Vec_PtrEntry(networks, 0).(*Gia_Man_t)
	for i := 0; i < Vec_PtrSize(networks); i++ {
		src := Vec_PtrEntry(networks, i).(*Gia_Man_t)
		if Gia_ManCiNum(src) != Gia_ManCiNum(first) || Gia_ManCoNum(src) != Gia_ManCoNum(first) || src.nRegs != first.nRegs {
			panic("gia: incompatible miter interfaces")
		}
		Gia_ManFillValue(src)
		Gia_ManConst0(src).Value = 0
	}
	dst := Gia_ManStart(int32(Vec_PtrSize(networks)) * first.nObjs)
	dst.pName, dst.PSpec = first.pName, first.PSpec
	for k := int32(0); k < Gia_ManCiNum(first); k++ {
		lit := Gia_ManAppendCi(dst)
		for i := 0; i < Vec_PtrSize(networks); i++ {
			Gia_ManCi(Vec_PtrEntry(networks, i).(*Gia_Man_t), k).Value = uint32(lit)
		}
	}
	Gia_ManHashAlloc(dst)
	for k := int32(0); k < Gia_ManCoNum(first); k++ {
		for i := 0; i < Vec_PtrSize(networks); i++ {
			src := Vec_PtrEntry(networks, i).(*Gia_Man_t)
			Hcd_ManChoiceMiter_rec(dst, src, Gia_ManCo(src, k))
		}
	}
	Gia_ManHashStop(dst)
	if Gia_ManHasDangling(dst) != 0 {
		panic("gia: dangling choice miter nodes")
	}
	return dst
}
func Hcd_ObjCheckTfi_rec(p *Gia_Man_t, old, node *Gia_Obj_t, visited *Vec_Ptr_t) int32 {
	if node == nil || Gia_ObjIsCi(node) != 0 {
		return 0
	}
	if node == old {
		return 1
	}
	if node.id == 0 || node.fMark0 != 0 {
		return 0
	}
	node.fMark0 = 1
	Vec_PtrPush(visited, node)
	if Hcd_ObjCheckTfi_rec(p, old, Gia_ObjFanin0(node), visited) != 0 {
		return 1
	}
	if Gia_ObjIsAnd(node) != 0 && Hcd_ObjCheckTfi_rec(p, old, Gia_ObjFanin1(node), visited) != 0 {
		return 1
	}
	if p.PNexts != nil {
		return Hcd_ObjCheckTfi_rec(p, old, Gia_ObjNextObj(p, node.id), visited)
	}
	return 0
}
func Hcd_ObjCheckTfi(p *Gia_Man_t, old, node *Gia_Obj_t) int32 {
	if Gia_IsComplement(old) != 0 || Gia_IsComplement(node) != 0 {
		panic("gia: complemented TFI node")
	}
	visited := Vec_PtrAlloc(100)
	result := Hcd_ObjCheckTfi_rec(p, old, node, visited)
	for i := 0; i < Vec_PtrSize(visited); i++ {
		Vec_PtrEntry(visited, i).(*Gia_Obj_t).fMark0 = 0
	}
	return result
}
func Hcd_ManAddNextEntry_rec(p *Gia_Man_t, old, node *Gia_Obj_t) {
	for Gia_ObjNext(p, old.id) != 0 {
		old = Gia_ObjNextObj(p, old.id)
	}
	Gia_ObjSetNext(p, old.id, node.id)
}
