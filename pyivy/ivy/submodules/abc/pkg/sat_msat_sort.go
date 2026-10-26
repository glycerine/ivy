package abc

// Learned-clause sorting from src/sat/msat/msatSort.c.

const msatClauseSortSeed = 91648253.0

func msatSortDRand(seed float64) float64 {
	seed *= 1389796
	quotient := int(seed / 2147483647)
	seed -= float64(quotient) * 2147483647
	return seed / 2147483647
}

func msatSortIRand(seed float64, size int) int {
	return int(msatSortDRand(seed) * float64(size))
}

func msatClauseActivityCompare(first, second *MsatClause) int {
	if first.activity < second.activity {
		return -1
	}
	if first.activity > second.activity {
		return 1
	}
	return 0
}

func msatSortClauseSelection(clauses []*MsatClause) {
	for index := 0; index < len(clauses)-1; index++ {
		best := index
		for candidate := index + 1; candidate < len(clauses); candidate++ {
			if clauses[candidate].activity < clauses[best].activity {
				best = candidate
			}
		}
		clauses[index], clauses[best] = clauses[best], clauses[index]
	}
}

func msatSortClauses(clauses []*MsatClause, seed float64) {
	if len(clauses) <= 15 {
		msatSortClauseSelection(clauses)
		return
	}
	pivot := clauses[msatSortIRand(seed, len(clauses))]
	left, right := -1, len(clauses)
	for {
		left++
		for clauses[left].activity < pivot.activity {
			left++
		}
		right--
		for pivot.activity < clauses[right].activity {
			right--
		}
		if left >= right {
			break
		}
		clauses[left], clauses[right] = clauses[right], clauses[left]
	}
	msatSortClauses(clauses[:left], seed)
	msatSortClauses(clauses[left:], seed)
}

func validateMsatSortClauses(clauses []*MsatClause) error {
	for _, clause := range clauses {
		if clause == nil || clause.closed {
			return ErrDimension
		}
	}
	return nil
}

// SortMsatClausesByActivity mutates clauses into the deterministic native
// MiniSat order. Equal and NaN activities retain the source algorithm's
// deterministic partition behavior rather than receiving a stable tie-break.
func SortMsatClausesByActivity(clauses []*MsatClause) error {
	if err := validateMsatSortClauses(clauses); err != nil {
		return err
	}
	msatSortClauses(clauses, msatClauseSortSeed)
	return nil
}

// SortByActivity sorts the active clause database in place.
func (v *MsatClauseVector) SortByActivity() error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if err := validateMsatSortClauses(v.values[:v.size]); err != nil {
		return err
	}
	msatSortClauses(v.values[:v.size], msatClauseSortSeed)
	return nil
}
