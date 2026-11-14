package gia

import "github.com/glycerine/abcgo/misc/vec"

type Emb_Dat_t = float32

// Emb_Obj_t stores adjacency explicitly instead of C's relative pointer offsets.
type Emb_Obj_t struct {
	fCi, fCo, fMark0, fMark1 uint32
	hHandle                  int32
	TravId, Value            uint32
	fanins, fanouts          []*Emb_Obj_t
}
type Emb_Man_t struct {
	pGia                   *Gia_Man_t
	vCis, vCos             *Vec_Int_t
	objects                []*Emb_Obj_t
	nObjs, nRegs, nTravIds int32
	fVerbose               int32
	pVecs                  []Emb_Dat_t
	nReached, nDistMax     int32
	pMatr, pEigen          [][]float32
	pSols                  []float32
	pPlacement             []uint16
}

func Emb_ManRegNum(p *Emb_Man_t) int32                 { return p.nRegs }
func Emb_ManCiNum(p *Emb_Man_t) int32                  { return int32(vec.Vec_IntSize(p.vCis)) }
func Emb_ManCoNum(p *Emb_Man_t) int32                  { return int32(vec.Vec_IntSize(p.vCos)) }
func Emb_ManPiNum(p *Emb_Man_t) int32                  { return Emb_ManCiNum(p) - p.nRegs }
func Emb_ManPoNum(p *Emb_Man_t) int32                  { return Emb_ManCoNum(p) - p.nRegs }
func Emb_ManObjNum(p *Emb_Man_t) int32                 { return p.nObjs }
func Emb_ManNodeNum(p *Emb_Man_t) int32                { return p.nObjs - Emb_ManCiNum(p) - Emb_ManCoNum(p) }
func Emb_ManObj(p *Emb_Man_t, handle int32) *Emb_Obj_t { return p.objects[handle] }
func Emb_ManCi(p *Emb_Man_t, i int32) *Emb_Obj_t {
	return Emb_ManObj(p, vec.Vec_IntEntry(p.vCis, int(i)))
}
func Emb_ManCo(p *Emb_Man_t, i int32) *Emb_Obj_t {
	return Emb_ManObj(p, vec.Vec_IntEntry(p.vCos, int(i)))
}
func Emb_ObjIsTerm(o *Emb_Obj_t) int32               { return giaBool(o.fCi != 0 || o.fCo != 0) }
func Emb_ObjIsCi(o *Emb_Obj_t) int32                 { return int32(o.fCi) }
func Emb_ObjIsCo(o *Emb_Obj_t) int32                 { return int32(o.fCo) }
func Emb_ObjIsNode(o *Emb_Obj_t) int32               { return giaBool(Emb_ObjIsTerm(o) == 0 && len(o.fanins) > 0) }
func Emb_ObjSize(o *Emb_Obj_t) int32                 { return 1 }
func Emb_ObjFaninNum(o *Emb_Obj_t) int32             { return int32(len(o.fanins)) }
func Emb_ObjFanoutNum(o *Emb_Obj_t) int32            { return int32(len(o.fanouts)) }
func Emb_ObjFanin(o *Emb_Obj_t, i int32) *Emb_Obj_t  { return o.fanins[i] }
func Emb_ObjFanout(o *Emb_Obj_t, i int32) *Emb_Obj_t { return o.fanouts[i] }
func Emb_ObjAddFanin(o, fanin *Emb_Obj_t) {
	o.fanins = append(o.fanins, fanin)
	fanin.fanouts = append(fanin.fanouts, o)
}

