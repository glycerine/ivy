package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

func Mf_ObjComputeBestCut(p *Mf_Man_t, id int32) {
	best := Mf_ManObj(p, id)
	set := Mf_ObjCutSet(p, id)
	if p.fUseEla != 0 && best.nMapRefs != 0 {
		Mf_CutDeref_rec(p, set[1:])
	}
	bestOffset := -1
	bestFlow := float32(math.MaxFloat32)
	bestTime := int32(math.MaxInt32)
	var time int32
	offset := 1
	for i := int32(0); i < set[0]; i++ {
		cut := set[offset:]
		size := Mf_CutSize(cut)
		if Mf_CutIsTriv(cut, id) != 0 || size > p.pPars.NLutSize {
			panic("gia: invalid mapping cut")
		}
		var flow float32
		if p.fUseEla != 0 {
			flow = float32(Mf_CutAreaDerefed2(p, cut))
		} else {
			flow = Mf_CutFlow(p, cut, &time)
		}
		if bestOffset < 0 || bestFlow > flow+MF_EPSILON || (bestFlow > flow-MF_EPSILON && bestTime > time) {
			bestOffset, bestFlow, bestTime = offset, flow, time
		}
		offset += int(size) + 1
	}
	if bestOffset < 0 {
		panic("gia: mapping object has no cuts")
	}
	if p.fUseEla != 0 && best.nMapRefs != 0 {
		Mf_CutRef_rec(p, set[bestOffset:])
	} else {
		best.nMapRefs = 0
	}
	if p.fUseEla != 0 {
		Mf_CutFlow(p, set[bestOffset:], &bestTime)
	}
	best.Delay = uint32(bestTime) & 0xffff
	best.Flow = bestFlow / best.nFlowRefs
	Mf_ObjSetBestCut(set, bestOffset)
}

func Mf_ManMappingFromMapping(p *Mf_Man_t) int32 {
	g := p.pGia0
	if g.VMapping != nil {
		panic("gia: mapping already exists")
	}
	mapping := vec.Vec_IntStart(int(g.nObjs))
	var count int32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		vec.Vec_IntWriteEntry(mapping, int(obj.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, Mf_CutSize(cut))
		for _, leaf := range cut[1 : 1+Mf_CutSize(cut)] {
			vec.Vec_IntPush(mapping, leaf)
		}
		vec.Vec_IntPush(mapping, obj.id)
		count++
	}
	g.VMapping = mapping
	fmt.Printf("Mapping is %.2fx larger than AIG manager.\n", float64(vec.Vec_IntSize(mapping))/float64(g.nObjs))
	return count
}
func Mf_ManDeriveMapping(p *Mf_Man_t) *Gia_Man_t {
	if p.pPars.FCutMin != 0 || p.pGia.VMapping != nil {
		panic("gia: invalid mapping derivation")
	}
	mapping := vec.Vec_IntStart(int(p.pGia.nObjs))
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		vec.Vec_IntWriteEntry(mapping, int(obj.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, Mf_CutSize(cut))
		for _, leaf := range cut[1 : 1+Mf_CutSize(cut)] {
			vec.Vec_IntPush(mapping, leaf)
		}
		vec.Vec_IntPush(mapping, obj.id)
	}
	p.pGia.VMapping = mapping
	return p.pGia
}
