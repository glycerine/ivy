// phase5_goals.go implements Phase 5 Batch 5.1 (Goal Utilities) and
// Batch 5.4 (Tactics/Remaining) functions from the Ivy port.
// These correspond to Python ivy_proof.py goal manipulation functions.
package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// === Batch 5.1: Goal Utilities ===

// AttribGoals copies line number information from a proof node to goals.
// Corresponds to Python's attrib_goals.
func AttribGoals(proof Node, goals []*LabeledFormula) []*LabeledFormula {
	if proof == nil {
		return goals
	}
	loc := proof.GetLineno()
	if loc.Line == 0 && loc.Filename == "" {
		return goals
	}
	for _, g := range goals {
		g.SetLineno(loc)
	}
	return goals
}

// IsGoal returns true if the node is a LabeledFormula (a goal).
// Corresponds to Python's is_goal.
func IsGoal(g Node) bool {
	_, ok := g.(*LabeledFormula)
	return ok
}

// GoalsSubst substitutes multiple subgoals into the first goal's conclusion.
// Corresponds to Python's goals_subst.
func GoalsSubst(cfg *AstConfig, goals []*LabeledFormula, subgoals []*LabeledFormula, loc Location) ([]*LabeledFormula, error) {
	if len(goals) == 0 || len(subgoals) == 0 {
		return goals, nil
	}
	var result []*LabeledFormula
	for _, sg := range subgoals {
		sub, err := GoalSubst(cfg, goals[0], sg, loc)
		if err != nil {
			return nil, err
		}
		result = append(result, sub)
	}
	result = append(result, goals[1:]...)
	return result, nil
}

// FreshLabel generates a fresh unique label not used in any goal.
// Corresponds to Python's fresh_label.
func FreshLabel(cfg *AstConfig, goals []*LabeledFormula) Node {
	var used []string
	for _, g := range goals {
		name := g.LabelName()
		if name != "" {
			used = append(used, name)
		}
	}
	rn := NewUniqueRenamer("", used)
	name := rn.Rename("")
	return cfg.NewAtom(name)
}

// GoalPrefixPrems adds premises to the beginning of a goal's premise list.
// Corresponds to Python's goal_prefix_prems.
func GoalPrefixPrems(cfg *AstConfig, goal *LabeledFormula, prems []Node, loc Location) *LabeledFormula {
	existingPrems := GoalPrems(goal)
	allPrems := make([]Node, 0, len(prems)+len(existingPrems))
	allPrems = append(allPrems, prems...)
	allPrems = append(allPrems, existingPrems...)
	return MakeGoal(cfg, loc, goal.Label, allPrems, GoalConc(goal))
}

// CheckNameClash checks that two goals have no overlapping symbol definitions.
// Corresponds to Python's check_name_clash.
func CheckNameClash(g1, g2 *LabeledFormula) error {
	xtracer.Trace("proof.CheckNameClash ENTER g1Label=%s g2Label=%s", g1.LabelForTrace(), g2.LabelForTrace())
	d1 := GoalDefns(g1)
	d2 := GoalDefns(g2)
	for k := range d1 {
		if _, exists := d2[k]; exists {
			xtracer.Trace("proof.CheckNameClash EXIT err=clash key=%s", string(k))
			return &ProofError{Msg: "premise of subgoal clashes with context"}
		}
	}
	xtracer.Trace("proof.CheckNameClash EXIT ok")
	return nil
}

// CheckPremisesProvided verifies that all non-proposition premises of g1
// are provided by g2.
// Corresponds to Python's check_premises_provided (ivy_proof.py:586-593).
func CheckPremisesProvided(g1, g2 *LabeledFormula, sig *Sig) error {
	defns := GoalDefns(g2)
	for k, thing := range GoalDefns(g1) {
		// Python: syms = [] if il.is_lambda(thing) else [thing]
		if _, isLam := thing.(*Lambda); isLam {
			continue
		}
		if _, ok := defns[k]; !ok && !sig.Contains(thing) {
			return &ProofError{Msg: fmt.Sprintf("premise %v does not match anything in the environment", thing)}
		}
	}
	return nil
}

