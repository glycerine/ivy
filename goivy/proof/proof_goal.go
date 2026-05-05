package proof

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/compiler"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
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
	if tm, ok := c.(*ast.AstTemporalModels); ok {
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
	if conc == nil {
		xtracer.Trace("proof.ApplyToConc ENTER concNil=true")
		return conc
	}
	xtracer.Trace("proof.ApplyToConc ENTER type=%s HASH canon=%v", iu.TypeName(conc), conc.Canon())
	if tm, ok := conc.(*ast.AstTemporalModels); ok {
		if innerExpr, ok := tm.Fmla.(lg.Expr); ok {
			result := tm.Clone([]ast.Node{fn(innerExpr)})
			xtracer.Trace("proof.ApplyToConc EXIT type=TemporalModels HASH canon=%v", result.Canon())
			return result
		}
		xtracer.Trace("proof.ApplyToConc EXIT type=TemporalModels passthrough")
		return tm
	}
	if expr, ok := conc.(lg.Expr); ok {
		result := fn(expr)
		if result != nil {
			xtracer.Trace("proof.ApplyToConc EXIT type=Expr HASH canon=%v", result.Canon())
		} else {
			xtracer.Trace("proof.ApplyToConc EXIT type=Expr result=nil")
		}
		return result
	}
	xtracer.Trace("proof.ApplyToConc EXIT type=%s passthrough", iu.TypeName(conc))
	return conc
}

// normalizeOpsConc mirrors Python's `il.normalize_ops(goal_conc(x))` in
// ivy_proof.py:770 normalize_goal. Python uses duck typing (.args/.clone)
// so normalize_ops handles TemporalModels transparently. In Go,
// il.NormalizeOps is typed func(lg.Expr) lg.Expr and cannot accept
// *ast.TemporalModels, so we unwrap/rewrap it here. Emits no xtrace —
// Python's normalize_goal does not call apply_to_conc.
func normalizeOpsConc(conc ast.Node) ast.Node {
	if conc == nil {
		return conc
	}
	if tm, ok := conc.(*ast.AstTemporalModels); ok {
		if inner, ok := tm.Fmla.(lg.Expr); ok {
			return tm.Clone([]ast.Node{il.NormalizeOps(inner)})
		}
		return tm
	}
	if expr, ok := conc.(lg.Expr); ok {
		return il.NormalizeOps(expr)
	}
	return conc
}

// WrapImplies mirrors Python's il.Implies(cond, formula) duck-typed wrapping
// used by if_tactic and let_tactic (ivy_proof.py:226, :414, :416). It wraps
// the ENTIRE formula in Implies(cond, formula) without descending into any
// *ast.TemporalModels or *ast.SchemaBody wrapper.
//
// Python's il.Implies accepts any formula as its second argument. Go's
// lg.Implies requires lg.Expr fields, so when formula is not an lg.Expr
// (i.e. *ast.TemporalModels or *ast.SchemaBody) we fall back to ast.Implies
// which has ast.Node fields. The common case (formula is lg.Expr) still
// produces *lg.Implies so downstream *lg.Implies type-switches keep matching.
func WrapImplies(cfg *ast.AstConfig, cond lg.Expr, formula ast.Node) ast.Node {
	xtracer.Trace("proof.WrapImplies ENTER formulaType=%s", iu.TypeName(formula))
	if e, ok := formula.(lg.Expr); ok {
		result := &lg.Implies{T1: cond, T2: e}
		xtracer.Trace("proof.WrapImplies EXIT type=lgImplies HASH canon=%v", result.Canon())
		return result
	}
	result := cfg.NewImplies(cond, formula)
	xtracer.Trace("proof.WrapImplies EXIT type=astImplies HASH canon=%v", result.Canon())
	return result
}