// Placement keeps the source's two stored pins (including repeated buffer
// pins). Real MUXes add a control pin stored outside the object itself.
func giaPlacementFanins(g *Gia_Man_t, obj *Gia_Obj_t) []int32 {
	fanins := []int32{Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id)}
	if Gia_ObjIsMux(g, obj) != 0 {
		fanins = append(fanins, Gia_ObjFaninId2(g, obj.id))
	}
	return fanins
}
func Emb_ManCleanTravId(p *Emb_Man_t) {
	for _, o := range p.objects {
		o.TravId = 0
	}
}
func Emb_ManSetValue(p *Emb_Man_t) {
	for i, o := range p.objects {
		o.Value = uint32(i)
	}
}
func Emb_ManResetTravId(p *Emb_Man_t)                     { Emb_ManCleanTravId(p); p.nTravIds = 1 }
func Emb_ManIncrementTravId(p *Emb_Man_t)                 { p.nTravIds++ }
func Emb_ObjSetTravId(o *Emb_Obj_t, id int32)             { o.TravId = uint32(id) }
func Emb_ObjSetTravIdCurrent(p *Emb_Man_t, o *Emb_Obj_t)  { o.TravId = uint32(p.nTravIds) }
func Emb_ObjSetTravIdPrevious(p *Emb_Man_t, o *Emb_Obj_t) { o.TravId = uint32(p.nTravIds - 1) }
func Emb_ObjIsTravIdCurrent(p *Emb_Man_t, o *Emb_Obj_t) int32 {
	return giaBool(o.TravId == uint32(p.nTravIds))
}
func Emb_ObjIsTravIdPrevious(p *Emb_Man_t, o *Emb_Obj_t) int32 {
	return giaBool(o.TravId == uint32(p.nTravIds-1))
}
func Emb_ManVec(p *Emb_Man_t, v int32) []Emb_Dat_t {
	start := v * p.nObjs
	return p.pVecs[start : start+p.nObjs]
}
func Emb_ManSol(p *Emb_Man_t, v int32) []float32 {
	start := v * p.nObjs
	return p.pSols[start : start+p.nObjs]
}

// Emb_ManStartSimple creates an embedding vertex for every GIA object.
func Emb_ManStartSimple(g *Gia_Man_t) *Emb_Man_t {
	p := &Emb_Man_t{pGia: g, nRegs: Gia_ManRegNum(g), vCis: vec.Vec_IntAlloc(int(Gia_ManCiNum(g))), vCos: vec.Vec_IntAlloc(int(Gia_ManCoNum(g)))}
	byID := make([]*Emb_Obj_t, int(g.nObjs))
	add := func(obj *Gia_Obj_t) *Emb_Obj_t {
		o := &Emb_Obj_t{hHandle: int32(len(p.objects))}
		p.objects = append(p.objects, o)
		byID[obj.id] = o
		Gia_ObjSetValue(obj, o.hHandle)
		return o
	}
	constant := add(Gia_ManConst0(g))
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		constant.fanins = append(constant.fanins, constant)
	}
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		o := add(Gia_ManCi(g, i))
		o.fCi = 1
		vec.Vec_IntPush(p.vCis, o.hHandle)
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		o := add(obj)
		for _, id := range giaPlacementFanins(g, obj) {
			Emb_ObjAddFanin(o, byID[id])
		}
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		o := add(obj)
		o.fCo = 1
		vec.Vec_IntPush(p.vCos, o.hHandle)
		Emb_ObjAddFanin(o, byID[Gia_ObjFaninId0(obj, obj.id)])
	}
	for i := int32(0); i < p.nRegs; i++ {
		Emb_ObjAddFanin(byID[Gia_ManRo(g, i).id], byID[Gia_ManRi(g, i).id])
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		o := byID[Gia_ManCo(g, i).id]
		o.fanouts = append(o.fanouts, o)
	}
	p.nObjs = int32(len(p.objects))
	return p
}

