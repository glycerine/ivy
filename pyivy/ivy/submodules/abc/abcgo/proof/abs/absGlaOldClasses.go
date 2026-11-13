package abs

import (
	"sync/atomic"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_GlaAbsCount(p *Gla_Man_t, ro, and int32) int32 {
	var count int32
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		obj := Gla_ManObj(p, id)
		if obj.fAbs == 0 {
			continue
		}
		if ro != 0 && obj.fRo == 0 || ro == 0 && and != 0 && obj.fAnd == 0 {
			continue
		}
		count++
	}
	return count
}
func Gla_ManTranslate_rec(g *gia.Gia_Man_t, obj *gia.Gia_Obj_t, classes *vec.Vec_Int_t, usage int32) int32 {
	if gia.Gia_ObjIsTravIdCurrent(g, obj) != 0 {
		return 1
	}
	gia.Gia_ObjSetTravIdCurrent(g, obj)
	if gia.Gia_ObjIsCi(obj) != 0 {
		return 0
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("invalid mapped abstraction cone")
	}
	a := Gla_ManTranslate_rec(g, gia.Gia_ObjFanin0(obj), classes, usage)
	b := Gla_ManTranslate_rec(g, gia.Gia_ObjFanin1(obj), classes, usage)
	if a != 0 || b != 0 {
		vec.Vec_IntAddToEntry(classes, int(gia.Gia_ObjId(g, obj)), usage)
		return 1
	}
	return 0
}
func Gla_ManTranslate(p *Gla_Man_t) *vec.Vec_Int_t {
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p.pGia)))
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		obj := Gla_ManObj(p, id)
		usage := vec.Vec_IntEntry(p.vCoreCounts, int(obj.iGiaObj))
		if usage < 0 {
			panic("negative core usage")
		}
		if usage == 0 {
			usage = 1
		}
		gobj := Gla_ManGiaObj(p, obj)
		if gia.Gia_ObjIsConst0(gobj) != 0 || gia.Gia_ObjIsRo(p.pGia, gobj) != 0 {
			vec.Vec_IntWriteEntry(classes, int(obj.iGiaObj), usage)
			continue
		}
		if gia.Gia_ObjIsAnd(gobj) == 0 {
			panic("invalid mapped abstraction object")
		}
		gia.Gia_ManIncrementTravId(p.pGia)
		for _, fanin := range obj.Fanins[:obj.nFanins] {
			gia.Gia_ObjSetTravIdCurrent(p.pGia, Gla_ManGiaObj(p, Gla_ManObj(p, fanin)))
		}
		Gla_ManTranslate_rec(p.pGia, gobj, classes, usage)
	}
	vec.Vec_IntWriteEntry(classes, 0, p.pPars.iFrame+1)
	if configs := gia.Gia_ManLutConfigs(p.pGia); configs != nil {
		original := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p.pGia0)))
		for i, value := range vec.Vec_IntArray(classes) {
			if value != 0 {
				vec.Vec_IntWriteEntry(original, int(vec.Vec_IntEntry(configs, i)), value)
			}
		}
		return original
	}
	return classes
}
func Gla_ManCollectPPis(p *Gla_Man_t, pis *vec.Vec_Int_t) *vec.Vec_Int_t {
	ppis := vec.Vec_IntAlloc(1000)
	if pis != nil {
		vec.Vec_IntClear(pis)
	}
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		obj := Gla_ManObj(p, id)
		if obj.fConst == 0 && obj.fRo == 0 && obj.fAnd == 0 {
			panic("invalid abstract object")
		}
		for _, fid := range obj.Fanins[:obj.nFanins] {
			fanin := Gla_ManObj(p, fid)
			if fanin.fAbs != 0 {
				continue
			}
			if fanin.fPi == 0 {
				vec.Vec_IntPush(ppis, fid)
			} else if pis != nil {
				vec.Vec_IntPush(pis, fid)
			}
		}
	}
	vec.Vec_IntUniqify(ppis)
	vec.Vec_IntReverseOrder(ppis)
	if pis != nil {
		vec.Vec_IntUniqify(pis)
	}
	return ppis
}
func Gla_ManCountPPis(p *Gla_Man_t) int32 { return int32(vec.Vec_IntSize(Gla_ManCollectPPis(p, nil))) }

var oldGlaExploreRound atomic.Uint64

func Gla_ManExplorePPis(p *Gla_Man_t, ppis *vec.Vec_Int_t) {
	round := oldGlaExploreRound.Add(1)
	if (round-1)%5 == 0 {
		return
	}
	write := 0
	for _, id := range vec.Vec_IntArray(ppis) {
		obj := Gla_ManObj(p, id)
		if obj.fAbs != 0 {
			panic("exploring an abstract object")
		}
		var count uint32
		for _, fid := range obj.Fanins[:obj.nFanins] {
			count += Gla_ManObj(p, fid).fAbs
		}
		if count == 0 || round&1 != 0 && count == 1 {
			continue
		}
		vec.Vec_IntWriteEntry(ppis, write, id)
		write++
	}
	vec.Vec_IntShrink(ppis, write)
}
