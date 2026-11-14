package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Kf_ObjTime(p *Kf_Man_t, id int32) int32   { return vec.Vec_IntEntry(&p.vTime, int(id)) }
func Kf_ObjArea(p *Kf_Man_t, id int32) float32 { return vec.Vec_FltEntry(&p.vArea, int(id)) }
func Kf_ObjRefs(p *Kf_Man_t, id int32) float32 { return vec.Vec_FltEntry(&p.vRefs, int(id)) }
func Kf_ObjSetCuts(p *Kf_Man_t, id int32, v *Vec_Int_t) {
	vec.Vec_IntWriteEntry(&p.vCuts, int(id), p.pMem.append(vec.Vec_IntArray(v), vec.Vec_IntSize(v)))
}
func Kf_ObjCuts(p *Kf_Man_t, id int32) []int32 {
	return p.pMem.get(vec.Vec_IntEntry(&p.vCuts, int(id)))
}
func Kf_ObjCuts0(p *Kf_Man_t, id int32) []int32 {
	return Kf_ObjCuts(p, Gia_ObjFaninId0(Gia_ManObj(p.pGia, id), id))
}
func Kf_ObjCuts1(p *Kf_Man_t, id int32) []int32 {
	return Kf_ObjCuts(p, Gia_ObjFaninId1(Gia_ManObj(p.pGia, id), id))
}
func Kf_ObjCutBest(p *Kf_Man_t, id int32) []int32 { cuts := Kf_ObjCuts(p, id); return cuts[cuts[1]:] }
func Kf_CutSize(cut []int32) int32                { return cut[0] }
func Kf_CutFunc(cut []int32) int32                { return cut[cut[0]+1] }
func Kf_CutLeaf(cut []int32, i int32) int32 {
	if i == 0 {
		panic("KF invalid cut index")
	}
	return cut[i] >> 1
}
func Kf_CutTime(p *Kf_Man_t, cut []int32) int32 {
	var result int32
	for i := int32(1); i <= Kf_CutSize(cut); i++ {
		result = max(result, Kf_ObjTime(p, Kf_CutLeaf(cut, i)))
	}
	return result + 1
}
func Kf_CutRef(p *Kf_Man_t, cut []int32) {
	for i := int32(1); i <= Kf_CutSize(cut); i++ {
		Gia_ObjRefIncId(p.pGia, Kf_CutLeaf(cut, i))
	}
}
func Kf_CutDeref(p *Kf_Man_t, cut []int32) {
	for i := int32(1); i <= Kf_CutSize(cut); i++ {
		Gia_ObjRefDecId(p.pGia, Kf_CutLeaf(cut, i))
	}
}
func Kf_CutPrint(cut []int32) {
	fmt.Printf("%d {", Kf_CutSize(cut))
	for i := int32(1); i <= Kf_CutSize(cut); i++ {
		fmt.Printf(" %d", Kf_CutLeaf(cut, i))
	}
	fmt.Printf(" } Func = %d\n", Kf_CutFunc(cut))
}
func Gia_CutSetPrint(cuts []int32) {
	off := KF_ADD_ON1
	for i := int32(0); i < cuts[0]; i++ {
		cut := cuts[off:]
		Kf_CutPrint(cut)
		off += int(cut[0]) + KF_ADD_ON2
	}
	fmt.Println()
}
func Kf_ManSetInitRefs(g *Gia_Man_t, refs *Vec_Flt_t) {
	vec.Vec_FltFill(refs, int(g.nObjs), 0)
	data := vec.Vec_FltArray(refs)
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		data[Gia_ObjFanin0(obj).id]++
		data[Gia_ObjFanin1(obj).id]++
		if Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var d1, d0 *Gia_Obj_t
		ctrl := Gia_ObjRecognizeMux(obj, &d1, &d0)
		data[Gia_Regular(ctrl).id]--
		if Gia_Regular(d0) == Gia_Regular(d1) {
			data[Gia_Regular(d0).id]--
		}
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		data[Gia_ObjFanin0(Gia_ManObj(g, id)).id]++
	}
	for i := range data {
		data[i] = max(data[i], 1)
	}
}
func Kf_ManAlloc(g *Gia_Man_t, pars *Jf_Par_t) *Kf_Man_t {
	if pars.NLutSize < 0 || pars.NLutSize > KF_LEAF_MAX || pars.NCutNum < 0 || pars.NCutNum > KF_CUT_MAX || pars.NProcNum < 0 || pars.NProcNum > KF_PROC_MAX {
		panic("KF invalid mapping limits")
	}
	g.vMapping = nil
	p := &Kf_Man_t{pGia: g, pPars: pars, clkStart: time.Now()}
	n := int(g.nObjs)
	vec.Vec_IntFill(&p.vCuts, n, 0)
	vec.Vec_IntFill(&p.vTime, n, 0)
	vec.Vec_FltFill(&p.vArea, n, 0)
	Kf_ManSetInitRefs(g, &p.vRefs)
	p.vTemp = vec.Vec_IntAlloc(1000)
	g.Refs = make([]int32, n)
	for i := int32(0); i < max(1, pars.NProcNum); i++ {
		s := &p.pSett[i]
		s.pMan = p
		s.nLutSize = uint16(pars.NLutSize)
		s.nCutNum = uint16(pars.NCutNum)
		s.TableMask = (1 << KF_LOG_TABLE) - 1
	}
	return p
}
func Kf_ManFree(p *Kf_Man_t) {
	if p == nil {
		return
	}
	p.pGia.Refs = nil
	p.vCuts = Vec_Int_t{}
	p.vTime = Vec_Int_t{}
	p.vArea = Vec_Flt_t{}
	p.vRefs = Vec_Flt_t{}
	p.vTemp = nil
	p.pMem.reset()
}
func Kf_ManComputeDelay(p *Kf_Man_t, evaluate int32) int32 {
	if evaluate != 0 {
		for _, obj := range p.pGia.pObjs {
			if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(p.pGia, obj) > 0 {
				vec.Vec_IntWriteEntry(&p.vTime, int(obj.id), Kf_CutTime(p, Kf_ObjCutBest(p, obj.id)))
			}
		}
	}
	var delay int32
	for _, id := range vec.Vec_IntArray(p.pGia.vCos) {
		driver := Gia_ObjFanin0(Gia_ManObj(p.pGia, id))
		if Gia_ObjRefNum(p.pGia, driver) <= 0 {
			panic("KF output driver is not referenced")
		}
		delay = max(delay, Kf_ObjTime(p, driver.id))
	}
	return delay
}
func Kf_ManComputeRefs(p *Kf_Man_t) int32 {
	g := p.pGia
	if g.Refs == nil {
		panic("KF missing references")
	}
	clear(g.Refs)
	p.pPars.Area = 0
	p.pPars.Edge = 0
	for i := len(g.pObjs) - 1; i >= 0; i-- {
		obj := g.pObjs[i]
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			Gia_ObjRefInc(g, Gia_ObjFanin0(obj))
		} else if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNum(g, obj) > 0 {
			cut := Kf_ObjCutBest(p, obj.id)
			Kf_CutRef(p, cut)
			p.pPars.Edge += word(Kf_CutSize(cut))
			p.pPars.Area++
		}
	}
	refs, flow := vec.Vec_FltArray(&p.vRefs), vec.Vec_FltArray(&p.vArea)
	for i, old := range refs {
		var n float32
		if p.pPars.FOptEdge != 0 {
			n = max(1, 0.8*old+0.2*float32(g.Refs[i]))
		} else {
			n = max(1, 0.2*old+0.8*float32(g.Refs[i]))
		}
		flow[i] *= old / n
		refs[i] = n
		if flow[i] < 0 {
			panic("KF negative flow")
		}
	}
	p.pPars.Delay = word(Kf_ManComputeDelay(p, 1))
	return int32(p.pPars.Area)
}
