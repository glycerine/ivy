// Public library, gate and pin accessors from mioApi.c.
package mio

import "math"

func Mio_LibraryReadName(pLib *Mio_Library_t) string             { return pLib.pName }
func Mio_LibraryReadGateNum(pLib *Mio_Library_t) int32           { return pLib.nGates }
func Mio_LibraryReadGates(pLib *Mio_Library_t) *Mio_Gate_t       { return pLib.pGates }
func Mio_LibraryReadGateArray(pLib *Mio_Library_t) []*Mio_Gate_t { return pLib.ppGatesName }
func Mio_LibraryReadBuf(pLib *Mio_Library_t) *Mio_Gate_t         { return pLib.pGateBuf }
func Mio_LibraryReadInv(pLib *Mio_Library_t) *Mio_Gate_t         { return pLib.pGateInv }
func Mio_LibraryReadConst0(pLib *Mio_Library_t) *Mio_Gate_t      { return pLib.pGate0 }
func Mio_LibraryReadConst1(pLib *Mio_Library_t) *Mio_Gate_t      { return pLib.pGate1 }
func Mio_LibraryReadNand2(pLib *Mio_Library_t) *Mio_Gate_t       { return pLib.pGateNand2 }
func Mio_LibraryReadAnd2(pLib *Mio_Library_t) *Mio_Gate_t        { return pLib.pGateAnd2 }
func Mio_LibraryReadNor2(pLib *Mio_Library_t) *Mio_Gate_t        { return pLib.pGateNor2 }
func Mio_LibraryReadOr2(pLib *Mio_Library_t) *Mio_Gate_t         { return pLib.pGateOr2 }
func Mio_GateReadName(pGate *Mio_Gate_t) string                  { return pGate.pName }
func Mio_GateReadOutName(pGate *Mio_Gate_t) string               { return pGate.pOutName }
func Mio_GateReadArea(pGate *Mio_Gate_t) float64                 { return pGate.dArea }
func Mio_GateReadForm(pGate *Mio_Gate_t) string                  { return pGate.pForm }
func Mio_GateReadPins(pGate *Mio_Gate_t) *Mio_Pin_t              { return pGate.pPins }
func Mio_GateReadLib(pGate *Mio_Gate_t) *Mio_Library_t           { return pGate.pLib }
func Mio_GateReadNext(pGate *Mio_Gate_t) *Mio_Gate_t             { return pGate.pNext }
func Mio_GateReadTwin(pGate *Mio_Gate_t) *Mio_Gate_t             { return pGate.pTwin }
func Mio_GateReadPinNum(pGate *Mio_Gate_t) int32                 { return pGate.nInputs }
func Mio_GateReadDelayMax(pGate *Mio_Gate_t) float64             { return pGate.dDelayMax }
func Mio_GateReadSop(pGate *Mio_Gate_t) string                   { return pGate.pSop }
func Mio_GateReadExpr(pGate *Mio_Gate_t) *Vec_Int_t              { return pGate.vExpr }
func Mio_GateReadValue(pGate *Mio_Gate_t) int32                  { return pGate.Value }
func Mio_GateReadCell(pGate *Mio_Gate_t) int32                   { return pGate.Cell }
func Mio_GateReadProfile(pGate *Mio_Gate_t) int32                { return pGate.Profile }
func Mio_GateReadProfile2(pGate *Mio_Gate_t) int32               { return pGate.Profile2 }
func Mio_PinReadName(pPin *Mio_Pin_t) string                     { return pPin.pName }
func Mio_PinReadPhase(pPin *Mio_Pin_t) Mio_PinPhase_t            { return pPin.Phase }
func Mio_PinReadInputLoad(pPin *Mio_Pin_t) float64               { return pPin.dLoadInput }
func Mio_PinReadMaxLoad(pPin *Mio_Pin_t) float64                 { return pPin.dLoadMax }
func Mio_PinReadDelayBlockRise(pPin *Mio_Pin_t) float64          { return pPin.dDelayBlockRise }
func Mio_PinReadDelayFanoutRise(pPin *Mio_Pin_t) float64         { return pPin.dDelayFanoutRise }
func Mio_PinReadDelayBlockFall(pPin *Mio_Pin_t) float64          { return pPin.dDelayBlockFall }
func Mio_PinReadDelayFanoutFall(pPin *Mio_Pin_t) float64         { return pPin.dDelayFanoutFall }
func Mio_PinReadDelayBlockMax(pPin *Mio_Pin_t) float64           { return pPin.dDelayBlockMax }
func Mio_PinReadNext(pPin *Mio_Pin_t) *Mio_Pin_t                 { return pPin.pNext }
func Mio_GateSetValue(g *Mio_Gate_t, v int32)                    { g.Value = v }
func Mio_GateSetCell(g *Mio_Gate_t, v int32)                     { g.Cell = v }
func Mio_GateSetProfile(g *Mio_Gate_t, v int32)                  { g.Profile = v }
func Mio_GateSetProfile2(g *Mio_Gate_t, v int32)                 { g.Profile2 = v }
func Mio_GateAddToProfile(g *Mio_Gate_t, v int32)                { g.Profile += v }
func Mio_GateAddToProfile2(g *Mio_Gate_t, v int32)               { g.Profile2 += v }
func Mio_GateIncProfile2(g *Mio_Gate_t)                          { g.Profile2++ }
func Mio_GateDecProfile2(g *Mio_Gate_t)                          { g.Profile2-- }
func Mio_LibraryReadDelayInvRise(p *Mio_Library_t) float32 {
	if p.pGateInv == nil || p.pGateInv.pPins == nil {
		return 0
	}
	return float32(p.pGateInv.pPins.dDelayBlockRise)
}
func Mio_LibraryReadDelayInvFall(p *Mio_Library_t) float32 {
	if p.pGateInv == nil || p.pGateInv.pPins == nil {
		return 0
	}
	return float32(p.pGateInv.pPins.dDelayBlockFall)
}
func Mio_LibraryReadDelayInvMax(p *Mio_Library_t) float32 {
	if p.pGateInv == nil || p.pGateInv.pPins == nil {
		return 0
	}
	return float32(p.pGateInv.pPins.dDelayBlockMax)
}
func Mio_LibraryReadDelayNand2Rise(p *Mio_Library_t) float32 {
	if p.pGateNand2 == nil || p.pGateNand2.pPins == nil {
		return 0
	}
	return float32(p.pGateNand2.pPins.dDelayBlockRise)
}
func Mio_LibraryReadDelayNand2Fall(p *Mio_Library_t) float32 {
	if p.pGateNand2 == nil || p.pGateNand2.pPins == nil {
		return 0
	}
	return float32(p.pGateNand2.pPins.dDelayBlockFall)
}
func Mio_LibraryReadDelayNand2Max(p *Mio_Library_t) float32 {
	if p.pGateNand2 == nil || p.pGateNand2.pPins == nil {
		return 0
	}
	return float32(p.pGateNand2.pPins.dDelayBlockMax)
}
func Mio_LibraryReadAreaInv(p *Mio_Library_t) float32 {
	if p.pGateInv == nil {
		return 0
	}
	return float32(p.pGateInv.dArea)
}
func Mio_LibraryReadAreaBuf(p *Mio_Library_t) float32 {
	if p.pGateBuf == nil {
		return 0
	}
	return float32(p.pGateBuf.dArea)
}
func Mio_LibraryReadAreaNand2(p *Mio_Library_t) float32 {
	if p.pGateNand2 == nil {
		return 0
	}
	return float32(p.pGateNand2.dArea)
}

