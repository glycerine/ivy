//go:build classic_z3 || js

package smt

// CheckVerificationCondition runs the native backend on the active assertions.
func (s *Z3Solver) CheckVerificationCondition() Z3CheckResult { return s.Check() }

// BeginVerificationChecks uses the native solver for each final condition.
func (s *Z3Solver) BeginVerificationChecks() (check func() Z3CheckResult, release func()) {
	return s.Check, func() {}
}

// The native backend checks final conditions in ordinary incremental scopes.
func (s *Z3Solver) BeginVerificationGoalChecks() (func() Z3CheckResult, func(Z3Expr) bool, func()) {
	return nil, nil, nil
}
