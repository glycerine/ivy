package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func cecInfo(info *Vec_Ptr_t, index int32) []uint32 {
	return vec.Vec_PtrEntry(info, int(index)).([]uint32)
}
func cecBit(data []uint32, bit int32) uint32 { return data[bit>>5] >> uint(bit&31) & 1 }
func cecWriteBit(data []uint32, bit int32, value uint32) {
	mask := uint32(1) << uint(bit&31)
	data[bit>>5] = data[bit>>5]&^mask | value<<uint(bit&31)
}

func Cec_ManSeqDeriveInfoFromCex(info *Vec_Ptr_t, p *Gia_Man_t, cex *Abc_Cex_t) {
	if cex.NBits != cex.NRegs+cex.NPis*(cex.IFrame+1) || int(cex.NBits-cex.NRegs+gia.Gia_ManRegNum(p)) > vec.Vec_PtrSize(info) {
		panic("cec: malformed sequential counterexample")
	}
	for bit := int32(0); bit < cex.NRegs; bit++ {
		if cecBit(cex.PData, bit) != 0 {
			fmt.Println("The CEX has flop values different from 0, but they are currently not used by resim.")
			break
		}
	}
	k := gia.Gia_ManRegNum(p)
	for i := int32(0); i < k; i++ {
		clear(cecInfo(info, i))
	}
	for i := cex.NRegs; i < cex.NBits; i++ {
		words := cecInfo(info, k)
		k++
		for w := range words {
			words[w] = gia.Gia_ManRandom(0)
		}
		words[0] = ((words[0] << 1) | cecBit(cex.PData, i)) << 1
	}
	for ; int(k) < vec.Vec_PtrSize(info); k++ {
		for w := range cecInfo(info, k) {
			cecInfo(info, k)[w] = gia.Gia_ManRandom(0)
		}
	}
}
func Cec_ManSeqDeriveInfoInitRandom(info *Vec_Ptr_t, p *Gia_Man_t, cex *Abc_Cex_t) {
	regs := gia.Gia_ManRegNum(p)
	if int(regs) > vec.Vec_PtrSize(info) || cex != nil && cex.NRegs != regs {
		panic("cec: initial-state width mismatch")
	}
	for k := int32(0); int(k) < vec.Vec_PtrSize(info); k++ {
		words := cecInfo(info, k)
		for w := range words {
			if k < regs {
				words[w] = 0
				if cex != nil && cecBit(cex.PData, k) != 0 {
					words[w] = ^uint32(0)
				}
			} else {
				words[w] = gia.Gia_ManRandom(0)
			}
		}
	}
}
func Cec_ManSeqResimulate(p *Cec_ManSim_t, info *Vec_Ptr_t) int32 {
	regs, pis, pos := gia.Gia_ManRegNum(p.pAig), gia.Gia_ManPiNum(p.pAig), gia.Gia_ManPoNum(p.pAig)
	if vec.Vec_PtrSize(info) != int(regs+pis*p.pPars.nFrames) {
		panic("cec: sequential simulation frame layout mismatch")
	}
	for k := int32(0); k < regs; k++ {
		copy(cecInfo(p.vCoSimInfo, pos+k), cecInfo(info, k))
	}
	k := regs
	for frame := int32(0); frame < p.pPars.nFrames; frame++ {
		if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			return 0
		}
		for i := int32(0); i < pis; i++ {
			copy(cecInfo(p.vCiSimInfo, i), cecInfo(info, k))
			k++
		}
		for i := int32(0); i < regs; i++ {
			copy(cecInfo(p.vCiSimInfo, pis+i), cecInfo(p.vCoSimInfo, pos+i))
		}
		if Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo) != 0 {
			return 1
		}
		if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			return 0
		}
	}
	return 0
}
func Cec_ManSeqResimulateInfo(p *Gia_Man_t, info *Vec_Ptr_t, state *Abc_Cex_t, check int32) int32 {
	regs, pis := gia.Gia_ManRegNum(p), gia.Gia_ManPiNum(p)
	remaining := int32(vec.Vec_PtrSize(info)) - regs
	if remaining < 0 || pis == 0 || remaining%pis != 0 {
		panic("cec: invalid sequential simulation layout")
	}
	var pars Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&pars)
	pars.nFrames = remaining / pis
	pars.nWords = int32(vec.Vec_PtrReadWordsSimInfo(info))
	pars.fCheckMiter = check
	cecSimRefs(p)
	sim := Cec_ManSimStart(p, &pars)
	sim.pBestState = state
	result := Cec_ManSeqResimulate(sim, info)
	sim.pBestState = nil
	Cec_ManSimStop(sim)
	return result
}
func Cec_ManSeqResimulateCounter(p *Gia_Man_t, pars *Cec_ParSim_t, cex *Abc_Cex_t) int32 {
	if cex == nil || !p.HasRepresentatives() || gia.Gia_ManRegNum(p) == 0 || gia.Gia_ManPiNum(p) != cex.NPis {
		return -1
	}
	gia.Gia_ManRandom(1)
	info := vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManRegNum(p)+gia.Gia_ManPiNum(p)*(pars.nFrames+cex.IFrame+1)), 1)
	Cec_ManSeqDeriveInfoFromCex(info, p, cex)
	return Cec_ManSeqResimulateInfo(p, info, nil, pars.fCheckMiter)
}
func Cec_ManCountNonConstOutputs(p *Gia_Man_t) int32 {
	if !p.HasRepresentatives() {
		return -1
	}
	count := int32(0)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		if gia.Gia_ObjIsConst(p, gia.Gia_ObjFaninId0p(p, gia.Gia_ManPo(p, i))) == 0 {
			count++
		}
	}
	return count
}
func Cec_ManCheckNonTrivialCands(p *Gia_Man_t) int32 {
	if !p.HasRepresentatives() {
		return 0
	}
	drivers := make(map[int32]bool)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		drivers[gia.Gia_ObjFaninId0p(p, gia.Gia_ManPo(p, i))] = true
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCand(o) != 0 && !drivers[id] && gia.Gia_ObjRepr(p, id) != gia.GIA_VOID {
			return 1
		}
	}
	return 0
}

