package sfm

import "github.com/glycerine/abcgo/sat/satoko"

func Sfm_NtkWindowToSolver(p *Sfm_Ntk_t) bool {
	p.solver = satoko.New()
	p.idVar = make(map[int]int, len(p.order))
	next := 1
	for _, id := range p.order {
		p.idVar[id] = next
		next++
	}
	oldRoots := make([]int, len(p.roots))
	for i, id := range p.roots {
		oldRoots[i] = p.idVar[id]
	}
	p.solver.SetVariables(next + len(p.tfo) + len(p.roots))
	addNode := func(id, pivot int) {
		fanins := make([]int, 0, len(p.Fanins[id])+1)
		for _, f := range p.Fanins[id] {
			fanins = append(fanins, p.idVar[f])
		}
		fanins = append(fanins, p.idVar[id])
		for _, c := range Sfm_TranslateCnf(p.cnfs[id], fanins, pivot) {
			p.solver.AddClause(c)
		}
	}
	for _, id := range p.order {
		if !Sfm_ObjIsPi(p, id) {
			addNode(id, -1)
		}
	}
	if len(p.tfo) > 0 {
		for _, id := range p.tfo {
			p.idVar[id] = next
			next++
		}
		for _, id := range p.tfo {
			addNode(id, p.idVar[p.pivot])
		}
		var clause []int
		for i, id := range p.roots {
			a, b, z := oldRoots[i]<<1, p.idVar[id]<<1, next<<1
			next++
			p.solver.AddClause([]int{a, b, z ^ 1})
			p.solver.AddClause([]int{a ^ 1, b ^ 1, z ^ 1})
			p.solver.AddClause([]int{a, b ^ 1, z})
			p.solver.AddClause([]int{a ^ 1, b, z})
			clause = append(clause, z)
		}
		p.solver.AddClause(clause)
	}
	return true
}
func (p *Sfm_Ntk_t) solve(assumptions []int) int {
	p.Stats.SATCalls++
	status := p.solver.SolveAssumptionsLimit(assumptions, p.Params.BTLimit)
	if status == satoko.SATOKO_UNDEC {
		p.Stats.Timeouts++
	}
	return status
}

// projectedSets enumerates observable on/off sets. Opposite target assignments
// with an identical projection produce one XOR-difference witness on all divisors.
func (p *Sfm_Ntk_t) projectedSets() ([2]uint64, int) {
	var result [2]uint64
	if len(p.divIDs) > 6 {
		panic("sfm: interpolation exceeds six variables")
	}
	n := 1 << uint(len(p.divIDs))
	models := make([][2][]bool, n)
	for a := 0; a < n; a++ {
		for polarity := 0; polarity < 2; polarity++ {
			assume := []int{p.idVar[p.pivot]<<1 | (polarity ^ 1)}
			for i, id := range p.divIDs {
				l := p.idVar[id] << 1
				if a&(1<<uint(i)) == 0 {
					l ^= 1
				}
				assume = append(assume, l)
			}
			status := p.solve(assume)
			if status == satoko.SATOKO_UNDEC {
				return result, status
			}
			if status == satoko.SATOKO_UNSAT {
				continue
			}
			result[polarity] |= uint64(1) << uint(a)
			values := make([]bool, len(p.divs))
			for i, id := range p.divs {
				values[i] = p.solver.ModelValue(p.idVar[id]) == 1
			}
			models[a][polarity] = values
			if result[polarity^1]&(uint64(1)<<uint(a)) != 0 {
				w := make([]bool, len(p.divs))
				for i := range w {
					w[i] = models[a][0][i] != models[a][1][i]
				}
				p.witnesses = append(p.witnesses, w)
				return result, satoko.SATOKO_SAT
			}
		}
	}
	return result, satoko.SATOKO_UNSAT
}
func Sfm_ComputeInterpolant(p *Sfm_Ntk_t) uint64 {
	sets, status := p.projectedSets()
	if status == satoko.SATOKO_UNDEC {
		return SFM_SAT_UNDEC
	}
	if status == satoko.SATOKO_SAT {
		return SFM_SAT_SAT
	}
	return stretch(sets[1], len(p.divIDs))
}
func Sfm_ComputeInterpolantInt(p *Sfm_Ntk_t) ([2]uint64, int) { return p.projectedSets() }