// GoalApplyToConc clones a goal with fn applied to its conclusion.
// Mirrors Python ivy_proof.py:1572-1573 goal_apply_to_conc.
// NOTE: fn is called with the raw conclusion (ast.Node) — fn is responsible
// for handling *ast.TemporalModels itself, OR the caller can wrap fn with
// ApplyToConc.
func GoalApplyToConc(cfg *ast.AstConfig, goal *ast.LabeledFormula, fn func(ast.Node) ast.Node) *ast.LabeledFormula {
	xtracer.Trace("proof.GoalApplyToConc ENTER label=%s", goal.LabelForTrace())
	result := CloneGoal(cfg, goal, GoalPrems(goal), fn(GoalConc(goal)))
	xtracer.Trace("proof.GoalApplyToConc EXIT HASH canon=%v", result.Canon())
	return result
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
	xtracer.Trace("proof.CloneGoal ENTER label=%s nprems=%d concType=%s", goal.LabelForTrace(), len(prems), iu.TypeName(conc))
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = cfg.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	result := goal.CloneWithFreshID([]ast.Node{goal.Label, formula})
	xtracer.Trace("proof.CloneGoal EXIT label=%s newID=%d", result.LabelForTrace(), result.ID)
	return result
}

// CloneGoalPreserveID is like CloneGoal but preserves the original goal's LF ID.
// Corresponds to Python's x.clone([x.label, fmla]) pattern (ivy_proof.py:988).
func CloneGoalPreserveID(cfg *ast.AstConfig, goal *ast.LabeledFormula, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	xtracer.Trace("proof.CloneGoalPreserveID ENTER label=%s nprems=%d concType=%s id=%d", goal.LabelForTrace(), len(prems), iu.TypeName(conc), goal.ID)
	var formula ast.Node
	if len(prems) > 0 {
		elems := make([]ast.Node, len(prems)+1)
		copy(elems, prems)
		elems[len(prems)] = conc
		formula = cfg.NewSchemaBody(elems...)
	} else {
		formula = conc
	}
	result := goal.Clone([]ast.Node{goal.Label, formula}).(*ast.LabeledFormula)
	xtracer.Trace("proof.CloneGoalPreserveID EXIT label=%s", result.LabelForTrace())
	return result
}

// MakeGoal creates a goal with the given label, premises, and conclusion.
// conc is ast.Node so it can carry *ast.TemporalModels (and any other ast type),
// mirroring Python's make_goal which is duck-typed.
func MakeGoal(cfg *ast.AstConfig, loc ast.Location, label ast.Node, prems []ast.Node, conc ast.Node) *ast.LabeledFormula {
	xtracer.Trace("proof.MakeGoal ENTER nprems=%d concType=%s", len(prems), iu.TypeName(conc))
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
	xtracer.Trace("proof.MakeGoal EXIT id=%d", lf.ID)
	return lf
}

// NormalizeGoal normalizes the subformulas of a goal so there are only
// binary conjunctions/disjunctions and single-variable quantifiers.
// Public API stays concrete (LF in, LF out). The actual body lives in
// normalizeGoalAny, which Python-style accepts any ast.Node so it can
// handle non-LabeledFormula premises (ConstantDecl / UninterpretedSort)
// via the goal_is_defn passthrough. Proof that the cast is safe:
// top-level callers always pass an LF, and for an LF input the body
// either hits GoalIsDefn (unreachable for LF — it matches only
// ConstantDecl / UninterpretedSort) and returns the LF unchanged, or
// falls through to CloneGoal which is typed to return LF.
func NormalizeGoal(cfg *ast.AstConfig, g *ast.LabeledFormula) *ast.LabeledFormula {
	return normalizeGoalAny(cfg, g).(*ast.LabeledFormula)
}

