package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManDupHashDfs_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	Gia_ManDupCofactorVar_rec(dst, src, obj)
	return int32(obj.Value)
}
func Gia_ManDupCleanDfs_rec(obj *Gia_Obj_t) {
	if obj.Value == ^uint32(0) || Gia_ObjIsConst0(obj) != 0 {
		return
	}
	obj.Value = ^uint32(0)
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid hashed cone node")
	}
	Gia_ManDupCleanDfs_rec(Gia_ObjFanin0(obj))
	if Gia_ObjIsBuf(obj) == 0 {
		Gia_ManDupCleanDfs_rec(Gia_ObjFanin1(obj))
	}
}
func Gia_ManIsoStrashReduceOne(dst, src *Gia_Man_t, obj *Gia_Obj_t, support *Vec_Int_t) int32 {
	if Gia_ObjIsCo(obj) == 0 || vec.Vec_IntSize(support) > int(Gia_ManCiNum(dst)) {
		panic("gia: invalid support mapping")
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for i, ci := range vec.Vec_IntArray(support) {
		if ci < 0 || ci >= Gia_ManCiNum(src) {
			panic("gia: invalid support input")
		}
		Gia_ManCi(src, ci).Value = uint32(Gia_ObjToLit(dst, Gia_ManCi(dst, int32(i))))
	}
	Gia_ManDupHashDfs_rec(dst, src, Gia_ObjFanin0(obj))
	return Gia_ObjFanin0Copy(obj)
}
func Gia_ManDupStrashReduce(src *Gia_Man_t, supports *Vec_Wec_t, coMap **Vec_Int_t) *Gia_Man_t {
	if Gia_ManRegNum(src) != 0 {
		panic("gia: strash reduction requires a combinational graph")
	}
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	for i := int32(0); i < Gia_ManCoSuppSizeMax(src, supports); i++ {
		Gia_ManAppendCi(out)
	}
	Gia_ManHashAlloc(out)
	results := make([]int32, Gia_ManCoNum(src))
	for i := range results {
		results[i] = Gia_ManIsoStrashReduceOne(out, src, Gia_ManCo(src, int32(i)), vec.Vec_WecEntry(supports, i))
	}
	Gia_ManHashStop(out)
	mapping := vec.Vec_IntAlloc(len(results))
	indices := make(map[int32]int32)
	for _, lit := range results {
		if coMap == nil {
			Gia_ManAppendCo(out, lit)
			continue
		}
		index, ok := indices[lit]
		if !ok {
			index = Gia_ManCoNum(out)
			indices[lit] = index
			Gia_ManAppendCo(out, lit)
		}
		vec.Vec_IntPush(mapping, index)
	}
	if coMap != nil {
		*coMap = mapping
	}
	return giaCleanComposed(out)
}
func Gia_ManIsoStrashReduce2(src *Gia_Man_t, groups **Vec_Ptr_t, verbose int32) *Gia_Man_t {
	supports := Gia_ManCreateCoSupps(src, verbose)
	var mapping *Vec_Int_t
	out := Gia_ManDupStrashReduce(src, supports, &mapping)
	if groups != nil {
		*groups = nil
	}
	return out
}
func Gia_ManIsoStrashReduceInt(src *Gia_Man_t, supports *Vec_Wec_t, verbose int32) *Vec_Wec_t {
	out := Gia_ManStart(src.nObjs)
	for i := int32(0); i < Gia_ManCoSuppSizeMax(src, supports); i++ {
		Gia_ManAppendCi(out)
	}
	Gia_ManHashAlloc(out)
	classes := vec.Vec_WecAlloc(0)
	indices := make(map[int32]int)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		lit := Gia_ManIsoStrashReduceOne(out, src, Gia_ManCo(src, i), vec.Vec_WecEntry(supports, int(i)))
		index, ok := indices[lit]
		if !ok {
			index = vec.Vec_WecSize(classes)
			indices[lit] = index
			vec.Vec_WecPushLevel(classes)
		}
		vec.Vec_IntPush(vec.Vec_WecEntry(classes, index), i)
	}
	Gia_ManHashStop(out)
	Gia_ManStop(out)
	return classes
}
func Gia_ManIsoStrashReduce(src *Gia_Man_t, groups **Vec_Ptr_t, verbose int32) *Gia_Man_t {
	supports := Gia_ManCreateCoSupps(src, verbose)
	classes := Gia_ManIsoStrashReduceInt(src, supports, verbose)
	firsts := vec.Vec_WecCollectFirsts(classes)
	out := Gia_ManDupCones(src, vec.Vec_IntArray(firsts), int32(vec.Vec_IntSize(firsts)), 0)
	if groups != nil {
		*groups = vec.Vec_WecConvertToVecPtr(classes)
	}
	return out
}
