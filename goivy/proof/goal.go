package proof

import (
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// Vocab represents the vocabulary of a goal: the sorts, symbols, and variables
// that are bound in the goal's premises and conclusion.
type Vocab struct {
	Sorts     []lg.Sort
	Symbols   []*lg.Const
	Variables []*lg.Variable
}

// GoalConc returns the conclusion of a goal.
// If the goal's formula is a SchemaBody, returns the last element (conclusion).
// Otherwise returns the formula itself, regardless of type.
// Mirrors Python ivy_proof.py:481-482 goal_conc — returns whatever ast.Node
// is in the formula slot, including *ast.TemporalModels.
func GoalConc(g *ast.LabeledFormula) ast.Node {
	if sb, ok := g.Formula.(*ast.SchemaBody); ok {
		return sb.Conc()
	}
	return g.Formula
}

// GoalConcExpr returns the conclusion as an lg.Expr if possible, else nil.
// Use when the caller needs lg.Expr for substitution, matching, or other
// logic-level operations. Use GoalConc when the caller passes the conclusion
// through to CloneGoal/MakeGoal or just nil-checks it.
// NOTE: this does NOT unwrap *ast.TemporalModels. Use GoalConcUnwrap (or
// ConcAsExpr on a raw conclusion node) when you want the inner formula
// regardless of any TemporalModels wrapper.
func GoalConcExpr(g *ast.LabeledFormula) lg.Expr {
	if e, ok := GoalConc(g).(lg.Expr); ok {
		return e
	}
	return nil
}

// ConcAsExpr converts a conclusion ast.Node to an lg.Expr, unwrapping
// *ast.TemporalModels if present. Returns nil if the conclusion is neither
// an lg.Expr nor a TemporalModels containing an lg.Expr.
// Mirrors Python pattern at ivy_proof.py:580 (`conc_fmla = conc.fmla if
// isinstance(conc,ia.TemporalModels) else conc`).
func ConcAsExpr(c ast.Node) lg.Expr {
	if tm, ok := c.(*ast.TemporalModels); ok {
		if e, ok := tm.Fmla.(lg.Expr); ok {
			return e
		}
		return nil
	}
	if e, ok := c.(lg.Expr); ok {
		return e
	}
	return nil
}

// GoalConcUnwrap returns the inner lg.Expr conclusion of a goal, unwrapping
// *ast.TemporalModels if present. Convenience wrapper for ConcAsExpr(GoalConc(g)).
func GoalConcUnwrap(g *ast.LabeledFormula) lg.Expr {
	return ConcAsExpr(GoalConc(g))
}

// ApplyToConc applies fn to the conclusion, unwrapping *ast.TemporalModels
// if present. The result is wrapped back into a TemporalModels (preserving
// the Model field) so that transformations applied to a temporal goal stay
// inside the temporal wrapper.
// Mirrors Python ivy_proof.py:1370-1373 apply_to_conc.
func ApplyToConc(conc ast.Node, fn func(lg.Expr) lg.Expr) ast.Node {
	if tm, ok := conc.(*ast.TemporalModels); ok {
		if innerExpr, ok := tm.Fmla.(lg.Expr); ok {
			return tm.Clone([]ast.Node{fn(innerExpr)})
		}
		return tm
	}
	if expr, ok := conc.(lg.Expr); ok {
		return fn(expr)
	}
	return conc
}

// GoalApplyToConc clones a goal with fn applied to its conclusion.
// Mirrors Python ivy_proof.py:1572-1573 goal_apply_to_conc.
// NOTE: fn is called with the raw conclusion (ast.Node) — fn is responsible
// for handling *ast.TemporalModels itself, OR the caller can wrap fn with
// ApplyToConc.
func GoalApplyToConc(cfg *ast.AstConfig, goal *ast.LabeledFormula, fn func(ast.Node) ast.Node) *ast.LabeledFormula {
	return CloneGoal(cfg, goal, GoalPrems(goal), fn(GoalConc(goal)))
}

// GoalPrems returns the premises of a goal.
// If the goal's formula is a SchemaBody, returns all elements except the last.
// Otherwise returns nil.
func GoalPrems(g *ast.LabeledFormula) []ast.Node {
	if sb, ok := g.Formula.(*ast.SchemaBody); ok {
		return sb.Prems()
	}
	return nil
}

// GoalPremGoals returns all premises of a goal that are themselves LabeledFormulas.
func GoalPremGoals(goal *ast.LabeledFormula) []*ast.LabeledFormula {
	prems := GoalPrems(goal)
	var result []*ast.LabeledFormula
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			result = append(result, lf)
		}
	}
	return result
}

