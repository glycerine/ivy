package gia

func Gia_ManDupDfs_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	_, _, _, kind := src.Gate(Gia_ObjId(src, obj))
	if kind != 2 && kind != 5 {
		panic("gia: DFS AIG copier requires ordinary AND/buffer nodes")
	}
	Gia_ManDupDfs_rec(dst, src, Gia_ObjFanin0(obj))
	Gia_ManDupDfs_rec(dst, src, Gia_ObjFanin1(obj))
	obj.Value = uint32(Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
}
func Gia_ManDupDfsCone(src *Gia_Man_t, root *Gia_Obj_t) *Gia_Man_t {
	if Gia_ObjIsCo(root) == 0 {
		panic("gia: DFS cone root must be an output")
	}
	Gia_ObjId(src, root)
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	Gia_ManDupDfs_rec(dst, src, Gia_ObjFanin0(root))
	Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(root))
	return dst
}
