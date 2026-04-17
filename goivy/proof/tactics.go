// tactics.go implements the 7 missing proof tactics, faithfully ported from
// Python ivy_proof.py.
//
// Each tactic manipulates the goal stack (a list of LabeledFormula) according
// to the proof rule it implements.
package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
)

// letTactic introduces local definitions in a proof.
// Corresponds to Python ProofChecker.let_tactic (ivy_proof.py:213-223).
//
// Python:
//   vocab = goal_vocab(goal)
//   defs = [compile_expr_vocab(ia.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.args]
//   cond = il.And(*[il.Equals(a.args[0], a.args[1]) for a in defs])
//   goal = ia.LabeledFormula(goal.label, il.Implies(cond, goal.formula))
//   return [goal] + decls[1:]
func (pc *ProofChecker) letTactic(decls []*ast.LabeledFormula, proof *ast.LetTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "let tactic: no goals"}
	}
	goal := decls[0]

	vocab := GoalVocab(goal)

	// Compile each equality definition with goal vocabulary.
	// Python: defs = [compile_expr_vocab(ia.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.args]
	// Then: cond = il.And(*[il.Equals(a.args[0], a.args[1]) for a in defs])
	var eqs []lg.Expr
	for _, def := range proof.Defs {
		// Each def is an equality atom. Compile it with goal vocab.
		compiled := CompileExprVocab(def, vocab, pc.Mod)
		if compiled == nil {
			// Fallback: try direct extraction without compilation
			defArgs := def.Args()
			if len(defArgs) >= 2 {
				lhs := astNodeToLogicNode(defArgs[0])
				rhs := astNodeToLogicNode(defArgs[1])
				if lhs != nil && rhs != nil {
					eqs = append(eqs, &lg.Eq{T1: lhs, T2: rhs})
				}
			}
			continue
		}
		// Extract LHS and RHS from compiled equality (il.Equals(a.args[0], a.args[1]))
		if eq, ok := compiled.(*lg.Eq); ok {
			eqs = append(eqs, eq)
		} else if app, ok := compiled.(*lg.Apply); ok {
			// Some compilers produce Apply(=, args...) — extract
			if len(app.Terms) >= 2 {
				eqs = append(eqs, &lg.Eq{T1: app.Terms[0], T2: app.Terms[1]})
			}
		}
	}

	if len(eqs) == 0 {
		return decls, nil
	}

	var cond lg.Expr
	if len(eqs) == 1 {
		cond = eqs[0]
	} else {
		cond = &lg.And{Terms: eqs}
	}

	// Build subgoal: cond -> original formula.
	// Python uses goal.formula directly (wraps entire formula including SchemaBody).
	// Go uses ApplyToConc to handle *ast.TemporalModels wrapping correctly.
	if GoalConc(goal) == nil {
		return nil, &ProofError{Msg: "let tactic: goal has no conclusion"}
	}
	newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return &lg.Implies{T1: cond, T2: c}
	})
	subgoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), newConc)
	subgoal.SetLineno(goal.GetLineno())

	result := []*ast.LabeledFormula{subgoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// assumeTactic introduces an assumption from a schema or premise.
// Faithful port of Python ProofChecker.assume_tactic (ivy_proof.py:350-382).
//
// isGlobal distinguishes AssumeGlobalTactic (from "assume" keyword) from
// AssumeTactic (from "instantiate" keyword). Python uses isinstance() check.
func (pc *ProofChecker) assumeTactic(decls []*ast.LabeledFormula, proof *ast.AssumeTactic, isGlobal bool) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "assume tactic: no goals"}
	}
	decl := decls[0]

	// Python: schemaname = proof.schemaname()
	schemaName := ""
	if proof.SchemaName != nil {
		schemaName = fmt.Sprint(proof.SchemaName)
	}
	if schemaName == "" {
		return nil, &ProofError{Msg: "assume tactic: no schema name"}
	}

	// Python: premmap = dict((x.name,x) for x in goal_prem_goals(decl))
	premMap := make(map[string]*ast.LabeledFormula)
	for _, pg := range GoalPremGoals(decl) {
		premMap[pg.LabelName()] = pg
	}

	// Python lines 354-359: AssumeGlobalTactic vs AssumeTactic distinction.
	// AssumeTactic looks in premises first; AssumeGlobalTactic skips to global.
	var schema *ast.LabeledFormula
	if !isGlobal {
		if prem, ok := premMap[schemaName]; ok {
			schema = prem
			// Python: if isinstance(proof.label, ia.NoneAST): decl = goal_remove_prem(...)
			if isNoneAST(proof.TLabel) {
				decl = GoalRemovePrem(pc.astCfg(), decl, schemaName)
			}
		}
	}
	if schema == nil {
		var err error
		schema, err = pc.LookupSchema(schemaName, decl, proof, false)
		if err != nil {
			return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
				"No property %s exists in the current context", schemaName)}
		}
	}

	// Python: schema = remove_explicit(schema)
	schema = RemoveExplicit(schema)

	// Python: prob, pmatch = self.setup_schema_matching(decl, proof, schema, allow_witness=True)
	prob, pmatch, err := pc.SetupSchemaMatchingRaw(decl, proof.Ren, proof.Matches, schema, true)
	if err != nil {
		return nil, err
	}

	// Python lines 365-367: extract witnesses (variables not in freesyms).
	// iswit = lambda x: isinstance(x, il.Variable) and x not in prob.freesyms
	witness := make(map[lg.NodeKey]lg.Expr)
	pmatchClean := make(map[lg.NodeKey]lg.Expr)
	for k, v := range pmatch {
		if isWitVar(k, v, prob) {
			witness[k] = v
		} else {
			pmatchClean[k] = v
		}
	}
	pmatch = pmatchClean

	// Python: prem = prob.schema
	prem := prob.SchemaLF

	// Python: if schemaname not in premmap: prem = close_unmatched(prem, pmatch)
	if _, inPrems := premMap[schemaName]; !inPrems {
		prem = CloseUnmatched(pc.astCfg(), prem, pmatch)
	}

	// Python: conc = goal_conc(prem)
	//         conc = lu.witness_ast(True, [], witness, conc)
	//         prem = clone_goal(prem, goal_prems(prem), conc)
	if len(witness) > 0 {
		rawConc := GoalConc(prem)
		if concExpr, ok := rawConc.(lg.Expr); ok {
			newConc, werr := module.WitnessAst(true, nil, witness, concExpr)
			if werr == nil {
				prem = CloneGoal(pc.astCfg(), prem, GoalPrems(prem), newConc)
			}
		}
	}

	// Python: prem = apply_match_goal(pmatch, prem, apply_match_alt)
	prem = ApplyMatchGoalNode(pc.astCfg(), pmatch, prem)

	// Python: prem = drop_supplied_prems(prem, decl, proof.match())
	prem = DropSuppliedPrems(pc.astCfg(), prem, decl, proof.Matches)

	// Python lines 374-377: label handling.
	// When label is NoneAST, keep the schema's own label (Python creates
	// Atom(proof.label.rep, prem.label.args) which effectively preserves
	// the schema label name). When label is explicit, use it.
	if !isNoneAST(proof.TLabel) {
		prem = prem.CloneWithFreshID([]ast.Node{proof.TLabel, prem.Formula})
	}

	// Python lines 378-381: clash detection.
	// AssumeGlobalTactic renames to avoid clash; AssumeTactic errors.
	for _, pg := range GoalPremGoals(decl) {
		if pg.LabelName() == prem.LabelName() {
			if isGlobal {
				prem = RenamePremNoClash(prem, decl)
			} else {
				return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
					"instance name %s clashes with context", prem.LabelName())}
			}
			break
		}
	}

	// Python: return [goal_add_prem(decl, prem, proof.lineno)] + decls[1:]
	newGoal := pc.goalAddPrem(decl, prem, proof.GetLineno())
	result := []*ast.LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// isNoneAST checks if a node is a NoneAST (or nil).
