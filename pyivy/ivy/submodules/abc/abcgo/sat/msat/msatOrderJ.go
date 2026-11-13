// The graph-boundary ordering from msatOrderJ.c is an alternative to the heap
// selected by module.make. BoundaryOrder exposes that alternative without C's
// mutually exclusive duplicate symbol definitions.
package msat

type BoundaryOrder struct {
	solver    *Msat_Solver_t
	variables []int
	present   []bool
}

func NewBoundaryOrder(p *Msat_Solver_t) *BoundaryOrder {
	return &BoundaryOrder{solver: p, present: make([]bool, len(p.pAssigns))}
}
func (p *BoundaryOrder) Assigned(v int) {
	p.remove(v)
	for _, neighbor := range p.solver.vAdjacents[v].data {
		if p.solver.vVarsUsed.data[neighbor] != 0 && p.solver.pAssigns[neighbor] < 0 && !p.present[neighbor] {
			p.present[neighbor] = true
			p.variables = append(p.variables, neighbor)
		}
	}
}
func (p *BoundaryOrder) remove(v int) {
	if !p.present[v] {
		return
	}
	for i, x := range p.variables {
		if x == v {
			copy(p.variables[i:], p.variables[i+1:])
			p.variables = p.variables[:len(p.variables)-1]
			break
		}
	}
	p.present[v] = false
}
func (p *BoundaryOrder) Unassigned(v int) {
	assert(!p.present[v], "unassigned variable is not already on boundary")
	for _, n := range p.solver.vAdjacents[v].data {
		if p.solver.pAssigns[n] != MSAT_VAR_UNASSIGNED {
			p.present[v] = true
			p.variables = append(p.variables, v)
			break
		}
	}
	for _, neighbor := range p.solver.vAdjacents[v].data {
		if !p.present[neighbor] {
			continue
		}
		assert(p.solver.pAssigns[neighbor] == MSAT_VAR_UNASSIGNED, "boundary neighbor is unassigned")
		hasAssigned := false
		for _, n := range p.solver.vAdjacents[neighbor].data {
			if p.solver.pAssigns[n] != MSAT_VAR_UNASSIGNED {
				hasAssigned = true
				break
			}
		}
		if !hasAssigned {
			p.remove(neighbor)
		}
	}
}
func (p *BoundaryOrder) Select() int {
	best, weight := MSAT_ORDER_UNKNOWN, -1.0
	for _, v := range p.variables {
		if p.solver.pdActivity[v] > weight {
			best, weight = v, p.solver.pdActivity[v]
		}
	}
	if best != MSAT_ORDER_UNKNOWN {
		assert(p.solver.vVarsUsed.data[best] != 0, "boundary variable is in cone")
	}
	return best
}
func (p *BoundaryOrder) Clear() { p.variables = p.variables[:0]; clear(p.present) }
