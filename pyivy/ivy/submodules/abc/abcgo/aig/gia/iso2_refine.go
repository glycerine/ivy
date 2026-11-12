package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

const iso2PrimeC uint32 = 49

type Gia_Iso2Man_t struct {
	pGia                                                           *Gia_Man_t
	nObjs, nUniques                                                int32
	vUniques, vTied, vTable, vPlaces                               *Vec_Int_t
	vSingles                                                       *Vec_Ptr_t
	vVec0, vVec1, vMap0, vMap1                                     *Vec_Int_t
	nIters                                                         int32
	timeStart, timeSim, timeRefine, timeSort, timeOther, timeTotal int64
}

func Gia_Iso2ManCollectTies(g *Gia_Man_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(Gia_ManCandNum(g)))
	for id := int32(0); id < Gia_ManObjNum(g); id++ {
		if Gia_ObjIsCand(Gia_ManObj(g, id)) != 0 {
			vec.Vec_IntPush(result, id)
		}
	}
	return result
}
func Gia_Iso2ManPrepare(g *Gia_Man_t) {
	for id := int32(0); id < Gia_ManObjNum(g); id++ {
		obj := Gia_ManObj(g, id)
		obj.Value = 0
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = 1 + max(Gia_ObjFanin0(obj).Value, Gia_ObjFanin1(obj).Value)
		}
	}
	Gia_ManConst0(g).Value = s_256Primes[ISO_MASK]
	for id := int32(1); id < Gia_ManObjNum(g); id++ {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = s_256Primes[obj.Value&ISO_MASK] + s_256Primes[ISO_MASK-10+Gia_ObjFaninC0(obj)+Gia_ObjFaninC1(obj)]
		} else if Gia_ObjIsPi(g, obj) != 0 {
			obj.Value = s_256Primes[ISO_MASK-1]
		} else if Gia_ObjIsRo(g, obj) != 0 {
			obj.Value = s_256Primes[ISO_MASK-2]
		}
	}
}
func Gia_Iso2ManPropagate(g *Gia_Man_t) {
	for id := int32(1); id < Gia_ManObjNum(g); id++ {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) != 0 {
			a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
			obj.Value += (uint32(Gia_ObjFaninC0(obj))+iso2PrimeC)*a.Value + (uint32(Gia_ObjFaninC1(obj))+iso2PrimeC)*b.Value
			if Gia_ObjFaninC0(obj) == Gia_ObjFaninC1(obj) && a.Value == b.Value {
				obj.Value += s_256Primes[ISO_MASK-11]
			}
		} else if Gia_ObjIsCo(obj) != 0 {
			obj.Value += (uint32(Gia_ObjFaninC0(obj)) + iso2PrimeC) * Gia_ObjFanin0(obj).Value
		}
	}
	for i := int32(0); i < Gia_ManRegNum(g); i++ {
		ri, ro := Gia_ManRi(g, i), Gia_ManRo(g, i)
		ro.Value += ri.Value
		if ro == Gia_ObjFanin0(ri) {
			ro.Value += s_256Primes[ISO_MASK-12]
		}
	}
}
func Gia_Iso2ManCone_rec(g *Gia_Man_t, id, level int32) uint32 {
	if level == 0 || Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(g, id)
	obj := Gia_ManObj(g, id)
	if Gia_ObjIsAnd(obj) != 0 {
		return obj.Value + Gia_Iso2ManCone_rec(g, Gia_ObjFaninId0(obj, id), level-1) + Gia_Iso2ManCone_rec(g, Gia_ObjFaninId1(obj, id), level-1)
	}
	if Gia_ObjIsPi(g, obj) != 0 {
		return obj.Value
	}
	if Gia_ObjIsRo(g, obj) != 0 {
		return obj.Value + Gia_Iso2ManCone_rec(g, Gia_ObjId(g, Gia_ObjFanin0(Gia_ObjRoToRi(g, obj))), level)
	}
	if Gia_ObjIsConst0(obj) == 0 {
		panic("gia: unexpected isomorphism cone object")
	}
	return obj.Value
}
func Gia_Iso2ManCone(g *Gia_Man_t, id, level int32) uint32 {
	Gia_ManIncrementTravId(g)
	return Gia_Iso2ManCone_rec(g, id, level)
}
func Gia_Iso2ManUpdate(p *Gia_Iso2Man_t, level int32) {
	for _, id := range vec.Vec_IntArray(p.vTied) {
		Gia_ManObj(p.pGia, id).Value += Gia_Iso2ManCone(p.pGia, id, level)
	}
}
func Gia_Iso2ManStart(g *Gia_Man_t) *Gia_Iso2Man_t {
	n := int(Gia_ManObjNum(g))
	p := &Gia_Iso2Man_t{pGia: g, nObjs: int32(n), vUniques: vec.Vec_IntStartFull(n), vTied: Gia_Iso2ManCollectTies(g), vTable: vec.Vec_IntStart(max(2, 2*n+1)), vPlaces: vec.Vec_IntAlloc(1000), vSingles: vec.Vec_PtrAlloc(1000), vVec0: vec.Vec_IntAlloc(10000), vVec1: vec.Vec_IntAlloc(10000), vMap0: vec.Vec_IntStart(n), vMap1: vec.Vec_IntStart(n)}
	vec.Vec_IntWriteEntry(p.vUniques, 0, 0)
	p.nUniques = 1
	return p
}
func Gia_Iso2ManStop(p *Gia_Iso2Man_t) {
	if p != nil {
		*p = Gia_Iso2Man_t{}
	}
}
func Gia_Iso2ManPrint(p *Gia_Iso2Man_t, elapsed int64, verbose int32) {
	if verbose == 0 {
		return
	}
	fmt.Printf("Iter %4d :  Entries =%8d.  Uniques =%8d.  Singles =%8d.  %9.2f sec\n", p.nIters, vec.Vec_IntSize(p.vTied), p.nUniques, vec.Vec_PtrSize(p.vSingles), float64(elapsed)/float64(time.Second))
	p.nIters++
}
func Gia_ObjCompareByValue2(a, b **Gia_Obj_t) int32 { return Gia_ObjCompareByValue(a, b) }