// normalizeGoalAny mirrors Python ivy_proof.normalize_goal exactly.
// Emits exactly one ENTER xtrace per invocation. Non-LF inputs
// (ConstantDecl, UninterpretedSort) are handled by the GoalIsDefn
// passthrough; LF inputs flow through to CloneGoal.
func normalizeGoalAny(cfg *ast.AstConfig, x ast.Node) ast.Node {
	xtracer.Trace("proof.NormalizeGoal ENTER label=%s", normalizeGoalLabel(x))
	if GoalIsDefn(x) {
		xtracer.Trace("proof.NormalizeGoal EXIT passthrough=isDefn")
		return x
	}
	// After GoalIsDefn early-exit, Python requires x to expose
	// .formula / .label / .clone_with_fresh_id — i.e. an LF.
	g := x.(*ast.LabeledFormula)
	prems := GoalPrems(g)
	normPrems := make([]ast.Node, len(prems))
	for i, p := range prems {
		normPrems[i] = normalizeGoalAny(cfg, p)
	}
	// Mirror Python ivy_proof.py:770: `il.normalize_ops(goal_conc(x))`.
	// Uses the private normalizeOpsConc helper (not ApplyToConc) so no
	// spurious ApplyToConc xtrace is emitted here — Python's normalize_goal
	// does not call apply_to_conc.
	newConc := normalizeOpsConc(GoalConc(g))
	result := CloneGoal(cfg, g, normPrems, newConc)
	xtracer.Trace("proof.NormalizeGoal EXIT HASH canon=%v", result.Canon())
	return result
}

// normalizeGoalLabel mirrors Python's
//
//	(x.label if hasattr(x,'label') else 'N/A')
//
// for the ENTER xtrace. Returns "N/A" for non-LabeledFormula inputs so
// the cross-language trace lines match.
func normalizeGoalLabel(x ast.Node) string {
	if lf, ok := x.(*ast.LabeledFormula); ok {
		return lf.LabelForTrace()
	}
	return "N/A"
}

// GoalIsDefn returns true if x is a non-lambda constant declaration
// or an uninterpreted sort.
//
// Go-specific note: Go's compile_schema_prem wraps compiled values in
// *ast.CompiledNode (a Go-only adapter) where Python returns bare
// il.UninterpretedSort / a ConstantDecl whose arg is a bare ivy_logic
// symbol. We unwrap CompiledNode here so the isinstance-style checks
// match Python's goal_is_defn.
func GoalIsDefn(x ast.Node) bool {
	x = unwrapCompiledNode(x)
	if cd, ok := x.(*ast.ConstantDecl); ok {
		args := cd.Args()
		if len(args) > 0 {
			// Check if the arg is a Lambda (unwrapping Go's CompiledNode
			// around the compiled symbol / definition body).
			if _, isLam := unwrapCompiledNode(args[0]).(*lg.Lambda); isLam {
				return false
			}
		}
		return true
	}
	// Python: return isinstance(x, il.UninterpretedSort)
	if _, ok := x.(*lg.UninterpretedSort); ok {
		return true
	}
	return false
}

// unwrapCompiledNode returns the inner ast.Node of a *ast.CompiledNode
// wrapper, or x unchanged if it isn't one. Python has no CompiledNode —
// compile_schema_prem there returns bare il.UninterpretedSort and bare
// symbols. This helper bridges the port so type-dispatch checks see
// through Go's wrapper.
func unwrapCompiledNode(x ast.Node) ast.Node {
	if cn, ok := x.(*ast.CompiledNode); ok {
		if inner, ok := cn.Node.(ast.Node); ok {
			return inner
		}
	}
	return x
}

// GoalDefns returns the symbols and types defined in the premises of a goal.
// Corresponds to Python goal_defns (ivy_proof.py:540-547).
func GoalDefns(goal *ast.LabeledFormula) map[lg.NodeKey]lg.Expr {
	res := make(map[lg.NodeKey]lg.Expr)
	for _, p := range GoalPrems(goal) {
		p = unwrapCompiledNode(p)
		// Python: isinstance(x, ia.ConstantDecl) and isinstance(x.args[0], il.Symbol)
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := unwrapCompiledNode(args[0]).(*lg.Const); ok {
					res[lg.Key(c)] = c
				}
			}
		}
		// Python: elif isinstance(x, il.UninterpretedSort): res.add(x)
		if us, ok := p.(*lg.UninterpretedSort); ok {
			res[lg.Key(us)] = us
		}
	}
	return res
}