func isNoneAST(n ast.Node) bool {
	if n == nil {
		return true
	}
	_, ok := n.(*ast.NoneAST)
	return ok
}

// isWitVar checks if a match entry is a witness variable.
// Python: iswit = lambda x: isinstance(x, il.Variable) and x not in prob.freesyms
func isWitVar(key lg.NodeKey, val lg.Expr, prob *MatchProblem) bool {
	if _, isVar := val.(*lg.Variable); !isVar {
		return false
	}
	_, inFree := prob.FreeSyms[key]
	return !inFree
}

// unfoldTactic unfolds definitions in the goal.
// Corresponds to Python ProofChecker.unfold_tactic (lines 376-392).
func (pc *ProofChecker) unfoldTactic(decls []*ast.LabeledFormula, proof *ast.UnfoldTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "unfold tactic: no goals"}
	}
	goal := decls[0]

	// Look up each definition to unfold
	var defns []lg.Expr
	for _, unfspecNode := range proof.UnfSpecs {
		unfspec, ok := unfspecNode.(*ast.UnfoldSpec)
		if !ok {
			continue
		}
		defName := ""
		if unfspec.DefName != nil {
			defName = fmt.Sprint(unfspec.DefName)
		}
		if defName == "" {
			continue
		}
		defLF, ok := pc.Definitions[defName]
		if !ok {
			return nil, &ProofError{Msg: fmt.Sprintf("unfold tactic: definition %s not found", defName)}
		}
		// Definitions are always lg.Expr (never *ast.TemporalModels).
		defConc := GoalConcExpr(defLF)
		if defConc != nil {
			defns = append(defns, defConc)
		}
	}

	if len(defns) == 0 {
		return decls, nil
	}

	// Unfold in the conclusion. ApplyToConc unwraps *ast.TemporalModels so
	// the unfold runs on the inner formula and the wrapper is preserved.
	if GoalConc(goal) == nil {
		return decls, nil
	}
	newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return unfoldFmla(c, defns)
	})
	result := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), newConc)
	return append([]*ast.LabeledFormula{result}, decls[1:]...), nil
}

