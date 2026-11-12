package bsat2

import "math"

func (s *Solver) search(budget int) LBool {
	s.Starts++
	conflicts := 0
	for {
		conflict := s.propagate()
		if conflict != CRefUndef {
			s.Conflicts++
			conflicts++
			if s.level() == 0 {
				return False
			}
			learned, backtrack := s.analyze(conflict)
			s.cancelUntil(backtrack)
			if len(learned) == 1 {
				s.uncheckedEnqueue(learned[0], CRefUndef)
			} else {
				h := s.ca.Alloc(learned, true)
				s.learnts = append(s.learnts, h)
				s.attach(h)
				s.bumpClause(s.ca.At(h))
				s.uncheckedEnqueue(learned[0], h)
			}
			s.variableInc /= s.VarDecay
			s.clauseInc /= s.ClauseDecay
			s.learntAdjustCount--
			if s.learntAdjustCount == 0 {
				s.learntAdjustConflicts *= 1.5
				s.learntAdjustCount = int(s.learntAdjustConflicts)
				s.maxLearnts *= 1.1
			}
			continue
		}
		if budget >= 0 && conflicts >= budget || !s.withinBudget() {
			s.progress = s.progressEstimate()
			s.cancelUntil(0)
			return Undef
		}
		if s.level() == 0 && !s.Simplify() {
			return False
		}
		if float64(len(s.learnts)-len(s.trail)) >= s.maxLearnts {
			s.reduceDB()
		}
		next := LitUndef
		for s.level() < len(s.assumptions) {
			l := s.assumptions[s.level()]
			switch s.Value(l) {
			case True:
				s.trailLimits = append(s.trailLimits, len(s.trail))
			case False:
				s.analyzeFinal(l.Not())
				return False
			default:
				next = l
			}
			if next != LitUndef {
				break
			}
		}
		if next == LitUndef {
			s.Decisions++
			next = s.pickBranch()
			if next == LitUndef {
				return True
			}
		}
		s.trailLimits = append(s.trailLimits, len(s.trail))
		s.uncheckedEnqueue(next, CRefUndef)
	}
}
func luby(base float64, index int) float64 {
	size, sequence := 1, 0
	for size < index+1 {
		size = 2*size + 1
		sequence++
	}
	for size-1 != index {
		size = (size - 1) / 2
		sequence--
		index %= size
	}
	return math.Pow(base, float64(sequence))
}
func (s *Solver) progressEstimate() float64 {
	if s.NumVars() == 0 {
		return 0
	}
	progress := float64(0)
	factor := 1 / float64(s.NumVars())
	for level := 0; level <= s.level(); level++ {
		begin := 0
		if level > 0 {
			begin = s.trailLimits[level-1]
		}
		end := len(s.trail)
		if level < s.level() {
			end = s.trailLimits[level]
		}
		progress += math.Pow(factor, float64(level)) * float64(end-begin)
	}
	return progress / float64(s.NumVars())
}
func (s *Solver) SolveLimited(assumptions ...Lit) LBool {
	s.Model = nil
	s.Conflict = nil
	if !s.ok {
		return False
	}
	for _, l := range assumptions {
		if l < 0 || int(l.Var()) >= s.NumVars() {
			panic("assumption outside solver")
		}
	}
	s.assumptions = append(s.assumptions[:0], assumptions...)
	s.Solves++
	s.maxLearnts = float64(len(s.clauses)) / 3
	s.learntAdjustConflicts = 100
	s.learntAdjustCount = 100
	status := Undef
	for restart := 0; status == Undef; restart++ {
		base := math.Pow(s.RestartIncrease, float64(restart))
		if s.LubyRestart {
			base = luby(s.RestartIncrease, restart)
		}
		status = s.search(int(base * float64(s.RestartFirst)))
		if !s.withinBudget() {
			break
		}
	}
	if status == True {
		s.Model = append([]LBool(nil), s.assigns...)
	} else if status == False && len(s.Conflict) == 0 {
		s.ok = false
	}
	s.cancelUntil(0)
	return status
}
func (s *Solver) Solve(assumptions ...Lit) LBool {
	s.BudgetOff()
	return s.SolveLimited(assumptions...)
}
