package bsat

import (
	"fmt"
	"math"
)

func sat_solver_count_assigned(s *sat_solver) int32 {
	if sat_solver_dl(s) != 0 {
		panic("count assigned above root")
	}
	var count int32
	for _, v := range s.assigns {
		if v != satUnassigned {
			count++
		}
	}
	return count
}
func sat_solver_progress(s *sat_solver) float64 {
	if s.size == 0 {
		return 0
	}
	factor := 1 / float64(s.size)
	progress := 0.0
	for v, value := range s.assigns {
		if value != satUnassigned {
			progress += math.Pow(factor, float64(s.levels[v]))
		}
	}
	return progress / float64(s.size)
}
func luby(y float64, x int32) float64 {
	if x < 0 {
		panic("negative restart index")
	}
	size, seq := int64(1), 0
	index := int64(x)
	for size < index+1 {
		seq++
		size = 2*size + 1
	}
	for size-1 != index {
		size = (size - 1) >> 1
		seq--
		index %= size
	}
	return math.Pow(y, float64(seq))
}
func sat_solver_search(s *sat_solver, budget int64) int32 {
	if s.root_level != sat_solver_dl(s) {
		panic("search started above assumption level")
	}
	randomFreq := float64(float32(0.02))
	if s.fNotUseRandom != 0 {
		randomFreq = 0
	}
	var conflictCount int64
	var learnt veci
	veci_new(&learnt)
	defer veci_delete(&learnt)
	s.nRestarts++
	s.stats.starts++
	if s.nRestarts&1 != 0 {
		for _, v := range veci_begin(&s.act_vars) {
			act_var_bump_factor(s, v)
		}
	}
	if s.pGlobalVars != nil {
		for _, v := range veci_begin(&s.act_vars) {
			act_var_bump_global(s, v)
		}
	}
	for {
		conflict := sat_solver_propagate(s)
		if conflict != 0 {
			s.stats.conflicts++
			conflictCount++
			if sat_solver_dl(s) == s.root_level {
				sat_solver_analyze_final(s, conflict, 0)
				return l_False
			}
			veci_resize(&learnt, 0)
			sat_solver_analyze(s, conflict, &learnt)
			backtrack := s.root_level
			if learnt.size > 1 {
				backtrack = max(backtrack, s.levels[lit_var(learnt.ptr[1])])
			}
			sat_solver_canceluntil(s, backtrack)
			sat_solver_record(s, &learnt)
			if learnt.size == 1 {
				s.levels[lit_var(learnt.ptr[0])] = 0
			}
			act_var_decay(s)
			act_clause_decay(s)
			continue
		}
		runtimeReached := s.stats.conflicts&63 == 0 && s.runtimeExpired()
		if s.fNoRestarts == 0 && budget >= 0 && conflictCount >= budget || runtimeReached || s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit || s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit {
			s.progress_estimate = sat_solver_progress(s)
			sat_solver_canceluntil(s, s.root_level)
			return l_Undef
		}
		if sat_solver_dl(s) == 0 && s.fSkipSimplify == 0 {
			sat_solver_simplify(s)
		}
		if s.nLearntMax != 0 && s.act_clas.size >= s.nLearntMax {
			sat_solver_reducedb(s)
		}
		s.stats.decisions++
		next := order_select(s, randomFreq)
		if next == var_Undef {
			for v, value := range s.assigns {
				if value == var1 {
					s.model[v] = l_True
				} else {
					s.model[v] = l_False
				}
			}
			sat_solver_canceluntil(s, s.root_level)
			return l_True
		}
		decision := lit_neg(toLit(next))
		if s.polarity[next] != 0 {
			decision = toLit(next)
		}
		sat_solver_decision(s, decision)
	}
}
func sat_solver_solve_internal(s *sat_solver) int32 {
	status := int32(l_Undef)
	restart := int32(0)
	veci_resize(&s.unit_lits, 0)
	s.nCalls++
	if s.verbosity >= 1 {
		fmt.Println("==============================[MINISAT]==============================")
	}
	for status == l_Undef {
		if s.runtimeExpired() {
			break
		}
		if s.verbosity >= 1 {
			fmt.Printf("conflicts=%d clauses=%d learned=%d progress=%.3f%%\n", s.stats.conflicts, s.stats.clauses, s.stats.learnts, s.progress_estimate*100)
		}
		budget := int64(100 * luby(2, restart))
		restart++
		status = sat_solver_search(s, budget)
		if s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit || s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit || s.runtimeExpired() || s.pFuncStop != nil && s.pFuncStop(s.RunId) != 0 {
			break
		}
	}
	sat_solver_canceluntil(s, s.root_level)
	if status == l_True {
		for _, v := range veci_begin(&s.user_vars) {
			veci_push(&s.user_values, sat_solver_var_value(s, v))
		}
	}
	return status
}