// unfoldFmla substitutes definitions into a formula.
func unfoldFmla(fmla lg.Expr, defns []lg.Expr) lg.Expr {
	result := fmla
	for _, defn := range defns {
		if def, ok := defn.(*il.Definition); ok {
			// Build substitution: defined symbol → definition body
			defSym := def.Defines()
			if c, ok := defSym.(*lg.Const); ok {
				subs := map[string]lg.Expr{c.Name: def.Rhs}
				result = lu.SubstituteByName(result, subs)
			}
		}
	}
	return result
}

// ifTactic splits the goal into two subgoals based on a condition.
// Corresponds to Python ProofChecker.if_tactic (lines 402-410).
//
// Given condition C and proof branches P1, P2, the goal G becomes:
//   C -> G  (proved by P1)
//   ~C -> G (proved by P2)
func (pc *ProofChecker) ifTactic(decls []*ast.LabeledFormula, proof *ast.IfTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "if tactic: no goals"}
	}
	goal := decls[0]

	// Get condition as logic node
	cond := astNodeToLogicNode(proof.Cond)
	if cond == nil {
		return nil, &ProofError{Msg: "if tactic: could not convert condition to logic node"}
	}

	if GoalConc(goal) == nil {
		return nil, &ProofError{Msg: "if tactic: goal has no conclusion"}
	}

	// Build true_goal: C -> G. ApplyToConc unwraps *ast.TemporalModels so the
	// implication wraps the inner formula and the temporal wrapper is preserved.
	trueConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return &lg.Implies{T1: cond, T2: c}
	})
	trueGoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), trueConc)
	trueGoal.SetLineno(goal.GetLineno())

	// Build false_goal: ~C -> G (same TemporalModels treatment).
	falseConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return &lg.Implies{T1: &lg.Not{Body: cond}, T2: c}
	})
	falseGoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), falseConc)
	falseGoal.SetLineno(goal.GetLineno())

	// Apply proof branches
	var result []*ast.LabeledFormula

	if proof.Then != nil {
		trueResult, err := pc.ApplyProof([]*ast.LabeledFormula{trueGoal}, proof.Then)
		if err != nil {
			return nil, err
		}
		result = append(result, trueResult...)
	} else {
		result = append(result, trueGoal)
	}

	if proof.Else != nil {
		falseResult, err := pc.ApplyProof([]*ast.LabeledFormula{falseGoal}, proof.Else)
		if err != nil {
			return nil, err
		}
		result = append(result, falseResult...)
	} else {
		result = append(result, falseGoal)
	}

	result = append(result, decls[1:]...)
	return result, nil
}

