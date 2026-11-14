package lpk

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lpk_FunAlloc(nVars int32) *Lpk_Fun_t {
	if nVars < 0 || nVars > 16 {
		return nil
	}
	words := int(kit.Kit_TruthWordNum(nVars))
	return &Lpk_Fun_t{truth: make([]uint32, words*3)}
}

func Lpk_FunFree(p *Lpk_Fun_t) {
	if p == nil {
		return
	}
	if p.vNodes != nil && int(p.Id) < vec.Vec_PtrSize(p.vNodes) && vec.Vec_PtrEntry(p.vNodes, int(p.Id)) == p {
		vec.Vec_PtrWriteEntry(p.vNodes, int(p.Id), nil)
	}
	p.truth = nil
	p.vNodes = nil
}

func Lpk_FunCreate(pNtk *Abc_Ntk_t, vLeaves *Vec_Ptr_t, truth []uint32, nLutK, areaLim, delayLim int32) *Lpk_Fun_t {
	if pNtk == nil || vLeaves == nil {
		return nil
	}
	nVars := int32(vec.Vec_PtrSize(vLeaves))
	if nVars > 16 || len(truth) < int(kit.Kit_TruthWordNum(nVars)) {
		return nil
	}
	p := Lpk_FunAlloc(nVars)
	p.Id = uint32(nVars)
	p.vNodes = vLeaves
	p.nVars = uint32(nVars)
	p.nLutK = uint32(nLutK)
	p.nAreaLim = uint32(areaLim)
	p.nDelayLim = uint32(delayLim)
	p.uSupp = kit.Kit_TruthSupport(truth, nVars)
	kit.Kit_TruthCopy(Lpk_FunTruth(p, 0), truth, nVars)
	for i := 0; i < int(nVars); i++ {
		p.pFanins[i] = byte(i)
		node, ok := vec.Vec_PtrEntry(vLeaves, i).(*abc.Object)
		if !ok {
			return nil
		}
		p.pDelays[i] = int32(node.Level)
	}
	vec.Vec_PtrPush(vLeaves, p)
	return p
}

func Lpk_FunDup(p *Lpk_Fun_t, truth []uint32) *Lpk_Fun_t {
	if p == nil || len(truth) < int(kit.Kit_TruthWordNum(int32(p.nVars))) {
		return nil
	}
	copyFun := Lpk_FunAlloc(int32(p.nVars))
	if copyFun == nil {
		return nil
	}
	copyFun.Id = uint32(vec.Vec_PtrSize(p.vNodes))
	copyFun.vNodes = p.vNodes
	copyFun.nVars = p.nVars
	copyFun.nLutK = p.nLutK
	copyFun.nAreaLim = p.nAreaLim
	copyFun.nDelayLim = p.nDelayLim
	copyFun.uSupp = kit.Kit_TruthSupport(truth, int32(p.nVars))
	kit.Kit_TruthCopy(Lpk_FunTruth(copyFun, 0), truth, int32(p.nVars))
	copy(copyFun.pFanins[:], p.pFanins[:])
	copy(copyFun.pDelays[:], p.pDelays[:])
	vec.Vec_PtrPush(p.vNodes, copyFun)
	return copyFun
}

func Lpk_FunSuppMinimize(p *Lpk_Fun_t) int32 {
	if p == nil {
		return 0
	}
	if p.uSupp == kit.Kit_BitMask(int32(p.nVars)) {
		return 0
	}
	p.fSupports = 0
	nVarsNew := kit.Kit_WordCountOnes(p.uSupp)
	kit.Kit_TruthShrink(Lpk_FunTruth(p, 1), Lpk_FunTruth(p, 0), nVarsNew, int32(p.nVars), p.uSupp, 1)
	k := 0
	for i := 0; i < int(p.nVars); i++ {
		if p.uSupp&(1<<uint(i)) != 0 {
			p.pFanins[k] = p.pFanins[i]
			p.pDelays[k] = p.pDelays[i]
			k++
		}
	}
	if int32(k) != nVarsNew {
		panic("lpk: support compaction count mismatch")
	}
	p.nVars = uint32(k)
	p.uSupp = kit.Kit_BitMask(int32(p.nVars))
	return 1
}

func Lpk_FunComputeCofSupps(p *Lpk_Fun_t) {
	if p == nil {
		return
	}
	assert(p.fSupports == 0, "p.fSupports == 0")
	truth, cof0, cof1 := Lpk_FunTruth(p, 0), Lpk_FunTruth(p, 1), Lpk_FunTruth(p, 2)
	for variable := int32(0); variable < int32(p.nVars); variable++ {
		kit.Kit_TruthCofactor0New(cof0, truth, int32(p.nVars), variable)
		kit.Kit_TruthCofactor1New(cof1, truth, int32(p.nVars), variable)
		p.puSupps[2*variable] = kit.Kit_TruthSupport(cof0, int32(p.nVars))
		p.puSupps[2*variable+1] = kit.Kit_TruthSupport(cof1, int32(p.nVars))
	}
	p.fSupports = 1
}

func Lpk_SuppDelay(support uint32, delays []int32) int32 {
	var max int32
	for v := 0; v < 16; v++ {
		if support&(1<<uint(v)) != 0 && v < len(delays) && delays[v] > max {
			max = delays[v]
		}
	}
	return max + 1
}

func Lpk_SuppToVars(boundSet uint32, vars []byte) int32 {
	n := 0
	for v := 0; v < 16; v++ {
		if boundSet&(1<<uint(v)) != 0 {
			if n >= len(vars) {
				return int32(n)
			}
			vars[n] = byte(v)
			n++
		}
	}
	return int32(n)
}

var _ = util.Abc_MaxInt
var _ = Lpk_LutNumVars
