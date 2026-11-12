package sfm

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/sat/satoko"
)

func (p *Sfm_Ntk_t) covered(i int) bool {
	for _, w := range p.witnesses {
		if !w[i] {
			return false
		}
	}
	return true
}
func Sfm_NodeResubSolve(p *Sfm_Ntk_t, target, pin int, removeOnly bool) bool {
	p.Stats.TryRemoves++
	p.witnesses = nil
	p.divIDs = nil
	removed := p.Fanins[target][pin]
	skip := -1
	if p.Fixed[removed] && len(p.Fanins[removed]) == 1 {
		skip = p.Fanins[removed][0]
	}
	for i, f := range p.Fanins[target] {
		if i != pin {
			p.divIDs = append(p.divIDs, f)
		}
	}
	truth := Sfm_ComputeInterpolant(p)
	if truth == SFM_SAT_UNDEC {
		return false
	}
	if truth != SFM_SAT_SAT {
		p.Stats.Removes++
		Sfm_NtkUpdate(p, target, pin, -1, truth)
		return true
	}
	if removeOnly || p.Params.RrOnly {
		return false
	}
	p.Stats.TryResubs++
	for {
		next := -1
		for i, id := range p.divs {
			if id != skip && p.covered(i) {
				next = i
				break
			}
		}
		if next < 0 {
			return false
		}
		replacement := p.divs[next]
		p.divIDs = append(p.divIDs, replacement)
		truth = Sfm_ComputeInterpolant(p)
		if truth == SFM_SAT_UNDEC {
			return false
		}
		if truth != SFM_SAT_SAT {
			p.Stats.Resubs++
			Sfm_NtkUpdate(p, target, pin, replacement, truth)
			return true
		}
		p.divIDs = p.divIDs[:len(p.divIDs)-1]
		if len(p.witnesses) >= 64 {
			return false
		}
	}
}
func words32(t uint64) []uint32 { return []uint32{uint32(t), uint32(t >> 32)} }
func cubeTruth(cube uint32, vars int) uint64 {
	value := ^uint64(0)
	for i := 0; i < vars; i++ {
		var elementary uint64
		for b := 0; b < 64; b++ {
			if b&(1<<uint(i)) != 0 {
				elementary |= 1 << uint(b)
			}
		}
		switch cube >> uint(i*2) & 3 {
		case 1:
			value &= ^elementary
		case 2:
			value &= elementary
		}
	}
	return value
}
func isopCompletion(on, off uint64, vars int) (uint64, int) {
	var memory []uint32
	compl := kit.Kit_TruthIsop2(words32(off), words32(on), int32(vars), &memory, 0, 0)
	if compl < 0 {
		panic("sfm: incomplete ISOP failed")
	}
	truth := uint64(0)
	for _, cube := range memory {
		truth |= cubeTruth(cube, vars)
	}
	return truth, len(memory)
}
func Sfm_ComputeInterpolant2(p *Sfm_Ntk_t) uint64 {
	sets, status := p.projectedSets()
	if status == satoko.SATOKO_UNDEC {
		return SFM_SAT_UNDEC
	}
	if status == satoko.SATOKO_SAT {
		return SFM_SAT_SAT
	}
	on, off := stretch(sets[1], len(p.divIDs)), stretch(sets[0], len(p.divIDs))
	a, ca := isopCompletion(on, off, len(p.divIDs))
	b, cb := isopCompletion(off, on, len(p.divIDs))
	if ca <= cb {
		return a
	}
	return ^b
}
func literalCost(t uint64, n int) int {
	var cover []uint32
	kit.Kit_TruthIsop(words32(t), int32(n), &cover, 0)
	cost := 0
	for _, c := range cover {
		for i := 0; i < n; i++ {
			if c>>uint(2*i)&3 != 0 {
				cost++
			}
		}
	}
	return cost
}
func Sfm_NodeResubOne(p *Sfm_Ntk_t, target int) bool {
	p.Stats.TryImproves++
	p.divIDs = append(p.divIDs[:0], p.Fanins[target]...)
	truth := Sfm_ComputeInterpolant2(p)
	if truth == SFM_SAT_UNDEC || truth == SFM_SAT_SAT || truth == p.Truths[target][0] {
		return false
	}
	if literalCost(truth, len(p.divIDs)) > literalCost(p.Truths[target][0], len(p.divIDs)) {
		return false
	}
	p.Truths[target][0] = truth
	p.cnfs[target] = Sfm_TruthToCnf(truth, nil, len(p.divIDs))
	p.Stats.Improves++
	return true
}
func Sfm_NodeResub(p *Sfm_Ntk_t, target int) bool {
	p.Stats.NodesTried++
	if !Sfm_NtkCreateWindow(p, target, false) || !Sfm_NtkWindowToSolver(p) {
		return false
	}
	for i, f := range p.Fanins[target] {
		if Sfm_ObjIsNode(p, f) && len(p.Fanouts[f]) == 1 {
			if Sfm_NodeResubSolve(p, target, i, false) {
				return true
			}
		}
	}
	if !p.Params.Area {
		for i, f := range p.Fanins[target] {
			if !(Sfm_ObjIsNode(p, f) && len(p.Fanouts[f]) == 1) {
				if Sfm_NodeResubSolve(p, target, i, true) {
					return true
				}
			}
		}
	}
	return p.Params.UseDcs && Sfm_NodeResubOne(p, target)
}
func Sfm_NtkPerform(p *Sfm_Ntk_t, pars *Sfm_Par_t) int {
	if pars != nil {
		p.Params = *pars
	}
	if p.Params.WinSizeMax < 0 || p.Params.FanoutMax < 0 || p.Params.TfoLevMax < 0 || p.Params.DepthMax < 0 || p.Params.BTLimit < 0 {
		panic("sfm: negative parameter")
	}
	if err := p.refresh(); err != nil {
		panic(err)
	}
	Sfm_NtkPrepare(p)
	p.Stats = Stats{}
	p.Stats.NodesBefore, p.Stats.EdgesBefore = p.counts()
	changed := 0
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if p.Fixed[id] || p.Deleted[id] || len(p.Fanins[id]) < 2 || len(p.Fanins[id]) > 6 || (p.Params.DepthMax > 0 && p.Levels[id] > p.Params.DepthMax) {
			continue
		}
		one := false
		for Sfm_NodeResub(p, id) {
			one = true
		}
		if one {
			changed++
		}
		if p.Params.NodesMax > 0 && changed >= p.Params.NodesMax {
			break
		}
	}
	p.Stats.NodesAfter, p.Stats.EdgesAfter = p.counts()
	if p.Params.Verbose {
		Sfm_NtkPrintStats(p)
	}
	return changed
}
func (p *Sfm_Ntk_t) counts() (int, int) {
	nodes, edges := 0, 0
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if !p.Deleted[id] && len(p.Fanins[id]) > 0 {
			nodes++
		}
		if !p.Deleted[id] {
			edges += len(p.Fanins[id])
		}
	}
	return nodes, edges
}
func Sfm_NtkPrintStats(p *Sfm_Ntk_t) {
	fmt.Printf("SFM: nodes %d -> %d, edges %d -> %d, removes %d, resubs %d, improves %d, SAT %d, timeouts %d\n", p.Stats.NodesBefore, p.Stats.NodesAfter, p.Stats.EdgesBefore, p.Stats.EdgesAfter, p.Stats.Removes, p.Stats.Resubs, p.Stats.Improves, p.Stats.SATCalls, p.Stats.Timeouts)
}
