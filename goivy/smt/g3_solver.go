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
	c, err := newG3Solver(ctx.c, logic)
	fail(err)
	p, err := ctx.c.NewParameterSet()
	fail(err)
	s := &Z3Solver{ctx: ctx, c: c, params: p, logic: logic}
	ctx.solvers = append(ctx.solvers, s)
	return s
}

// GoIvy also builds against the pinned g3 v0.2.0. Newer g3 contexts defer
// the unused default tactic portfolio when constructing verification solvers.
func newG3Solver(ctx *api.SMTContext, logic string) (*api.APISimpleSolver, error) {
	if logic != "" {
		return ctx.NewSolverForLogic(logic)
	}
	if factory, ok := any(ctx).(interface {
		NewSolverForIncrementalChecks() (*api.APISimpleSolver, error)
	}); ok {
		return factory.NewSolverForIncrementalChecks()
	}
	return ctx.NewSolver()
}

func (s *Z3Solver) Assert(e Z3Expr)      { fail(s.c.Assert(e.c)) }
func (s *Z3Solver) Check() Z3CheckResult { return s.CheckAssumptions(nil) }

// SimplifyVerificationCondition eliminates equality-defined quantifier variables
// before the incremental push. Older g3 versions provide existential reduction.
func (s *Z3Solver) SimplifyVerificationCondition(e Z3Expr) Z3Expr {
	if rewriter, ok := any(s.ctx.c).(interface {
		SimplifyVerificationCondition(ast.ASTHandle) (ast.ASTRef, error)
	}); ok {
		return s.ctx.refExpr(rewriter.SimplifyVerificationCondition(e.c))
	}
	rewriter, ok := any(s.ctx.c).(interface {
		EliminateExistentialDefinitions(ast.ASTHandle) (ast.ASTRef, error)
	})
	if !ok {
		return e
	}
	return s.ctx.refExpr(rewriter.EliminateExistentialDefinitions(e.c))
}

// CheckVerificationCondition may use an independent refutation of the active
// assertions. On every inconclusive result it checks the original solver so a
// failing Ivy obligation has its concrete counterexample ready for Model.
func (s *Z3Solver) CheckVerificationCondition() Z3CheckResult {
	return s.checkVerificationCondition(s.c.ProveUnsatByOrderAbstraction)
}

// BeginVerificationChecks uses one private proof kernel for the current set
// of Ivy final conditions. Release it when the checker leaves that set.
func (s *Z3Solver) BeginVerificationChecks() (check func() Z3CheckResult, release func()) {
	batch := s.c.BeginOrderProofBatch()
	return func() Z3CheckResult { return s.checkVerificationCondition(batch.Prove) },
		func() { fail(batch.Close()) }
}

// BeginVerificationGoalChecks can refute the current base plus a temporary
// goal before the first incremental push. Older g3 releases support only
// proofs over assertions already in the solver.
func (s *Z3Solver) BeginVerificationGoalChecks() (check func() Z3CheckResult, proveGoal func(Z3Expr) bool, release func()) {
	return s.BeginVerificationGoalBatch(0)
}

// BeginVerificationGoalBatch supplies the number of sibling conditions already
// available to the verifier. Older g3 releases use an unsized proof batch.
func (s *Z3Solver) BeginVerificationGoalBatch(count int) (check func() Z3CheckResult, proveGoal func(Z3Expr) bool, release func()) {
	if s.logic != "" {
		return nil, nil, nil
	}
	var batch *api.OrderProofBatch
	if sized, ok := any(s.c).(interface {
		BeginOrderProofBatchForGoals(int) *api.OrderProofBatch
	}); ok {
		batch = sized.BeginOrderProofBatchForGoals(count)
	} else {
		batch = s.c.BeginOrderProofBatch()
	}
	withGoal, ok := any(batch).(interface {
		ProveWithGoal(ast.ASTHandle) (bool, error)
	})
	if !ok {
		fail(batch.Close())
		return nil, nil, nil
	}
	return func() Z3CheckResult { return s.checkVerificationCondition(batch.Prove) },
		func(goal Z3Expr) bool {
			var proved bool
			s.ctx.do(func() {
				yes, err := withGoal.ProveWithGoal(goal.c)
				proved = yes && err == nil
			})
			return proved
		}, func() { fail(batch.Close()) }
}

func (s *Z3Solver) checkVerificationCondition(prove func() (bool, error)) Z3CheckResult {
	var result Z3CheckResult
	s.ctx.do(func() {
		// The optional solver can exhaust its own setup resources even when
		// the concrete solver already has everything it needs to decide.
		if proved, err := prove(); err == nil && proved {
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
	s.c, err = newG3Solver(s.ctx.c, s.logic)
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