func Cec_ManSeqSemiformal(p *Gia_Man_t, pars *Cec_ParSmf_t) int32 {
	if !p.HasRepresentatives() || gia.Gia_ManRegNum(p) == 0 {
		return -1
	}
	gia.Gia_ManRandom(1)
	state := util.Abc_CexAlloc(gia.Gia_ManRegNum(p), 0, 0)
	state.IFrame = -1
	state.IPo = -1
	var sat Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&sat)
	sat.nBTLimit = pars.nBTLimit
	sat.fVerbose = pars.fVerbose
	noRefinements, result := int32(0), int32(0)
	for round := int32(0); round < pars.nRounds; round++ {
		if Cec_ManCheckNonTrivialCands(p) == 0 {
			break
		}
		var outputs *Vec_Int_t
		var model *Gia_Man_t
		target := pars.nMinOutputs
		if target == 0 {
			for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
				if gia.Gia_ObjHasRepr(p, id) != 0 {
					target++
				}
			}
			target = target/2 + 1
		}
		frames := int32(1)
		for {
			model = cecCorrUnrollInit(p, frames, 0, 0, &outputs, pars.fDualOut, state)
			if gia.Gia_ManCoNum(model) >= target || frames >= pars.nFrames {
				break
			}
			if gia.Gia_ManAndNum(model) > 500000 {
				return result
			}
			frames++
		}
		info := vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManRegNum(p)+gia.Gia_ManPiNum(p)*(frames+16)), int(pars.nWords))
		Cec_ManSeqDeriveInfoInitRandom(info, p, state)
		var patterns int32
		Cec_ManSatSolveSeq(info, model, &sat, gia.Gia_ManRegNum(p), &patterns)

		result = Cec_ManSeqResimulateInfo(p, info, state, pars.fCheckMiter)
		state.IPo = -1
		if result != 0 {
			break
		}
		if patterns > 0 {
			noRefinements = 0
		} else {
			noRefinements++
			if noRefinements == pars.nNonRefines {
				break
			}
		}
	}
	return result
}
