package gia

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/misc/vec"
)

type Gia_Sim2_t struct {
	pAig                 *Gia_Man_t
	pPars                *Gia_ParSim_t
	nWords               int32
	pDataSim             []uint32
	vClassOld, vClassNew *Vec_Int_t
}

func Gia_Sim2Data(p *Gia_Sim2_t, id int32) []uint32 {
	start := int(id * p.nWords)
	return p.pDataSim[start : start+int(p.nWords)]
}
func Gia_Sim2Delete(p *Gia_Sim2_t) { p.vClassOld = nil; p.vClassNew = nil; p.pDataSim = nil }
func Gia_Sim2Create(g *Gia_Man_t, pars *Gia_ParSim_t) *Gia_Sim2_t {
	if g == nil || pars == nil || pars.NWords <= 0 {
		panic("gia: invalid equivalence simulator")
	}
	p := &Gia_Sim2_t{pAig: g, pPars: pars, nWords: pars.NWords, pDataSim: make([]uint32, int(pars.NWords)*int(g.nObjs)), vClassOld: vec.Vec_IntAlloc(100), vClassNew: vec.Vec_IntAlloc(100)}
	if pars.FVerbose != 0 {
		fmt.Printf("Memory: AIG = %7.2f MB.  SimInfo = %7.2f MB.\n", 12*float64(g.nObjs)/(1<<20), 4*float64(pars.NWords)*float64(g.nObjs)/(1<<20))
	}
	Gia_ManSetPhase(g)
	for _, obj := range g.pObjs {
		obj.Value = uint32(obj.id)
	}
	return p
}
func Gia_Sim2InfoRandom(p *Gia_Sim2_t, data []uint32) {
	for i := p.nWords - 1; i >= 0; i-- {
		data[i] = Gia_ManRandom(0)
	}
}
func Gia_Sim2InfoZero(p *Gia_Sim2_t, data []uint32) { clear(data[:p.nWords]) }
func Gia_Sim2InfoOne(p *Gia_Sim2_t, data []uint32) {
	for i := int32(0); i < p.nWords; i++ {
		data[i] = ^uint32(0)
	}
}
func Gia_Sim2InfoCopy(p *Gia_Sim2_t, dst, src []uint32) { copy(dst[:p.nWords], src[:p.nWords]) }
func Gia_Sim2SimulateCo(p *Gia_Sim2_t, obj *Gia_Obj_t) {
	dst, src := Gia_Sim2Data(p, obj.id), Gia_Sim2Data(p, Gia_ObjFaninId0(obj, obj.id))
	for i, v := range src {
		if obj.fCompl0 != 0 {
			v = ^v
		}
		dst[i] = v
	}
}
func Gia_Sim2SimulateNode(p *Gia_Sim2_t, obj *Gia_Obj_t) {
	dst, a, b := Gia_Sim2Data(p, obj.id), Gia_Sim2Data(p, Gia_ObjFaninId0(obj, obj.id)), Gia_Sim2Data(p, Gia_ObjFaninId1(obj, obj.id))
	for i := range dst {
		x, y := a[i], b[i]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		dst[i] = x & y
	}
}
func Gia_Sim2InfoTransfer(p *Gia_Sim2_t) {
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		Gia_Sim2InfoCopy(p, Gia_Sim2Data(p, Gia_ManRo(p.pAig, i).id), Gia_Sim2Data(p, Gia_ManRi(p.pAig, i).id))
	}
}
func Gia_Sim2SimulateRound(p *Gia_Sim2_t) {
	Gia_Sim2InfoZero(p, Gia_Sim2Data(p, 0))
	for i := int32(0); i < Gia_ManPiNum(p.pAig); i++ {
		Gia_Sim2InfoRandom(p, Gia_Sim2Data(p, Gia_ManPi(p.pAig, i).id))
	}
	for _, obj := range p.pAig.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_Sim2SimulateNode(p, obj)
		}
	}
	for _, id := range vec.Vec_IntArray(p.pAig.vCos) {
		Gia_Sim2SimulateCo(p, Gia_ManObj(p.pAig, id))
	}
}
func Gia_Sim2CompareEqual(a, b []uint32, n, compl int32) int32 {
	for i := int32(0); i < n; i++ {
		v := b[i]
		if compl != 0 {
			v = ^v
		}
		if a[i] != v {
			return 0
		}
	}
	return 1
}
func Gia_Sim2CompareZero(a []uint32, n, compl int32) int32 {
	want := uint32(0)
	if compl != 0 {
		want = ^want
	}
	for i := int32(0); i < n; i++ {
		if a[i] != want {
			return 0
		}
	}
	return 1
}
func Gia_Sim2InfoIsZero(p *Gia_Sim2_t, data []uint32) int32 {
	for i, v := range data[:p.nWords] {
		if v != 0 {
			return int32(32*i + bits.TrailingZeros32(v))
		}
	}
	return -1
}
func Gia_Sim2CheckPos(p *Gia_Sim2_t, po, pat *int32) int32 {
	for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
		if bit := Gia_Sim2InfoIsZero(p, Gia_Sim2Data(p, Gia_ManPo(p.pAig, i).id)); bit >= 0 {
			*po, *pat = i, bit
			return 1
		}
	}
	return 0
}
