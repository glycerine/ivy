package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManDupConeSupp_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t, visited *Vec_Int_t) {
	if Gia_ObjCopyArray(src, obj.id) >= 0 {
		return
	}
	a, b, c, kind := src.Gate(obj.id)
	read := func(lit int32) int32 {
		node := src.pObjs[lit>>1]
		Gia_ManDupConeSupp_rec(dst, src, node, visited)
		return Gia_ObjCopyArray(src, node.id) ^ (lit & 1)
	}
	var result int32
	switch kind {
	case 2:
		result = Gia_ManAppendAnd(dst, read(a), read(b))
	case 3:
		result = Gia_ManAppendXorReal(dst, read(a), read(b))
	case 4:
		result = Gia_ManAppendMuxReal(dst, read(a), read(b), read(c))
	case 5:
		result = Gia_ManAppendBuf(dst, read(a))
	default:
		panic("gia: cone support does not cover a leaf")
	}
	Gia_ObjSetCopyArray(src, obj.id, result)
	vec.Vec_IntPush(visited, obj.id)
}
func Gia_ManDupConeSupp(src *Gia_Man_t, lit int32, ciIDs *Vec_Int_t) *Gia_Man_t {
	if lit < 0 || lit>>1 >= src.nObjs {
		panic("gia: invalid cone literal")
	}
	vec.Vec_IntFillExtra(&src.vCopies, int(src.nObjs), -1)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	visited := vec.Vec_IntAlloc(16)
	constOld := Gia_ObjCopyArray(src, 0)
	Gia_ObjSetCopyArray(src, 0, 0)
	mapped := make([]int32, 0, vec.Vec_IntSize(ciIDs))
	defer func() {
		for _, id := range mapped {
			Gia_ObjSetCopyArray(src, id, -1)
		}
		for _, id := range vec.Vec_IntArray(visited) {
			Gia_ObjSetCopyArray(src, id, -1)
		}
		Gia_ObjSetCopyArray(src, 0, constOld)
	}()
	for _, index := range vec.Vec_IntArray(ciIDs) {
		if index < 0 || index >= Gia_ManCiNum(src) {
			panic("gia: invalid cone CI")
		}
		obj := Gia_ManCi(src, index)
		if Gia_ObjCopyArray(src, obj.id) >= 0 {
			panic("gia: duplicate or occupied cone CI")
		}
		Gia_ObjSetCopyArray(src, obj.id, Gia_ManAppendCi(out))
		mapped = append(mapped, obj.id)
	}
	root := src.pObjs[lit>>1]
	Gia_ManDupConeSupp_rec(out, src, root, visited)
	Gia_ManAppendCo(out, Gia_ObjCopyArray(src, root.id)^lit&1)
	return out
}
func Gia_ManDupConeBack_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	Gia_ManDupCofactorVar_rec(dst, src, obj)
}
func Gia_ManDupConeBack(dst, src *Gia_Man_t, ciIDs *Vec_Int_t) int32 {
	if vec.Vec_IntSize(ciIDs) != int(Gia_ManCiNum(src)) {
		panic("gia: invalid cone mapping")
	}
	ids := vec.Vec_IntAlloc(vec.Vec_IntSize(ciIDs))
	for _, index := range vec.Vec_IntArray(ciIDs) {
		if index < 0 || index >= Gia_ManCiNum(dst) {
			panic("gia: invalid cone CI mapping")
		}
		vec.Vec_IntPush(ids, Gia_ManCi(dst, index).id)
	}
	return Gia_ManDupConeBackObjs(dst, src, ids)
}
func Gia_ManDupConeBackObjs(dst, src *Gia_Man_t, ids *Vec_Int_t) int32 {
	if vec.Vec_IntSize(ids) != int(Gia_ManCiNum(src)) || Gia_ManCoNum(src) < 1 {
		panic("gia: invalid cone mapping")
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for i, id := range vec.Vec_IntArray(ids) {
		if id < 0 || id >= dst.nObjs || Gia_ObjIsCo(dst.pObjs[id]) != 0 {
			panic("gia: invalid mapped cone object")
		}
		Gia_ManCi(src, int32(i)).Value = uint32(id << 1)
	}
	root := Gia_ManCo(src, 0)
	Gia_ManDupConeBack_rec(dst, src, Gia_ObjFanin0(root))
	return Gia_ObjFanin0Copy(root)
}
