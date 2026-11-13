// Learned-clause sort from msatSort.c.
package msat

func Msat_SolverSortDB(p *Msat_Solver_t) {
	msatSolverSort(p.vLearned.data, 91648253)
}

func msatSolverSort(clauses []*Msat_Clause_t, seed float64) {
	if len(clauses) <= 15 {
		for i := 0; i < len(clauses)-1; i++ {
			best := i
			for j := i + 1; j < len(clauses); j++ {
				if clauses[j].activity < clauses[best].activity {
					best = j
				}
			}
			clauses[i], clauses[best] = clauses[best], clauses[i]
		}
		return
	}
	value := seed * 1389796
	q := int(value / 2147483647)
	value -= float64(q) * 2147483647
	pivot := clauses[int(value/2147483647*float64(len(clauses)))]
	i, j := -1, len(clauses)
	for {
		for {
			i++
			if !(clauses[i].activity < pivot.activity) {
				break
			}
		}
		for {
			j--
			if !(pivot.activity < clauses[j].activity) {
				break
			}
		}
		if i >= j {
			break
		}
		clauses[i], clauses[j] = clauses[j], clauses[i]
	}
	msatSolverSort(clauses[:i], seed)
	msatSolverSort(clauses[i:], seed)
}