// GoalIsTemporal checks if a goal has temporal properties.
// Corresponds to Python's goal_is_temporal (ivy_proof.py:603-605):
//
//	def goal_is_temporal(x):
//	    conc = goal_conc(x)
//	    return conc.temoral or isinstance(conc.formula,ia.TemporalModels)
//
// Returns true when the goal's conclusion is a *ast.TemporalModels OR a
// NamedBinder for "globally"/"eventually".
func GoalIsTemporal(x *LabeledFormula) bool {
	if x.IsTemporal() {
		return true
	}
	conc := GoalConc(x)
	if conc == nil {
		return false
	}
	if _, ok := conc.(*TemporalModels); ok {
		return true
	}
	if nb, ok := conc.(*LogicNamedBinder); ok {
		return nb.Name == "globally" || nb.Name == "eventually"
	}
	return false
}

// GoalDefines extracts what a definition premise defines.
// Corresponds to Python's goal_defines.
func GoalDefines(x Node) Expr {
	if cd, ok := x.(*ConstantDecl); ok {
		if len(cd.Args()) > 0 {
			if sym, ok := cd.Args()[0].(Expr); ok {
				return sym
			}
		}
	}
	if n, ok := x.(Expr); ok {
		return n
	}
	return nil
}

// GetUnprovidedDefns gets definitions from g1 that aren't provided by g2.
// Corresponds to Python's get_unprovided_defns.
func GetUnprovidedDefns(g1, g2 *LabeledFormula) []Node {
	defns := GoalDefns(g2)
	free := GoalFree(g2)
	var result []Node
	for _, prem := range GoalPrems(g1) {
		if GoalIsDefn(prem) {
			sym := GoalDefines(prem)
			if sym == nil {
				continue
			}
			k := Key(sym)
			if _, inDefns := defns[k]; !inDefns {
				if _, inFree := free[k]; !inFree {
					result = append(result, prem)
				}
			}
		}
	}
	return result
}

// GoalSubgoals extracts subgoals from a schema instantiation.
// Corresponds to Python's goal_subgoals.
func GoalSubgoals(cfg *AstConfig, schema, goal *LabeledFormula, loc Location) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.GoalSubgoals ENTER schemaLabel=%s goalLabel=%s", schema.LabelForTrace(), goal.LabelForTrace())
	if err := CheckConcsMatch(schema, goal); err != nil {
		xtracer.Trace("proof.GoalSubgoals EXIT err=concMismatch err=%v", err)
		return nil, err
	}
	upds := GetUnprovidedDefns(schema, goal)
	g := CloneGoal(cfg, goal, upds, GoalConc(goal))
	resultGoal, err := GoalSubst(cfg, goal, g, loc)
	if err != nil {
		xtracer.Trace("proof.GoalSubgoals EXIT err=goalSubst err=%v", err)
		return nil, err
	}

	gpms := GoalPremGoals(resultGoal)

	var subgoals []*LabeledFormula
	for _, sp := range GoalPremGoals(schema) {
		alreadyPresent := false
		for _, gpm := range gpms {
			if GoalsEqModAlpha(sp, gpm) {
				alreadyPresent = true
				break
			}
		}
		if !alreadyPresent {
			sg, err := GoalSubst(cfg, resultGoal, sp, loc)
			if err != nil {
				xtracer.Trace("proof.GoalSubgoals EXIT err=subgoalSubst err=%v", err)
				return nil, err
			}
			if !TrivialGoal(sg) {
				subgoals = append(subgoals, sg)
			}
		}
	}
	xtracer.Trace("proof.GoalSubgoals EXIT nsubgoals=%d", len(subgoals))
	return subgoals, nil
}

// FmlaVocab gets the free vocabulary of a formula, including sorts, symbols, and variables.
// Corresponds to Python's fmla_vocab.
func FmlaVocab(fmla Expr) map[NodeKey]Expr {
	result := make(map[NodeKey]Expr)
	// Python: lu.used_sorts_ast(fmla)
	for k, s := range SortsAst(fmla) {
		result[k] = s
	}
	// Use clauseops.UsedSymbolsAST for symbols
	for k, v := range UsedSymbolsAST(fmla).All() {
		result[k] = v
	}
	// Use clauseops.VariablesAST for variables
	for _, v := range VariablesAST(fmla) {
		result[Key(v)] = v
	}
	return result
}

// CheckSchemaCapture verifies that free variables in schema are not captured by goal bindings.
// Corresponds to Python's check_schema_capture.
func CheckSchemaCapture(schema, goal *LabeledFormula) error {
	gvocab := GoalVocab(goal)
	fvocab := GoalFree(schema)
	gSymSet := make(map[string]bool)
	for _, s := range gvocab.Symbols {
		gSymSet[s.Name] = true
	}
	gSortSet := make(map[string]bool)
	for _, s := range gvocab.Sorts {
		gSortSet[IvySortName(s)] = true
	}
	for _, sym := range fvocab {
		name := fmt.Sprint(sym)
		if gSymSet[name] || gSortSet[name] {
			return &CaptureError{Msg: fmt.Sprintf("%q is captured when importing %q", sym, schema.LabelName())}
		}
	}
	return nil
}

