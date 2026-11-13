package gia

// Gia_ManRecordSatVariable updates the optional source-style SAT-to-object
// mapping. If mapping is not enabled, it has no effect.
func Gia_ManRecordSatVariable(p *Gia_Man_t, variable int, object int32) {
	if p.vVar2Obj == nil {
		return
	}
	if Vec_IntSize(p.vVar2Obj) != variable {
		panic("gia: nonsequential SAT variable mapping")
	}
	Vec_IntPush(p.vVar2Obj, object)
}