// GoalVocab returns the vocabulary of a goal: the sorts, symbols, and FREE
// variables in the goal's premises and conclusion. Mirrors Python
// `goal_vocab(goal, bound=False)` (ivy_proof.py:572-583) which uses
// `lu.used_variables_asts = apply_gen_to_list(variables_ast)` — and
// `variables_ast` (ivy_logic_utils.py:523-535) yields ONLY variables that
// are NOT bound by any enclosing binder in each fmla.
//
// IMPORTANT: returning all variables (including bound) is incorrect —
// callers that need bound vars (e.g. compilation of terms that reference
// them) should use `GoalVocabBound` or augment locally (see
// `CompileMatchList` which adds used_variables(conc) when allow_witness=True).
// Putting bound vars into the default vocab corrupts `prob.FreeSyms`
// because `buildMatchProblem` copies `vocab.Variables` there; downstream
// `assume_tactic`'s iswit check then misclassifies bound-variable match
// keys. See the `ifabric_rw_fair_ax` divergence in log.golden.2hr at index
// 2411549 for the failure mode this mismatch caused.
func GoalVocab(goal *ast.LabeledFormula) *Vocab {
	xtracer.Trace("proof.GoalVocab ENTER label=%s", goal.LabelForTrace())
	v := goalVocabRaw(goal)
	xtracer.Trace("proof.GoalVocab EXIT nsorts=%d nsymbols=%d nvariables=%d", len(v.Sorts), len(v.Symbols), len(v.Variables))
	return v
}

// goalVocabRaw builds the vocab without emitting any xtrace — for use
// by callers that emit their own trace wrapper (GoalVocabBound).
// Python: goal_vocab(goal, bound=False) computes the same three lists.
func goalVocabRaw(goal *ast.LabeledFormula) *Vocab {
	prems := GoalPrems(goal)
	conc := GoalConc(goal)

	var symbols []*lg.Const
	var sorts []lg.Sort
	var fmlas []lg.Expr

	for _, p := range prems {
		p = unwrapCompiledNode(p)
		// Python: sorts = [s for s in prems if isinstance(s, il.UninterpretedSort)]
		if us, ok := p.(*lg.UninterpretedSort); ok {
			sorts = append(sorts, us)
		}
		// Python: symbols = [x.args[0] for x in prems if isinstance(x, ia.ConstantDecl)]
		if cd, ok := p.(*ast.ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if cc, ok := unwrapCompiledNode(args[0]).(*lg.Const); ok {
					symbols = append(symbols, cc)
				}
			}
		}
		// Python: fmlas = [x.formula for x in prems if isinstance(x, ia.LabeledFormula)]
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if f, ok := lf.Formula.(lg.Expr); ok {
				fmlas = append(fmlas, f)
			}
		}
	}
	if concExpr := ConcAsExpr(conc); concExpr != nil {
		fmlas = append(fmlas, concExpr)
	}

	// Python: variables = list(lu.used_variables_asts(fmlas))
	// used_variables_asts = apply_gen_to_list(variables_ast) — yields each
	// free-variable OCCURRENCE (not deduped). Count / order match Python.
	variables := lu.UsedVariablesAsts(fmlas)

	return &Vocab{
		Sorts:     sorts,
		Symbols:   symbols,
		Variables: variables,
	}
}

