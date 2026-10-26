package abc

// Port of src/sat/bsat/satSolver.c: variable heap, drand, luby and reducedb.

import "slices"

func (s *SatSolver) heapUp(index int) {
	variable := s.order[index]
	for index > 0 {
		parent := (index - 1) / 2
		if s.activity[variable] <= s.activity[s.order[parent]] {
			break
		}
		s.order[index] = s.order[parent]
		s.orderpos[s.order[index]] = index
		index = parent
	}
	s.order[index] = variable
	s.orderpos[variable] = index
}
func (s *SatSolver) heapInsert(variable int) {
	if s.orderpos[variable] >= 0 {
		return
	}
	s.orderpos[variable] = len(s.order)
	s.order = append(s.order, variable)
	s.heapUp(len(s.order) - 1)
}
func (s *SatSolver) heapPop() int {
	variable := s.order[0]
	last := s.order[len(s.order)-1]
	s.order = s.order[:len(s.order)-1]
	s.orderpos[variable] = -1
	if len(s.order) > 0 {
		index := 0
		for {
			child := 2*index + 1
			if child >= len(s.order) {
				break
			}
			if child+1 < len(s.order) && s.activity[s.order[child]] < s.activity[s.order[child+1]] {
				child++
			}
			if s.activity[last] >= s.activity[s.order[child]] {
				break
			}
			s.order[index] = s.order[child]
			s.orderpos[s.order[index]] = index
			index = child
		}
		s.order[index] = last
		s.orderpos[last] = index
	}
	return variable
}
func (s *SatSolver) random() float64 {
	s.randomSeed *= 1389796
	quotient := int64(s.randomSeed / 2147483647)
	s.randomSeed -= float64(quotient) * 2147483647
	return s.randomSeed / 2147483647
}
func (s *SatSolver) selectVariable() int {
	r := s.random()
	if s.randomDecisions && r < 0.02 && len(s.assigns) > 0 {
		v := int(s.random() * float64(len(s.assigns)))
		if s.assigns[v] < 0 {
			return v
		}
	}
	for len(s.order) > 0 {
		v := s.heapPop()
		if s.assigns[v] < 0 {
			return v
		}
	}
	return -1
}
func bsatLuby(index uint64) uint64 {
	size, sequence := uint64(1), uint(0)
	for size < index+1 {
		size = 2*size + 1
		sequence++
	}
	for size-1 != index {
		size = (size - 1) / 2
		sequence--
		index %= size
	}
	return uint64(1) << sequence
}

type SatLearnedReductionStats struct {
	Before, Kept, Deleted, Limit, Reduction int
	Skipped                                 bool
}

func satLearnedScore(clause bsatClause) uint32 {
	lbd := min(uint32(clause.lbd), 7)
	return (7-lbd)<<28 | clause.activity>>4
}

func (s *SatSolver) nextLearnedLimit() int {
	s.dbReductions++
	maximum := int(^uint(0) >> 1)
	if s.learntDelta != 0 && s.dbReductions > (maximum-s.learntStart)/s.learntDelta {
		return maximum
	}
	return s.learntStart + s.learntDelta*s.dbReductions
}

func (s *SatSolver) reduceDBExact() (SatLearnedReductionStats, error) {
	learned := make([]int, 0, s.learntCount)
	for id := 1; id < len(s.clauses); id++ {
		if s.clauses[id].learnt && !s.clauses[id].deleted {
			learned = append(learned, id)
		}
	}
	stats := SatLearnedReductionStats{Before: len(learned), Limit: s.nextLearnedLimit(), Reduction: s.dbReductions}
	s.learntLimit = stats.Limit
	if len(learned) == 0 {
		return stats, nil
	}
	selected := len(learned) * s.learntRatio / 100
	if selected <= 0 {
		stats.Kept, stats.Skipped = len(learned), true
		return stats, nil
	}
	scores := make([]uint32, len(learned))
	ordered := make([]uint32, len(learned))
	for index, id := range learned {
		scores[index] = satLearnedScore(s.clauses[id])
		ordered[index] = scores[index]
	}
	slices.Sort(ordered)
	cutoff := ordered[len(ordered)-selected]
	locked := make([]bool, len(s.clauses))
	for _, reason := range s.reasons {
		if reason < 0 || reason >= len(s.clauses) {
			return SatLearnedReductionStats{}, ErrInvariant
		}
		if reason != 0 {
			locked[reason] = true
		}
	}
	counterStart := len(learned) - stats.Limit/20
	remove := make([]bool, len(s.clauses))
	deletions := 0
	for index, id := range learned {
		clause := s.clauses[id]
		preserve := index > counterStart || len(clause.lits) < 2 || scores[index] >= cutoff || locked[id]
		if !preserve {
			remove[id] = true
			deletions++
		}
	}
	if deletions < len(learned)/6 {
		stats.Kept, stats.Skipped = len(learned), true
		return stats, nil
	}
	oldToNew := make([]int, len(s.clauses))
	clauses := make([]bsatClause, 1, len(s.clauses)-deletions)
	for id := 1; id < len(s.clauses); id++ {
		if remove[id] {
			continue
		}
		oldToNew[id] = len(clauses)
		clauses = append(clauses, s.clauses[id])
	}
	proofBefore := s.proof.clone()
	resolutionBefore := s.resolutionProof.clone()
	if err := s.proofCompactLearned(remove, oldToNew, clauses); err != nil {
		return SatLearnedReductionStats{}, err
	}
	if err := s.resolutionProofCompact(remove, oldToNew, clauses); err != nil {
		s.proof = proofBefore
		s.resolutionProof = resolutionBefore
		return SatLearnedReductionStats{}, err
	}
	for variable, reason := range s.reasons {
		if reason == 0 {
			continue
		}
		if oldToNew[reason] == 0 {
			return SatLearnedReductionStats{}, ErrInvariant
		}
		s.reasons[variable] = oldToNew[reason]
	}
	units := make([]int, 0, len(s.units))
	for _, id := range s.units {
		if id <= 0 || id >= len(oldToNew) || oldToNew[id] == 0 {
			return SatLearnedReductionStats{}, ErrInvariant
		}
		units = append(units, oldToNew[id])
	}
	watches := make([][]int, len(s.watches))
	for literal, bucket := range s.watches {
		watches[literal] = make([]int, 0, len(bucket))
		for _, id := range bucket {
			if id <= 0 || id >= len(oldToNew) {
				return SatLearnedReductionStats{}, ErrInvariant
			}
			if oldToNew[id] != 0 {
				watches[literal] = append(watches[literal], oldToNew[id])
			}
		}
	}
	s.clauses, s.units, s.watches = clauses, units, watches
	s.learntCount -= deletions
	s.stats.Deleted += uint64(deletions)
	stats.Kept, stats.Deleted = len(learned)-deletions, deletions
	return stats, nil
}

func (s *SatSolver) reduceDB() {
	_, _ = s.reduceDBExact()
}

func (s *SatSolver) ReduceLearnedDatabase() (SatLearnedReductionStats, error) {
	if s == nil {
		return SatLearnedReductionStats{}, ErrDimension
	}
	if s.solving || len(s.clauses) == 0 || s.learntLimit == 0 {
		return SatLearnedReductionStats{}, ErrInvariant
	}
	trial := s.cloneState()
	trial.bookmark = s.bookmark
	stats, err := trial.reduceDBExact()
	if err != nil {
		return SatLearnedReductionStats{}, err
	}
	*s = trial
	return stats, nil
}
