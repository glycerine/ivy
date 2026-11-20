//go:build !js && !classic_z3

package smt

import (
	"strconv"

	"github.com/glycerine/g3/pkg/api"
	"github.com/glycerine/g3/pkg/ast"
)

type Z3CheckResult int

const (
	Sat     Z3CheckResult = 1
	Unsat   Z3CheckResult = -1
	Unknown Z3CheckResult = 0
)

func (r Z3CheckResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}

type Z3Solver struct {
	ctx    *Z3Context
	c      *api.APISimpleSolver
	params *api.ParameterSet
	logic  string
}

func (ctx *Z3Context) NewZ3Solver() *Z3Solver { return NewZ3SolverForLogic(ctx, "") }
func NewZ3SolverForLogic(ctx *Z3Context, logic string) *Z3Solver {
	var c *api.APISimpleSolver
	var err error
	if logic == "" {
		c, err = ctx.c.NewSolver()
	} else {
		c, err = ctx.c.NewSolverForLogic(logic)
	}
	fail(err)
	p, err := ctx.c.NewParameterSet()
	fail(err)
	s := &Z3Solver{ctx: ctx, c: c, params: p, logic: logic}
	ctx.solvers = append(ctx.solvers, s)
	return s
}
func (s *Z3Solver) Assert(e Z3Expr)      { fail(s.c.Assert(e.c)) }
func (s *Z3Solver) Check() Z3CheckResult { return s.CheckAssumptions(nil) }

// CheckVerificationCondition may use an independent refutation of the active
// assertions. On every inconclusive result it checks the original solver so a
// failing Ivy obligation has its concrete counterexample ready for Model.
func (s *Z3Solver) CheckVerificationCondition() Z3CheckResult {
	var result Z3CheckResult
	s.ctx.do(func() {
		proved, err := s.c.ProveUnsatByOrderAbstraction()
		fail(err)
		if proved {
			result = Unsat
			return
		}
		r, err := s.c.Check()
		fail(err)
		result = Z3CheckResult(r)
	})
	return result
}
func (s *Z3Solver) CheckAssumptions(es []Z3Expr) Z3CheckResult {
	var result Z3CheckResult
	s.ctx.do(func() {
		r, err := s.c.Check(exprHandles(es)...)
		fail(err)
		result = Z3CheckResult(r)
	})
	return result
}
func (s *Z3Solver) Push() { fail(s.c.Push()) }
func (s *Z3Solver) Pop()  { fail(s.c.Pop(1)) }
func (s *Z3Solver) Reset() {
	fail(s.c.Close())
	var err error
	if s.logic == "" {
		s.c, err = s.ctx.c.NewSolver()
	} else {
		s.c, err = s.ctx.c.NewSolverForLogic(s.logic)
	}
	fail(err)
	fail(s.c.SetParameters(s.params))
}
func (s *Z3Solver) String() string       { return textResult(s.c.String()) }
func (s *Z3Solver) assertions() []Z3Expr { return s.refs(s.c.Assertions()) }
func (s *Z3Solver) refs(refs []ast.ASTRef, err error) []Z3Expr {
	fail(err)
	es := make([]Z3Expr, len(refs))
	for i, ref := range refs {
		es[i] = s.ctx.refExpr(ref, nil)
	}
	return es
}
func (s *Z3Solver) UnsatCore() []Z3Expr   { return s.refs(s.c.UnsatCore()) }
func (s *Z3Solver) ReasonUnknown() string { return textResult(s.c.ReasonUnknown()) }
func (s *Z3Solver) ToSMT2() string {
	es := s.assertions()
	formula := s.ctx.BoolVal(true)
	var assumptions []ast.ASTHandle
	if len(es) > 0 {
		formula = es[len(es)-1]
		assumptions = exprHandles(es[:len(es)-1])
	}
	return textResult(s.ctx.c.BenchmarkSMT2(formula.c, api.APIBenchmarkConfig{Name: "benchmark generated from python API", Status: "unknown", Assumptions: assumptions}))
}
func (s *Z3Solver) SetParam(key, value string) {
	switch value {
	case "true":
		fail(s.params.SetBool(key, true))
	case "false":
		fail(s.params.SetBool(key, false))
	default:
		if v, err := strconv.ParseUint(value, 10, 32); err == nil {
			fail(s.params.SetUInt(key, uint32(v)))
		} else {
			fail(s.params.SetSymbol(key, value))
		}
	}
	fail(s.c.SetParameters(s.params))
}

type Model struct {
	ctx *Z3Context
	c   *api.APIModel
}

func (s *Z3Solver) Model() *Model {
	c, err := s.c.Model()
	if err != nil || c == nil {
		return nil
	}
	m := &Model{ctx: s.ctx, c: c}
	s.ctx.models = append(s.ctx.models, m)
	return m
}
func (m *Model) Eval(e Z3Expr, completion bool) (Z3Expr, bool) {
	ref, err := m.c.Evaluate(e.c, completion)
	if err != nil {
		return Z3Expr{}, false
	}
	return m.ctx.refExpr(ref, nil), true
}
func (m *Model) Sorts() []Z3Sort {
	n := number(m.c.NumSorts())
	ss := make([]Z3Sort, n)
	for i := range ss {
		ss[i] = m.ctx.newSort(handle(m.c.Sort(uint32(i))))
	}
	return ss
}
func (m *Model) SortUniverse(s Z3Sort) []Z3Expr {
	u, err := m.c.Universe(s.c)
	fail(err)
	if u == nil {
		return nil
	}
	defer u.Close()
	es := make([]Z3Expr, u.Size())
	for i := range es {
		es[i] = m.ctx.newExpr(handle(u.At(uint32(i))))
	}
	return es
}
func (m *Model) String() string { return textResult(m.c.String()) }
func (m *Model) Release() {
	if m != nil {
		fail(m.c.Close())
	}
}