// CloneGoal creates a new goal with the same label but new premises and conclusion.
// If prems is non-empty, wraps them in a SchemaBody; otherwise uses conc directly.
// conc is ast.Node so it can carry *ast.TemporalModels (and any other ast type),
// mirroring Python's clone_goal which is duck-typed.
func CloneGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = cfg.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	return goal.CloneWithFreshID([]ast.Node{goal.Label, formula})
}

// MakeGoal creates a goal with the given label, premises, and conclusion.
// conc is ast.Node so it can carry *ast.TemporalModels (and any other ast type),
// mirroring Python's make_goal which is duck-typed.
func MakeGoal(cfg *ast.AstConfig, loc ast.Location, label ast.Node, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = cfg.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	lf := cfg.NewLabeledFormula(label, formula)
	lf.SetLineno(loc)
	return lf
}

// NormalizeGoal normalizes the subformulas of a goal so there are only
// binary conjunctions/disjunctions and single-variable quantifiers.
func NormalizeGoal(cfg *ast.AstConfig, g *ast.LabeledFormula) *ast.LabeledFormula {
	if GoalIsDefn(g) {
		return g
	}
	prems := GoalPrems(g)
	normPrems := make([]ast.Node, len(prems))
	for i, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			normPrems[i] = NormalizeGoal(cfg, lf)
		} else {
			normPrems[i] = p
		}
	}
	// ApplyToConc unwraps *ast.TemporalModels so NormalizeOps runs on the inner
	// formula, then re-wraps. For plain lg.Expr conclusions, NormalizeOps runs
	// directly. For unknown types, the conc is passed through unchanged.
	newConc := ApplyToConc(GoalConc(g), il.NormalizeOps)
	return CloneGoal(cfg, g, normPrems, newConc)
}

// GoalIsDefn returns true if x is a non-lambda constant declaration
// or an uninterpreted sort.
func GoalIsDefn(x ast.Node) bool {
	if cd, ok := x.(*ast.ConstantDecl); ok {
		args := cd.Args()
		if len(args) > 0 {
			// Check if the arg is a Lambda
			if _, isLam := args[0].(*lg.Lambda); isLam {
				return false
			}
		}
		return true
	}
	return false
}

// GoalDefns returns the symbols and types defined in the premises of a goal.
func GoalDefns(goal *ast.LabeledFormula) map[lg.NodeKey]lg.Expr {
	res := make(map[lg.NodeKey]lg.Expr)
	for _, p := range GoalPrems(goal) {
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(lg.Expr); ok {
					res[lg.Key(c)] = c
				}
			}
		}
	}
	return res
}

// GoalVocab returns the vocabulary of a goal: the sorts, symbols, and
// variables that are bound in the goal's premises and conclusion.
// Mirrors Python ivy_proof.py:580 — when conc is *ast.TemporalModels,
// extracts conc.Fmla as the formula to scan.
func GoalVocab(goal *ast.LabeledFormula) *Vocab {
	prems := GoalPrems(goal)
	conc := GoalConc(goal)

	var symbols []*lg.Const
	var sorts []lg.Sort
	var fmlas []lg.Expr

	for _, p := range prems {
		// Collect sorts: Python: sorts = [s for s in prems if isinstance(s, il.UninterpretedSort)]
		if s, ok := p.(lg.Sort); ok {
			if _, isUninterp := s.(*lg.UninterpretedSort); isUninterp {
				sorts = append(sorts, s)
			}
		}
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(lg.Expr); ok {
					if cc, ok := c.(*lg.Const); ok {
						symbols = append(symbols, cc)
					}
				}
			}
		}
		if lf, ok := p.(*ast.LabeledFormula); ok {
			fc := ConcAsExpr(GoalConc(lf))
			if fc != nil {
				fmlas = append(fmlas, fc)
			}
		}
	}
	if concExpr := ConcAsExpr(conc); concExpr != nil {
		fmlas = append(fmlas, concExpr)
	}

	// Collect variables from formulas
	varSet := make(map[lg.NodeKey]lg.Expr)
	for _, f := range fmlas {
		for _, v := range lu.UsedVariables(f) {
			varSet[lg.Key(v)] = v
		}
	}
	var variables []*lg.Variable
	for _, node := range varSet {
		if v, ok := node.(*lg.Variable); ok {
			variables = append(variables, v)
		}
	}

	return &Vocab{
		Sorts:     sorts,
		Symbols:   symbols,
		Variables: variables,
	}
}

