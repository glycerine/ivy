package gia

import "github.com/glycerine/abcgo/misc/vec"

// Cof_Obj_t stores explicit links instead of offsets into C's packed arena.
type Cof_Obj_t struct {
	Id                            int32
	fTerm, fPhase, fMark0, fMark1 uint32
	nFanoutsM                     uint32
	Value                         uint32
	refs                          int32
	iNext, iLit                   int32
	fanins, fanouts               []*Cof_Obj_t
	complements                   []bool
}
type Cof_Fan_t struct{ iFan, fCompl uint32 }
type Cof_Man_t struct {
	pGia                    *Gia_Man_t
	vCis, vCos              *Vec_Int_t
	objects                 []*Cof_Obj_t
	nObjs, nNodes, nTravIds int32
	pLevels                 []int32
	nLevels                 int32
}

func Gia_ObjHandle(obj *Gia_Obj_t) uint32 { return obj.Value }
func Cof_ObjLevel(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	return Gia_ObjLevel(p.pGia, Gia_ManObj(p.pGia, obj.Id))
}
func Cof_ObjHandle(p *Cof_Man_t, obj *Cof_Obj_t) uint32 { return uint32(obj.Id) }
func Cof_ObjHandleDiff(obj, fanin *Cof_Obj_t) uint32    { return uint32(obj.Id - fanin.Id) }
func Cof_ObjIsTerm(obj *Cof_Obj_t) int32                { return giaBool(obj.fTerm != 0) }
func Cof_ObjIsCi(obj *Cof_Obj_t) int32                  { return giaBool(obj.fTerm != 0 && len(obj.fanins) == 0) }
func Cof_ObjIsCo(obj *Cof_Obj_t) int32                  { return giaBool(obj.fTerm != 0 && len(obj.fanins) == 1) }
func Cof_ObjIsNode(obj *Cof_Obj_t) int32                { return giaBool(obj.fTerm == 0 && len(obj.fanins) > 0) }
func Cof_ObjIsConst0(obj *Cof_Obj_t) int32              { return giaBool(obj.fTerm == 0 && len(obj.fanins) == 0) }
func Cof_ObjFaninNum(obj *Cof_Obj_t) int32              { return int32(len(obj.fanins)) }
func Cof_ObjFanoutNum(obj *Cof_Obj_t) int32             { return int32(len(obj.fanouts)) }
func Cof_ObjSize(obj *Cof_Obj_t) int32                  { return 1 }
func Cof_ManObj(p *Cof_Man_t, id uint32) *Cof_Obj_t     { return p.objects[id] }
func Cof_ObjFanin(obj *Cof_Obj_t, i int32) *Cof_Obj_t   { return obj.fanins[i] }
func Cof_ObjFanout(obj *Cof_Obj_t, i int32) *Cof_Obj_t  { return obj.fanouts[i] }
func Cof_ManObjNum(p *Cof_Man_t) int32                  { return p.nObjs }
func Cof_ManNodeNum(p *Cof_Man_t) int32                 { return p.nNodes }
func Cof_ManResetTravId(p *Cof_Man_t)                   { Cof_ManCleanValue(p); p.nTravIds = 1 }
func Cof_ManIncrementTravId(p *Cof_Man_t) {
	p.nTravIds++
	if p.nTravIds <= 0 {
		Cof_ManResetTravId(p)
	}
}
func Cof_ObjSetTravId(obj *Cof_Obj_t, id int32)             { obj.Value = uint32(id) }
func Cof_ObjSetTravIdCurrent(p *Cof_Man_t, obj *Cof_Obj_t)  { Cof_ObjSetTravId(obj, p.nTravIds) }
func Cof_ObjSetTravIdPrevious(p *Cof_Man_t, obj *Cof_Obj_t) { Cof_ObjSetTravId(obj, p.nTravIds-1) }
func Cof_ObjIsTravIdCurrent(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	return giaBool(int32(obj.Value) == p.nTravIds)
}
func Cof_ObjIsTravIdPrevious(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	return giaBool(int32(obj.Value) == p.nTravIds-1)
}

