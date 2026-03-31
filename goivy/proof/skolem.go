package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// SkolemizeGoal converts a goal to skolem normal form:
// premises are in universal prenex form and the conclusion is in
// existential prenex form.
// If prenex is false, don't convert to prenex form.
func SkolemizeGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, prenex bool) *ast.LabeledFormula {
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
		// Replace free variables with fresh skolem constants
		var variables []*lg.Variable
		for _, freeNode := range free {
			if vv, ok := freeNode.(*lg.Variable); ok {
				variables = append(variables, vv)
			}
		}
		sks := make([]*lg.Const, len(variables))
		subs := make(map[lg.NodeKey]lg.Expr)
		for i, v := range variables {
			name := renamer.Rename("_" + v.Name)
			sk := lg.NewConst(name, v.VSort)
			sks[i] = sk
			subs[lg.Key(v)] = sk
		}
		if len(subs) > 0 {
			goal = varSubstGoal(cfg, goal, subs)
		}
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
	return CloneGoal(cfg, goal, newPrems, GoalConc(goal))
}

// SkolemizeFmla skolemizes a formula.
// pos indicates whether the formula appears in positive position.
// renamer generates unique names for skolem functions.
// skfuns accumulates the skolem function constants.
// If prenex is true, universally quantified variables are collected
// into a single prenex quantifier.
func SkolemizeFmla(fmla lg.Expr, pos bool, renamer *iu.UniqueRenamer, skfuns *[]*lg.Const, prenex bool) lg.Expr {
	var univs []*lg.Variable
	var outer []*lg.Variable

	var rec func(lg.Expr, bool) lg.Expr
	rec = func(fmla lg.Expr, pos bool) lg.Expr {
		switch f := fmla.(type) {
		case *lg.Not:
			return &lg.Not{Body: rec(f.Body, !pos)}
		case *lg.Implies:
			return &lg.Implies{
				T1: rec(f.T1, !pos),
				T2: rec(f.T2, pos),
			}
		case *lg.And:
			terms := make([]lg.Expr, len(f.Terms))
			for i, t := range f.Terms {
				terms[i] = rec(t, pos)
			}
			return &lg.And{Terms: terms}
		case *lg.Or:
			terms := make([]lg.Expr, len(f.Terms))
			for i, t := range f.Terms {
				terms[i] = rec(t, pos)
			}
			return &lg.Or{Terms: terms}
		}

		isE := il.IsExists(fmla)
		isA := il.IsForall(fmla)

		// Skolemize: forall in positive / exists in negative position
		if (isA && pos) || (isE && !pos) {
			vars := il.BinderVars(fmla)
			body := il.BinderBody(fmla)

			// Collect outer universal variables for the skolem function domain
			fvs := outerVarsInFormula(fmla, outer)

			for _, v := range vars {
				domSorts := make([]lg.Sort, len(fvs)+1)
				for j, fv := range fvs {
					domSorts[j] = fv.VSort
				}
				domSorts[len(fvs)] = v.VSort
				skSort := il.FuncConstSort(domSorts...)

				name := renamer.Rename("_" + v.Name)
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
				}
			}
			return rec(body, pos)
		}

		// Universalize: exists in positive / forall in negative position
		if (isE && pos) || (isA && !pos) {
			vars := il.BinderVars(fmla)
			body := il.BinderBody(fmla)

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
			}
			res := rec(body, pos)
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
		return fmla
	}

	body := rec(fmla, pos)
	if len(univs) > 0 {
		if pos {
			body = il.Exists(univs, body)
		} else {
			body = il.ForAll(univs, body)
		}
	}
	return body
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
func varSubstGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, subs map[lg.NodeKey]lg.Expr) *ast.LabeledFormula {
	prems := GoalPrems(goal)
	newPrems := make([]ast.Node, len(prems))
	for i, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			newPrems[i] = varSubstGoal(cfg, lf, subs)
		} else {
			newPrems[i] = p
		}
	}
	conc := GoalConc(goal)
	if conc != nil {
		newConc, err := lu.Substitute(conc, subs)
		if err == nil {
			conc = newConc
		}
	}
	return CloneGoal(cfg, goal, newPrems, conc)
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
