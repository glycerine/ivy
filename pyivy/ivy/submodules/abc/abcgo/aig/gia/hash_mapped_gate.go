package gia

// The source transforms operate on ordinary AIGs. Lower real gates through the
// existing Boolean hash operations when copying into an ordinary hashed AIG.
func giaHashMappedGate(dst, src *Gia_Man_t, obj *Gia_Obj_t, read func(int32) int32) int32 {
	a, b, c, kind := src.Gate(obj.id)
	switch kind {
	case 2:
		return Gia_ManHashAnd(dst, read(a), read(b))
	case 3:
		return Gia_ManHashXor(dst, read(a), read(b))
	case 4:
		return Gia_ManHashMux(dst, read(a), read(b), read(c))
	case 5:
		return read(a)
	default:
		panic("gia: expected an internal graph gate")
	}
}
