package goivy

import "github.com/glycerine/ivy/goivy/smt"

// guardedEqualityCount finds consecutive, static counterexample conditions of
// the form !(guard => equality) with a shared conjunctive guard. Leave other
// checkers in their original callback and assertion order.
func guardedEqualityCount(conditions []FinalCond) int {
	var first Expr
	for i, condition := range conditions {
		if i == 16 {
			return i
		}
		checker, ok := condition.(*ConjChecker)
		if !ok || checker.FC == nil || len(checker.FC.Defs) != 0 || len(checker.FC.Fmlas) != 1 {
			return i
		}
		neg, ok := checker.FC.Fmlas[0].(*LogicNot)
		if !ok {
			return i
		}
		implication, ok := neg.Body.(*LogicImplies)
		if !ok {
			return i
		}
		guard, ok := implication.T1.(*LogicAnd)
		if !ok || len(guard.Terms) < 2 {
			return i
		}
		switch implication.T2.(type) {
		case *Eq, *LogicIff:
		default:
			return i
		}
		if i == 0 {
			first = guard
		} else if !first.Equal(guard) {
			return i
		}
	}
	return len(conditions)
}

// An unsatisfiable disjunction refutes every original counterexample condition.
// Keep each condition intact, including constraints introduced during translation.
// An inconclusive group or a satisfiable group goes back to the individual
// checks so the first failing obligation can return its own model.
func (s *Solver) proveGuardedEqualities(zs *smt.Z3Solver, first smt.Z3Expr, rest []FinalCond) bool {
	goals := make([]smt.Z3Expr, 1, 1+len(rest))
	goals[0] = first
	for _, condition := range rest {
		goal, err := s.ClausesToZ3(condition.(*ConjChecker).FC)
		if err != nil {
			return false
		}
		goals = append(goals, zs.SimplifyVerificationCondition(goal))
	}
	zs.Push()
	defer zs.Pop()
	zs.Assert(s.tr.Ctx.Or(goals...))
	return zs.Check() == smt.Unsat
}