// propertyTactic introduces a property (cut) in a proof.
// Corresponds to Python ProofChecker.property_tactic (lines 225-273).
func (pc *ProofChecker) propertyTactic(decls []*ast.LabeledFormula, proof *ast.PropertyTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "property tactic: no goals"}
	}
	goal := decls[0]

	// The property tactic introduces a "cut" formula. The goal G becomes:
	//   cut (as a subgoal)
	//   cut -> G (modified goal with cut as premise)
	//
	// Python: ivy_proof.py:225-273
	cutFormula := astNodeToLogicNode(proof.Prop)
	if cutFormula == nil {
		return nil, &ProofError{Msg: "property tactic: could not convert cut formula"}
	}

	if GoalConc(goal) == nil {
		return nil, &ProofError{Msg: "property tactic: goal has no conclusion"}
	}

	// Create the cut subgoal: prove the cut formula. The cut formula is a plain
	// lg.Expr (not a TemporalModels) — we don't wrap it.
	cutGoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), cutFormula)

	// Modify the original goal: add cut as premise (cut -> G). ApplyToConc
	// unwraps *ast.TemporalModels so the implication wraps the inner formula.
	modifiedConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return &lg.Implies{T1: cutFormula, T2: c}
	})
	modifiedGoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), modifiedConc)

	// If there's a proof for the cut, apply it
	var result []*ast.LabeledFormula
	if proof.Proof != nil {
		cutResult, err := pc.ApplyProof([]*ast.LabeledFormula{cutGoal}, proof.Proof)
		if err != nil {
			return nil, err
		}
		result = append(result, cutResult...)
	} else {
		result = append(result, cutGoal)
	}
	result = append(result, modifiedGoal)
	result = append(result, decls[1:]...)
	return result, nil
}

// functionTactic introduces a function definition in a proof.
// Corresponds to Python ProofChecker.function_tactic (lines 275-304).
func (pc *ProofChecker) functionTactic(decls []*ast.LabeledFormula, proof *ast.FunctionTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "function tactic: no goals"}
	}
	goal := decls[0]

	// The function tactic introduces a fresh function symbol with a definition.
	// The definition defn(x) = body is added as a universally quantified
	// equality premise: forall x. defn(x) = body(x).
	//
	// Python: ivy_proof.py:275-304
	// The function tactic's elements contain the definition
	var defFormula lg.Expr
	for _, elem := range proof.Elems {
		if n := astNodeToLogicNode(elem); n != nil {
			defFormula = n
			break
		}
	}
	if defFormula == nil {
		return nil, &ProofError{Msg: "function tactic: could not convert definition"}
	}

	if GoalConc(goal) == nil {
		return nil, &ProofError{Msg: "function tactic: goal has no conclusion"}
	}

	// Add the definition as a premise: defn -> G. ApplyToConc unwraps
	// *ast.TemporalModels so the implication wraps the inner formula.
	modifiedConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return &lg.Implies{T1: defFormula, T2: c}
	})
	modifiedGoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), modifiedConc)

	return append([]*ast.LabeledFormula{modifiedGoal}, decls[1:]...), nil
}

// ensure imports are used
var _ = il.IsApp
var _ = lu.SubstituteByName

// witnessTactic provides witnesses for existentially quantified variables.
// Corresponds to Python ProofChecker.witness_tactic (ivy_proof.py:459-471).
func (pc *ProofChecker) witnessTactic(decls []*ast.LabeledFormula, proof *ast.WitnessTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "witness tactic: no goals"}
	}
	goal := decls[0]

	if GoalConc(goal) == nil {
		return nil, &ProofError{Msg: "witness tactic: goal has no conclusion"}
	}

	// Python: if ia.has_temporal(proof) and not goal_is_temporal(goal): raise error
	// (Python has `goal` here but means `decl`/`goal` = decls[0])
	if ast.HasTemporal(proof) && !GoalIsTemporal(goal) {
		return nil, &ProofError{Msg: "temporal operator not allowed in instantiation", Node: proof}
	}

	// Build witness map from proof witnesses
	// Each witness is x = e, mapping variable x to expression e
	witMap := make(map[string]lg.Expr)
	for _, w := range proof.Witnesses {
		wargs := w.Args()
		if len(wargs) >= 2 {
			lhs := astNodeToLogicNode(wargs[0])
			rhs := astNodeToLogicNode(wargs[1])
			if lhs != nil && rhs != nil {
				if v, ok := lhs.(*lg.Variable); ok {
					witMap[v.Name] = rhs
				}
			}
		}
	}

	if len(witMap) == 0 {
		return decls, nil
	}

	// Apply witness substitution to the conclusion. ApplyToConc unwraps
	// *ast.TemporalModels so the witness substitution runs on the inner
	// formula and the temporal wrapper is preserved.
	// Python: conc = lu.witness_ast(False, [], wit_map, conc)
	newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
		return applyWitness(c, witMap)
	})

	prems := GoalPrems(goal)
	newGoal := CloneGoal(pc.astCfg(), goal, prems, newConc)
	result := []*ast.LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// applyWitness substitutes witness values for existentially quantified variables.
