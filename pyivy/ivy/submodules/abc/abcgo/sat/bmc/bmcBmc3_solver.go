package bmc

import (
	"time"

	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/glucose"
	"github.com/glycerine/abcgo/sat/satoko"
)

// bmc3Solver is the common incremental surface used by the frame manager.
// Literals use ABC's variable<<1|complement representation in all backends.
type bmc3Solver interface {
	NumVars() int32
	NewVariable() int32
	AddClause(...bsat.Literal) bool
	SetDeadline(time.Time)
	SetLimits(int64, int64)
	SetLearnedPolicy(int32, int32, int32)
	SetNoRestarts(bool)
	Solve(...bsat.Literal) bsat.Status
	Value(int32) (bool, bool)
}

func newBmc3Solver(p *Saig_ParBmc_t) bmc3Solver {
	if p.FUseSatoko != 0 {
		s := satoko.New()
		s.SetRunID(int(p.RunId))
		if p.pFuncStop != nil {
			s.SetStopFunc(func(id int) int { return int(p.pFuncStop(int32(id))) })
		}
		return &bmc3Satoko{Solver: s}
	}
	if p.FUseGlucose != 0 {
		return &bmc3Glucose{Solver: glucose.NewSolver()}
	}
	s := bsat.NewSolver()
	s.SetLearnedPolicy(p.NLearnedStart, p.NLearnedDelta, p.NLearnedPerce)
	s.SetNoRestarts(p.FNoRestarts != 0)
	s.SetStopFunc(p.RunId, p.pFuncStop)
	return s
}

type bmc3Satoko struct {
	*satoko.Solver
	limit int64
}

func (s *bmc3Satoko) NumVars() int32     { return int32(s.VariableCount()) }
func (s *bmc3Satoko) NewVariable() int32 { return int32(s.AddVariable(false)) }
func (s *bmc3Satoko) AddClause(c ...bsat.Literal) bool {
	x := make([]int, len(c))
	for i, v := range c {
		x[i] = int(v)
	}
	return s.Solver.AddClause(x)
}
func (s *bmc3Satoko) SetLimits(conflicts, _ int64)         { s.limit = conflicts }
func (s *bmc3Satoko) SetLearnedPolicy(int32, int32, int32) {}
func (s *bmc3Satoko) SetNoRestarts(bool)                   {}
func (s *bmc3Satoko) Solve(a ...bsat.Literal) bsat.Status {
	x := make([]int, len(a))
	for i, v := range a {
		x[i] = int(v)
	}
	switch s.SolveAssumptionsLimit(x, s.limit) {
	case satoko.SATOKO_SAT:
		return bsat.Satisfiable
	case satoko.SATOKO_UNSAT:
		return bsat.Unsatisfiable
	default:
		return bsat.Unknown
	}
}
func (s *bmc3Satoko) Value(v int32) (bool, bool) {
	x := s.ModelValue(int(v))
	if x < 0 {
		return false, false
	}
	return x != 0, true
}

type bmc3Glucose struct {
	*glucose.Solver
	limit    int64
	deadline time.Time
}

func (s *bmc3Glucose) NumVars() int32     { return int32(s.NVars()) }
func (s *bmc3Glucose) NewVariable() int32 { return int32(s.NewVar(false, true)) }
func (s *bmc3Glucose) AddClause(c ...bsat.Literal) bool {
	x := make([]glucose.Lit, len(c))
	for i, v := range c {
		x[i] = glucose.Lit(v)
	}
	return s.Solver.AddClause(x)
}
func (s *bmc3Glucose) SetDeadline(d time.Time)              { s.deadline = d }
func (s *bmc3Glucose) SetLimits(conflicts, _ int64)         { s.limit = conflicts }
func (s *bmc3Glucose) SetLearnedPolicy(int32, int32, int32) {}
func (s *bmc3Glucose) SetNoRestarts(bool)                   {}
func (s *bmc3Glucose) Solve(a ...bsat.Literal) bsat.Status {
	x := make([]glucose.Lit, len(a))
	for i, v := range a {
		x[i] = glucose.Lit(v)
	}
	if s.limit > 0 {
		s.SetConfBudget(s.limit)
	} else {
		s.BudgetOff()
	}
	var timer *time.Timer
	if !s.deadline.IsZero() {
		remaining := time.Until(s.deadline)
		if remaining <= 0 {
			return bsat.Unknown
		}
		timer = time.AfterFunc(remaining, s.Interrupt)
	}
	result := s.SolveLimited(x)
	if timer != nil {
		timer.Stop()
	}
	s.ClearInterrupt()
	s.BudgetOff()
	s.deadline = time.Time{}
	switch result {
	case glucose.LTrue:
		return bsat.Satisfiable
	case glucose.LFalse:
		return bsat.Unsatisfiable
	default:
		return bsat.Unknown
	}
}
func (s *bmc3Glucose) Value(v int32) (bool, bool) {
	if s.Model == nil || int(v) >= len(s.Model) {
		return false, false
	}
	x := s.ModelValueVar(glucose.Var(v))
	if x.IsUndef() {
		return false, false
	}
	return x == glucose.LTrue, true
}
