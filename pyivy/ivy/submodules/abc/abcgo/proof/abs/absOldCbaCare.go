package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Saig_ManCba_t struct {
	pAig, pFrames          *aig.Graph
	pCex                   *util.Abc_Cex_t
	nInputs, fVerbose      int32
	vMapPiF2A              *vec.Vec_Int_t
	vReg2Frame, vReg2Value *vec.Vec_Vec_t
	phase, visited         []bool
}

// Saig_ManCbaReason2Inputs maps frame inputs to distinct additional inputs,
// numbered relative to the first abstracted register input.
func Saig_ManCbaReason2Inputs(p *Saig_ManCba_t, reasons *vec.Vec_Int_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(len(p.pAig.Inputs()))
	seen := make(map[int32]bool)
	for _, reason := range vec.Vec_IntArray(reasons) {
		input := vec.Vec_IntEntry(p.vMapPiF2A, int(2*reason))
		if input < p.nInputs || int(input) >= len(p.pAig.Inputs()) {
			panic("invalid CBA reason input")
		}
		if !seen[input] {
			vec.Vec_IntPush(result, input-p.nInputs)
			seen[input] = true
		}
	}
	return result
}
func Saig_ManCbaReason2Cex(p *Saig_ManCba_t, reasons *vec.Vec_Int_t) *util.Abc_Cex_t {
	care := util.Abc_CexDup(p.pCex, p.pCex.NRegs)
	clear(care.PData)
	for _, pi := range vec.Vec_IntArray(reasons) {
		if pi < 0 || int(pi) >= len(p.pFrames.Inputs()) {
			panic("invalid CBA frame input")
		}
		input, frame := vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi)), vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi+1))
		util.Abc_InfoSetBit(care.PData, care.NRegs+care.NPis*frame+input)
	}
	return care
}

func Saig_ManCbaStart(g *aig.Graph, cex *util.Abc_Cex_t, nInputs, verbose int32) *Saig_ManCba_t {
	return &Saig_ManCba_t{pAig: g, pCex: cex, nInputs: nInputs, fVerbose: verbose}
}
func Saig_ManCbaStop(p *Saig_ManCba_t) {
	p.pFrames = nil
	p.vMapPiF2A = nil
	p.vReg2Frame = nil
	p.vReg2Value = nil
}