// CheckAlphaCapture checks that alpha renaming doesn't cause variable capture.
// Corresponds to Python's check_alpha_capture.
func CheckAlphaCapture(goal *LabeledFormula, match map[NodeKey]Expr) error {
	revMatch := make(map[NodeKey]Expr)
	for k, v := range match {
		revMatch[Key(v)] = match[k]
	}
	free := GoalFree(goal)
	defns := GoalDefns(goal)
	for k, s := range free {
		if _, inRev := revMatch[k]; inRev {
			if _, inMatch := match[k]; !inMatch {
				return &CaptureError{Msg: fmt.Sprintf("%q is captured by renaming", s)}
			}
		}
	}
	for k, s := range defns {
		if _, inRev := revMatch[k]; inRev {
			if _, inMatch := match[k]; !inMatch {
				return &CaptureError{Msg: fmt.Sprintf("%q is captured by renaming", s)}
			}
		}
	}
	return nil
}

// CheckRenaming validates a renaming specification for consistency.
// Corresponds to Python's check_renaming.
func CheckRenaming(goal *LabeledFormula, renaming Node) error {
	fwd := make(map[string]string)
	rev := make(map[string]string)
	for _, arg := range renaming.Args() {
		defn, ok := arg.(*Definition)
		if !ok {
			continue
		}
		lAtom, lok := defn.Lhs.(*Atom)
		rAtom, rok := defn.Rhs.(*Atom)
		if !lok || !rok {
			continue
		}
		l := lAtom.Rep
		r := rAtom.Rep
		if prev, exists := fwd[l]; exists {
			return &ProofError{Msg: fmt.Sprintf("%q is renamed to both %q and %q", l, prev, r)}
		}
		if prev, exists := rev[r]; exists {
			return &ProofError{Msg: fmt.Sprintf("both %q and %q are renamed to %q", prev, l, r)}
		}
		fwd[l] = r
		rev[r] = l
	}
	return nil
}

// GoalIsProperty returns true if the goal is a property (not a schema).
// Corresponds to Python's goal_is_property.
func GoalIsProperty(x *LabeledFormula) bool {
	_, isSchema := x.Formula.(*SchemaBody)
	return !isSchema
}

// GoalIsSchema returns true if the goal is a schema.
// Corresponds to Python's goal_is_schema.
func GoalIsSchema(x *LabeledFormula) bool {
	_, isSchema := x.Formula.(*SchemaBody)
	return isSchema
}

// GoalsEqModAlpha checks if two goals are equivalent modulo alpha renaming.
// Corresponds to Python's goals_eq_mod_alpha.
func GoalsEqModAlpha(x, y *LabeledFormula) bool {
	xIsSchema := GoalIsSchema(x)
	yIsSchema := GoalIsSchema(y)
	if xIsSchema {
		if !yIsSchema {
			return false
		}
		xps := GoalPrems(x)
		yps := GoalPrems(y)
		if len(xps) != len(yps) {
			return false
		}
		for i := range xps {
			xlf, xIsLF := xps[i].(*LabeledFormula)
			ylf, yIsLF := yps[i].(*LabeledFormula)
			if xIsLF != yIsLF {
				return false
			}
			if xIsLF && yIsLF {
				if !GoalsEqModAlpha(xlf, ylf) {
					return false
				}
			}
		}
	} else {
		if yIsSchema {
			return false
		}
	}
	// Unwrap *ast.TemporalModels if present so equality compares the inner
	// formulas. Mirrors Python's duck-typed access via goal_conc.
	xConc := GoalConcUnwrap(x)
	yConc := GoalConcUnwrap(y)
	if xConc == nil || yConc == nil {
		return xConc == yConc
	}
	return EqualModAlpha(xConc, yConc)
}

// IsLambdaPrem checks if a premise is a lambda function definition.
// Corresponds to Python's is_lambda.
func IsLambdaPrem(p Node) bool {
	cd, ok := p.(*ConstantDecl)
	if !ok {
		return false
	}
	args := cd.Args()
	if len(args) == 0 {
		return false
	}
	// Check if the first arg is a lambda (wrapped as lg.Expr in ast.Node)
	if n, ok := args[0].(Expr); ok {
		_, isLam := n.(*Lambda)
		return isLam
	}
	return false
}

