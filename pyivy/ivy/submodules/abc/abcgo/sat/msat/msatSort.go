// Go port of learnt-clause sorting in msatSort.c.
package msat

import "sort"

func Msat_SolverSortDB(p *Msat_Solver_t) {
	sort.Slice(p.vLearned.data, func(i, j int) bool { return p.vLearned.data[i].activity < p.vLearned.data[j].activity })
}
