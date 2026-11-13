package msat

// Solver is the Go spelling of the original C solver type. Its implementation
// is the independent MSAT algorithm, not an adapter to another solver package.
type Solver = Msat_Solver_t

func New(variables int) *Solver {
	p := Msat_SolverAlloc(variables, 1, 1, 1, 1, 0)
	Msat_SolverClean(p, variables)
	msatInitializeOrder(p)
	return p
}
func (p *Msat_Solver_t) AddClause(literals []int) bool {
	for i := range p.pModel {
		p.pModel[i] = -1
	}
	if !p.okay {
		return false
	}
	if Msat_SolverAddClause(p, &Msat_IntVec_t{data: append([]int(nil), literals...)}) == 0 {
		p.okay = false
		return false
	}
	return true
}
func (p *Msat_Solver_t) Solve(assumptions []int, backtrackLimit, seconds int) Msat_Type_t {
	if !p.okay {
		return MSAT_FALSE
	}
	if Msat_SolverPropagate(p) != nil {
		p.okay = false
		return MSAT_FALSE
	}
	result := Msat_Type_t(Msat_SolverSolve(p, &Msat_IntVec_t{data: assumptions}, backtrackLimit, seconds))
	if result == MSAT_FALSE && len(assumptions) == 0 {
		p.okay = false
	}
	return result
}
func (p *Msat_Solver_t) ModelValue(variable int) int {
	if variable < 0 || variable >= p.nVars || p.pModel[variable] < 0 {
		return -1
	}
	return 1 - (p.pModel[variable] & 1)
}
