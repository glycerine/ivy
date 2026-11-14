package if_

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

func (p *If_Par_t_) LUTSize() int32          { return p.nLutSize }
func (p *If_Par_t_) SetLUTSize(n int32)      { p.nLutSize = n }
func (p *If_Par_t_) GateSize() int32         { return p.nGateSize }
func (p *If_Par_t_) UsesBdds() bool          { return p.fUseBdds != 0 }
func (p *If_Par_t_) UsesCnfs() bool          { return p.fUseCnfs != 0 }
func (p *If_Par_t_) UsesMv() bool            { return p.fUseMv != 0 }
func (p *If_Par_t_) UsesSops() bool          { return p.fUseSops != 0 }
func (p *If_Par_t_) UsesUserSesLib() bool    { return p.fUserSesLib != 0 }
func (p *If_Par_t_) UsesBuffs() bool         { return p.fUseBuffs != 0 }
func (p *If_Par_t_) UsesTtPerm() bool        { return p.fUseTtPerm != 0 }
func (p *If_Par_t_) UsesDsd() bool           { return p.fUseDsd != 0 }
func (p *If_Par_t_) LUTStructure() string    { return p.pLutStruct }
func (p *If_Par_t_) DelayOptimization() bool { return p.fDelayOpt != 0 }
func (p *If_Par_t_) DsdBalance() bool        { return p.fDsdBalance != 0 }
func (p *If_Par_t_) UserRecLibrary() bool    { return p.fUserRecLib != 0 }
func (p *If_Par_t_) LatchPaths() bool        { return p.fLatchPaths != 0 }
func (p *If_Par_t_) TruthMode() bool         { return p.fTruth != 0 }
func (p *If_Par_t_) Verbose() bool           { return p.fVerbose != 0 }
func (p *If_Par_t_) SetCutMin(value bool) {
	p.fCutMin = 0
	if value {
		p.fCutMin = 1
	}
}
func (p *If_Par_t_) Bidec() bool { return p.fBidec != 0 }
func (p *If_Par_t_) SetArrivalTimes(arrivals []float32) {
	p.pTimesArr = append([]float32(nil), arrivals...)
}

// SetCutCostFunction installs the source-compatible per-cut cost callback.
func (p *If_Par_t_) SetCutCostFunction(cost func(*If_Man_t, *If_Cut_t) int32) {
	p.pFuncCost = cost
}

// SetRenodeConfig applies the complete option block assembled by ABC's
// Abc_NtkRenode before invoking the mapper.
func (p *If_Par_t_) SetRenodeConfig(lutSize, cutsMax, flowIters, areaIters, area, bdds, sops, cnfs, mv, verbose int32, reo any, cost func(*If_Man_t, *If_Cut_t) int32) {
	*p = If_Par_t_{
		nLutSize: lutSize, nCutsMax: cutsMax, nFlowIters: flowIters, nAreaIters: areaIters,
		DelayTarget: -1, Epsilon: 0.005, fPreprocess: 1, fArea: area, fVerbose: verbose,
		fTruth: 1, fUsePerm: 1, fUseBdds: bdds, fUseSops: sops, fUseCnfs: cnfs, fUseMv: mv,
		pFuncCost: cost, pReoMan: reo,
	}
}
func (p *If_Par_t_) ArrivalTimes() []float32 { return append([]float32(nil), p.pTimesArr...) }
func (p *If_Par_t_) SetRequiredTimes(required []float32) {
	p.pTimesReq = append([]float32(nil), required...)
}
func (p *If_Par_t_) RequiredTimes() []float32               { return append([]float32(nil), p.pTimesReq...) }
func (p *If_Man_t_) Parameters() *If_Par_t                  { return p.pPars }
func (p *If_Man_t_) TimingManager() *tim.Tim_Man_t          { return p.pManTim }
func (p *If_Man_t_) SetTimingManager(timing *tim.Tim_Man_t) { p.pManTim = timing }
func (p *If_Man_t_) SetDsdManager(manager *If_DsdMan_t)     { p.pIfDsdMan = manager }

func (o *If_Obj_t_) Equivalent() *If_Obj_t { return o.pEquiv }
func (o *If_Obj_t_) Phase() bool           { return o.fPhase != 0 }
func If_ObjSetSkipCut(o *If_Obj_t, skip bool) {
	o.fSkipCut = 0
	if skip {
		o.fSkipCut = 1
	}
}

func (p *If_Par_t_) Power() bool { return p.fPower != 0 }
func (p *If_Man_t_) SetSwitchingProbabilities(probabilities []float32) {
	p.vSwitching = vec.Vec_IntAlloc(len(probabilities))
	for _, probability := range probabilities {
		vec.Vec_IntPush(p.vSwitching, int32(math.Float32bits(probability)))
	}
}
func (p *If_Man_t_) HasSwitchingVector() bool             { return p.vSwitching != nil }
func (p *If_Man_t_) SetSwitchingVector(vector *Vec_Int_t) { p.vSwitching = vector }
