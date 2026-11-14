package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManStart(nObjsMax int32) *Gia_Man_t {
	if nObjsMax <= 0 {
		panic("gia: object capacity must be positive")
	}
	p := &Gia_Man_t{nObjsAlloc: nObjsMax, pObjs: make([]*Gia_Obj_t, 0, int(nObjsMax)), vCis: vec.Vec_IntAlloc(int(nObjsMax / 20)), vCos: vec.Vec_IntAlloc(int(nObjsMax / 20))}
	p.pObjs = append(p.pObjs, &Gia_Obj_t{iDiff0: objectNoFanin, iDiff1: objectNoFanin})
	p.store = &objectStore{objects: &p.pObjs}
	p.pObjs[0].store = p.store
	p.nObjs = 1
	return p
}

// All manager-owned arrays are managed by Go. Dropping references releases the
// data and leaves a stopped manager empty, including its auxiliary graphs.
func Gia_ManStop(p *Gia_Man_t) {
	if p != nil {
		*p = Gia_Man_t{}
	}
}
func Gia_ManStopP(p **Gia_Man_t) {
	if p != nil {
		Gia_ManStop(*p)
		*p = nil
	}
}
func Gia_ManConst0(p *Gia_Man_t) *Gia_Obj_t { return Gia_ManObj(p, 0) }
func Gia_ManObj(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if id < 0 || id >= p.nObjs {
		panic("gia: invalid object index")
	}
	return p.pObjs[id]
}
func Gia_ObjId(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj == nil || obj.id < 0 || obj.id >= p.nObjs || p.pObjs[obj.id] != obj {
		panic("gia: object does not belong to manager")
	}
	return obj.id
}
func Gia_ManObjIsConst0(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return giaBool(obj == Gia_ManConst0(p)) }
func Gia_ManAppendObj(p *Gia_Man_t) *Gia_Obj_t {
	if p.nObjs >= int32(objectNoFanin) {
		panic("gia: object index exceeds backward-edge representation")
	}
	obj := &Gia_Obj_t{id: p.nObjs, store: p.store}
	p.pObjs = append(p.pObjs, obj)
	p.nObjs++
	p.nObjsAlloc = int32(cap(p.pObjs))
	if p.PMuxes != nil && len(p.PMuxes) < int(p.nObjs) {
		p.PMuxes = append(p.PMuxes, make([]uint32, int(p.nObjs)-len(p.PMuxes))...)
	}
	if vec.Vec_IntSize(&p.vHTable) != 0 {
		vec.Vec_IntPush(&p.vHash, 0)
	}
	return obj
}

func Gia_ManAppendCi(p *Gia_Man_t) int32 {
	obj := Gia_ManAppendObj(p)
	obj.fTerm = 1
	obj.iDiff0 = objectNoFanin
	obj.iDiff1 = uint32(vec.Vec_IntSize(p.vCis))
	vec.Vec_IntPush(p.vCis, obj.id)
	return obj.id << 1
}
func giaCheckFanin(p *Gia_Man_t, lit int32) {
	if lit < 0 || lit>>1 >= p.nObjs {
		panic("gia: invalid fanin literal")
	}
	if Gia_ObjIsCo(Gia_ManObj(p, lit>>1)) != 0 {
		panic("gia: CO cannot drive a node")
	}
}
func Gia_ManAppendAnd(p *Gia_Man_t, a, b int32) int32 {
	giaCheckFanin(p, a)
	giaCheckFanin(p, b)
	if p.fGiaSimple == 0 && a>>1 == b>>1 {
		panic("gia: equal fanin objects require simplified append")
	}
	if a > b {
		a, b = b, a
	}
	obj := Gia_ManAppendObj(p)
	obj.iDiff0 = uint32(obj.id - (a >> 1))
	obj.fCompl0 = uint32(a & 1)
	obj.iDiff1 = uint32(obj.id - (b >> 1))
	obj.fCompl1 = uint32(b & 1)
	if p.pFanData != nil {
		Gia_ObjAddFanout(p, Gia_ObjFanin0(obj), obj)
		Gia_ObjAddFanout(p, Gia_ObjFanin1(obj), obj)
	}
	if p.fSweeper != 0 || p.fBuiltInSim != 0 {
		f0, f1 := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		if p.fSweeper != 0 {
			if f0.fMark0 != 0 {
				f0.fMark1 = 1
			} else {
				f0.fMark0 = 1
			}
			if f1.fMark0 != 0 {
				f1.fMark1 = 1
			} else {
				f1.fMark0 = 1
			}
		}
		obj.fPhase = (f0.fPhase ^ obj.fCompl0) & (f1.fPhase ^ obj.fCompl1)
		if p.fBuiltInSim != 0 {
			Gia_ManBuiltInSimPerform(p, obj.id)
		}
	}
	if p.vSuppWords != nil {
		Gia_ManQuantSetSuppAnd(p, obj)
	}
	return obj.id << 1
}
func Gia_ManAppendCo(p *Gia_Man_t, lit int32) int32 {
	giaCheckFanin(p, lit)
	obj := Gia_ManAppendObj(p)
	obj.fTerm = 1
	obj.iDiff0 = uint32(obj.id - (lit >> 1))
	obj.fCompl0 = uint32(lit & 1)
	obj.iDiff1 = uint32(vec.Vec_IntSize(p.vCos))
	vec.Vec_IntPush(p.vCos, obj.id)
	if p.pFanData != nil {
		Gia_ObjAddFanout(p, Gia_ObjFanin0(obj), obj)
	}
	return obj.id << 1
}
func Gia_ManAppendOr(p *Gia_Man_t, a, b int32) int32 { return Gia_ManAppendAnd(p, a^1, b^1) ^ 1 }
func Gia_ManAppendMux(p *Gia_Man_t, control, yes, no int32) int32 {
	a, b := Gia_ManAppendAnd(p, control^1, no), Gia_ManAppendAnd(p, control, yes)
	return Gia_ManAppendAnd(p, a^1, b^1) ^ 1
}
func Gia_ManAppendMaj(p *Gia_Man_t, a, b, c int32) int32 {
	return Gia_ManAppendOr(p, Gia_ManAppendAnd(p, a, Gia_ManAppendOr(p, b, c)), Gia_ManAppendAnd(p, b, c))
}
func Gia_ManAppendXor(p *Gia_Man_t, a, b int32) int32 { return Gia_ManAppendMux(p, a, b^1, b) }
