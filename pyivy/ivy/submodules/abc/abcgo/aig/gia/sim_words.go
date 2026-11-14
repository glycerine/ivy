package gia

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_SimData(p *Gia_ManSim_t, i int32) []uint32 {
	start := int(i * p.nWords)
	return p.pDataSim[start : start+int(p.nWords)]
}
func Gia_SimDataCi(p *Gia_ManSim_t, i int32) []uint32 {
	start := int(i * p.nWords)
	return p.pDataSimCis[start : start+int(p.nWords)]
}
func Gia_SimDataCo(p *Gia_ManSim_t, i int32) []uint32 {
	start := int(i * p.nWords)
	return p.pDataSimCos[start : start+int(p.nWords)]
}
func Gia_SimDataExt(p *Gia_ManSim_t, i int32) []uint32   { return Gia_SimData(p, i) }
func Gia_SimDataCiExt(p *Gia_ManSim_t, i int32) []uint32 { return Gia_SimDataCi(p, i) }
func Gia_SimDataCoExt(p *Gia_ManSim_t, i int32) []uint32 { return Gia_SimDataCo(p, i) }

func Gia_ManSimCreate(g *Gia_Man_t, pars *Gia_ParSim_t) *Gia_ManSim_t {
	if g == nil || pars == nil || pars.NWords <= 0 {
		panic("gia: invalid simulator parameters")
	}
	p := &Gia_ManSim_t{pPars: pars, nWords: pars.NWords}
	if pars.FVerbose != 0 {
		p.vConsts = Gia_ManSimDeriveResets(g)
	}
	p.pAig = Gia_ManFront(g)
	p.pDataSim = make([]uint32, int(p.nWords)*int(p.pAig.nFront))
	p.pDataSimCis = make([]uint32, int(p.nWords)*int(Gia_ManCiNum(p.pAig)))
	p.pDataSimCos = make([]uint32, int(p.nWords)*int(Gia_ManCoNum(p.pAig)))
	p.vCis2Ids = vec.Vec_IntAlloc(int(Gia_ManCiNum(g)))
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		vec.Vec_IntPush(p.vCis2Ids, i)
	}
	if pars.FVerbose != 0 {
		fmt.Printf("AIG = %7.2f MB.   Front mem = %7.2f MB.  Other mem = %7.2f MB.\n", 12.0*float64(p.pAig.nObjs)/(1<<20), 4.0*float64(p.nWords)*float64(p.pAig.nFront)/(1<<20), 4.0*float64(p.nWords)*float64(Gia_ManCiNum(p.pAig)+Gia_ManCoNum(p.pAig))/(1<<20))
	}
	return p
}
func Gia_ManSimDelete(p *Gia_ManSim_t) {
	if p == nil {
		return
	}
	if p.pAig != nil {
		Gia_ManStop(p.pAig)
	}
	p.pAig = nil
	p.vConsts = nil
	p.vCis2Ids = nil
	p.pDataSim = nil
	p.pDataSimCis = nil
	p.pDataSimCos = nil
}
func Gia_ManSimInfoRandom(p *Gia_ManSim_t, info []uint32) {
	for i := p.nWords - 1; i >= 0; i-- {
		info[i] = Gia_ManRandom(0)
	}
}
func Gia_ManSimInfoZero(p *Gia_ManSim_t, info []uint32) { clear(info[:p.nWords]) }
func Gia_ManSimInfoOne(p *Gia_ManSim_t, info []uint32) {
	for i := int32(0); i < p.nWords; i++ {
		info[i] = ^uint32(0)
	}
}
func Gia_ManSimInfoCopy(p *Gia_ManSim_t, dst, src []uint32) { copy(dst[:p.nWords], src[:p.nWords]) }
func Gia_ManSimInfoIsZero(p *Gia_ManSim_t, info []uint32) int32 {
	for i, v := range info[:p.nWords] {
		if v != 0 {
			return int32(32*i + bits.TrailingZeros32(v))
		}
	}
	return -1
}
func Gia_ManSimulateCi(p *Gia_ManSim_t, obj *Gia_Obj_t, i int32) {
	copy(Gia_SimData(p, Gia_ObjValue(obj)), Gia_SimDataCi(p, i))
}
func giaSimulateCo(p *Gia_ManSim_t, i int32, obj *Gia_Obj_t) {
	dst, src := Gia_SimDataCo(p, i), Gia_SimData(p, Gia_ObjDiff0(obj))
	for j, v := range src {
		if Gia_ObjFaninC0(obj) != 0 {
			v = ^v
		}
		dst[j] = v
	}
}
func giaSimulateNode(p *Gia_ManSim_t, obj *Gia_Obj_t) {
	dst, a, b := Gia_SimData(p, Gia_ObjValue(obj)), Gia_SimData(p, Gia_ObjDiff0(obj)), Gia_SimData(p, Gia_ObjDiff1(obj))
	for i := range dst {
		x, y := a[i], b[i]
		if Gia_ObjFaninC0(obj) != 0 {
			x = ^x
		}
		if Gia_ObjFaninC1(obj) != 0 {
			y = ^y
		}
		dst[i] = x & y
	}
}
func Gia_ManSimInfoInit(p *Gia_ManSim_t) {
	for i, id := range vec.Vec_IntArray(p.vCis2Ids) {
		if id < Gia_ManPiNum(p.pAig) {
			Gia_ManSimInfoRandom(p, Gia_SimDataCi(p, int32(i)))
		} else {
			Gia_ManSimInfoZero(p, Gia_SimDataCi(p, int32(i)))
		}
	}
}
func Gia_ManSimInfoTransfer(p *Gia_ManSim_t) {
	for i, id := range vec.Vec_IntArray(p.vCis2Ids) {
		if id < Gia_ManPiNum(p.pAig) {
			Gia_ManSimInfoRandom(p, Gia_SimDataCi(p, int32(i)))
		} else {
			Gia_ManSimInfoCopy(p, Gia_SimDataCi(p, int32(i)), Gia_SimDataCo(p, Gia_ManPoNum(p.pAig)+id-Gia_ManPiNum(p.pAig)))
		}
	}
}
func Gia_ManSimulateRound(p *Gia_ManSim_t) {
	if p.pAig.nFront <= 0 || Gia_ManConst0(p.pAig).Value != 0 {
		panic("gia: invalid simulation frontier")
	}
	Gia_ManSimInfoZero(p, Gia_SimData(p, 0))
	var ci, co int32
	for _, obj := range p.pAig.pObjs[1:] {
		if Gia_ObjIsAnd(obj) != 0 {
			giaSimulateNode(p, obj)
		} else if Gia_ObjIsCo(obj) != 0 {
			giaSimulateCo(p, co, obj)
			co++
		} else {
			Gia_ManSimulateCi(p, obj, ci)
			ci++
		}
	}
	if ci != Gia_ManCiNum(p.pAig) || co != Gia_ManCoNum(p.pAig) {
		panic("gia: simulation interface mismatch")
	}
}
func Gia_ManCheckPos(p *Gia_ManSim_t, po, pattern *int32) int32 {
	for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
		if pat := Gia_ManSimInfoIsZero(p, Gia_SimDataCo(p, i)); pat >= 0 {
			*po, *pattern = i, pat
			return 1
		}
	}
	return 0
}