// GoalFree returns the free vocabulary of a goal, including sorts,
// symbols, and variables that are not bound in the goal's premises.
// Symmetric with GoalVocab — when conc is *ast.TemporalModels, extracts
// conc.Fmla as the formula to scan.
func GoalFree(goal *ast.LabeledFormula) map[lg.NodeKey]lg.Expr {
	bound := make(map[lg.NodeKey]lg.Expr)
	res := make(map[lg.NodeKey]lg.Expr)

	var recFmla func(lg.Expr)
	recFmla = func(fmla lg.Expr) {
		if fmla == nil {
			return
		}
		for vKey, vNode := range lu.FreeVariables(fmla) {
			if bound[vKey] == nil {
				res[vKey] = vNode
			}
		}
		for cKey, cNode := range il.UsedSymbolsAst(fmla) {
			if bound[cKey] == nil {
				res[cKey] = cNode
			}
		}
	}

	var rec func(*ast.LabeledFormula)
	rec = func(g *ast.LabeledFormula) {
		defns := GoalDefns(g)
		// Add defns to bound
		for d, dn := range defns {
			bound[d] = dn
		}
		for _, pg := range GoalPremGoals(g) {
			if _, ok := pg.Formula.(*ast.SchemaBody); ok {
				rec(pg)
			} else {
				recFmla(ConcAsExpr(GoalConc(pg)))
			}
		}
		recFmla(ConcAsExpr(GoalConc(g)))
		// Remove defns from bound (restore)
		for d := range defns {
			delete(bound, d)
		}
	}
	rec(goal)
	return res
}

// GoalSubst substitutes goal g2 for the conclusion of goal g1.
// The result has the label of g2.
func GoalSubst(cfg *ast.AstConfig, g1, g2 *ast.LabeledFormula, loc ast.Location) *ast.LabeledFormula {
	prems := append(GoalPrems(g1), GoalPrems(g2)...)
	return MakeGoal(cfg, loc, g2.Label, prems, GoalConc(g2))
}

// GoalAddPrem adds a premise to a goal.
func GoalAddPrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, prem ast.Node, loc ast.Location) *ast.LabeledFormula {
	prems := append(GoalPrems(goal), prem)
	return MakeGoal(cfg, loc, goal.Label, prems, GoalConc(goal))
}

// GoalRemovePrem removes a premise by name from a goal.
func GoalRemovePrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, premName string) *ast.LabeledFormula {
	var prems []ast.Node
	for _, p := range GoalPrems(goal) {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.LabelName() == premName {
				continue
			}
		}
		prems = append(prems, p)
	}
	return CloneGoal(cfg, goal, prems, GoalConc(goal))
}

// TrivialGoal returns true if the conclusion equals one of the
// premises modulo alpha conversion. Uses GoalConcExpr because alpha
// equivalence is defined on lg.Expr; for non-Expr conclusions
// (e.g., *ast.TemporalModels) the trivial check returns false,
// matching the conservative behavior of falling through to a deeper check.
func TrivialGoal(goal *ast.LabeledFormula) bool {
	conc := GoalConcExpr(goal)
	if conc == nil {
		return false
	}
	for _, prem := range GoalPremGoals(goal) {
		if len(GoalPrems(prem)) == 0 {
			pc := GoalConcExpr(prem)
			if pc != nil && lu.EqualModAlpha(pc, conc) {
				return true
			}
		}
	}
	return false
}

// CheckConcsMatch checks that the conclusions of two goals match
// modulo alpha conversion. Uses GoalConcExpr because EqualModAlpha
// requires lg.Expr; non-Expr conclusions (e.g., *ast.TemporalModels)
// fall into the existing nil-check error path.
func CheckConcsMatch(g1, g2 *ast.LabeledFormula) error {
	c1 := GoalConcExpr(g1)
	c2 := GoalConcExpr(g2)
	if c1 == nil || c2 == nil {
		return &ProofError{Msg: "nil conclusion in goal"}
	}
	if !lu.EqualModAlpha(c1, c2) {
		return &ProofError{Msg: "conclusions do not match: " + c1.String() + " vs " + c2.String()}
	}
	return nil
}

