package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"time"
)

// Ssw_SmlSimulateOne evaluates all time frames using the current input words.
func Ssw_SmlSimulateOne(p *Ssw_Sml_t) { sswSimulateOne(p, nil) }

// sswSimulateOne stops between node batches when a deadline expires.
func sswSimulateOne(p *Ssw_Sml_t, pars *Ssw_Pars_t) bool {
	start := time.Now()
	nodes := aig.Aig_ManDfsAll(p.pAig)
	outputs := p.pAig.Outputs()
	inputs := p.pAig.Inputs()
	nRegs := int(p.pAig.RegisterCount())
	for frame := int32(0); frame < p.nFrames; frame++ {
		if sswDeadlineExpired(pars) {
			return false
		}
		for i, obj := range nodes {
			if i&255 == 0 && sswDeadlineExpired(pars) {
				return false
			}
			switch p.pAig.Node(obj).Kind {
			case aig.And, aig.ExclusiveOr, aig.Buffer:
				Ssw_SmlNodeSimulate(p, obj, frame)
			}
		}
		for _, obj := range outputs {
			Ssw_SmlNodeCopyFanin(p, obj, frame)
		}
		if frame+1 < p.nFrames {
			for i := 0; i < nRegs; i++ {
				Ssw_SmlNodeTransferNext(p, outputs[len(outputs)-nRegs+i], inputs[len(inputs)-nRegs+i], frame)
			}
		}
	}
	p.timeSim += time.Since(start)
	p.nSimRounds++
	return true
}

// Ssw_SmlUnnormalize converts the first frame's normalized internal words.
func Ssw_SmlUnnormalize(p *Ssw_Sml_t) {
	words := sswFrameWords(p, p.pAig.Const1(), 0)
	for i := range words {
		words[i] = ^words[i]
	}
	for _, obj := range p.pAig.LiveEdges() {
		node := p.pAig.Node(obj)
		if (node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer) && node.Phase {
			words = sswFrameWords(p, obj, 0)
			for i := range words {
				words[i] = ^words[i]
			}
		}
	}
}