// === Batch 5.4: Tactics and Remaining ===

// VarSubstGoal applies a variable substitution to a goal.
// Corresponds to Python's var_subst_goal.
func VarSubstGoal(cfg *AstConfig, goal *LabeledFormula, subst map[NodeKey]Expr) *LabeledFormula {
	var prems []Node
	for _, prem := range GoalPrems(goal) {
		if premLF, ok := prem.(*LabeledFormula); ok {
			prems = append(prems, VarSubstGoal(cfg, premLF, subst))
		} else {
			prems = append(prems, prem)
		}
	}
	conc := GoalConc(goal)
	if conc != nil && !GoalIsSchema(goal) {
		conc = ApplyToConc(conc, func(x Expr) Expr {
			return SubstituteAstByName(x, nodeMapToStringMap(subst))
		})
	}
	return CloneGoal(cfg, goal, prems, conc)
}

// nodeMapToStringMap converts lg.NodeKey->lg.Expr map to string->lg.Expr for substitution.
func nodeMapToStringMap(m map[NodeKey]Expr) map[string]Expr {
	result := make(map[string]Expr, len(m))
	for _, v := range m {
		if variable, ok := v.(*LogicVariable); ok {
			result[variable.Name] = v
		}
	}
	return result
}

// (ApplyToConc moved to proof/goal.go and broadened to take ast.Node so it
// handles *ast.TemporalModels as well as plain lg.Expr conclusions. The old
// NamedBinder special case for "globally"/"eventually" was redundant —
// SubstituteAstByName already recurses into NamedBinder bodies — and Python's
// apply_to_conc only handles TemporalModels (ivy_proof.py:1370-1373).)

// RemoveUnusedDefinitionsGoal removes definitions that aren't referenced
// in the goal's conclusion formulas.
// concFmlas is the list of formulas to scan for initial used symbols.
// For TemporalModels conclusions, this should be model.Fmlas() + [conc.Fmla].
// For non-temporal conclusions, this should be [conc].
// Corresponds to Python's remove_unused_definitions_goal (ivy_proof.py:1508-1528).
func RemoveUnusedDefinitionsGoal(cfg *AstConfig, goal *LabeledFormula, concFmlas []Expr) *LabeledFormula {
	prems := GoalPrems(goal)
	conc := GoalConc(goal)
	if conc == nil {
		return goal
	}

	// Build initial symbol set from concFmlas.
	// Python: syms = lu.used_symbols_asts(fmlas)
	usedSyms := make(map[NodeKey]Expr)
	for _, f := range concFmlas {
		for k, v := range UsedSymbolsAST(f).All() {
			usedSyms[k] = v
		}
	}

	reversed := make([]Node, len(prems))
	for i, p := range prems {
		reversed[len(prems)-1-i] = p
	}
	var newPrems []Node
	for _, x := range reversed {
		// Case 1: property definitions (Python: goal_is_property(x) and x.definition)
		if lf, ok := x.(*LabeledFormula); ok && GoalIsProperty(lf) && lf.IsDefinition {
			if fExpr, ok := lf.Formula.(Expr); ok {
				// Python: sym = il.drop_universals(x.formula).args[0].rep
				df := IvyDropUniversals(fExpr)
				dfArgs := df.Args()
				if len(dfArgs) > 0 {
					if argExpr, ok := dfArgs[0].(Expr); ok {
						sym := IvyNodeRep(argExpr)
						if sym != nil {
							if _, used := usedSyms[Key(sym)]; !used {
								continue
							}
						}
					}
				}
			}
		} else if GoalIsDefn(x) {
			// Case 2: ConstantDecl/UninterpretedSort (existing logic)
			sym := GoalDefines(x)
			if sym == nil {
				newPrems = append([]Node{x}, newPrems...)
				continue
			}
			if _, used := usedSyms[Key(sym)]; !used {
				continue
			}
		}
		if n, ok := x.(Expr); ok {
			for k, v := range UsedSymbolsAST(n).All() {
				usedSyms[k] = v
			}
		}
		newPrems = append([]Node{x}, newPrems...)
	}
	return CloneGoal(cfg, goal, newPrems, conc)
}

