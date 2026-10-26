package abc

// Dynamic flop occurrence priorities from src/proof/pdr/pdrCore.c:
// Pdr_ManSortByPriority and priority updates in Pdr_ManBlockCube.

import "slices"

// The default C priorities start at zero and increment by a common power of
// two. Counts give identical comparisons without carrying that scale factor.
func (p *pdrEngine) orderCube(cube *PdrCube) ([]PdrLit, error) {
	if err := cube.Check(); err != nil {
		return nil, err
	}
	order := slices.Clone(cube.state)
	for _, l := range order {
		if uint64(l.Index()) >= uint64(len(p.flopPriority)) {
			return nil, ErrDimension
		}
	}
	// Preserve the C selection-sort swaps, including their tie behavior.
	for i := 0; i < len(order)-1; i++ {
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		best := i
		for j := i + 1; j < len(order); j++ {
			if p.flopPriority[order[j].Index()] < p.flopPriority[order[best].Index()] {
				best = j
			}
		}
		order[i], order[best] = order[best], order[i]
	}
	return order, nil
}
func (p *pdrEngine) bumpCubePriority(cube *PdrCube) error {
	if err := cube.Check(); err != nil {
		return err
	}
	for _, l := range cube.state {
		if uint64(l.Index()) >= uint64(len(p.flopPriority)) {
			return ErrDimension
		}
	}
	// Avoid wrapping a heavily used register to lowest priority. Rescaling is
	// deterministic; the C signed counters do not protect this extreme case.
	step := p.priorityStep
	if step == 0 {
		step = 1
	}
	for _, l := range cube.state {
		if p.flopPriority[l.Index()] > ^uint64(0)-step {
			for i := range p.flopPriority {
				p.flopPriority[i] >>= 1
			}
			break
		}
	}
	for _, l := range cube.state {
		p.flopPriority[l.Index()] += step
	}
	return nil
}

// queryCubeLiterals ports Vec_IntSelectSortPrioReverseLit on an owned copy.
// Canonical cube storage and exact predecessor witness bits remain unchanged.
func (p *pdrEngine) queryCubeLiterals(cube *PdrCube, next bool) ([]SatLit, error) {
	if !next || !p.options.OrderAssumptions {
		return p.cubeLiterals(cube, next, false)
	}
	if err := cube.Check(); err != nil {
		return nil, err
	}
	order := slices.Clone(cube.state)
	for _, l := range order {
		if uint64(l.Index()) >= uint64(len(p.flopPriority)) {
			return nil, ErrDimension
		}
	}
	for i := 0; i < len(order)-1; i++ {
		if err := p.checkStop(); err != nil {
			return nil, err
		}
		best := i
		for j := i + 1; j < len(order); j++ {
			if p.flopPriority[order[j].Index()] > p.flopPriority[order[best].Index()] {
				best = j
			}
		}
		order[i], order[best] = order[best], order[i]
	}
	// Do not canonicalize this query-only view: its order is the heuristic.
	return p.cubeLiterals(&PdrCube{state: order}, true, false)
}
