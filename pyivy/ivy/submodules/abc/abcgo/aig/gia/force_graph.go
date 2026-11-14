package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

type Frc_Obj_t struct {
	fCi, fCo, fMark0, fMark1 uint32
	hHandle, pPlace, iFanout int32
	fEdgeCenter              float32
	fanins, fanouts          []*Frc_Obj_t
}
type Frc_Man_t struct {
	pGia                                     *Gia_Man_t
	vCis, vCos                               *Vec_Int_t
	objects                                  []*Frc_Obj_t
	nObjs, nRegs, fVerbose, nCutCur, nCutMax int32
}

func Frc_ManRegNum(p *Frc_Man_t) int32             { return p.nRegs }
func Frc_ManCiNum(p *Frc_Man_t) int32              { return int32(vec.Vec_IntSize(p.vCis)) }
func Frc_ManCoNum(p *Frc_Man_t) int32              { return int32(vec.Vec_IntSize(p.vCos)) }
func Frc_ManPiNum(p *Frc_Man_t) int32              { return Frc_ManCiNum(p) - p.nRegs }
func Frc_ManPoNum(p *Frc_Man_t) int32              { return Frc_ManCoNum(p) - p.nRegs }
func Frc_ManObjNum(p *Frc_Man_t) int32             { return p.nObjs }
func Frc_ManNodeNum(p *Frc_Man_t) int32            { return p.nObjs - Frc_ManCiNum(p) - Frc_ManCoNum(p) }
func Frc_ManObj(p *Frc_Man_t, id int32) *Frc_Obj_t { return p.objects[id] }
func Frc_ManCi(p *Frc_Man_t, id int32) *Frc_Obj_t {
	return Frc_ManObj(p, vec.Vec_IntEntry(p.vCis, int(id)))
}
func Frc_ManCo(p *Frc_Man_t, id int32) *Frc_Obj_t {
	return Frc_ManObj(p, vec.Vec_IntEntry(p.vCos, int(id)))
}
func Frc_ObjIsTerm(o *Frc_Obj_t) int32               { return giaBool(o.fCi != 0 || o.fCo != 0) }
func Frc_ObjIsCi(o *Frc_Obj_t) int32                 { return int32(o.fCi) }
func Frc_ObjIsCo(o *Frc_Obj_t) int32                 { return int32(o.fCo) }
func Frc_ObjIsPi(o *Frc_Obj_t) int32                 { return giaBool(o.fCi != 0 && len(o.fanins) == 0) }
func Frc_ObjIsPo(o *Frc_Obj_t) int32                 { return giaBool(o.fCo != 0 && len(o.fanouts) == 0) }
func Frc_ObjIsNode(o *Frc_Obj_t) int32               { return giaBool(Frc_ObjIsTerm(o) == 0 && len(o.fanins) > 0) }
func Frc_ObjIsConst0(o *Frc_Obj_t) int32             { return giaBool(Frc_ObjIsTerm(o) == 0 && len(o.fanins) == 0) }
func Frc_ObjSize(o *Frc_Obj_t) int32                 { return 1 }
func Frc_ObjFaninNum(o *Frc_Obj_t) int32             { return int32(len(o.fanins)) }
func Frc_ObjFanoutNum(o *Frc_Obj_t) int32            { return int32(len(o.fanouts)) }
func Frc_ObjFanin(o *Frc_Obj_t, i int32) *Frc_Obj_t  { return o.fanins[i] }
func Frc_ObjFanout(o *Frc_Obj_t, i int32) *Frc_Obj_t { return o.fanouts[i] }
func Frc_ObjAddFanin(o, fanin *Frc_Obj_t) {
	o.fanins = append(o.fanins, fanin)
	fanin.fanouts = append(fanin.fanouts, o)
}
func Frc_ManCollectSuper_rec(p *Gia_Man_t, obj *Gia_Obj_t, super, visit *Vec_Int_t) {
	Emb_ManCollectSuper_rec(p, obj, super, visit)
}
func Frc_ManCollectSuper(p *Gia_Man_t, obj *Gia_Obj_t, super, visit *Vec_Int_t) {
	Emb_ManCollectSuper(p, obj, super, visit)
}
func Frc_ManCreateRefsSpecial(p *Gia_Man_t) { Emb_ManCreateRefsSpecial(p) }
func Frc_ManTransformRefs(p *Gia_Man_t, nObjs, nFanios *int32) {
	Emb_ManTransformRefs(p, nObjs, nFanios)
}
func frcManStart(g *Gia_Man_t, cluster bool) *Frc_Man_t {
	var expected int32
	if cluster {
		Frc_ManCreateRefsSpecial(g)
		var fanios int32
		Frc_ManTransformRefs(g, &expected, &fanios)
	}
	Gia_ManFillValue(g)
	p := &Frc_Man_t{pGia: g, nRegs: g.nRegs, vCis: vec.Vec_IntAlloc(int(Gia_ManCiNum(g))), vCos: vec.Vec_IntAlloc(int(Gia_ManCoNum(g)))}
	byID := make([]*Frc_Obj_t, g.nObjs)
	add := func(obj *Gia_Obj_t) *Frc_Obj_t {
		o := &Frc_Obj_t{hHandle: int32(len(p.objects))}
		p.objects = append(p.objects, o)
		byID[obj.id] = o
		Gia_ObjSetValue(obj, o.hHandle)
		return o
	}
	add(Gia_ManConst0(g))
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		o := add(Gia_ManCi(g, i))
		o.fCi = 1
		vec.Vec_IntPush(p.vCis, o.hHandle)
	}
	super, visit := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || cluster && obj.fMark0 == 0 {
			continue
		}
		var fanins []int32
		if cluster {
			Frc_ManCollectSuper(g, obj, super, visit)
			fanins = vec.Vec_IntArray(super)
		} else {
			fanins = giaPlacementFanins(g, obj)
		}
		o := add(obj)
		for _, id := range fanins {
			if byID[id] == nil {
				panic("gia: FORCE graph not topological")
			}
			Frc_ObjAddFanin(o, byID[id])
		}
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		o := add(obj)
		o.fCo = 1
		vec.Vec_IntPush(p.vCos, o.hHandle)
		Frc_ObjAddFanin(o, byID[Gia_ObjFaninId0(obj, obj.id)])
	}
	p.nObjs = int32(len(p.objects))
	if cluster {
		Gia_ManCleanMark0(g)
		g.Refs = nil
		if expected != p.nObjs {
			panic("gia: FORCE object count mismatch")
		}
	}
	return p
}
func Frc_ManStartSimple(g *Gia_Man_t) *Frc_Man_t { return frcManStart(g, false) }
func Frc_ManStart(g *Gia_Man_t) *Frc_Man_t       { return frcManStart(g, true) }
func Frc_ManPrintStats(p *Frc_Man_t) {
	fmt.Printf("i/o =%7d/%7d  ", Frc_ManPiNum(p), Frc_ManPoNum(p))
	if p.nRegs != 0 {
		fmt.Printf("ff =%7d  ", p.nRegs)
	}
	fmt.Printf("node =%8d  obj =%8d\n", Frc_ManNodeNum(p), p.nObjs)
}
func Frc_ManStop(p *Frc_Man_t) {
	if p != nil {
		*p = Frc_Man_t{}
	}
}