// MatchFromDefn extracts a match from a definition formula.
// Corresponds to Python's match_from_defn.
func MatchFromDefn(defn *LabeledFormula) (map[NodeKey]Expr, error) {
	conc := GoalConc(defn)
	if conc == nil {
		return nil, &ProofError{Msg: "not a definition"}
	}
	fmla := conc
	for {
		if fa, ok := fmla.(*ForAll); ok {
			fmla = fa.Body
		} else {
			break
		}
	}
	if eq, ok := fmla.(*Eq); ok {
		if app, ok := eq.T1.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok {
				vars := nodesToVarsPhase5(app.Terms)
				// Python: if iu.distinct(lhs.args)
				if !distinctVars(vars) {
					return nil, &ProofError{Msg: "not a definition: duplicate parameters"}
				}
				lam := &Lambda{Variables: vars, Body: eq.T2}
				result := make(map[NodeKey]Expr)
				result[Key(c)] = lam
				return result, nil
			}
		}
	}
	if iff, ok := fmla.(*LogicIff); ok {
		if app, ok := iff.T1.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok {
				vars := nodesToVarsPhase5(app.Terms)
				// Python: if iu.distinct(lhs.args)
				if !distinctVars(vars) {
					return nil, &ProofError{Msg: "not a definition: duplicate parameters"}
				}
				lam := &Lambda{Variables: vars, Body: iff.T2}
				result := make(map[NodeKey]Expr)
				result[Key(c)] = lam
				return result, nil
			}
		}
	}
	return nil, &ProofError{Msg: "not a definition", Node: fmla}
}

func nodesToVarsPhase5(nodes []Expr) []*LogicVariable {
	var result []*LogicVariable
	for _, n := range nodes {
		if v, ok := n.(*LogicVariable); ok {
			result = append(result, v)
		}
	}
	return result
}

// distinctVars returns true if all variables have distinct NodeKeys.
// Corresponds to Python's iu.distinct(lhs.args) in match_from_defn.
func distinctVars(vars []*LogicVariable) bool {
	seen := make(map[NodeKey]bool, len(vars))
	for _, v := range vars {
		k := Key(v)
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	return true
}

// ExprListOrLambdaUnion holds either a single lambda or a list of lambdas
// for multi-definition unfolding. Used by MatchFromDefns.
// Does NOT implement lg.Expr — stays localized to the unfold path.
//
// Python: match_from_defns returns {sym: [lambda1, lambda2, ...]}
// Python: match_get pops the first element on each access:
//
//	val = save[0]; if len(save) > 1: del save[0]
type ExprListOrLambdaUnion struct {
	Items []Expr // always len >= 1; Pop returns next
}

// Pop returns the first item. If more than one item, removes it.
// Mirrors Python: val = save[0]; if len(save) > 1: del save[0]
func (u *ExprListOrLambdaUnion) Pop() Expr {
	if len(u.Items) == 0 {
		return nil
	}
	val := u.Items[0]
	if len(u.Items) > 1 {
		u.Items = u.Items[1:]
	}
	return val
}

// MatchFromDefns extracts matches from multiple definition formulas.
// Corresponds to Python's match_from_defns (ivy_proof.py:1543-1547).
//
// Python:
//
//	matches = [match_from_defn(d) for d in defns]
//	lhs = list(matches[0].keys())[0]
//	assert all(lhs in m for m in matches)
//	return {lhs: [m[lhs] for m in matches]}
func MatchFromDefns(defns []*LabeledFormula) (NodeKey, *ExprListOrLambdaUnion, error) {
	if len(defns) == 0 {
		return "", nil, &ProofError{Msg: "no definitions"}
	}
	// Python: matches = [match_from_defn(d) for d in defns]
	matches := make([]map[NodeKey]Expr, len(defns))
	for i, d := range defns {
		m, err := MatchFromDefn(d)
		if err != nil {
			return "", nil, err
		}
		matches[i] = m
	}
	// Python: lhs = list(matches[0].keys())[0]
	var lhsKey NodeKey
	for k := range matches[0] {
		lhsKey = k
		break
	}
	// Python: assert all(lhs in m for m in matches)
	for _, m := range matches[1:] {
		if _, ok := m[lhsKey]; !ok {
			return "", nil, &ProofError{Msg: "match_from_defns: definitions have different LHS symbols"}
		}
	}
	// Python: return {lhs: [m[lhs] for m in matches]}
	items := make([]Expr, len(matches))
	for i, m := range matches {
		items[i] = m[lhsKey]
	}
	return lhsKey, &ExprListOrLambdaUnion{Items: items}, nil
}

// unfoldRhsVars computes free vars from all lambdas in a union.
// Python: match_rhs_vars handles list values via:
//
//	for v in w if isinstance(w, list) else [w]: ...
func unfoldRhsVars(union *ExprListOrLambdaUnion) map[NodeKey]Expr {
	result := make(map[NodeKey]Expr)
	for _, item := range union.Items {
		for k, sym := range FmlaVocab(item) {
			result[k] = sym
		}
	}
	return result
}

// applyUnfoldRec recursively unfolds a formula with destructive pop.
// Simplified version of applyMatchAltRec for the single-key unfold case.
// Each occurrence of the unfold key pops the next lambda from the union.
func applyUnfoldRec(key NodeKey, union *ExprListOrLambdaUnion, fmla Expr) Expr {
	if fmla == nil {
		return nil
	}
	args := NodeArgs(fmla)
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		newArgs[i] = applyUnfoldRec(key, union, a)
	}
	// App: if function matches the unfold key, pop and beta-reduce
	if IsApp(fmla) {
		if app, ok := fmla.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok && Key(c) == key {
				lam := union.Pop()
				if l, ok := lam.(*Lambda); ok {
					result, err := LambdaApply(l, newArgs)
					if err != nil {
						return fmla // capture — return original
					}
					return result
				}
				// Non-lambda: apply as function to args.
				// Python: Symbol.__call__(*args) creates Apply(fun, args).
				if len(newArgs) > 0 {
					app, err := NewApply(lam, newArgs...)
					if err == nil {
						return app
					}
				}
				return lam
			}
		}
	}
	// Binder: clone with processed vars and body
	if IsQuantifier(fmla) && len(newArgs) > 0 {
		vars := BinderVars(fmla)
		return CloneBinder(fmla, vars, newArgs[0])
	}
	if len(args) == 0 {
		return fmla
	}
	return CloneNode(fmla, newArgs)
}