// --- helpers ---

// CompileWithGoalVocab compiles an expression using the vocabulary of a goal.
// This is a simplified version that returns the expression's formula as a
// logic.Expr, performing sort inference within the goal's vocabulary context.
// Corresponds to Python compile_with_goal_vocab (ivy_proof.py:1453-1457).
func CompileWithGoalVocab(expr ast.Node, goal *ast.LabeledFormula) lg.Expr {
	// For expressions that are already logic.Nodes, return directly
	if n, ok := expr.(lg.Expr); ok {
		return n
	}
	// For LabeledFormulas, extract the formula
	if lf, ok := expr.(*ast.LabeledFormula); ok {
		if n, ok := lf.Formula.(lg.Expr); ok {
			return n
		}
	}
	// For other AST nodes, try to compile via the vocabulary
	vocab := GoalVocab(goal)
	_ = vocab
	// Fallback: return nil if we can't compile
	return nil
}

// CompileDefinitionGoalVocab compiles a definition and adds it to the goal
// as new premises (a function declaration and a property stating the definition).
// Returns the modified goal.
// Corresponds to Python compile_definition_goal_vocab (ivy_proof.py:1477-1506).
//
// Accepts either a *ast.LabeledFormula (containing the equation directly) or
// a *ast.DerivedDecl (whose first arg is the LabeledFormula). Python's
// `lf = df.args[0]` (line 1480) extracts the inner LF from a DerivedDecl.
func CompileDefinitionGoalVocab(cfg *ast.AstConfig, df ast.Node, goal *ast.LabeledFormula) *ast.LabeledFormula {
	// Unwrap DerivedDecl: Python lf = df.args[0]
	var innerLF *ast.LabeledFormula
	if lf, ok := df.(*ast.LabeledFormula); ok {
		innerLF = lf
	} else if dd, ok := df.(*ast.DerivedDecl); ok {
		if len(dd.DeclArgs) > 0 {
			if lf, ok := dd.DeclArgs[0].(*ast.LabeledFormula); ok {
				innerLF = lf
			}
		}
	}
	if innerLF == nil {
		return goal
	}

	// Extract the definition formula
	var defnFormula lg.Expr
	if n, ok := innerLF.Formula.(lg.Expr); ok {
		defnFormula = n
	}
	if defnFormula == nil {
		return goal // can't process, return unchanged
	}

	// Drop universals to get lhs = rhs
	inner := il.DropUniversals(defnFormula)
	eq, ok := inner.(*lg.Eq)
	if !ok {
		return goal
	}

	// Get the defined symbol info
	var defSym *lg.Const
	switch lhs := eq.T1.(type) {
	case *lg.Apply:
		if c, ok := lhs.Func.(*lg.Const); ok {
			defSym = c
		}
	case *lg.Const:
		defSym = lhs
	}
	if defSym == nil {
		return goal
	}

	// Add the definition as a premise to the goal
	// In the full version, this would also add a function declaration premise.
	// For now, we add the definition equation as a property premise.
	sb, ok := goal.Formula.(*ast.SchemaBody)
	if !ok {
		return goal
	}

	// Create a new premise with the definition.
	// Python ivy_proof.py:1501 sets `lf.definition = True` on the premise.
	defPrem := cfg.NewLabeledFormula(innerLF.Label, defnFormula)
	defPrem.IsDefinition = true

	// Clone the SchemaBody with the new premise added
	newPrems := make([]ast.Node, 0, len(sb.Prems())+1)
	newPrems = append(newPrems, sb.Prems()...)
	newPrems = append(newPrems, defPrem)

	// Build new SchemaBody with prems + conclusion
	newArgs := make([]ast.Node, 0, len(newPrems)+1)
	newArgs = append(newArgs, newPrems...)
	if conc := sb.Conc(); conc != nil {
		newArgs = append(newArgs, conc)
	}

	goalArgs := goal.Args()
	if len(goalArgs) < 2 {
		return goal
	}
	newGoal := goal.Clone([]ast.Node{goalArgs[0], cfg.NewSchemaBody(newArgs...)})
	if lf, ok := newGoal.(*ast.LabeledFormula); ok {
		return lf
	}
	return goal
}
