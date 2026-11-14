package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lf_ObjSetCiArrival(p *Lf_Man_t, i, t int32) { vec.Vec_IntWriteEntry(&p.vCiArrivals, int(i), t) }
func Lf_ObjCiArrival(p *Lf_Man_t, i int32) int32 { return vec.Vec_IntEntry(&p.vCiArrivals, int(i)) }
func Lf_ObjArrival_rec(p *Lf_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsBuf(obj) != 0 {
		return Lf_ObjArrival_rec(p, Gia_ObjFanin0(obj))
	}
	if Gia_ObjIsAnd(obj) != 0 {
		return Lf_ObjReadBest(p, obj.id).Delay[0]
	}
	if Gia_ObjIsCi(obj) != 0 {
		return Lf_ObjCiArrival(p, Gia_ObjCioId(obj))
	}
	return 0
}
func Lf_ObjCoArrival(p *Lf_Man_t, i int32) int32 {
	return Lf_ObjArrival_rec(p, Gia_ObjFanin0(Gia_ManCo(p.pGia, i)))
}
func Lf_ObjCoArrival2_rec(p *Lf_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsBuf(obj) != 0 {
		return Lf_ObjCoArrival2_rec(p, Gia_ObjFanin0(obj))
	}
	if Gia_ObjIsAnd(obj) != 0 {
		best := Lf_ObjReadBest(p, obj.id)
		index := Lf_BestCutIndex(best)
		if index == 2 && Gia_ObjIsMux(p.pGia, obj) == 0 {
			panic("LF unmapped non-MUX driver")
		}
		return best.Delay[index]
	}
	if Gia_ObjIsCi(obj) != 0 {
		return Lf_ObjCiArrival(p, Gia_ObjCioId(obj))
	}
	return 0
}
func Lf_ObjCoArrival2(p *Lf_Man_t, i int32) int32 {
	return Lf_ObjCoArrival2_rec(p, Gia_ObjFanin0(Gia_ManCo(p.pGia, i)))
}
func Lf_ManAnalyzeCoDrivers(g *Gia_Man_t, drivers, inverts *int32) {
	marks := make([]int32, g.nObjs)
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		marks[Gia_ObjFanin0(co).id] |= 1 << uint(Gia_ObjFaninC0(co))
	}
	*drivers = 0
	*inverts = 0
	for _, mark := range marks {
		if mark != 0 {
			*drivers++
		}
		if mark == 3 {
			*inverts++
		}
	}
}
func Lf_CutSwitches(p *Lf_Man_t, c *Lf_Cut_t) float32 {
	var result float32
	for _, id := range c.pLeaves[:c.nLeaves] {
		result += Lf_ObjSwitches(p, id)
	}
	return result
}
func Lf_CutArea(p *Lf_Man_t, c *Lf_Cut_t) float32 {
	if c.nLeaves < 2 || c.fMux7 != 0 {
		return 0
	}
	if p.pPars.FPower != 0 {
		return float32(c.nLeaves) + Lf_CutSwitches(p, c)
	}
	group := int32(1)
	if p.pPars.FCutGroup != 0 && int32(c.nLeaves) > p.pPars.NLutSize/2 {
		group++
	}
	if p.pPars.FOptEdge != 0 {
		return float32((int32(c.nLeaves) + p.pPars.NAreaTuner) * group)
	}
	return float32(group)
}
func Lf_CutPrint(p *Lf_Man_t, c *Lf_Cut_t) {
	fmt.Printf("%d  {", c.nLeaves)
	for _, id := range c.pLeaves[:c.nLeaves] {
		fmt.Printf(" %d", id)
	}
	fmt.Printf("  }   Late = %d  D = %4d  A = %9.4f  F = %6d\n", c.fLate, c.Delay, c.Flow, c.iFunc)
}
func Lf_CutIsMux(p *Lf_Man_t, c *Lf_Cut_t, obj *Gia_Obj_t) int32 {
	if c.nLeaves != 3 {
		return 0
	}
	if Gia_ObjIsMux(p.pGia, obj) == 0 {
		panic("LF expected MUX")
	}
	if Gia_ObjIsCi(Gia_ObjFanin0(obj)) != 0 || Gia_ObjIsCi(Gia_ObjFanin1(obj)) != 0 {
		return 0
	}
	for _, id := range []int32{Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id), Gia_ObjFaninId2(p.pGia, obj.id)} {
		found := false
		for _, leaf := range c.pLeaves[:3] {
			if leaf == id {
				found = true
				break
			}
		}
		if !found {
			return 0
		}
	}
	return 1
}
func Lf_MemLoadMuxCut(p *Lf_Man_t, id int32, c *Lf_Cut_t) *Lf_Cut_t {
	obj := Gia_ManObj(p.pGia, id)
	if Gia_ObjIsMux(p.pGia, obj) == 0 {
		panic("LF expected MUX")
	}
	c.iFunc = -1
	if p.pPars.FCutMin != 0 {
		c.iFunc = 4
	}
	c.pLeaves[0] = Gia_ObjFaninId0(obj, id)
	c.pLeaves[1] = Gia_ObjFaninId1(obj, id)
	c.pLeaves[2] = Gia_ObjFaninId2(p.pGia, id)
	c.nLeaves = 3
	c.fMux7 = 1
	return c
}
func Lf_ObjCutMux(p *Lf_Man_t, id int32) *Lf_Cut_t { return Lf_MemLoadMuxCut(p, id, new(Lf_Cut_t)) }
func Lf_ObjCutBest(p *Lf_Man_t, id int32) *Lf_Cut_t {
	c := new(Lf_Cut_t)
	best := Lf_ObjReadBest(p, id)
	index := Lf_BestCutIndex(best)
	c.Delay = best.Delay[index]
	c.Flow = best.Flow[index]
	if index == 2 {
		return Lf_MemLoadMuxCut(p, id, c)
	}
	return Lf_MemLoadCut(&p.vStoreOld, int32(best.Cut[index].Handle), id, c, p.pPars.FCutMin, 0)
}
func Lf_ObjCutBestNew(p *Lf_Man_t, id int32, c *Lf_Cut_t) *Lf_Cut_t {
	best := Lf_ObjReadBest(p, id)
	index := Lf_BestCutIndex(best)
	c.Delay = best.Delay[index]
	c.Flow = best.Flow[index]
	if index == 2 {
		return Lf_MemLoadMuxCut(p, id, c)
	}
	return Lf_MemLoadCut(&p.vStoreNew, int32(best.Cut[index].Handle), id, c, 0, 0)
}
