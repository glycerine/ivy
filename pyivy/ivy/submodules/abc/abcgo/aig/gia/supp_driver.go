package gia

// Gia_Man2SupportAnd uses simulation and SAT to reduce provably false ANDs.
func Gia_Man2SupportAnd(p *Gia_Man2Min_t, a, b int32) int32 {
	p.nCalls++
	p.iLits = [2]int32{a, b}
	if a < 2 || b < 2 || a>>1 == b>>1 || Gia_ManHashLookupInt(p.pGia, a, b) != 0 ||
		Gia_Min2ManGatherSupp(p) == 0 || Gia_Min2ManSimulate(p) == 0 || Gia_Min2ManSolve(p) == 0 {
		return Gia_ManHashAnd(p.pGia, a, b)
	}
	p.nSatUnsat++
	return 0
}

// Gia_Man2SupportAndTest rebuilds a graph while applying support minimization.
func Gia_Man2SupportAndTest(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManRandomW(1)
	Gia_ManFillValue(src)
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManHashAlloc(dst)
	Gia_ManConst0(src).Value = 0
	manager := Gia_Man2SuppStart(dst)
	for _, obj := range src.pObjs[1:] {
		var lit int32
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			lit = Gia_Man2SupportAnd(manager, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		case Gia_ObjIsCi(obj) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj))
		default:
			panic("gia: unsupported support-minimization node")
		}
		obj.Value = uint32(lit)
	}
	Gia_Man2SuppStop(manager)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return Gia_ManCleanup(dst)
}
