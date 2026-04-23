package proof

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// SkolemizeGoal converts a goal to skolem normal form:
// premises are in universal prenex form and the conclusion is in
// existential prenex form.
// If prenex is false, don't convert to prenex form.
func SkolemizeGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, prenex bool) *ast.LabeledFormula {
	xtracer.Trace("proof.SkolemizeGoal ENTER prenex=%v label=%s HASH canon=%v", prenex, goal.LabelForTrace(), goal.Canon())
	vocab := GoalVocab(goal)
	usedNames := make(map[string]struct{})
	for _, s := range vocab.Symbols {
		usedNames[s.Name] = struct{}{}
	}
	free := GoalFree(goal)
	for _, node := range free {
		if c, ok := node.(*lg.Const); ok {
			usedNames[c.Name] = struct{}{}
		}
		if v, ok := node.(*lg.Variable); ok {
			usedNames[v.Name] = struct{}{}
		}
	}

	usedSlice := make([]string, 0, len(usedNames))
	for n := range usedNames {
		usedSlice = append(usedSlice, n)
	}
	renamer := iu.NewUniqueRenamer("", usedSlice)
	var skfuns []*lg.Const

	if !prenex {
		// Python: variables = goal_free_vars(goal)
		// — distinct from goal_free; walks prems+conc for variables in DFS order.
		variables := GoalFreeVars(goal)
		xtracer.Trace("proof.SkolemizeGoal freeVarsPass nvariables=%d", len(variables))
		sks := make([]*lg.Const, len(variables))
		subs := make(map[lg.NodeKey]lg.Expr)
		for i, v := range variables {
			name := renamer.Rename("_" + v.Name)
			sk := lg.NewConst(name, v.VSort)
			sks[i] = sk
			subs[lg.Key(v)] = sk
			xtracer.Trace("proof.SkolemizeGoal substitute v=%s sk=%s", v.Name, name)
		}
		goal = varSubstGoal(cfg, goal, subs)
		skfuns = append(skfuns, sks...)
	}

	var rec func(g *ast.LabeledFormula, pos bool) *ast.LabeledFormula
	rec = func(g *ast.LabeledFormula, pos bool) *ast.LabeledFormula {
		prems := GoalPrems(g)
		newPrems := make([]ast.Node, len(prems))
		for i, p := range prems {
			if lf, ok := p.(*ast.LabeledFormula); ok {
				newPrems[i] = rec(lf, !pos)
			} else {
				newPrems[i] = p
			}
		}
		conc := GoalConc(g)
		if conc != nil {
			conc = SkolemizeFmla(conc, pos, renamer, &skfuns, prenex)
		}
		return CloneGoal(cfg, g, newPrems, conc)
	}

	goal = rec(goal, true)

	// Prepend skolem function declarations
	var newPrems []ast.Node
	for _, sk := range skfuns {
		cd := cfg.NewConstantDecl(sk)
		newPrems = append(newPrems, cd)
	}
	newPrems = append(newPrems, GoalPrems(goal)...)
	result := CloneGoal(cfg, goal, newPrems, GoalConc(goal))
	xtracer.Trace("proof.SkolemizeGoal EXIT nskfuns=%d HASH canon=%v", len(skfuns), result.Canon())
	return result
}

