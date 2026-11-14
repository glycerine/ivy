package gia

func Gia_ManCreateRefs(p *Gia_Man_t) {
	if p.Refs != nil {
		panic("gia: reference counts already initialized")
	}
	p.Refs = make([]int32, p.nObjs)
	for _, obj := range p.pObjs {
		for _, fanin := range giaObjectFanins(p, obj) {
			p.Refs[fanin]++
		}
	}
}
func Gia_ObjRefNumId(p *Gia_Man_t, id int32) int32     { return p.Refs[id] }
func Gia_ObjRefIncId(p *Gia_Man_t, id int32) int32     { old := p.Refs[id]; p.Refs[id]++; return old }
func Gia_ObjRefDecId(p *Gia_Man_t, id int32) int32     { p.Refs[id]--; return p.Refs[id] }
func Gia_ObjRefNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjRefNumId(p, Gia_ObjId(p, obj)) }
func Gia_ObjRefInc(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjRefIncId(p, Gia_ObjId(p, obj)) }
func Gia_ObjRefDec(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjRefDecId(p, Gia_ObjId(p, obj)) }
func Gia_ObjRefFanin0Inc(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefInc(p, Gia_ObjFanin0(obj)) }
func Gia_ObjRefFanin1Inc(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefInc(p, Gia_ObjFanin1(obj)) }
func giaMuxFanin(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	if int(obj.id) >= len(p.PMuxes) || p.PMuxes[obj.id] == 0 {
		panic("gia: expected MUX object")
	}
	return Gia_ManObj(p, int32(p.PMuxes[obj.id]>>1))
}
func Gia_ObjRefFanin2Inc(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefInc(p, giaMuxFanin(p, obj)) }
func Gia_ObjRefFanin0Dec(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefDec(p, Gia_ObjFanin0(obj)) }
func Gia_ObjRefFanin1Dec(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefDec(p, Gia_ObjFanin1(obj)) }
func Gia_ObjRefFanin2Dec(p *Gia_Man_t, obj *Gia_Obj_t) { Gia_ObjRefDec(p, giaMuxFanin(p, obj)) }
func Gia_ObjLutRefNumId(p *Gia_Man_t, id int32) int32  { return p.pLutRefs[id] }
func Gia_ObjLutRefIncId(p *Gia_Man_t, id int32) int32 {
	old := p.pLutRefs[id]
	p.pLutRefs[id]++
	return old
}
func Gia_ObjLutRefDecId(p *Gia_Man_t, id int32) int32     { p.pLutRefs[id]--; return p.pLutRefs[id] }
func Gia_ObjLutRefNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjLutRefNumId(p, obj.id) }
func Gia_ObjLutRefInc(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjLutRefIncId(p, obj.id) }
func Gia_ObjLutRefDec(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjLutRefDecId(p, obj.id) }

func (p *Gia_Man_t) LUTReferenceCounts() []int32 { return append([]int32(nil), p.pLutRefs...) }
func (p *Gia_Man_t) SetLUTReferenceCounts(refs []int32) {
	if refs == nil {
		p.pLutRefs = nil
		return
	}
	if len(refs) != len(p.pObjs) {
		panic("gia: LUT reference count size mismatch")
	}
	for _, count := range refs {
		if count < 0 {
			panic("gia: negative LUT reference count")
		}
	}
	p.pLutRefs = append([]int32(nil), refs...)
}
