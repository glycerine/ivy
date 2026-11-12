package gia

import (
	"fmt"
	"time"
)

func gliFlip(p *Gli_Man_t, o *Gli_Obj_t) {
	Vec_IntPush(p.vCisChanged, o.Handle)
	o.fPhase ^= 1
	o.fPhase2 ^= 1
	o.nSwitches++
	o.nGlitches++
}
func gliRandomTransition(prob float32) bool { return float32(Gia_ManRandom(0)&0xffff)/65536 < prob }
func Gli_ManSetPiRandom(p *Gli_Man_t, prob float32) {
	if prob <= 0 || prob >= 1 {
		panic("glitch: invalid transition probability")
	}
	Vec_IntClear(p.vCisChanged)
	for i := int32(0); i < Gli_ManCiNum(p); i++ {
		o := Gli_ManCi(p, i)
		if gliRandomTransition(prob) {
			gliFlip(p, o)
		}
	}
}
func Gli_ManSetPiFromSaved(p *Gli_Man_t, bit int32) {
	if bit < 0 || bit >= 32 || len(p.pSimInfoPrev) != int(Gli_ManCiNum(p)) {
		panic("glitch: invalid saved simulation")
	}
	Vec_IntClear(p.vCisChanged)
	for i := int32(0); i < Gli_ManCiNum(p); i++ {
		o := Gli_ManCi(p, i)
		if (p.pSimInfoPrev[i]^o.uSimInfo)&(uint32(1)<<uint(bit)) != 0 {
			gliFlip(p, o)
		}
	}
}
func Gli_ManSwitching(p *Gli_Man_t) {
	for _, o := range p.objects {
		if Gli_ObjIsNode(o) != 0 && int32(o.fPhase) != Gli_NodeComputeValue(o) {
			o.fPhase ^= 1
			o.nSwitches++
		}
	}
}
func Gli_ManGlitching(p *Gli_Man_t) {
	Vec_IntClear(p.vAffected)
	for _, id := range Vec_IntArray(p.vCisChanged) {
		Vec_IntPush(p.vAffected, id)
	}
	for Vec_IntSize(p.vAffected) > 0 {
		Vec_IntClear(p.vFrontier)
		for _, id := range Vec_IntArray(p.vAffected) {
			for _, fanout := range Gli_ManObj(p, id).fanouts {
				if Gli_ObjIsCo(fanout) != 0 || fanout.fMark != 0 {
					continue
				}
				fanout.fMark = 1
				Vec_IntPush(p.vFrontier, fanout.Handle)
			}
		}
		Vec_IntClear(p.vAffected)
		for _, id := range Vec_IntArray(p.vFrontier) {
			o := Gli_ManObj(p, id)
			o.fMark = 0
			if int32(o.fPhase2) == Gli_NodeComputeValue2(o) {
				continue
			}
			o.fPhase2 ^= 1
			o.nGlitches++
			Vec_IntPush(p.vAffected, id)
		}
	}
}
func Gli_ManVerify(p *Gli_Man_t) {
	for _, o := range p.objects {
		if o.fPhase != o.fPhase2 || o.nGlitches < o.nSwitches {
			panic("glitch: inconsistent switching state")
		}
	}
}
func Gli_ManSimulateSeqNode(_ *Gli_Man_t, o *Gli_Obj_t) uint32 {
	var result uint32
	for bit := uint(0); bit < 32; bit++ {
		var phase uint32
		for i, fanin := range o.fanins {
			phase |= ((fanin.uSimInfo >> bit) & 1) << uint(i)
		}
		if int(phase/64) >= len(o.pTruth) {
			panic("glitch: short truth table")
		}
		result |= uint32(o.pTruth[phase/64]>>uint(phase%64)&1) << bit
	}
	return result
}
func Gli_ManUpdateRandomInput(info uint32, prob float32) uint32 {
	if prob == 0.5 {
		return Gia_ManRandom(0)
	}
	for i := uint(0); i < 32; i++ {
		if gliRandomTransition(prob) {
			info ^= uint32(1) << i
		}
	}
	return info
}
func Gli_ManSimulateSeqPref(p *Gli_Man_t, frames int32) {
	for i := int32(0); i < Gli_ManPiNum(p); i++ {
		o := Gli_ManPi(p, i)
		o.uSimInfo = Gli_ManUpdateRandomInput(o.uSimInfo, 0.5)
	}
	for i := int32(0); i < p.nRegs; i++ {
		Gli_ManRo(p, i).uSimInfo = 0
	}
	for frame := int32(0); frame < frames; frame++ {
		for _, o := range p.objects {
			if Gli_ObjIsNode(o) != 0 {
				o.uSimInfo = Gli_ManSimulateSeqNode(p, o)
			}
		}
		for i := int32(0); i < p.nRegs; i++ {
			o := Gli_ManRi(p, i)
			o.uSimInfo = o.fanins[0].uSimInfo
		}
		for i := int32(0); i < Gli_ManPiNum(p); i++ {
			o := Gli_ManPi(p, i)
			o.uSimInfo = Gli_ManUpdateRandomInput(o.uSimInfo, 0.5)
		}
		for i := int32(0); i < p.nRegs; i++ {
			Gli_ManRo(p, i).uSimInfo = Gli_ManRi(p, i).uSimInfo
		}
	}
	p.pSimInfoPrev = make([]uint32, Gli_ManCiNum(p))
	for i := range p.pSimInfoPrev {
		p.pSimInfoPrev[i] = Gli_ManCi(p, int32(i)).uSimInfo
	}
}
func Gli_ManSetDataSaved(p *Gli_Man_t, bit int32) {
	if bit < 0 || bit >= 32 || len(p.pSimInfoPrev) != int(Gli_ManCiNum(p)) {
		panic("glitch: invalid saved simulation")
	}
	for i := int32(0); i < Gli_ManCiNum(p); i++ {
		o := Gli_ManCi(p, i)
		o.fPhase = p.pSimInfoPrev[i] >> uint(bit) & 1
		o.fPhase2 = o.fPhase
	}
	for _, o := range p.objects {
		if Gli_ObjIsNode(o) != 0 {
			o.fPhase = uint32(Gli_NodeComputeValue(o))
			o.fPhase2 = o.fPhase
		}
	}
}
func Gli_ManSetPiRandomSeq(p *Gli_Man_t, prob float32) {
	if prob <= 0 || prob >= 1 {
		panic("glitch: invalid transition probability")
	}
	for i := int32(0); i < Gli_ManCoNum(p); i++ {
		o := Gli_ManCo(p, i)
		o.fPhase = o.fanins[0].fPhase
		o.fPhase2 = o.fPhase
	}
	Vec_IntClear(p.vCisChanged)
	for i := int32(0); i < Gli_ManPiNum(p); i++ {
		o := Gli_ManPi(p, i)
		if gliRandomTransition(prob) {
			gliFlip(p, o)
		}
	}
	for i := int32(0); i < p.nRegs; i++ {
		o := Gli_ManRo(p, i)
		if Gli_ManRi(p, i).fPhase != o.fPhase {
			gliFlip(p, o)
		}
	}
}
func Gli_ManSwitchesAndGlitches(p *Gli_Man_t, patterns int32, prob float32, verbose int32) {
	start := time.Now()
	Gia_ManRandom(1)
	Gli_ManFinalize(p)
	if p.nRegs == 0 {
		for i := int32(0); i < patterns; i++ {
			Gli_ManSetPiRandom(p, prob)
			Gli_ManSwitching(p)
			Gli_ManGlitching(p)
		}
	} else {
		iterations := (patterns + 31) / 32
		Gli_ManSimulateSeqPref(p, 16)
		for bit := int32(0); bit < 32; bit++ {
			Gli_ManSetDataSaved(p, bit)
			for i := int32(0); i < iterations; i++ {
				Gli_ManSetPiRandomSeq(p, prob)
				Gli_ManSwitching(p)
				Gli_ManGlitching(p)
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Simulated %d patterns. Input transition probability %.2f. Time %s\n", patterns, prob, time.Since(start))
	}
}
