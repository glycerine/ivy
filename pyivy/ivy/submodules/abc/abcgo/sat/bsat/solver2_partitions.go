package bsat

import "os"

// SetNumVars exposes source solver2 variable allocation to CNF drivers.
func (p *Solver2) SetNumVars(n int32) {
	if n < 0 {
		panic("negative solver2 variable count")
	}
	p.solver.lastStatus = Unknown
	sat_solver2_setnvars(p.solver, n)
}

func (p *Solver2) SetVariablePartitionA(variable int32, partA bool) {
	if variable < 0 || variable >= p.NumVars() {
		panic("partition variable outside solver2")
	}
	var_set_partA(p.solver, variable, solver2Bool(partA))
}

// AddPartitionClause retains the C solver2 insertion contract: problem clauses
// must be nonempty and not already contradicted by root units. It marks the
// returned clause after insertion, exactly as the source interpolation driver.
func (p *Solver2) AddPartitionClause(partA bool, literals ...Literal) {
	p.AddPartitionClauseWithID(p.nextID, partA, literals...)
}

// AddPartitionClauseWithID preserves the explicit source clause-ID contract,
// including solver2's rejection of ID zero.
func (p *Solver2) AddPartitionClauseWithID(id int32, partA bool, literals ...Literal) {
	p.solver.lastStatus = Unknown
	handle := sat_solver2_addclause(p.solver, literals, id)
	p.nextID = max(p.nextID, id+1)
	clause2_set_partA(p.solver, handle, solver2Bool(partA))
}

func (p *Solver2) PrintStatistics(file *os.File) { Sat_Solver2PrintStats(file, p.solver) }
func (p *Solver2) Close()                        { sat_solver2_delete(p.solver) }