// GoalVocabBound is like GoalVocab but also includes bound variables from
// the conclusion, matching Python's goal_vocab(goal, bound=True)
// (ivy_proof.py:579-582).
func GoalVocabBound(goal *ast.LabeledFormula) *Vocab {
	xtracer.Trace("proof.GoalVocabBound ENTER label=%s", goal.LabelForTrace())
	// Python: goal_vocab(goal, bound=True) — single function, one trace pair.
	// Use goalVocabRaw to avoid emitting the inner GoalVocab ENTER/EXIT.
	v := goalVocabRaw(goal)

	// Python: conc_fmla = conc.fmla if isinstance(conc,ia.TemporalModels) else conc
	conc := GoalConc(goal)
	concFmla := ConcAsExpr(conc)
	if concFmla == nil {
		xtracer.Trace("proof.GoalVocabBound EXIT concNil nvariables=%d", len(v.Variables))
		return v
	}

	// Python: variables = variables + [x for x in logic_util.bound_variables(conc_fmla)
	//                                  if x not in variables]
	existing := make(map[lg.NodeKey]bool, len(v.Variables))
	for _, vr := range v.Variables {
		existing[lg.Key(vr)] = true
	}
	for bk, bn := range lu.BoundVariables(concFmla) {
		if !existing[bk] {
			if bv, ok := bn.(*lg.Variable); ok {
				v.Variables = append(v.Variables, bv)
				existing[bk] = true
			}
		}
	}
	xtracer.Trace("proof.GoalVocabBound EXIT nvariables=%d", len(v.Variables))
	return v
}

