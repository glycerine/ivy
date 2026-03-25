// phase5_goals.go implements Phase 5 Batch 5.1 (Goal Utilities) and
// Batch 5.4 (Tactics/Remaining) functions from the Ivy port.
// These correspond to Python ivy_proof.py goal manipulation functions.
package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// === Batch 5.1: Goal Utilities ===

// AttribGoals copies line number information from a proof node to goals.
// Corresponds to Python's attrib_goals.
func AttribGoals(proof ast.Node, goals []*ast.LabeledFormula) []*ast.LabeledFormula {
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
func IsGoal(g ast.Node) bool {
	_, ok := g.(*ast.LabeledFormula)
	return ok
}

// GoalsSubst substitutes multiple subgoals into the first goal's conclusion.
// Corresponds to Python's goals_subst.
func GoalsSubst(cfg *ast.AstConfig, goals []*ast.LabeledFormula, subgoals []*ast.LabeledFormula, loc ast.Location) []*ast.LabeledFormula {
	if len(goals) == 0 || len(subgoals) == 0 {
		return goals
	}
	var result []*ast.LabeledFormula
	for _, sg := range subgoals {
		result = append(result, GoalSubst(cfg, goals[0], sg, loc))
	}
	result = append(result, goals[1:]...)
	return result
}

// FreshLabel generates a fresh unique label not used in any goal.
// Corresponds to Python's fresh_label.
func FreshLabel(cfg *ast.AstConfig, goals []*ast.LabeledFormula) ast.Node {
	var used []string
	for _, g := range goals {
		name := g.LabelName()
		if name != "" {
			used = append(used, name)
		}
	}
	rn := iu.NewUniqueRenamer("", used)
	name := rn.Rename("")
	return cfg.NewAtom(name)
}

// GoalPrefixPrems adds premises to the beginning of a goal's premise list.
// Corresponds to Python's goal_prefix_prems.
func GoalPrefixPrems(cfg *ast.AstConfig, goal *ast.LabeledFormula, prems []ast.Node, loc ast.Location) *ast.LabeledFormula {
	existingPrems := GoalPrems(goal)
	allPrems := make([]ast.Node, 0, len(prems)+len(existingPrems))
	allPrems = append(allPrems, prems...)
	allPrems = append(allPrems, existingPrems...)
	return MakeGoal(cfg, loc, goal.Label, allPrems, GoalConc(goal))
}

// CheckNameClash checks that two goals have no overlapping symbol definitions.
// Corresponds to Python's check_name_clash.
func CheckNameClash(g1, g2 *ast.LabeledFormula) error {
	d1 := GoalDefns(g1)
	d2 := GoalDefns(g2)
	for k := range d1 {
		if _, exists := d2[k]; exists {
			return &ProofError{Msg: "premise of subgoal clashes with context"}
		}
	}
	return nil
}

// CheckPremisesProvided verifies that all non-proposition premises of g1
// are provided by g2.
// Corresponds to Python's check_premises_provided (ivy_proof.py:586-593).
func CheckPremisesProvided(g1, g2 *ast.LabeledFormula, sig *il.Sig) error {
	defns := GoalDefns(g2)
	for k, thing := range GoalDefns(g1) {
		// Python: syms = [] if il.is_lambda(thing) else [thing]
		if _, isLam := thing.(*lg.Lambda); isLam {
			continue
		}
		if _, ok := defns[k]; !ok && !sig.Contains(thing) {
			return &ProofError{Msg: fmt.Sprintf("premise %v does not match anything in the environment", thing)}
		}
	}
	return nil
}

// GoalIsTemporal checks if a goal has temporal properties.
// Corresponds to Python's goal_is_temporal.
func GoalIsTemporal(x *ast.LabeledFormula) bool {
	if x.IsTemporal() {
		return true
	}
	conc := GoalConc(x)
	if conc == nil {
		return false
	}
	if nb, ok := conc.(*lg.NamedBinder); ok {
		return nb.Name == "globally" || nb.Name == "eventually"
	}
	return false
}

// GoalDefines extracts what a definition premise defines.
// Corresponds to Python's goal_defines.
func GoalDefines(x ast.Node) lg.Expr {
	if cd, ok := x.(*ast.ConstantDecl); ok {
		if len(cd.Args()) > 0 {
			if sym, ok := cd.Args()[0].(lg.Expr); ok {
				return sym
			}
		}
	}
	if n, ok := x.(lg.Expr); ok {
		return n
	}
	return nil
}

// GetUnprovidedDefns gets definitions from g1 that aren't provided by g2.
// Corresponds to Python's get_unprovided_defns.
func GetUnprovidedDefns(g1, g2 *ast.LabeledFormula) []ast.Node {
	defns := GoalDefns(g2)
	free := GoalFree(g2)
	var result []ast.Node
	for _, prem := range GoalPrems(g1) {
		if GoalIsDefn(prem) {
			sym := GoalDefines(prem)
			if sym == nil {
				continue
			}
			k := lg.Key(sym)
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
func GoalSubgoals(cfg *ast.AstConfig, schema, goal *ast.LabeledFormula, loc ast.Location) []*ast.LabeledFormula {
	if err := CheckConcsMatch(schema, goal); err != nil {
		return nil
	}
	upds := GetUnprovidedDefns(schema, goal)
	g := CloneGoal(cfg, goal, upds, GoalConc(goal))
	resultGoal := GoalSubst(cfg, goal, g, loc)

	gpms := GoalPremGoals(resultGoal)

	var subgoals []*ast.LabeledFormula
	for _, sp := range GoalPremGoals(schema) {
		alreadyPresent := false
		for _, gpm := range gpms {
			if GoalsEqModAlpha(sp, gpm) {
				alreadyPresent = true
				break
			}
		}
		if !alreadyPresent {
			sg := GoalSubst(cfg, resultGoal, sp, loc)
			if !TrivialGoal(sg) {
				subgoals = append(subgoals, sg)
			}
		}
	}
	return subgoals
}

// FmlaVocab gets the free vocabulary of a formula, including sorts, symbols, and variables.
// Corresponds to Python's fmla_vocab.
func FmlaVocab(fmla lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	// Use clauseops.UsedSymbolsAST for symbols
	for k, v := range co.UsedSymbolsAST(fmla) {
		result[k] = v
	}
	// Use clauseops.VariablesAST for variables
	for _, v := range co.VariablesAST(fmla) {
		result[lg.Key(v)] = v
	}
	return result
}

// CheckSchemaCapture verifies that free variables in schema are not captured by goal bindings.
// Corresponds to Python's check_schema_capture.
func CheckSchemaCapture(schema, goal *ast.LabeledFormula) error {
	gvocab := GoalVocab(goal)
	fvocab := GoalFree(schema)
	gSymSet := make(map[string]bool)
	for _, s := range gvocab.Symbols {
		gSymSet[s.Name] = true
	}
	gSortSet := make(map[string]bool)
	for _, s := range gvocab.Sorts {
		gSortSet[il.SortName(s)] = true
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
func CheckAlphaCapture(goal *ast.LabeledFormula, match map[lg.NodeKey]lg.Expr) error {
	revMatch := make(map[lg.NodeKey]lg.Expr)
	for k, v := range match {
		revMatch[lg.Key(v)] = match[k]
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
func CheckRenaming(goal *ast.LabeledFormula, renaming ast.Node) error {
	fwd := make(map[string]string)
	rev := make(map[string]string)
	for _, arg := range renaming.Args() {
		defn, ok := arg.(*ast.Definition)
		if !ok {
			continue
		}
		lAtom, lok := defn.Lhs.(*ast.Atom)
		rAtom, rok := defn.Rhs.(*ast.Atom)
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
func GoalIsProperty(x *ast.LabeledFormula) bool {
	_, isSchema := x.Formula.(*ast.SchemaBody)
	return !isSchema
}

// GoalIsSchema returns true if the goal is a schema.
// Corresponds to Python's goal_is_schema.
func GoalIsSchema(x *ast.LabeledFormula) bool {
	_, isSchema := x.Formula.(*ast.SchemaBody)
	return isSchema
}

// GoalsEqModAlpha checks if two goals are equivalent modulo alpha renaming.
// Corresponds to Python's goals_eq_mod_alpha.
func GoalsEqModAlpha(x, y *ast.LabeledFormula) bool {
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
			xlf, xIsLF := xps[i].(*ast.LabeledFormula)
			ylf, yIsLF := yps[i].(*ast.LabeledFormula)
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
	xConc := GoalConc(x)
	yConc := GoalConc(y)
	if xConc == nil || yConc == nil {
		return xConc == yConc
	}
	return lu.EqualModAlpha(xConc, yConc)
}

// IsLambdaPrem checks if a premise is a lambda function definition.
// Corresponds to Python's is_lambda.
func IsLambdaPrem(p ast.Node) bool {
	cd, ok := p.(*ast.ConstantDecl)
	if !ok {
		return false
	}
	args := cd.Args()
	if len(args) == 0 {
		return false
	}
	// Check if the first arg is a lambda (wrapped as lg.Expr in ast.Node)
	if n, ok := args[0].(lg.Expr); ok {
		_, isLam := n.(*lg.Lambda)
		return isLam
	}
	return false
}

// === Batch 5.4: Tactics and Remaining ===

// VarSubstGoal applies a variable substitution to a goal.
// Corresponds to Python's var_subst_goal.
func VarSubstGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, subst map[lg.NodeKey]lg.Expr) *ast.LabeledFormula {
	var prems []ast.Node
	for _, prem := range GoalPrems(goal) {
		if premLF, ok := prem.(*ast.LabeledFormula); ok {
			prems = append(prems, VarSubstGoal(cfg, premLF, subst))
		} else {
			prems = append(prems, prem)
		}
	}
	conc := GoalConc(goal)
	if conc != nil && !GoalIsSchema(goal) {
		conc = ApplyToConc(conc, func(x lg.Expr) lg.Expr {
			return co.SubstituteAstByName(x, nodeMapToStringMap(subst))
		})
	}
	return CloneGoal(cfg, goal, prems, conc)
}

// nodeMapToStringMap converts lg.NodeKey->lg.Expr map to string->lg.Expr for substitution.
func nodeMapToStringMap(m map[lg.NodeKey]lg.Expr) map[string]lg.Expr {
	result := make(map[string]lg.Expr, len(m))
	for _, v := range m {
		if variable, ok := v.(*lg.Variable); ok {
			result[variable.Name] = v
		}
	}
	return result
}

// ApplyToConc applies a function to a goal's conclusion, handling temporal models.
// Corresponds to Python's apply_to_conc.
func ApplyToConc(conc lg.Expr, fn func(lg.Expr) lg.Expr) lg.Expr {
	if nb, ok := conc.(*lg.NamedBinder); ok {
		if nb.Name == "globally" || nb.Name == "eventually" {
			newBody := fn(nb.Body)
			return &lg.NamedBinder{Name: nb.Name, Variables: nb.Variables, Environ: nb.Environ, Body: newBody}
		}
	}
	return fn(conc)
}

// RemoveUnusedDefinitionsGoal removes definitions that aren't referenced
// in the goal's conclusion.
// Corresponds to Python's remove_unused_definitions_goal.
func RemoveUnusedDefinitionsGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula) *ast.LabeledFormula {
	prems := GoalPrems(goal)
	conc := GoalConc(goal)
	if conc == nil {
		return goal
	}
	usedSyms := co.UsedSymbolsAST(conc)
	var newPrems []ast.Node
	reversed := make([]ast.Node, len(prems))
	for i, p := range prems {
		reversed[len(prems)-1-i] = p
	}
	for _, x := range reversed {
		if GoalIsDefn(x) {
			sym := GoalDefines(x)
			if sym == nil {
				newPrems = append([]ast.Node{x}, newPrems...)
				continue
			}
			if _, used := usedSyms[lg.Key(sym)]; !used {
				continue
			}
		}
		if n, ok := x.(lg.Expr); ok {
			for k, v := range co.UsedSymbolsAST(n) {
				usedSyms[k] = v
			}
		}
		newPrems = append([]ast.Node{x}, newPrems...)
	}
	return CloneGoal(cfg, goal, newPrems, conc)
}

// MatchFromDefn extracts a match from a definition formula.
// Corresponds to Python's match_from_defn.
func MatchFromDefn(defn *ast.LabeledFormula) (map[lg.NodeKey]lg.Expr, error) {
	conc := GoalConc(defn)
	if conc == nil {
		return nil, &ProofError{Msg: "not a definition"}
	}
	fmla := conc
	for {
		if fa, ok := fmla.(*lg.ForAll); ok {
			fmla = fa.Body
		} else {
			break
		}
	}
	if eq, ok := fmla.(*lg.Eq); ok {
		if app, ok := eq.T1.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Symbol); ok {
				lam := &lg.Lambda{Variables: nodesToVarsPhase5(app.Terms), Body: eq.T2}
				result := make(map[lg.NodeKey]lg.Expr)
				result[lg.Key(c)] = lam
				return result, nil
			}
		}
	}
	if iff, ok := fmla.(*lg.Iff); ok {
		if app, ok := iff.T1.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Symbol); ok {
				lam := &lg.Lambda{Variables: nodesToVarsPhase5(app.Terms), Body: iff.T2}
				result := make(map[lg.NodeKey]lg.Expr)
				result[lg.Key(c)] = lam
				return result, nil
			}
		}
	}
	return nil, &ProofError{Msg: "not a definition", Node: fmla}
}

func nodesToVarsPhase5(nodes []lg.Expr) []*lg.Variable {
	var result []*lg.Variable
	for _, n := range nodes {
		if v, ok := n.(*lg.Variable); ok {
			result = append(result, v)
		}
	}
	return result
}

// MatchFromDefns extracts matches from multiple definition formulas.
// Corresponds to Python's match_from_defns.
func MatchFromDefns(defns []*ast.LabeledFormula) (map[lg.NodeKey]lg.Expr, error) {
	if len(defns) == 0 {
		return nil, &ProofError{Msg: "no definitions"}
	}
	return MatchFromDefn(defns[0])
}

// UnfoldGoal unfolds definitions in a goal.
// Corresponds to Python's unfold_goal.
func UnfoldGoal(goal *ast.LabeledFormula, defns [][]*ast.LabeledFormula) *ast.LabeledFormula {
	for _, rdefs := range defns {
		match, err := MatchFromDefns(rdefs)
		if err != nil {
			continue
		}
		goal = ApplyMatchGoalNode(match, goal)
	}
	return goal
}

// UnfoldFmla unfolds definitions in a formula.
// Corresponds to Python's unfold_fmla.
func UnfoldFmla(fmla lg.Expr, defns [][]*ast.LabeledFormula) lg.Expr {
	for _, rdefs := range defns {
		match, err := MatchFromDefns(rdefs)
		if err != nil {
			continue
		}
		fmla = ApplyMatchAlt(match, fmla, nil)
	}
	return fmla
}

// GoalApplyToPrem applies a function to a specific premise by name.
// Corresponds to Python's goal_apply_to_prem.
func GoalApplyToPrem(cfg *ast.AstConfig, goal *ast.LabeledFormula, premName string, fn func(*ast.LabeledFormula) *ast.LabeledFormula) *ast.LabeledFormula {
	prems := GoalPrems(goal)
	for i, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if lf.LabelName() == premName {
				newPrems := make([]ast.Node, len(prems))
				copy(newPrems, prems)
				newPrems[i] = fn(lf)
				return CloneGoal(cfg, goal, newPrems, GoalConc(goal))
			}
		}
	}
	return nil
}

// GoalApplyToConc applies a function to the conclusion of a goal.
// Corresponds to Python's goal_apply_to_conc.
func GoalApplyToConc(cfg *ast.AstConfig, goal *ast.LabeledFormula, fn func(lg.Expr) lg.Expr) *ast.LabeledFormula {
	return CloneGoal(cfg, goal, GoalPrems(goal), fn(GoalConc(goal)))
}

// CloseUnmatched universally quantifies unmatched free variables in the conclusion.
// Corresponds to Python's close_unmatched.
func CloseUnmatched(cfg *ast.AstConfig, goal *ast.LabeledFormula, match map[lg.NodeKey]lg.Expr) *ast.LabeledFormula {
	conc := GoalConc(goal)
	if conc == nil {
		return goal
	}
	premVars := make(map[lg.NodeKey]bool)
	for _, pg := range GoalPremGoals(goal) {
		pgConc := GoalConc(pg)
		if pgConc != nil {
			for _, v := range co.VariablesAST(pgConc) {
				premVars[lg.Key(v)] = true
			}
		}
	}
	concVars := co.VariablesAST(conc)
	var toClose []*lg.Variable
	for _, v := range concVars {
		k := lg.Key(v)
		if _, inMatch := match[k]; !inMatch {
			if !premVars[k] {
				toClose = append(toClose, v)
			}
		}
	}
	for i := len(toClose) - 1; i >= 0; i-- {
		conc = il.ForAll([]*lg.Variable{toClose[i]}, conc)
	}
	return CloneGoal(cfg, goal, GoalPrems(goal), conc)
}

// DropSuppliedPrems removes premises from schema that are supplied by goal.
// Corresponds to Python's drop_supplied_prems.
func DropSuppliedPrems(cfg *ast.AstConfig, schema, goal *ast.LabeledFormula, proofMatch []ast.Node) *ast.LabeledFormula {
	gprems := GoalPremsByName(goal)
	pmap := make(map[string]string)
	for _, m := range proofMatch {
		defn, ok := m.(*ast.Definition)
		if !ok {
			continue
		}
		if lAtom, ok := defn.Lhs.(*ast.Atom); ok && len(lAtom.Terms) == 0 {
			if rAtom, ok := defn.Rhs.(*ast.Atom); ok && len(rAtom.Terms) == 0 {
				pmap[lAtom.Rep] = rAtom.Rep
			}
		}
	}
	isSupplied := func(prem ast.Node) bool {
		lf, ok := prem.(*ast.LabeledFormula)
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
	var newPrems []ast.Node
	for _, p := range GoalPrems(schema) {
		if !isSupplied(p) {
			newPrems = append(newPrems, p)
		}
	}
	return CloneGoal(cfg, schema, newPrems, GoalConc(schema))
}

// RemoveExplicit clears the explicit flag on a goal.
// Corresponds to Python's remove_explicit.
func RemoveExplicit(goal *ast.LabeledFormula) *ast.LabeledFormula {
	if goal.Explicit {
		newGoal := goal.CloneWithFreshID(goal.Args())
		newGoal.Explicit = false
		return newGoal
	}
	return goal
}

// RenamePremNoClash renames a premise to avoid clashing with existing premise names.
// Corresponds to Python's rename_prem_no_clash.
func RenamePremNoClash(prem, decl *ast.LabeledFormula) *ast.LabeledFormula {
	var used []string
	for _, pg := range GoalPremGoals(decl) {
		used = append(used, pg.LabelName())
	}
	rn := iu.NewUniqueRenamer("", used)
	newName := rn.Rename(prem.LabelName())
	return prem.Rename(newName)
}

// GoalPremsByName creates a dict mapping premise names to premise goals.
// Corresponds to Python's goal_prems_by_name.
func GoalPremsByName(goal *ast.LabeledFormula) map[string]*ast.LabeledFormula {
	result := make(map[string]*ast.LabeledFormula)
	for _, p := range GoalPremGoals(goal) {
		name := p.LabelName()
		if name != "" {
			result[name] = p
		}
	}
	return result
}
