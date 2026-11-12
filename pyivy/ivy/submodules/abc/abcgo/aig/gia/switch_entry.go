package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math"
)

func Saig_ManComputeSwitchProbs(graph *aig.Graph, frames, prefix, probOne int32) *Vec_Int_t {
	converted := Gia_ManFromAigSwitch(graph)
	values := Gia_ManComputeSwitchProbs(converted, frames, prefix, probOne)
	result := Vec_IntStart(int(aig.Aig_ManObjNumMax(graph)))
	for i := int32(0); i < aig.Aig_ManObjNumMax(graph); i++ {
		obj := aig.Edge(i) << 1
		if aig.Aig_ObjIsNone(graph, obj) {
			continue
		}
		copy := aig.Aig_ObjCopy(graph, obj)
		if copy == aig.InvalidEdge {
			panic("gia: missing AIG switching copy")
		}
		Vec_IntWriteEntry(result, int(i), Vec_IntEntry(values, copy.NodeID()))
	}
	Gia_ManStop(converted)
	return result
}
func Gia_ManEvaluateSwitching(p *Gia_Man_t) float32 {
	if len(p.pSwitching) != int(Gia_ManObjNum(p)) {
		panic("gia: missing switching data")
	}
	p.Refs = nil
	Gia_ManCreateRefs(p)
	var total float32
	for _, obj := range p.pObjs {
		total += float32(Gia_ObjRefNum(p, obj)) * float32(p.pSwitching[obj.id]) / 255
	}
	return total
}
func Gia_ManPrintOutputProb(p *Gia_Man_t) *Vec_Flt_t {
	copy := Gia_ManDup(p)
	frames := int32(1)
	if Gia_ManRegNum(p) != 0 {
		frames = 16
	}
	values := Gia_ManComputeSwitchProbs(copy, frames, 0, 1)
	result := Vec_FltStart(Vec_IntSize(values))
	for i, bits := range Vec_IntArray(values) {
		Vec_FltWriteEntry(result, i, math.Float32frombits(uint32(bits)))
	}
	Gia_ManStop(copy)
	return result
}