// GoalFree returns the free vocabulary of a goal, including sorts,
// symbols, and variables that are not bound in the goal's premises.
// Symmetric with GoalVocab — when conc is *ast.TemporalModels, extracts
// conc.Fmla as the formula to scan.
func GoalFree(goal *ast.LabeledFormula) map[lg.NodeKey]lg.Expr {
	xtracer.Trace("proof.GoalFree ENTER label=%s", goal.LabelForTrace())
	bound := make(map[lg.NodeKey]lg.Expr)
	res := make(map[lg.NodeKey]lg.Expr)

	// Python fmla_vocab: collects used_sorts_ast, used_symbols_ast, used_variables_ast
	var recFmla func(lg.Expr)
	recFmla = func(fmla lg.Expr) {
		if fmla == nil {
			return
		}
		// Python: lu.used_sorts_ast(fmla)
		for sKey, s := range lu.SortsAst(fmla) {
			if sExpr, ok := s.(lg.Expr); ok {
				if bound[sKey] == nil {
					res[sKey] = sExpr
				}
			}
		}
		// Python: lu.used_symbols_ast(fmla)
		for cKey, cNode := range il.UsedSymbolsAst(fmla).All() {
			if bound[cKey] == nil {
				res[cKey] = cNode
			}
		}
		// Python: lu.used_variables_ast(fmla)
		for vKey, vNode := range lu.FreeVariables(fmla).All() {
			if bound[vKey] == nil {
				res[vKey] = vNode
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
	{
		items := make([]string, 0, len(res))
		for _, v := range res {
			items = append(items, fmt.Sprint(v))
		}
		sort.Strings(items)
		xtracer.Trace("proof.GoalFree items=[%s]", strings.Join(items, ","))
	}
	xtracer.Trace("proof.GoalFree EXIT nfree=%d", len(res))
	return res
}

// GoalSubst substitutes goal g2 for the conclusion of goal g1.
// The result has the label of g2.
// Python ivy_proof.py:506-508: calls check_name_clash(g1,g2) before substituting.
func GoalSubst(cfg *ast.AstConfig, g1, g2 *ast.LabeledFormula, loc ast.Location) (*ast.LabeledFormula, error) {
	xtracer.Trace("proof.GoalSubst ENTER g1Label=%s g2Label=%s", g1.LabelForTrace(), g2.LabelForTrace())
	if err := CheckNameClash(g1, g2); err != nil {
		xtracer.Trace("proof.GoalSubst EXIT err=%v", err)
		return nil, err
	}
	prems := append(GoalPrems(g1), GoalPrems(g2)...)
	result := MakeGoal(cfg, loc, g2.Label, prems, GoalConc(g2))
	xtracer.Trace("proof.GoalSubst EXIT HASH canon=%v", result.Canon())
	return result, nil
}

// GoalAddPrem adds a premise to a goal.
func GoalAddPrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, prem ast.Node, loc ast.Location) *ast.LabeledFormula {
	xtracer.Trace("proof.GoalAddPrem ENTER goalLabel=%s premType=%s", goal.LabelForTrace(), iu.TypeName(prem))
	prems := append(GoalPrems(goal), prem)
	result := MakeGoal(cfg, loc, goal.Label, prems, GoalConc(goal))
	xtracer.Trace("proof.GoalAddPrem EXIT HASH canon=%v", result.Canon())
	return result
}

// GoalRemovePrem removes a premise by name from a goal.
func GoalRemovePrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, premName string) *ast.LabeledFormula {
	xtracer.Trace("proof.GoalRemovePrem ENTER goalLabel=%s premName=%s", goal.LabelForTrace(), premName)
	var prems []ast.Node
	for _, p := range GoalPrems(goal) {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.LabelName() == premName {
				continue
			}
		}
		prems = append(prems, p)
	}
	result := CloneGoal(cfg, goal, prems, GoalConc(goal))
	xtracer.Trace("proof.GoalRemovePrem EXIT nprems=%d", len(prems))
	return result
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
// Corresponds to Python compile_with_goal_vocab (ivy_proof.py:1461-1465):
//
//	the_goal_vocab = goal_vocab(goal)
//	return compile_expr_vocab_ext(expr, the_goal_vocab)
func CompileWithGoalVocab(expr ast.Node, goal *ast.LabeledFormula, mod *module.Module) lg.Expr {
	vocab := GoalVocab(goal)
	return CompileExprVocabExt(expr, vocab, mod)
}

// CompileDefinitionGoalVocab compiles a definition and adds it to the goal
// as new premises (a ConstantDecl and a LabeledFormula stating the definition).
// Returns the modified goal.
// Corresponds to Python compile_definition_goal_vocab (ivy_proof.py:1477-1506).
//
// Accepts either a *ast.LabeledFormula (containing the equation directly) or
// a *ast.DerivedDecl (whose first arg is the LabeledFormula). Python's
// `lf = df.args[0]` (line 1480) extracts the inner LF from a DerivedDecl.
func CompileDefinitionGoalVocab(cfg *ast.AstConfig, df ast.Node, goal *ast.LabeledFormula, mod *module.Module) (*ast.LabeledFormula, error) {
	// Python: vocab = goal_vocab(goal); free = goal_free(goal)
	vocab := GoalVocab(goal)
	free := GoalFree(goal)

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
		return goal, nil
	}

	// Python: lhs = lf.formula.args[0]
	// The formula is an *ast.Definition with Lhs and Rhs.
	defFormula, ok := innerLF.Formula.(*ast.AstDefinition)
	if !ok {
		return goal, nil
	}

	// Get the LHS atom and its variable args.
	// Python: lhs = lf.formula.args[0]; vars = lf.formula.args[0].args
	lhsAtom, ok := defFormula.Lhs.(*ast.Atom)
	if !ok {
		return goal, nil
	}
	vars := lhsAtom.Terms // Python: lhs.args — the variable bindings

	// Python: ts = il.TopFunctionSort(len(lhs.args))
	ts := il.TopFunctionSort(len(vars))
	// Python: newsym = il.Symbol(lhs.rep, ts)
	newsym := lg.NewConst(lhsAtom.Rep, ts)

	// Python: with il.WithSymbols([newsym]):
	sig := getSigFrom(mod)
	ws := il.NewWithSymbols(sig, []*lg.Const{newsym})
	ws.Enter()
	defer ws.Exit()

	// Python: body = ia.Atom('=', lf.formula.args)
	// lf.formula.args is [lhs, rhs] in Python; in Go that's defFormula.Lhs, defFormula.Rhs.
	body := cfg.NewAtom("=", defFormula.Lhs, defFormula.Rhs)

	// Python: fmla = ia.Forall(vars, body) if vars else body
	var fmla ast.Node
	hasVars := len(vars) > 0
	if hasVars {
		fmla = cfg.NewForall(vars, body)
	} else {
		fmla = body
	}

	// Python line 1488: elf = lf.clone([lf.label, fmla])  → LF.clone PRESERVE #1
	elf := innerLF.Clone([]ast.Node{innerLF.Label, fmla}).(*ast.LabeledFormula)

	// Python line 1489: lf = compile_expr_vocab(elf, vocab)
	// compile_expr_vocab pushes vocab symbols/sorts, sets TopSort default,
	// calls elf.compile() (which is _labeled_formula_cmpl → CompileLF → PRESERVE #2),
	// then runs sort_infer_list on [compiled] + vocab.variables.
	compiledLF, err := compileExprVocabLF(elf, vocab, mod)
	if err != nil {
		return nil, err
	}

	// Python line 1490: thing = lf.formula.body if vars else lf.formula
	compiledFormula, ok := compiledLF.Formula.(lg.Expr)
	if !ok {
		return nil, fmt.Errorf("CompileDefinitionGoalVocab: compiled formula is not lg.Expr (type %T)", compiledLF.Formula)
	}
	var thing lg.Expr
	if hasVars {
		fa, ok := compiledFormula.(*lg.ForAll)
		if !ok {
			return nil, fmt.Errorf("CompileDefinitionGoalVocab: expected ForAll after compiling vars, got %T", compiledFormula)
		}
		thing = fa.Body
	} else {
		thing = compiledFormula
	}

	// Python line 1491: thing = il.normalize_ops(thing)
	thing = il.NormalizeOps(thing)

	// Python line 1493: lf = lf.clone([lf.label, lf.formula.clone([thing]) if vars else thing])
	// → LF.clone PRESERVE #3
	var newFormula ast.Node
	if hasVars {
		// Clone the ForAll with normalized body: lf.formula.clone([thing])
		newFormula = compiledFormula.(ast.Node).Clone([]ast.Node{thing})
	} else {
		newFormula = thing
	}
	lf := compiledLF.Clone([]ast.Node{compiledLF.Label, newFormula}).(*ast.LabeledFormula)

	// Python line 1494: sym = thing.args[0].rep
	// thing is the body (an equality). Get the LHS symbol.
	var sym *lg.Const
	eq, isEq := thing.(*lg.Eq)
	if isEq {
		switch lhs := eq.T1.(type) {
		case *lg.Apply:
			if c, ok := lhs.Func.(*lg.Const); ok {
				sym = c
			}
		case *lg.Const:
			sym = lhs
		}
	}
	if sym == nil {
		return nil, fmt.Errorf("CompileDefinitionGoalVocab: cannot extract defined symbol from compiled definition")
	}

	// Python line 1495-1496: deps = list(lu.symbols_ilu_ast(thing.args[1]))
	// Check for recursion: if sym appears in the RHS deps.
	symKey := lg.Key(sym)
	if isEq {
		for dep := range il.SymbolsIluAst(eq.T2) {
			if c, ok := dep.(*lg.Const); ok && lg.Key(c) == symKey {
				return nil, &NoMatch{Node: lf, Msg: "no proof given for recursive definition"}
			}
		}
	}

	// Python line 1499-1500: cd = ia.ConstantDecl(sym); cd.lineno = lf.lineno
	cd := cfg.NewConstantDecl(sym)
	cd.SetLineno(lf.GetLineno())

	// Python line 1501: lf.definition = True
	lf.IsDefinition = true

	// Python line 1502-1503: goal = goal_add_prem(goal, cd, lf.lineno); goal = goal_add_prem(goal, lf, lf.lineno)
	loc := lf.GetLineno()
	goal = GoalAddPrem(cfg, goal, cd, loc)
	goal = GoalAddPrem(cfg, goal, lf, loc)

	// Python line 1504-1505: if sym in vocab.sorts or sym in vocab.symbols or sym in free:
	//     raise Redefinition(df, "redefinition of {}".format(sym))
	if isInVocabOrFree(sym, vocab, free) {
		return nil, &Redefinition{Node: df, Msg: fmt.Sprintf("redefinition of %s", sym.Name)}
	}

	return goal, nil
}