func Mio_LibraryReadDelayAnd2Max(p *Mio_Library_t) float32 {
	if p.pGateAnd2 == nil {
		return 0
	}
	return float32(p.pGateAnd2.dDelayMax)
}
func Mio_LibraryReadDelayAigNode(p *Mio_Library_t) float32 {
	if p.pGateAnd2 != nil {
		return Mio_LibraryReadDelayAnd2Max(p)
	}
	return Mio_LibraryReadDelayNand2Max(p)
}
func Mio_LibrarySetName(p *Mio_Library_t, name string) { p.pName = name }
func Mio_LibraryReadGateById(p *Mio_Library_t, id int32) *Mio_Gate_t {
	if id < 0 || int(id) >= len(p.ppGates0) {
		return nil
	}
	return p.ppGates0[id]
}
func Mio_LibraryReadGateByName(p *Mio_Library_t, name, out string) *Mio_Gate_t {
	g := p.tName2Gate[name]
	if g == nil || out == "" || g.pOutName == out {
		return g
	}
	if g.pTwin != nil && g.pTwin.pOutName == out {
		return g.pTwin
	}
	return nil
}
func Mio_LibraryReadGateByTruth(p *Mio_Library_t, t uint64) *Mio_Gate_t {
	for g := p.pGates; g != nil; g = g.pNext {
		if g.nInputs <= 6 && g.uTruth == t {
			return g
		}
	}
	return nil
}
func Mio_LibraryReadSopByName(p *Mio_Library_t, name string) string {
	g := p.tName2Gate[name]
	if g == nil {
		return ""
	}
	return g.pSop
}
func Mio_LibraryReadGateNameMax(p *Mio_Library_t) int32 {
	var n int
	for g := p.pGates; g != nil; g = g.pNext {
		n = max(n, len(g.pName))
	}
	return int32(n)
}
func Mio_GateReadTruth(g *Mio_Gate_t) uint64 {
	if g.nInputs > 6 {
		return 0
	}
	return g.uTruth
}
func Mio_GateReadTruthP(g *Mio_Gate_t) []uint64 {
	if g.nInputs <= 6 {
		return []uint64{g.uTruth}
	}
	return g.pTruth
}
func Mio_GateIsInv(g *Mio_Gate_t) int32 {
	return mioBool(g.nInputs == 1 && g.uTruth == 0x5555555555555555)
}
func Mio_GateReadPinName(g *Mio_Gate_t, i int32) string {
	if i < 0 {
		return ""
	}
	for p := g.pPins; p != nil; p = p.pNext {
		i--
		if i < 0 {
			return p.pName
		}
	}
	return ""
}
func Mio_GateReadPinDelay(g *Mio_Gate_t, i int32) float32 {
	if i < 0 {
		return float32(math.Inf(1))
	}
	for p := g.pPins; p != nil; p = p.pNext {
		if i == 0 {
			return float32((p.dDelayBlockRise + p.dDelayBlockFall) / 2)
		}
		i--
	}
	return float32(math.Inf(1))
}