func Cof_ManCreateLogicSimple(g *Gia_Man_t) *Cof_Man_t {
	p := &Cof_Man_t{pGia: g, nObjs: g.nObjs, vCis: vec.Vec_IntAlloc(int(Gia_ManCiNum(g))), vCos: vec.Vec_IntAlloc(int(Gia_ManCoNum(g))), objects: make([]*Cof_Obj_t, g.nObjs)}
	for id := int32(0); id < g.nObjs; id++ {
		a, b, c, kind := g.Gate(id)
		obj := &Cof_Obj_t{Id: id}
		p.objects[id] = obj
		g.pObjs[id].Value = uint32(id)
		var lits []int32
		switch kind {
		case 0:
		case 1:
			obj.fTerm = 1
			vec.Vec_IntPush(p.vCis, id)
		case 2, 3:
			lits = []int32{a, b}
			p.nNodes++
		case 4:
			lits = []int32{a, b, c}
			p.nNodes++
		case 5:
			lits = []int32{a}
			p.nNodes++
		case 6:
			obj.fTerm = 1
			lits = []int32{a}
			vec.Vec_IntPush(p.vCos, id)
		default:
			panic("gia: invalid cofactor graph object")
		}
		for i, lit := range lits {
			fanin := p.objects[lit>>1]
			if fanin == nil {
				panic("gia: cofactor graph is not topological")
			}
			obj.fanins = append(obj.fanins, fanin)
			obj.complements = append(obj.complements, lit&1 != 0)
			fanin.fanouts = append(fanin.fanouts, obj)
			if kind == 4 && i == 0 {
				fanin.nFanoutsM++
			}
		}
	}
	for _, obj := range p.objects {
		obj.refs = int32(len(obj.fanouts))
	}
	return p
}
func Cof_ManStop(p *Cof_Man_t) {
	if p != nil {
		*p = Cof_Man_t{}
	}
}
func Cof_ManCleanValue(p *Cof_Man_t) {
	for _, obj := range p.objects {
		obj.Value = 0
	}
}

func Cof_ManTfoSize_rec(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	if Cof_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Cof_ObjSetTravIdCurrent(p, obj)
	if Cof_ObjIsCo(obj) != 0 {
		return 0
	}
	var count int32 = 1
	for _, next := range obj.fanouts {
		count += Cof_ManTfoSize_rec(p, next)
	}
	return count
}
func Cof_ManTfoSize(p *Cof_Man_t, objects []*Cof_Obj_t, n int32) int32 {
	Cof_ManIncrementTravId(p)
	var count int32
	for _, obj := range objects[:n] {
		count += Cof_ManTfoSize_rec(p, obj) - 1
	}
	return count
}
func Cof_ManTfiSize_rec(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	if Cof_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Cof_ObjSetTravIdCurrent(p, obj)
	if Cof_ObjIsCi(obj) != 0 || Cof_ObjIsConst0(obj) != 0 {
		return 0
	}
	var count int32 = 1
	for _, next := range obj.fanins {
		count += Cof_ManTfiSize_rec(p, next)
	}
	return count
}
func Cof_ManTfiSize(p *Cof_Man_t, objects []*Cof_Obj_t, n int32) int32 {
	Cof_ManIncrementTravId(p)
	var count int32
	for _, obj := range objects[:n] {
		if Cof_ObjIsCo(obj) != 0 {
			obj = obj.fanins[0]
		}
		count += Cof_ManTfiSize_rec(p, obj)
	}
	return count
}
func Cof_ManSuppSize_rec(p *Cof_Man_t, obj *Cof_Obj_t) int32 {
	if Cof_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Cof_ObjSetTravIdCurrent(p, obj)
	if Cof_ObjIsCi(obj) != 0 {
		return 1
	}
	if Cof_ObjIsConst0(obj) != 0 {
		return 0
	}
	var count int32
	for _, next := range obj.fanins {
		count += Cof_ManSuppSize_rec(p, next)
	}
	return count
}
func Cof_ManSuppSize(p *Cof_Man_t, objects []*Cof_Obj_t, n int32) int32 {
	Cof_ManIncrementTravId(p)
	var count int32
	for _, obj := range objects[:n] {
		if Cof_ObjIsCo(obj) != 0 {
			obj = obj.fanins[0]
		}
		count += Cof_ManSuppSize_rec(p, obj)
	}
	return count
}
