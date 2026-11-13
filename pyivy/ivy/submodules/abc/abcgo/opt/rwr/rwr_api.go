package rwr

import (
	"fmt"
	"os"
)

// Rwr_NodeRewrite runs the source algorithm over the caller's cuts in order.
// The network adapter keeps this package independent of ABC's concrete graph
// types. fPlaceEnable is accepted for source API parity; placement scoring is
// disabled in the reference implementation's active evaluation path.
func Rwr_NodeRewrite(p *Rwr_Man_t, net Rwr_Network, root any, cuts []Rwr_Cut, required int32, fUseZeros, fPlaceEnable int32) int32 {
	_ = fPlaceEnable
	if p == nil {
		panic("rwr: nil rewrite manager")
	}
	if p.rewriteLibrary == nil {
		p.rewriteLibrary = rwrStartLibrary()
	}
	lib := p.rewriteLibrary
	choice := lib.rwrRewriteCuts(net, root, cuts, required, fUseZeros != 0)
	p.nNodesConsidered = int32(lib.nNodesConsidered)
	p.nNodesRewritten = int32(lib.nNodesRewritten)
	p.nNodesGained = int32(lib.nNodesGained)
	p.nCutsGood = int32(lib.nCutsGood)
	p.nCutsBad = int32(lib.nCutsBad)
	p.nSubgraphs = int32(lib.nSubgraphs)
	for i, score := range lib.classScores {
		p.nScores[i] = int32(score)
	}
	if choice.graph == nil {
		return -1
	}
	p.pGraph = choice.graph
	p.fCompl = 0
	if choice.complement {
		p.fCompl = 1
	}
	p.vFanins = append(p.vFanins[:0], choice.fanins...)
	return choice.gain
}

// Rwr_ScoresClean clears per-library rewrite scoring counters.
func Rwr_ScoresClean(p *Rwr_Man_t) {
	for i := range p.nScores {
		p.nScores[i] = 0
	}
	if p.rewriteLibrary != nil {
		for i := range p.rewriteLibrary.classScores {
			p.rewriteLibrary.classScores[i] = 0
		}
		for i := range p.rewriteLibrary.scores {
			p.rewriteLibrary.scores[i] = 0
			p.rewriteLibrary.gains[i] = 0
			p.rewriteLibrary.added[i] = 0
		}
	}
}

// Rwr_ScoresReport reports NPN classes selected by the rewrite pass.
func Rwr_ScoresReport(p *Rwr_Man_t) {
	for class, score := range p.nScores {
		if score != 0 {
			fmt.Fprintf(os.Stdout, "CLASS %3d: Rewrites = %6d\n", class, score)
		}
	}
}