// applyUnfoldGoal applies unfold to a goal (premises + conclusion).
// Recurses into premises, then unfolds the conclusion with destructive pop.
// Corresponds to Python's apply_match_goal called from unfold_goal.
func applyUnfoldGoal(cfg *AstConfig, key NodeKey, union *ExprListOrLambdaUnion, freeVars map[NodeKey]Expr, goal *LabeledFormula) *LabeledFormula {
	// Process premises recursively
	prems := GoalPrems(goal)
	var newPrems []Node
	for _, p := range prems {
		if lf, ok := p.(*LabeledFormula); ok {
			newPrems = append(newPrems, applyUnfoldGoal(cfg, key, union, freeVars, lf))
		} else {
			newPrems = append(newPrems, p)
		}
	}
	// Filter ConstantDecl premises matching the unfold key.
	// Python: apply_match_goal transforms ConstantDecl via apply_match_func_alt
	// (which pops a lambda via match_get), then filters lambda-typed premises:
	//   prems = [p for p in prems if not is_lambda(p)]
	// For the unfold path, a ConstantDecl whose symbol matches the unfold key
	// would become lambda-typed and then be filtered out.
	filtered := make([]Node, 0, len(newPrems))
	for _, p := range newPrems {
		if cd, ok := p.(*ConstantDecl); ok {
			args := cd.Args()
			if len(args) > 0 {
				if c, ok := args[0].(*Const); ok && Key(c) == key {
					continue // Python would make this lambda-typed, then filter
				}
			}
		}
		filtered = append(filtered, p)
	}
	newPrems = filtered

	// Unfold the conclusion with alpha-avoid + destructive pop
	newConc := ApplyToConc(GoalConc(goal), func(c Expr) Expr {
		c = AlphaAvoidMap(c, freeVars)
		return applyUnfoldRec(key, union, c)
	})
	return CloneGoal(cfg, goal, newPrems, newConc)
}

// UnfoldGoal unfolds definitions in a goal.
// Corresponds to Python's unfold_goal (ivy_proof.py:1549-1553).
func UnfoldGoal(cfg *AstConfig, goal *LabeledFormula, defns [][]*LabeledFormula) *LabeledFormula {
	for _, rdefs := range defns {
		key, union, err := MatchFromDefns(rdefs)
		if err != nil {
			continue
		}
		freeVars := unfoldRhsVars(union)
		goal = applyUnfoldGoal(cfg, key, union, freeVars, goal)
	}
	return goal
}