// SkolemizeFmla skolemizes a formula.
// pos indicates whether the formula appears in positive position.
// renamer generates unique names for skolem functions.
// skfuns accumulates the skolem function constants.
// If prenex is true, universally quantified variables are collected
// into a single prenex quantifier.
//
// Takes/returns ast.Node so it can handle *ast.TemporalModels (mirroring
// Python ivy_proof.py:1443-1450). For lg.Expr inputs, behaves identically
// to the previous lg.Expr-only signature.
func SkolemizeFmla(fmla ast.Node, pos bool, renamer *iu.UniqueRenamer, skfuns *[]*lg.Const, prenex bool) ast.Node {
	xtracer.Trace("proof.SkolemizeFmla ENTER pos=%v prenex=%v type=%s", pos, prenex, iu.TypeName(fmla))
	var univs []*lg.Variable
	var outer []*lg.Variable

	var rec func(ast.Node, bool) ast.Node
	rec = func(fmla ast.Node, pos bool) ast.Node {
		// Mirror Python ivy_proof.py:1443-1444:
		//   if isinstance(fmla,ia.TemporalModels):
		//       return fmla.clone([rec(fmla.args[0],pos)])
		if tm, ok := fmla.(*ast.TemporalModels); ok {
			xtracer.Trace("proof.SkolemizeFmla branch type=TemporalModels pos=%v", pos)
			return tm.Clone([]ast.Node{rec(tm.Args()[0], pos)})
		}

		switch f := fmla.(type) {
		case *lg.Not:
			xtracer.Trace("proof.SkolemizeFmla branch type=Not pos=%v", pos)
			return &lg.Not{Body: recExpr(rec, f.Body, !pos)}
		case *lg.Implies:
			xtracer.Trace("proof.SkolemizeFmla branch type=Implies pos=%v", pos)
			return &lg.Implies{
				T1: recExpr(rec, f.T1, !pos),
				T2: recExpr(rec, f.T2, pos),
			}
		case *lg.And:
			xtracer.Trace("proof.SkolemizeFmla branch type=And pos=%v nterms=%d", pos, len(f.Terms))
			terms := make([]lg.Expr, len(f.Terms))
			for i, t := range f.Terms {
				terms[i] = recExpr(rec, t, pos)
			}
			return &lg.And{Terms: terms}
		case *lg.Or:
			xtracer.Trace("proof.SkolemizeFmla branch type=Or pos=%v nterms=%d", pos, len(f.Terms))
			terms := make([]lg.Expr, len(f.Terms))
			for i, t := range f.Terms {
				terms[i] = recExpr(rec, t, pos)
			}
			return &lg.Or{Terms: terms}
		}

		// IsExists/IsForall take lg.Expr; guard with type assertion.
		var isE, isA bool
		if expr, ok := fmla.(lg.Expr); ok {
			isE = il.IsExists(expr)
			isA = il.IsForall(expr)
		}

		// Skolemize: forall in positive / exists in negative position
		if (isA && pos) || (isE && !pos) {
			expr := fmla.(lg.Expr) // safe because isE/isA imply lg.Expr
			vars := il.BinderVars(expr)
			body := il.BinderBody(expr)
			if xtracer.Enabled {
				vnames := make([]string, len(vars))
				for vi, vv := range vars {
					vnames[vi] = "'" + vv.Name + "'"
				}
				xtracer.Trace("proof.SkolemizeFmla branch type=Skolemize pos=%v isE=%v isA=%v nvars=%d varNames=[%s]", pos, isE, isA, len(vars), strings.Join(vnames, ", "))
			}

			// Collect outer universal variables for the skolem function domain
			fvs := outerVarsInFormula(expr, outer)

			for _, v := range vars {
				domSorts := make([]lg.Sort, len(fvs)+1)
				for j, fv := range fvs {
					domSorts[j] = fv.VSort
				}
				domSorts[len(fvs)] = v.VSort
				skSort := il.FuncConstSort(domSorts...)

				name := renamer.Rename("_" + v.Name)
				xtracer.Trace("proof.SkolemizeFmla skolemize v=%s sk=%s", v.Name, name)
				sym := lg.NewConst(name, skSort)
				*skfuns = append(*skfuns, sym)

				var term lg.Expr
				if len(fvs) > 0 {
					args := make([]lg.Expr, len(fvs))
					for j, fv := range fvs {
						args[j] = fv
					}
					app, err := lg.NewApply(sym, args...)
					if err != nil {
						term = sym
					} else {
						term = app
					}
				} else {
					term = sym
				}
				subs := map[lg.NodeKey]lg.Expr{lg.Key(v): term}
				newBody, err := lu.Substitute(body, subs)
				if err == nil {
					body = newBody
				} else {
					xtracer.Trace("proof.SkolemizeFmla skolemize substitute err=%v v=%s", err, v.Name)
				}
			}
			return rec(body, pos)
		}

		// Universalize: exists in positive / forall in negative position
		if (isE && pos) || (isA && !pos) {
			expr := fmla.(lg.Expr) // safe because isE/isA imply lg.Expr
			vars := il.BinderVars(expr)
			body := il.BinderBody(expr)
			xtracer.Trace("proof.SkolemizeFmla branch type=Universalize pos=%v isE=%v isA=%v nvars=%d", pos, isE, isA, len(vars))

			vu := il.NewVariableUniqifier(keysFromRenamer(renamer))
			for _, v := range vars {
				u := uniquifyVar(vu, v)
				if prenex {
					univs = append(univs, u)
				}
				outer = append(outer, u)
				subs := map[lg.NodeKey]lg.Expr{lg.Key(v): u}
				newBody, err := lu.Substitute(body, subs)
				if err == nil {
					body = newBody
				}
				xtracer.Trace("proof.SkolemizeFmla universalize v=%s u=%s", v.Name, u.Name)
			}
			res := recExpr(rec, body, pos)
			if !prenex {
				// Wrap in same quantifier type with the new variables
				tail := outer[len(outer)-len(vars):]
				if isE {
					res = il.Exists(tail, res)
				} else {
					res = il.ForAll(tail, res)
				}
			}
			outer = outer[:len(outer)-len(vars)]
			return res
		}
		xtracer.Trace("proof.SkolemizeFmla branch type=passthrough pos=%v astType=%s", pos, iu.TypeName(fmla))
		return fmla
	}

	body := rec(fmla, pos)
	// Final univs wrapping. Mirror Python ivy_proof.py:1447-1453:
	//   if isinstance(body,ia.TemporalModels):
	//       body = body.clone([quant(univs,body.args[0])])
	//   else:
	//       body = quant(univs,body)
	if len(univs) > 0 {
		xtracer.Trace("proof.SkolemizeFmla univsWrap pos=%v nunivs=%d", pos, len(univs))
		if tm, ok := body.(*ast.TemporalModels); ok {
			innerExpr, _ := tm.Args()[0].(lg.Expr)
			var quantBody lg.Expr
			if pos {
				quantBody = il.Exists(univs, innerExpr)
			} else {
				quantBody = il.ForAll(univs, innerExpr)
			}
			body = tm.Clone([]ast.Node{quantBody})
		} else if expr, ok := body.(lg.Expr); ok {
			if pos {
				body = il.Exists(univs, expr)
			} else {
				body = il.ForAll(univs, expr)
			}
		}
	}
	xtracer.Trace("proof.SkolemizeFmla EXIT nskfuns=%d", len(*skfuns))
	return body
}