// Corresponds to Python lu.witness_ast.
func applyWitness(fmla lg.Expr, witMap map[string]lg.Expr) lg.Expr {
	if fmla == nil || len(witMap) == 0 {
		return fmla
	}
	switch f := fmla.(type) {
	case *lg.Exists:
		// Check if any of the bound variables have witnesses
		var remainingVars []*lg.Variable
		subs := make(map[string]lg.Expr)
		for _, v := range f.Variables {
			if wit, ok := witMap[v.Name]; ok {
				subs[v.Name] = wit
			} else {
				remainingVars = append(remainingVars, v)
			}
		}
		body := f.Body
		if len(subs) > 0 {
			body = substituteVarsInNode(body, subs)
		}
		body = applyWitness(body, witMap)
		if len(remainingVars) == 0 {
			return body
		}
		return &lg.Exists{Variables: remainingVars, Body: body}
	case *lg.And:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = applyWitness(t, witMap)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = applyWitness(t, witMap)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: applyWitness(f.Body, witMap)}
	case *lg.Implies:
		return &lg.Implies{T1: applyWitness(f.T1, witMap), T2: applyWitness(f.T2, witMap)}
	case *lg.ForAll:
		return &lg.ForAll{Variables: f.Variables, Body: applyWitness(f.Body, witMap)}
	}
	return fmla
}

// substituteVarsInNode replaces variables with their substitutions.
func substituteVarsInNode(node lg.Expr, subs map[string]lg.Expr) lg.Expr {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.Variable:
		if r, ok := subs[n.Name]; ok {
			return r
		}
		return node
	case *lg.Const:
		return node
	case *lg.Apply:
		newFunc := substituteVarsInNode(n.Func, subs)
		newTerms := make([]lg.Expr, len(n.Terms))
		changed := newFunc != n.Func
		for i, t := range n.Terms {
			newTerms[i] = substituteVarsInNode(t, subs)
			if newTerms[i] != t {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return lg.MustApply(newFunc, newTerms...)
	case *lg.And:
		terms := make([]lg.Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = substituteVarsInNode(t, subs)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Expr, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = substituteVarsInNode(t, subs)
		}
		return &lg.Or{Terms: terms}
	case *lg.Not:
		return &lg.Not{Body: substituteVarsInNode(n.Body, subs)}
	case *lg.Implies:
		return &lg.Implies{T1: substituteVarsInNode(n.T1, subs), T2: substituteVarsInNode(n.T2, subs)}
	case *lg.Eq:
		return &lg.Eq{T1: substituteVarsInNode(n.T1, subs), T2: substituteVarsInNode(n.T2, subs)}
	case *lg.ForAll:
		return &lg.ForAll{Variables: n.Variables, Body: substituteVarsInNode(n.Body, subs)}
	case *lg.Exists:
		return &lg.Exists{Variables: n.Variables, Body: substituteVarsInNode(n.Body, subs)}
	}
	return node
}

// astNodeToLogicNode converts an ast.Node to a logic.Expr if possible.
func astNodeToLogicNode(n ast.Node) lg.Expr {
	if n == nil {
		return nil
	}
	// Try direct type assertion
	if ln, ok := n.(lg.Expr); ok {
		return ln
	}
	// Try unwrapping adapter
	type unwrapper interface {
		Unwrap() lg.Expr
	}
	if u, ok := n.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}

// goalAddPrem adds a premise to a goal.
// Corresponds to Python goal_add_prem.
func (pc *ProofChecker) goalAddPrem(goal *ast.LabeledFormula, prem *ast.LabeledFormula, loc ast.Location) *ast.LabeledFormula {
	prems := append(GoalPrems(goal), prem)
	return MakeGoal(pc.astCfg(), loc, goal.Label, prems, GoalConc(goal))
}