// Gia_Iso2ManUniqify assigns labels to unique signatures and retains ties.
func Gia_Iso2ManUniqify(p *Gia_Iso2Man_t) int32 {
	counts := make(map[uint32]int)
	for _, id := range vec.Vec_IntArray(p.vTied) {
		counts[Gia_ManObj(p.pGia, id).Value]++
	}
	var tied []int32
	var singles []*Gia_Obj_t
	for _, id := range vec.Vec_IntArray(p.vTied) {
		obj := Gia_ManObj(p.pGia, id)
		if counts[obj.Value] == 1 {
			singles = append(singles, obj)
		} else {
			tied = append(tied, id)
		}
	}
	vec.Vec_IntClear(p.vTied)
	for _, id := range tied {
		vec.Vec_IntPush(p.vTied, id)
	}
	vec.Vec_PtrClear(p.vSingles)
	sort.SliceStable(singles, func(i, j int) bool { return int32(singles[i].Value) < int32(singles[j].Value) })
	for _, obj := range singles {
		id := Gia_ObjId(p.pGia, obj)
		obj.Value += s_256Primes[p.nUniques&ISO_MASK]
		if vec.Vec_IntEntry(p.vUniques, int(id)) != -1 {
			panic("gia: duplicate unique label")
		}
		vec.Vec_IntWriteEntry(p.vUniques, int(id), p.nUniques)
		p.nUniques++
		vec.Vec_PtrPush(p.vSingles, obj)
	}
	return int32(len(singles))
}
func Gia_Iso2ManDerivePoClasses(g *Gia_Man_t) *Vec_Wec_t {
	classes := vec.Vec_WecAlloc(int(Gia_ManPoNum(g)))
	ids := make(map[uint32]int)
	for i := int32(0); i < Gia_ManPoNum(g); i++ {
		value := Gia_ManPo(g, i).Value
		index, ok := ids[value]
		if !ok {
			index = len(ids)
			ids[value] = index
			vec.Vec_WecPushLevel(classes)
		}
		vec.Vec_WecPush(classes, index, i)
	}
	return classes
}
func Gia_Iso2ManPerform(g *Gia_Man_t, verbose int32) *Vec_Wec_t {
	start := time.Now()
	p := Gia_Iso2ManStart(g)
	defer Gia_Iso2ManStop(p)
	Gia_Iso2ManPrepare(g)
	Gia_Iso2ManPropagate(g)
	Gia_Iso2ManPrint(p, time.Since(start).Nanoseconds(), verbose)
	for Gia_Iso2ManUniqify(p) != 0 {
		Gia_Iso2ManPrint(p, time.Since(start).Nanoseconds(), verbose)
		Gia_Iso2ManPropagate(g)
	}
	Gia_Iso2ManPrint(p, time.Since(start).Nanoseconds(), verbose)
	return Gia_Iso2ManDerivePoClasses(g)
}
