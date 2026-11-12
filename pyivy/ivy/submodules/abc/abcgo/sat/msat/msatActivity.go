// Go port of msatActivity.c by Alan Mishchenko, UC Berkeley (2004).
package msat

func Msat_SolverVarDecayActivity(p *Msat_Solver_t) {
	if p.dVarDecay >= 0 {
		p.dVarInc *= p.dVarDecay
	}
}
func Msat_SolverVarRescaleActivity(p *Msat_Solver_t) {
	for i := 0; i < p.nVars; i++ {
		p.pdActivity[i] *= 1e-100
	}
	p.dVarInc *= 1e-100
}
func Msat_SolverClaDecayActivity(p *Msat_Solver_t) { p.dClaInc *= p.dClaDecay }
func Msat_SolverClaRescaleActivity(p *Msat_Solver_t) {
	for _, c := range p.vLearned.data {
		c.activity *= 1e-20
	}
	p.dClaInc *= 1e-20
}
func Msat_SolverVarBumpActivity(p *Msat_Solver_t, lit int) {
	if p.dVarDecay < 0 {
		return
	}
	v := lit >> 1
	p.pdActivity[v] += p.dVarInc
	if p.pdActivity[v] > 1e100 {
		Msat_SolverVarRescaleActivity(p)
	}
	Msat_OrderUpdate(p.pOrder, v)
}
func Msat_SolverClaBumpActivity(p *Msat_Solver_t, c *Msat_Clause_t) {
	if !c.learned {
		return
	}
	if float64(c.activity)+p.dClaInc > 1e20 {
		Msat_SolverClaRescaleActivity(p)
	}
	c.activity += float32(p.dClaInc)
}
