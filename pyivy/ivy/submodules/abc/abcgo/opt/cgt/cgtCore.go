package cgt

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Cgt_ClockGatingRangeCheck(p *Manager, start, count int) {
	regs, _, pos := layout(p.Graph)
	if start < 0 || count < 0 || start+count > regs {
		panic("cgt: invalid gating range")
	}
	originalOutputs := p.Graph.Outputs()
	frameOutputs := p.Frame.Outputs()
	for r := start; r < start+count; r++ {
		next := driver(p.Graph, originalOutputs[pos+r])
		candidates := Cgt_ManDetectCandidates(p.Graph, p.Useful, next, p.Params.LevelMax)
		if p.Params.CandidateMax > 0 && len(candidates) > p.Params.CandidateMax {
			candidates = candidates[:p.Params.CandidateMax]
		}
		miter := p.partCopies[frameOutputs[r].NodeID()]
		for _, candidate := range candidates {
			frameGate := p.frameCopies[candidate.NodeID()]
			partGate := cloneEdge(p.Part, p.Frame, frameGate, p.partCopies, func(aig.Edge) aig.Edge { return p.Part.NewInput() })
			for phase := 0; phase < 2; phase++ {
				gate := partGate ^ aig.Edge(phase)
				if !Cgt_SimulationFilter(p, gate, miter) {
					p.Stats.Filtered++
					continue
				}
				status := Cgt_CheckImplication(p, gate, miter)
				if status == 1 {
					p.AllGates[r] = append(p.AllGates[r], candidate^aig.Edge(phase))
					break
				}
				if status == 0 {
					Cgt_SimulationRecord(p)
				}
			}
		}
	}
}
func Cgt_ClockGatingRange(p *Manager, start int) int {
	Cgt_ManClean(p)
	p.Part, p.partCopies, _, p.careCount = cgtPartition(p.Frame, p.Params.VarsMin, p.Params.FlopsMin, start, p.Care)
	count := p.Part.OutputCount() - p.careCount
	p.encodePart()
	outs := p.Part.Outputs()
	for i := 0; i < p.careCount; i++ {
		p.solver.AddClause([]int{int(outs[i])})
	}
	Cgt_ClockGatingRangeCheck(p, start, count)
	p.Stats.Recycles++
	return start + count
}
func Cgt_ClockGatingCandidates(g, care *aig.Graph, parameters *Cgt_Par_t, useful []bool) (Gates, error) {
	p, err := Cgt_ManCreate(g, care, parameters)
	if err != nil {
		return nil, err
	}
	defer Cgt_ManStop(p)
	if useful != nil && len(useful) < len(g.Nodes()) {
		return nil, fmt.Errorf("cgt: useful-node mask is too short")
	}
	p.Useful = append([]bool(nil), useful...)
	Cgt_ManDeriveAigForGating(p)
	for start := 0; start < p.Frame.OutputCount(); {
		start = Cgt_ClockGatingRange(p, start)
	}
	result := p.AllGates
	p.AllGates = nil
	return result, nil
}
func Cgt_ClockGatingInt(g, care *aig.Graph, parameters *Cgt_Par_t, useful []bool) (Gates, error) {
	pars := DefaultParams()
	if parameters != nil {
		pars = *parameters
	}
	all, err := Cgt_ClockGatingCandidates(g, care, &pars, useful)
	if err != nil {
		return nil, err
	}
	if pars.AreaOnly {
		return Cgt_ManDecideArea(g, all, pars.ODCMax, pars.Verbose), nil
	}
	return Cgt_ManDecideSimple(g, all, pars.ODCMax, pars.Verbose), nil
}
func Cgt_ClockGating(g, care *aig.Graph, parameters *Cgt_Par_t) (*aig.Graph, error) {
	pars := DefaultParams()
	if parameters != nil {
		pars = *parameters
	}
	gates, err := Cgt_ClockGatingInt(g, care, &pars, nil)
	if err != nil {
		return nil, err
	}
	result, _, err := Cgt_ManDeriveGatedAig(g, gates, pars.AreaOnly)
	return result, err
}
