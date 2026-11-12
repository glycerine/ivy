package gia

func Gia_ManConst1(p *Gia_Man_t) *Gia_Obj_t { return Gia_Not(Gia_ManConst0(p)) }
func Gia_ObjCopyF(p *Gia_Man_t, frame int32, obj *Gia_Obj_t) int32 {
	return Vec_IntEntry(&p.vCopies, int(p.nObjs*frame+Gia_ObjId(p, obj)))
}
func Gia_ObjSetCopyF(p *Gia_Man_t, frame int32, obj *Gia_Obj_t, lit int32) {
	Vec_IntWriteEntry(&p.vCopies, int(p.nObjs*frame+Gia_ObjId(p, obj)), lit)
}
func Gia_ObjCopyArray(p *Gia_Man_t, id int32) int32    { return Vec_IntEntry(&p.vCopies, int(id)) }
func Gia_ObjSetCopyArray(p *Gia_Man_t, id, lit int32)  { Vec_IntWriteEntry(&p.vCopies, int(id), lit) }
func Gia_ManCleanCopyArray(p *Gia_Man_t)               { Vec_IntFill(&p.vCopies, int(p.nObjs), -1) }
func Gia_ObjCopy2Array(p *Gia_Man_t, id int32) int32   { return Vec_IntEntry(&p.vCopies2, int(id)) }
func Gia_ObjSetCopy2Array(p *Gia_Man_t, id, lit int32) { Vec_IntWriteEntry(&p.vCopies2, int(id), lit) }
func Gia_ManCleanCopy2Array(p *Gia_Man_t)              { Vec_IntFill(&p.vCopies2, int(p.nObjs), -1) }
func Gia_ObjFanin0CopyF(p *Gia_Man_t, frame int32, obj *Gia_Obj_t) int32 {
	return Gia_ObjCopyF(p, frame, Gia_ObjFanin0(obj)) ^ Gia_ObjFaninC0(obj)
}
func Gia_ObjFanin1CopyF(p *Gia_Man_t, frame int32, obj *Gia_Obj_t) int32 {
	return Gia_ObjCopyF(p, frame, Gia_ObjFanin1(obj)) ^ Gia_ObjFaninC1(obj)
}
func Gia_ObjFanin0CopyArray(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjCopyArray(p, Gia_ObjFaninId0p(p, obj)) ^ Gia_ObjFaninC0(obj)
}
func Gia_ObjFanin1CopyArray(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjCopyArray(p, Gia_ObjFaninId1p(p, obj)) ^ Gia_ObjFaninC1(obj)
}
func Gia_ObjFromLit(p *Gia_Man_t, lit int32) *Gia_Obj_t { return Gia_Lit2Obj(p, lit) }
func Gia_ObjToLit(p *Gia_Man_t, obj *Gia_Obj_t) int32   { return Gia_Obj2Lit(p, obj) }
func Gia_ObjPhaseRealLit(p *Gia_Man_t, lit int32) int32 {
	return Gia_ObjPhaseReal(Gia_ObjFromLit(p, lit))
}

const giaInfinity int32 = 1_000_000_000

func Gia_ObjHasNumId(p *Gia_Man_t, id int32) int32 {
	return giaBool(Vec_IntEntry(p.vTtNums, int(id)) > -giaInfinity)
}
func Gia_ObjNumId(p *Gia_Man_t, id int32) int32     { return Vec_IntEntry(p.vTtNums, int(id)) }
func Gia_ObjNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjNumId(p, Gia_ObjId(p, obj)) }
func Gia_ObjSetNumId(p *Gia_Man_t, id, value int32) { Vec_IntWriteEntry(p.vTtNums, int(id), value) }
func Gia_ObjSetNum(p *Gia_Man_t, obj *Gia_Obj_t, value int32) {
	Gia_ObjSetNumId(p, Gia_ObjId(p, obj), value)
}
func Gia_ObjResetNumId(p *Gia_Man_t, id int32) { Gia_ObjSetNumId(p, id, -giaInfinity) }

func Gia_ManTimeClean(p *Gia_Man_t) {
	if p.vTiming == nil {
		panic("gia: timing storage is not started")
	}
	Vec_FltFill(p.vTiming, 3*int(p.nObjs), 0)
	for i := 0; i < int(p.nObjs); i++ {
		Vec_FltWriteEntry(p.vTiming, 3*i+1, float32(giaInfinity))
	}
}
func Gia_ManTimeStart(p *Gia_Man_t) {
	if p.vTiming != nil {
		panic("gia: timing storage already started")
	}
	p.vTiming = Vec_FltAlloc(0)
	Gia_ManTimeClean(p)
}
func Gia_ManTimeStop(p *Gia_Man_t) {
	if p.vTiming == nil {
		panic("gia: timing storage is not started")
	}
	p.vTiming = nil
}
func Gia_ObjTimeArrival(p *Gia_Man_t, id int32) float32  { return Vec_FltEntry(p.vTiming, 3*int(id)) }
func Gia_ObjTimeRequired(p *Gia_Man_t, id int32) float32 { return Vec_FltEntry(p.vTiming, 3*int(id)+1) }
func Gia_ObjTimeSlack(p *Gia_Man_t, id int32) float32    { return Vec_FltEntry(p.vTiming, 3*int(id)+2) }
func Gia_ObjTimeArrivalObj(p *Gia_Man_t, obj *Gia_Obj_t) float32 {
	return Gia_ObjTimeArrival(p, Gia_ObjId(p, obj))
}
func Gia_ObjTimeRequiredObj(p *Gia_Man_t, obj *Gia_Obj_t) float32 {
	return Gia_ObjTimeRequired(p, Gia_ObjId(p, obj))
}
func Gia_ObjTimeSlackObj(p *Gia_Man_t, obj *Gia_Obj_t) float32 {
	return Gia_ObjTimeSlack(p, Gia_ObjId(p, obj))
}
func Gia_ObjSetTimeArrival(p *Gia_Man_t, id int32, value float32) {
	Vec_FltWriteEntry(p.vTiming, 3*int(id), value)
}
func Gia_ObjSetTimeRequired(p *Gia_Man_t, id int32, value float32) {
	Vec_FltWriteEntry(p.vTiming, 3*int(id)+1, value)
}
func Gia_ObjSetTimeSlack(p *Gia_Man_t, id int32, value float32) {
	Vec_FltWriteEntry(p.vTiming, 3*int(id)+2, value)
}
func Gia_ObjSetTimeArrivalObj(p *Gia_Man_t, obj *Gia_Obj_t, value float32) {
	Gia_ObjSetTimeArrival(p, Gia_ObjId(p, obj), value)
}
func Gia_ObjSetTimeRequiredObj(p *Gia_Man_t, obj *Gia_Obj_t, value float32) {
	Gia_ObjSetTimeRequired(p, Gia_ObjId(p, obj), value)
}
func Gia_ObjSetTimeSlackObj(p *Gia_Man_t, obj *Gia_Obj_t, value float32) {
	Gia_ObjSetTimeSlack(p, Gia_ObjId(p, obj), value)
}
