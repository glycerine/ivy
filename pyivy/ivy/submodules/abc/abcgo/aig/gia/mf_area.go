package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

func Mf_CutArea(p *Mf_Man_t, n, f int32) int32 {
	if n < 2 {
		return 0
	}
	if p.pPars.FGenCnf != 0 || p.pPars.FGenLit != 0 {
		return vec.Vec_IntEntry(&p.vCnfSizes, int(uint32(f)>>1))
	}
	if p.pPars.FOptEdge != 0 {
		return n + p.pPars.NAreaTuner
	}
	return 1
}
func Mf_CutParams(p *Mf_Man_t, c *Mf_Cut_t, refs float32) {
	if c.nLeaves > uint32(p.pPars.NLutSize) {
		panic("gia: cut exceeds LUT size")
	}
	c.Delay, c.Flow = 0, 0
	for _, id := range c.pLeaves[:c.nLeaves] {
		best := Mf_ManObj(p, id)
		c.Delay = max(c.Delay, int32(best.Delay))
		c.Flow += best.Flow
	}
	if c.nLeaves > 1 {
		c.Delay++
	}
	c.Flow = (c.Flow + float32(Mf_CutArea(p, int32(c.nLeaves), int32(c.iFunc)))) / refs
}
func mfCutReference(p *Mf_Man_t, cut []int32, add bool, temp *Vec_Int_t, limit int32) int32 {
	count := Mf_CutArea(p, Mf_CutSize(cut), Mf_CutFunc(cut))
	if limit == 0 {
		return count
	}
	for _, id := range cut[1 : 1+Mf_CutSize(cut)] {
		if temp != nil {
			vec.Vec_IntPush(temp, id)
		}
		var refs int32
		if add {
			refs = Mf_ObjMapRefInc(p, id)
		} else {
			refs = Mf_ObjMapRefDec(p, id)
		}
		if refs == 0 && Mf_ManObj(p, id).iCutSet != 0 {
			count += mfCutReference(p, Mf_ObjCutBest(p, id), add, temp, limit-1)
		}
	}
	return count
}
func Mf_CutRef_rec(p *Mf_Man_t, c []int32) int32 {
	return mfCutReference(p, c, true, nil, math.MaxInt32)
}
func Mf_CutDeref_rec(p *Mf_Man_t, c []int32) int32 {
	return mfCutReference(p, c, false, nil, math.MaxInt32)
}
func Mf_CutAreaRefed(p *Mf_Man_t, c []int32) int32 {
	a := Mf_CutDeref_rec(p, c)
	if Mf_CutRef_rec(p, c) != a {
		panic("gia: inconsistent cut reference area")
	}
	return a
}
func Mf_CutAreaDerefed(p *Mf_Man_t, c []int32) int32 {
	a := Mf_CutRef_rec(p, c)
	if Mf_CutDeref_rec(p, c) != a {
		panic("gia: inconsistent cut reference area")
	}
	return a
}
func Mf_CutAreaMffc(p *Mf_Man_t, id int32) int32 {
	if Mf_ObjMapRefNum(p, id) != 0 {
		return Mf_CutAreaRefed(p, Mf_ObjCutBest(p, id))
	}
	return Mf_CutAreaDerefed(p, Mf_ObjCutBest(p, id))
}
func Mf_CutRef2_rec(p *Mf_Man_t, c []int32, temp *Vec_Int_t, limit int32) int32 {
	return mfCutReference(p, c, true, temp, limit)
}
func Mf_CutDeref2_rec(p *Mf_Man_t, c []int32, temp *Vec_Int_t, limit int32) int32 {
	return mfCutReference(p, c, false, temp, limit)
}
func mfCutAreaLimited(p *Mf_Man_t, c []int32, referenced bool) int32 {
	vec.Vec_IntClear(&p.vTemp)
	area := mfCutReference(p, c, !referenced, &p.vTemp, 8)
	for _, id := range vec.Vec_IntArray(&p.vTemp) {
		if referenced {
			Mf_ObjMapRefInc(p, id)
		} else {
			Mf_ObjMapRefDec(p, id)
		}
	}
	return area
}
func Mf_CutAreaRefed2(p *Mf_Man_t, c []int32) int32   { return mfCutAreaLimited(p, c, true) }
func Mf_CutAreaDerefed2(p *Mf_Man_t, c []int32) int32 { return mfCutAreaLimited(p, c, false) }
func Mf_CutAreaRefed2Multi(p *Mf_Man_t, id int32, cuts [][]int32, n int32) int32 {
	vec.Vec_IntClear(&p.vTemp)
	var area int32
	for _, c := range cuts[:n] {
		area += Mf_CutDeref2_rec(p, c, &p.vTemp, math.MaxInt32)
	}
	if Mf_ObjMapRefNum(p, id) != 0 {
		panic("gia: expected unreferenced object")
	}
	for _, v := range vec.Vec_IntArray(&p.vTemp) {
		Mf_ObjMapRefInc(p, v)
	}
	return area
}
func Mf_CutFlow(p *Mf_Man_t, c []int32, time *int32) float32 {
	var flow float32
	var delay int32
	for _, id := range c[1 : 1+Mf_CutSize(c)] {
		obj := Mf_ManObj(p, id)
		delay = max(delay, int32(obj.Delay))
		flow += obj.Flow
	}
	*time = delay + 1
	return flow + float32(Mf_CutArea(p, Mf_CutSize(c), Mf_CutFunc(c)))
}
