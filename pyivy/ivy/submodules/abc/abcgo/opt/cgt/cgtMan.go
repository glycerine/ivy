package cgt

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Cgt_ManCreate(g, care *aig.Graph, parameters *Cgt_Par_t) (*Manager, error) {
	if g == nil {
		return nil, fmt.Errorf("cgt: nil graph")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	regs, _, _ := layout(g)
	if regs == 0 {
		return nil, fmt.Errorf("cgt: clock gating requires registers")
	}
	pars := DefaultParams()
	if parameters != nil {
		pars = *parameters
	}
	if pars.LevelMax < 0 || pars.CandidateMax < 0 || pars.ODCMax < 0 || pars.ConflictMax < 0 || pars.VarsMin <= 0 || pars.FlopsMin <= 0 {
		return nil, fmt.Errorf("cgt: invalid parameters")
	}
	if care != nil {
		if err := care.Check(); err != nil {
			return nil, err
		}
		if care.InputCount() != g.InputCount() {
			return nil, fmt.Errorf("cgt: care and AIG input counts differ")
		}
		if aig.Aig_ManRegNum(care) != 0 {
			return nil, fmt.Errorf("cgt: care graph must be combinational")
		}
	}
	return &Manager{Params: pars, Graph: g, Care: care, AllGates: make(Gates, regs), futureInputs: make(map[int]aig.Edge)}, nil
}
func Cgt_ManClean(p *Manager) {
	p.Part = nil
	p.partCopies = nil
	p.solver = nil
	p.encoded = nil
	p.patterns = nil
	p.lastModel = nil
	p.careCount = 0
}
func Cgt_ManPrintStats(p *Manager) {
	fmt.Printf("CGT: calls=%d unsat=%d sat=%d unknown=%d filtered=%d partitions=%d\n", p.Stats.Calls, p.Stats.UNSAT, p.Stats.SAT, p.Stats.Unknown, p.Stats.Filtered, p.Stats.Recycles)
}
func Cgt_ManStop(p *Manager) {
	if p == nil {
		return
	}
	if p.Params.Verbose {
		Cgt_ManPrintStats(p)
	}
	Cgt_ManClean(p)
	p.Frame = nil
	p.frameCopies = nil
	p.AllGates = nil
	p.Useful = nil
	p.futureInputs = nil
	p.Graph = nil
	p.Care = nil
}
