// Compatibility names from src/sat/lsat/solver.h; license in solver.go.
package lsat

type solver = Solver
type solver_t = Solver
type solver_Var = Variable
type solver_Lit = Literal
type solver_lbool = Truth

const (
	solver_l_True  = True
	solver_l_False = False
	solver_l_Undef = Undefined
)

func solver_new() *solver                                 { return New() }
func solver_delete(s *solver)                             { *s = Solver{} }
func solver_newVar(s *solver) solver_Var                  { return s.NewVariable() }
func solver_newLit(s *solver) solver_Lit                  { return s.NewLiteral() }
func solver_mkLit(v solver_Var) solver_Lit                { return MakeLiteral(v, false) }
func solver_mkLit_args(v solver_Var, sign int) solver_Lit { return MakeLiteral(v, sign != 0) }
func solver_negate(p solver_Lit) solver_Lit               { return Negate(p) }
func solver_var(p solver_Lit) solver_Var                  { return LiteralVariable(p) }
func solver_sign(p solver_Lit) int                        { return p & 1 }
func asInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func solver_addClause(s *solver, n int, lits []solver_Lit) int   { return asInt(s.AddClause(lits[:n])) }
func solver_addClause_begin(s *solver)                           { s.BeginClause() }
func solver_addClause_addLit(s *solver, lit solver_Lit)          { s.AddClauseLiteral(lit) }
func solver_addClause_commit(s *solver) int                      { return asInt(s.CommitClause()) }
func solver_simplify(s *solver) int                              { return asInt(s.Simplify()) }
func solver_solve(s *solver, n int, lits []solver_Lit) int       { return asInt(s.Solve(lits[:n])) }
func solver_solve_begin(s *solver)                               { s.BeginSolve() }
func solver_solve_addLit(s *solver, lit solver_Lit)              { s.AddAssumption(lit) }
func solver_solve_commit(s *solver) int                          { return asInt(s.CommitSolve()) }
func solver_okay(s *solver) int                                  { return asInt(s.Okay()) }
func solver_setPolarity(s *solver, v solver_Var, b int)          { s.SetPolarity(v, b != 0) }
func solver_setDecisionVar(s *solver, v solver_Var, b int)       { s.SetDecisionVariable(v, b != 0) }
func solver_get_l_True() solver_lbool                            { return True }
func solver_get_l_False() solver_lbool                           { return False }
func solver_get_l_Undef() solver_lbool                           { return Undefined }
func solver_value_Var(s *solver, v solver_Var) solver_lbool      { return s.Value(v) }
func solver_value_Lit(s *solver, p solver_Lit) solver_lbool      { return s.LiteralValue(p) }
func solver_modelValue_Var(s *solver, v solver_Var) solver_lbool { return s.ModelValue(v) }
func solver_modelValue_Lit(s *solver, p solver_Lit) solver_lbool { return s.ModelLiteralValue(p) }
func solver_num_assigns(s *solver) int                           { return s.AssignmentCount() }
func solver_num_clauses(s *solver) int                           { return s.ClauseCount() }
func solver_num_learnts(s *solver) int                           { return s.LearntCount() }
func solver_num_vars(s *solver) int                              { return s.VariableCount() }
func solver_num_freeVars(s *solver) int                          { return s.FreeVariableCount() }
func solver_conflict_len(s *solver) int                          { return len(s.Conflict()) }
func solver_conflict_nthLit(s *solver, i int) solver_Lit         { return s.Conflict()[i] }
func solver_set_verbosity(s *solver, v int)                      { s.SetVerbosity(v) }
func solver_num_conflicts(s *solver) int                         { return int(s.ConflictCount()) }
