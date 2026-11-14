package gia

// Choice links are traversed before their roots, so each resulting link points
// backward in the new graph. Values are literal IDs rather than C pointers.
func Gia_ManDupOrderDfsChoices_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected internal choice node")
	}
	var next *Gia_Obj_t
	if id := src.PNexts[obj.id]; id > 0 {
		next = Gia_ManObj(src, id)
		Gia_ManDupOrderDfsChoices_rec(dst, src, next)
	}
	for _, id := range giaObjectFanins(src, obj) {
		Gia_ManDupOrderDfsChoices_rec(dst, src, Gia_ManObj(src, id))
	}
	Gia_ManDupOrderDfs_rec(dst, src, obj)
	if next != nil {
		dst.PNexts[obj.Value>>1] = int32(next.Value >> 1)
	}
}
func Gia_ManDupOrderDfsChoices(src *Gia_Man_t) *Gia_Man_t {
	if src.PReprs == nil || len(src.PNexts) != int(src.nObjs) {
		panic("gia: missing choice classes")
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	dst.PNexts = make([]int32, src.nObjs)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		Gia_ManDupOrderDfsChoices_rec(dst, src, Gia_ObjFanin0(obj))
		giaDupMappedOutput(dst, src, obj)
	}
	dst.PNexts = dst.PNexts[:dst.nObjs]
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return dst
}
