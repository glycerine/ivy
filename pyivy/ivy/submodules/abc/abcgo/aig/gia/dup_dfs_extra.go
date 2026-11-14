package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupDfs2_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	if src.pReprsOld != nil {
		if len(src.pReprsOld) != int(src.nObjs) {
			panic("gia: invalid old representatives")
		}
		repr := src.pReprsOld[obj.id]
		if repr >= 0 {
			if repr >= obj.id {
				panic("gia: invalid old representative")
			}
			other := src.pObjs[repr]
			lit := Gia_ManDupDfs2_rec(dst, src, other) ^ (Gia_ObjPhaseReal(other) ^ Gia_ObjPhaseReal(obj))
			obj.Value = uint32(lit)
			return lit
		}
	}
	a, b, c, kind := src.Gate(obj.id)
	if kind == 1 {
		obj.Value = uint32(Gia_ManAppendCi(dst))
		return int32(obj.Value)
	}
	read := func(lit int32) int32 { return Gia_ManDupDfs2_rec(dst, src, src.pObjs[lit>>1]) ^ (lit & 1) }
	var lit int32
	switch kind {
	case 2:
		if Vec_IntSize(&dst.vHTable) > 0 {
			lit = Gia_ManHashAnd(dst, read(a), read(b))
		} else {
			lit = Gia_ManAppendAnd(dst, read(a), read(b))
		}
	case 3:
		if Vec_IntSize(&dst.vHTable) > 0 {
			lit = Gia_ManHashXor(dst, read(a), read(b))
		} else {
			lit = Gia_ManAppendXorReal(dst, read(a), read(b))
		}
	case 4:
		if Vec_IntSize(&dst.vHTable) > 0 {
			lit = Gia_ManHashMux(dst, read(a), read(b), read(c))
		} else {
			lit = Gia_ManAppendMuxReal(dst, read(a), read(b), read(c))
		}
	case 5:
		lit = Gia_ManAppendBuf(dst, read(a))
	case 6:
		lit = Gia_ManAppendCo(dst, read(a))
	default:
		panic("gia: invalid DFS gate")
	}
	obj.Value = uint32(lit)
	return lit
}
func Gia_ManDupDfs2(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupDfs2_rec(out, src, Gia_ManCo(src, i))
	}
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		obj := Gia_ManCi(src, i)
		if obj.Value == ^uint32(0) {
			obj.Value = uint32(Gia_ManAppendCi(out))
		}
	}
	Gia_ManDupRemapCis(out, src)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupDfs(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(out, src, Gia_ObjFanin0(Gia_ManCo(src, i)))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	out.nConstrs = src.nConstrs
	if src.pCexSeq != nil {
		out.pCexSeq = util.Abc_CexDup(src.pCexSeq, Gia_ManRegNum(src))
	}
	return out
}
func Gia_ManDupDfsOnePo(src *Gia_Man_t, po int32) *Gia_Man_t {
	if po < 0 || po >= Gia_ManPoNum(src) {
		panic("gia: output index out of range")
	}
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	roots := []*Gia_Obj_t{Gia_ManPo(src, po)}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		roots = append(roots, Gia_ManRi(src, i))
	}
	for _, obj := range roots {
		Gia_ManDupOrderDfs_rec(out, src, Gia_ObjFanin0(obj))
	}
	for _, obj := range roots {
		giaDupMappedOutput(out, src, obj)
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}
func Gia_ManDupDfsSkip(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		if obj.fMark1 == 0 {
			Gia_ManDupOrderDfs_rec(out, src, obj)
		}
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
func Gia_ManDupDfs3_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) { Gia_ManDupOrderDfs_rec(dst, src, obj) }
func Gia_ManDupDfsNode(src *Gia_Man_t, root *Gia_Obj_t) *Gia_Man_t {
	if Gia_ObjIsAnd(root) == 0 {
		panic("gia: expected internal cone root")
	}
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	Gia_ManDupDfs3_rec(out, src, root)
	Gia_ManAppendCo(out, int32(root.Value))
	return out
}
func Gia_ManDupDfsLitArray(src *Gia_Man_t, lits *Vec_Int_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for _, lit := range vec.Vec_IntArray(lits) {
		if lit < 0 || lit>>1 >= src.nObjs {
			panic("gia: invalid cone literal")
		}
		obj := src.pObjs[lit>>1]
		Gia_ManDupDfs2_rec(out, src, obj)
		Gia_ManAppendCo(out, int32(obj.Value)^lit&1)
	}
	return out
}

func Gia_ManDupDfsCiMap(src *Gia_Man_t, ciMap []int32, lits *Vec_Int_t) *Gia_Man_t {
	if len(ciMap) != int(Gia_ManCiNum(src)) {
		panic("gia: invalid CI mapping")
	}
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	for i, mapping := range ciMap {
		obj := Gia_ManCi(src, int32(i))
		obj.Value = uint32(Gia_ManAppendCi(out))
		if mapping >= 0 {
			id := mapping >> 1
			if id >= src.nObjs || src.pObjs[id].Value == ^uint32(0) {
				panic("gia: CI mapping references an unmapped object")
			}
			obj.Value = uint32(int32(src.pObjs[id].Value) ^ mapping&1)
		}
	}
	Gia_ManHashAlloc(out)
	if lits != nil {
		for _, lit := range vec.Vec_IntArray(lits) {
			if lit < 0 || lit>>1 >= src.nObjs {
				panic("gia: invalid selected literal")
			}
			mapped := Gia_ManDupDfs2_rec(out, src, src.pObjs[lit>>1])
			Gia_ManAppendCo(out, mapped^lit&1)
		}
	} else {
		for i := int32(0); i < Gia_ManCoNum(src); i++ {
			obj := Gia_ManCo(src, i)
			Gia_ManDupDfs2_rec(out, src, Gia_ObjFanin0(obj))
			Gia_ManAppendCo(out, Gia_ObjFanin0Copy(obj))
		}
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