// UnfoldFmla unfolds definitions in a formula.
// Corresponds to Python's unfold_fmla (ivy_proof.py:1557-1561).
//
// Takes/returns ast.Node so it can handle *ast.TemporalModels (mirroring
// Python's duck-typed recursion via apply_match_alt at
// ivy_proof.py:1150-1168 which does
// `fmla.clone([apply_match_alt_rec(match,f,env) for f in fmla.args])`).
// For lg.Expr inputs, behaves identically to the previous lg.Expr-only
// signature. This mirrors the ast.Node broadening already applied to
// SkolemizeFmla (proof/skolem.go:100) and WitnessAst (module/skolem.go).
func UnfoldFmla(fmla Node, defns [][]*LabeledFormula) Node {
	// TemporalModels — recurse into the wrapped inner formula and rewrap,
	// mirroring Python's generic apply_match_alt_rec recursion.
	if tm, ok := fmla.(*TemporalModels); ok {
		innerExpr, ok := tm.Fmla.(Expr)
		if !ok {
			return tm
		}
		return tm.Clone([]Node{UnfoldFmla(innerExpr, defns)})
	}
	expr, ok := fmla.(Expr)
	if !ok {
		return fmla
	}
	for _, rdefs := range defns {
		key, union, err := MatchFromDefns(rdefs)
		if err != nil {
			continue
		}
		// Alpha-avoid: rename bound vars that clash with free vars in ALL lambdas.
		// Python: apply_match_alt calls match_rhs_vars then alpha_avoid.
		freeVars := unfoldRhsVars(union)
		expr = AlphaAvoidMap(expr, freeVars)
		// Single-pass unfold with destructive pop
		expr = applyUnfoldRec(key, union, expr)
	}
	return expr
}

// GoalApplyToPrem applies a function to a specific premise by name.
// Corresponds to Python's goal_apply_to_prem.
func GoalApplyToPrem(cfg *AstConfig, goal *LabeledFormula, premName string, fn func(*LabeledFormula) *LabeledFormula) *LabeledFormula {
	xtracer.Trace("proof.GoalApplyToPrem ENTER goalLabel=%s premName=%s", goal.LabelForTrace(), premName)
	prems := GoalPrems(goal)
	for i, p := range prems {
		if lf, ok := p.(*LabeledFormula); ok {
			if lf.LabelName() == premName {
				newPrems := make([]Node, len(prems))
				copy(newPrems, prems)
				newPrems[i] = fn(lf)
				result := CloneGoal(cfg, goal, newPrems, GoalConc(goal))
				xtracer.Trace("proof.GoalApplyToPrem EXIT HASH canon=%v", result.Canon())
				return result
			}
		}
	}
	xtracer.Trace("proof.GoalApplyToPrem EXIT notFound premName=%s", premName)
	return nil
}

// (GoalApplyToConc moved to proof/goal.go. The new version takes
// fn func(ast.Node) ast.Node — fn is responsible for handling any concrete
// type, OR the caller wraps fn with ApplyToConc to get TemporalModels
// unwrapping for free.)

// CloseUnmatched universally quantifies unmatched free variables in the conclusion.
// Corresponds to Python's close_unmatched (ivy_proof.py:1835).
//
// Python wraps the conclusion directly in il.IvyForAll. For plain logic
// conclusions we use the logic ForAll. For raw AST conclusions such as
// TemporalModels, use the AST Forall so the quantifier wraps the raw
// conclusion instead of being pushed inside it.
func CloseUnmatched(cfg *AstConfig, goal *LabeledFormula, match map[NodeKey]Expr) *LabeledFormula {
	xtracer.Trace("proof.CloseUnmatched ENTER label=%s nmatch=%d", goal.LabelForTrace(), len(match))
	rawConc := GoalConc(goal)
	if rawConc == nil {
		xtracer.Trace("proof.CloseUnmatched EXIT rawConcNil")
		return goal
	}
	concExpr := ConcAsExpr(rawConc)
	if concExpr == nil {
		xtracer.Trace("proof.CloseUnmatched EXIT concNotExpr")
		return goal
	}
	// Python: prem_vars = lu.used_variables_asts(goal_prem_goals(goal))
	premVars := make(map[NodeKey]bool)
	for _, pg := range GoalPremGoals(goal) {
		pgConcExpr := ConcAsExpr(GoalConc(pg))
		if pgConcExpr != nil {
			for _, v := range VariablesAST(pgConcExpr) {
				premVars[Key(v)] = true
			}
		}
	}
	// Python: conc_vars = [x for x in iu.unique(lu.variables_ast(conc))
	//                     if x not in match and x not in prem_vars]
	concVars := VariablesAST(concExpr)
	var toClose []*LogicVariable
	for _, v := range concVars {
		k := Key(v)
		if _, inMatch := match[k]; !inMatch {
			if !premVars[k] {
				toClose = append(toClose, v)
			}
		}
	}
	var finalConc Node = rawConc
	if len(toClose) > 0 {
		if _, isExpr := rawConc.(Expr); isExpr {
			newConc := concExpr
			for i := len(toClose) - 1; i >= 0; i-- {
				newConc = IvyForAll([]*LogicVariable{toClose[i]}, newConc)
			}
			finalConc = newConc
		} else {
			for i := len(toClose) - 1; i >= 0; i-- {
				finalConc = cfg.NewForall([]Node{toClose[i]}, finalConc)
			}
		}
	}
	result := CloneGoal(cfg, goal, GoalPrems(goal), finalConc)
	xtracer.Trace("proof.CloseUnmatched EXIT ntoClose=%d HASH canon=%v", len(toClose), result.Canon())
	return result
}