// recExpr is a helper that wraps an ast.Node-returning rec function to
// produce an lg.Expr result by type-asserting. Used inside SkolemizeFmla
// for the cases that need to construct an lg.Expr (Not, Implies, And, Or).
// If the recursive result is not lg.Expr (shouldn't happen for these cases),
// returns the original input as a fallback.
func recExpr(rec func(ast.Node, bool) ast.Node, fmla lg.Expr, pos bool) lg.Expr {
	r := rec(fmla, pos)
	if e, ok := r.(lg.Expr); ok {
		return e
	}
	return fmla
}

// --- helpers ---

// outerVarsInFormula returns the outer universal variables that appear
// free in the given formula.
func outerVarsInFormula(fmla lg.Expr, outer []*lg.Variable) []*lg.Variable {
	if len(outer) == 0 {
		return nil
	}
	used := lu.UsedVariables(fmla)
	outerSet := make(map[lg.NodeKey]lg.Expr, len(outer))
	for _, v := range outer {
		outerSet[lg.Key(v)] = v
	}
	var result []*lg.Variable
	// preserve order
	seen := make(map[lg.NodeKey]lg.Expr)
	for vKey, vNode := range used {
		if outerSet[vKey] != nil && seen[vKey] == nil {
			if vv, ok := vNode.(*lg.Variable); ok {
				result = append(result, vv)
			}
			seen[vKey] = vNode
		}
	}
	return result
}

// varSubstGoal applies a variable substitution to a goal.
// Mirrors Python ivy_proof.py:1364-1368 var_subst_goal — uses apply_to_conc
// so the substitution runs on the inner formula of *ast.TemporalModels and
// the wrapper is preserved.
func varSubstGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, subs map[lg.NodeKey]lg.Expr) *ast.LabeledFormula {
	xtracer.Trace("proof.varSubstGoal ENTER label=%s nsubs=%d", goal.LabelForTrace(), len(subs))
	prems := GoalPrems(goal)
	newPrems := make([]ast.Node, len(prems))
	for i, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			newPrems[i] = varSubstGoal(cfg, lf, subs)
		} else {
			newPrems[i] = p
		}
	}
	newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		result, err := lu.Substitute(c, subs)
		if err != nil {
			return c
		}
		return result
	})
	result := CloneGoal(cfg, goal, newPrems, newConc)
	xtracer.Trace("proof.varSubstGoal EXIT HASH canon=%v", result.Canon())
	return result
}

// keysFromRenamer extracts the used names from a UniqueRenamer.
func keysFromRenamer(rn *iu.UniqueRenamer) []string {
	result := make([]string, 0, len(rn.Used))
	for k := range rn.Used {
		result = append(result, k)
	}
	return result
}

// uniquifyVar creates a new variable with a unique name using a VariableUniqifier.
func uniquifyVar(vu *il.VariableUniqifier, v *lg.Variable) *lg.Variable {
	// Use the uniqifier to generate a fresh name
	fmla := il.ForAll([]*lg.Variable{v}, v)
	result := vu.Uniquify(fmla)
	if fa, ok := result.(*lg.ForAll); ok && len(fa.Variables) > 0 {
		return fa.Variables[0]
	}
	// fallback
	nv, _ := lg.NewVariable(v.Name+fmt.Sprintf("_%d", len(vu.InvMap)), v.VSort)
	return nv
}
