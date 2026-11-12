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
	if p.solver.vVarsUsed.data[v] != 0 && !p.present[v] {
		for _, n := range p.solver.vAdjacents[v].data {
			if p.solver.pAssigns[n] >= 0 {
				p.present[v] = true
				p.variables = append(p.variables, v)
				break
			}
		}
	}
}
func (p *BoundaryOrder) Select() int {
	best := -1
	for _, v := range p.variables {
		if p.solver.pAssigns[v] < 0 && (best < 0 || p.solver.pdActivity[v] > p.solver.pdActivity[best]) {
			best = v
		}
	}
	if best < 0 {
		for _, v := range p.solver.vConeVars.data {
			if p.solver.pAssigns[v] < 0 {
				return v
			}
		}
		return MSAT_ORDER_UNKNOWN
	}
	p.remove(best)
	return best
}
func (p *BoundaryOrder) Clear() { p.variables = p.variables[:0]; clear(p.present) }