func Emb_ManCollectSuper_rec(p *Gia_Man_t, obj *Gia_Obj_t, super, visit *Vec_Int_t) {
	if obj.fMark1 != 0 {
		return
	}
	obj.fMark1 = 1
	vec.Vec_IntPush(visit, obj.id)
	if obj.fMark0 != 0 {
		vec.Vec_IntPush(super, obj.id)
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unmarked supernode leaf")
	}
	for _, id := range giaPlacementFanins(p, obj) {
		Emb_ManCollectSuper_rec(p, Gia_ManObj(p, id), super, visit)
	}
}
func Emb_ManCollectSuper(p *Gia_Man_t, obj *Gia_Obj_t, super, visit *Vec_Int_t) {
	vec.Vec_IntClear(super)
	vec.Vec_IntClear(visit)
	if obj.fMark0 == 0 {
		panic("gia: unmarked supernode root")
	}
	obj.fMark0 = 0
	Emb_ManCollectSuper_rec(p, obj, super, visit)
	obj.fMark0 = 1
	for _, id := range vec.Vec_IntArray(visit) {
		p.pObjs[id].fMark1 = 0
	}
}
func Emb_ManCreateRefsSpecial(p *Gia_Man_t) {
	Gia_ManCleanMark0(p)
	Gia_ManCreateRefs(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		if Gia_ObjIsCi(a) != 0 || a.fMark0 != 0 || Gia_ObjIsCi(b) != 0 || b.fMark0 != 0 || Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		obj.fMark0 = 1
		a.fMark0 = 1
		b.fMark0 = 1
		var d1, d0 *Gia_Obj_t
		control := Gia_ObjRecognizeMux(obj, &d1, &d0)
		Gia_ObjRefDec(p, Gia_Regular(control))
		if Gia_Regular(d0) == Gia_Regular(d1) {
			Gia_ObjRefDec(p, Gia_Regular(d0))
		}
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(p, obj) <= 0 {
			panic("gia: dangling AND in embedding")
		}
	}
	Gia_ManCleanMark0(p)
}
func Emb_ManTransformRefs(p *Gia_Man_t, nObjs, nFanios *int32) {
	if p.Refs == nil {
		panic("gia: missing reference counts")
	}
	Gia_ManCleanMark0(p)
	Gia_ManConst0(p).fMark0 = 1
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).fMark0 = 1
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(p, obj) > 1 {
			obj.fMark0 = 1
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ObjFanin0(Gia_ManCo(p, i)).fMark0 = 1
	}
	*nObjs = Gia_ManCoNum(p)
	for _, obj := range p.pObjs {
		*nObjs += int32(obj.fMark0)
	}
	p.Refs = make([]int32, p.nObjs)
	super, visit := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	Gia_ManCleanMark1(p)
	var count int32
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 == 0 {
			continue
		}
		Emb_ManCollectSuper(p, obj, super, visit)
		for _, id := range vec.Vec_IntArray(super) {
			Gia_ObjRefIncId(p, id)
		}
		count += int32(vec.Vec_IntSize(super))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ObjRefInc(p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	*nFanios = count + Gia_ManCoNum(p)
}
func Emb_ManStart(g *Gia_Man_t) *Emb_Man_t {
	Emb_ManCreateRefsSpecial(g)
	var nObjs, nFanios int32
	Emb_ManTransformRefs(g, &nObjs, &nFanios)
	Gia_ManFillValue(g)
	p := &Emb_Man_t{pGia: g, nRegs: Gia_ManRegNum(g), vCis: vec.Vec_IntAlloc(int(Gia_ManCiNum(g))), vCos: vec.Vec_IntAlloc(int(Gia_ManCoNum(g)))}
	byID := make([]*Emb_Obj_t, int(g.nObjs))
	add := func(obj *Gia_Obj_t) *Emb_Obj_t {
		o := &Emb_Obj_t{hHandle: int32(len(p.objects))}
		p.objects = append(p.objects, o)
		byID[obj.id] = o
		Gia_ObjSetValue(obj, o.hHandle)
		return o
	}
	constant := add(Gia_ManConst0(g))
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		constant.fanins = append(constant.fanins, constant)
	}
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		o := add(Gia_ManCi(g, i))
		o.fCi = 1
		vec.Vec_IntPush(p.vCis, o.hHandle)
	}
	super, visit := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 == 0 {
			continue
		}
		Emb_ManCollectSuper(g, obj, super, visit)
		o := add(obj)
		for _, id := range vec.Vec_IntArray(super) {
			if byID[id] == nil {
				panic("gia: supernode is not topologically ordered")
			}
			Emb_ObjAddFanin(o, byID[id])
		}
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		o := add(obj)
		o.fCo = 1
		vec.Vec_IntPush(p.vCos, o.hHandle)
		Emb_ObjAddFanin(o, byID[Gia_ObjFaninId0(obj, obj.id)])
	}
	for i := int32(0); i < p.nRegs; i++ {
		Emb_ObjAddFanin(byID[Gia_ManRo(g, i).id], byID[Gia_ManRi(g, i).id])
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		o := byID[Gia_ManCo(g, i).id]
		o.fanouts = append(o.fanouts, o)
	}
	p.nObjs = int32(len(p.objects))
	Gia_ManCleanMark0(g)
	g.Refs = nil
	if p.nObjs != nObjs {
		panic("gia: incorrect embedding object count")
	}
	return p
}