// compileExprVocabLF compiles a LabeledFormula using a goal's vocabulary,
// returning the compiled LabeledFormula (not just the inner formula).
// This matches Python's compile_expr_vocab when called with a LabeledFormula:
// it calls elf.compile() → _labeled_formula_cmpl → lf.clone([...]) → PRESERVE,
// then runs sort_infer_list on [compiled_formula] + vocab.variables.
func compileExprVocabLF(lf *ast.LabeledFormula, vocab *Vocab, mod *module.Module) (*ast.LabeledFormula, error) {
	// Python's compile_expr_vocab emits this ENTER trace (ivy_proof.py:914).
	// The LF-specialized Go helper must mirror it.
	xtracer.Trace("proof.CompileExprVocab ENTER exprType=%s", iu.TypeName(lf))

	sig := getSigFrom(mod)

	// Python: with il.WithSymbols(vocab.symbols):
	ws := il.NewWithSymbols(sig, vocab.Symbols)
	ws.Enter()
	defer ws.Exit()

	// Python: with il.WithSorts(vocab.sorts):
	wso := il.NewWithSorts(sig, vocab.Sorts)
	wso.Enter()
	defer wso.Exit()

	// Python: with il.top_sort_as_default():
	tsDefault := il.TopSortAsDefault(sig)
	tsDefault.Enter()
	defer tsDefault.Exit()

	// Python: expr = il.sort_infer_list([expr.compile()] + vocab.variables)[0]
	// expr.compile() for LabeledFormula → _labeled_formula_cmpl → CompileLF
	if mod == nil {
		mod = module.New()
	}
	c := compiler.NewCompiler(sig, mod)
	compiled, err := c.ThingLF(lf)
	if err != nil {
		return nil, err
	}

	// Python: sort_infer_list([expr.compile()] + vocab.variables)
	// Python passes the entire LabeledFormula to sort_infer_list.
	// infer_sorts hits hasattr(t,'clone'), recursively processes args,
	// returns lambda: t.clone(inferred_args). check_concretely_sorted
	// on the LabeledFormula wrapper doesn't see unsorted vars inside.
	// The real sort inference already happened in SortifyWithInference
	// during CompileLF. So here we just clone the LF (matching Python).
	if formula, ok := compiled.Formula.(lg.Expr); ok {
		terms := make([]lg.Expr, 0, 1+len(vocab.Variables))
		terms = append(terms, formula)
		for _, v := range vocab.Variables {
			terms = append(terms, v)
		}
		inferred, inferErr := il.SortInferList(terms, nil, nil)
		if inferErr == nil && len(inferred) > 0 {
			// Sort inference succeeded — use inferred formula in clone.
			cloneArgs := compiled.Args()
			cloneArgs[1] = inferred[0]
			compiled = compiled.Clone(cloneArgs).(*ast.LabeledFormula)
		} else {
			// Sort inference failed or returned empty. Python still
			// clones the LF (the LF wrapper shields check_concretely_sorted
			// from seeing unsorted vars). Clone with original formula.
			compiled = compiled.Clone(compiled.Args()).(*ast.LabeledFormula)
		}
	}

	// Python compile_expr_vocab emits EXIT HASH canon=... right before return.
	xtracer.Trace("proof.CompileExprVocab EXIT HASH canon=%v", compiled.Canon())
	return compiled, nil
}

// isInVocabOrFree checks whether sym is in the vocab's sorts/symbols or in the free set.
// Python: sym in vocab.sorts or sym in vocab.symbols or sym in free
func isInVocabOrFree(sym *lg.Const, vocab *Vocab, free map[lg.NodeKey]lg.Expr) bool {
	symKey := lg.Key(sym)
	for _, s := range vocab.Sorts {
		if s != nil && s.Sexp() == sym.Sexp() {
			return true
		}
	}
	for _, c := range vocab.Symbols {
		if lg.Key(c) == symKey {
			return true
		}
	}
	if _, ok := free[symKey]; ok {
		return true
	}
	return false
}