// DropSuppliedPrems removes premises from schema that are supplied by goal.
// Corresponds to Python's drop_supplied_prems.
func DropSuppliedPrems(cfg *AstConfig, schema, goal *LabeledFormula, proofMatch []Node) *LabeledFormula {
	xtracer.Trace("proof.DropSuppliedPrems ENTER schemaLabel=%s goalLabel=%s nMatch=%d", schema.LabelForTrace(), goal.LabelForTrace(), len(proofMatch))
	gprems := GoalPremsByName(goal)
	pmap := make(map[string]string)
	for _, m := range proofMatch {
		defn, ok := m.(*Definition)
		if !ok {
			continue
		}
		if lAtom, ok := defn.Lhs.(*Atom); ok && len(lAtom.Terms) == 0 {
			if rAtom, ok := defn.Rhs.(*Atom); ok && len(rAtom.Terms) == 0 {
				pmap[lAtom.Rep] = rAtom.Rep
			}
		}
	}
	isSupplied := func(prem Node) bool {
		lf, ok := prem.(*LabeledFormula)
		if !ok {
			return false
		}
		gname, ok := pmap[lf.LabelName()]
		if !ok {
			return false
		}
		gprem, ok := gprems[gname]
		if !ok {
			return false
		}
		return GoalsEqModAlpha(lf, gprem)
	}
	var newPrems []Node
	for _, p := range GoalPrems(schema) {
		if !isSupplied(p) {
			newPrems = append(newPrems, p)
		}
	}
	result := CloneGoal(cfg, schema, newPrems, GoalConc(schema))
	xtracer.Trace("proof.DropSuppliedPrems EXIT nnewPrems=%d", len(newPrems))
	return result
}

// RemoveExplicit clears the explicit flag on a goal.
// Corresponds to Python's remove_explicit.
func RemoveExplicit(goal *LabeledFormula) *LabeledFormula {
	xtracer.Trace("proof.RemoveExplicit ENTER label=%s explicit=%v", goal.LabelForTrace(), goal.Explicit)
	if goal.Explicit {
		newGoal := goal.Clone(goal.Args()).(*LabeledFormula)
		newGoal.Explicit = false
		xtracer.Trace("proof.RemoveExplicit EXIT cleared")
		return newGoal
	}
	xtracer.Trace("proof.RemoveExplicit EXIT passthrough")
	return goal
}

// RenamePremNoClash renames a premise to avoid clashing with existing premise names.
// Corresponds to Python's rename_prem_no_clash.
func RenamePremNoClash(prem, decl *LabeledFormula) *LabeledFormula {
	xtracer.Trace("proof.RenamePremNoClash ENTER premLabel=%s declLabel=%s", prem.LabelForTrace(), decl.LabelForTrace())
	var used []string
	for _, pg := range GoalPremGoals(decl) {
		used = append(used, pg.LabelName())
	}
	rn := NewUniqueRenamer("", used)
	newName := rn.Rename(prem.LabelName())
	result := prem.Rename(newName)
	xtracer.Trace("proof.RenamePremNoClash EXIT newName=%s", newName)
	return result
}

// GoalPremsByName creates a dict mapping premise names to premise goals.
// Corresponds to Python's goal_prems_by_name.
func GoalPremsByName(goal *LabeledFormula) map[string]*LabeledFormula {
	result := make(map[string]*LabeledFormula)
	for _, p := range GoalPremGoals(goal) {
		name := p.LabelName()
		if name != "" {
			result[name] = p
		}
	}
	return result
}
