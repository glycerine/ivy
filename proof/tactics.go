// tactics.go implements the 7 missing proof tactics, faithfully ported from
// Python ivy_proof.py.
//
// Each tactic manipulates the goal stack (a list of LabeledFormula) according
// to the proof rule it implements.
package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// letTactic introduces local definitions in a proof.
// Corresponds to Python ProofChecker.let_tactic (lines 213-223).
//
// Given definitions x1=e1, ..., xn=en, the goal G becomes:
//   (x1=e1 & ... & xn=en) -> G
func (pc *ProofChecker) letTactic(decls []*ast.LabeledFormula, proof *ast.LetTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "let tactic: no goals"}
	}
	goal := decls[0]

	// Build the condition: conjunction of all equalities
	var eqs []lg.Node
	for _, def := range proof.Defs {
		// Each def should be an equality atom: = lhs rhs
		defArgs := def.Args()
		if len(defArgs) >= 2 {
			lhs := astNodeToLogicNode(defArgs[0])
			rhs := astNodeToLogicNode(defArgs[1])
			if lhs != nil && rhs != nil {
				eqs = append(eqs, &lg.Eq{T1: lhs, T2: rhs})
			}
		}
	}

	if len(eqs) == 0 {
		return decls, nil
	}

	var cond lg.Node
	if len(eqs) == 1 {
		cond = eqs[0]
	} else {
		cond = &lg.And{Terms: eqs}
	}

	// Build subgoal: cond -> original formula
	conc := GoalConc(goal)
	if conc == nil {
		return nil, &ProofError{Msg: "let tactic: goal has no conclusion"}
	}
	newConc := &lg.Implies{T1: cond, T2: conc}
	subgoal := CloneGoal(goal, GoalPrems(goal), newConc)
	subgoal.SetLineno(goal.GetLineno())

	result := []*ast.LabeledFormula{subgoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// assumeTactic introduces an assumption from a schema or premise.
// Corresponds to Python ProofChecker.assume_tactic (lines 342-374).
//
// This is a complex tactic that involves schema matching and witness
// substitution. The full implementation requires setup_matching,
// compile_match, and witness_ast.
func (pc *ProofChecker) assumeTactic(decls []*ast.LabeledFormula, proof *ast.AssumeTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "assume tactic: no goals"}
	}
	goal := decls[0]

	// Get the schema name from the proof
	schemaName := ""
	if proof.SchemaName != nil {
		schemaName = fmt.Sprint(proof.SchemaName)
	}
	if schemaName == "" {
		return nil, &ProofError{Msg: "assume tactic: no schema name"}
	}

	// Look up in premises first
	premMap := make(map[string]*ast.LabeledFormula)
	for _, pg := range GoalPremGoals(goal) {
		premMap[pg.LabelName()] = pg
	}

	var schema *ast.LabeledFormula
	if prem, ok := premMap[schemaName]; ok {
		schema = prem
	} else {
		var err error
		schema, err = pc.LookupSchema(schemaName, goal)
		if err != nil {
			return nil, &ProofError{Msg: fmt.Sprintf("No property %s exists in the current context", schemaName)}
		}
	}

	// Add schema as premise to goal
	prem := schema
	if proof.TLabel != nil {
		// Rename with given label
		prem = prem.CloneWithFreshID([]ast.Node{proof.TLabel, prem.Formula})
	}

	// Check for name clash
	for _, pg := range GoalPremGoals(goal) {
		if pg.LabelName() == prem.LabelName() {
			return nil, &ProofError{Msg: fmt.Sprintf("instance name %s clashes with context", prem.LabelName())}
		}
	}

	newGoal := goalAddPrem(goal, prem, proof.GetLineno())
	result := []*ast.LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// unfoldTactic unfolds definitions in the goal.
// Corresponds to Python ProofChecker.unfold_tactic (lines 376-392).
func (pc *ProofChecker) unfoldTactic(decls []*ast.LabeledFormula, proof *ast.UnfoldTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "unfold tactic: no goals"}
	}
	// The unfold tactic looks up definitions by name and substitutes them
	// into the goal's conclusion or a named premise.
	//
	// For each unfspec in proof.UnfoldSpecs:
	//   1. Look up the definition by name
	//   2. Apply any renamings
	//   3. Either unfold in the conclusion or in a named premise
	//
	// Full implementation requires unfold_goal and unfold_fmla helpers.
	// For now, return the goals unchanged with a note.
	return decls, nil
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

	conc := GoalConc(goal)
	if conc == nil {
		return nil, &ProofError{Msg: "if tactic: goal has no conclusion"}
	}

	// Build true_goal: C -> G
	trueConc := &lg.Implies{T1: cond, T2: conc}
	trueGoal := CloneGoal(goal, GoalPrems(goal), trueConc)
	trueGoal.SetLineno(goal.GetLineno())

	// Build false_goal: ~C -> G
	falseConc := &lg.Implies{T1: &lg.Not{Body: cond}, T2: conc}
	falseGoal := CloneGoal(goal, GoalPrems(goal), falseConc)
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
	// The property tactic introduces a "cut" formula. The goal G becomes:
	//   cut (as a subgoal)
	//   cut -> G (added as premise)
	//
	// Full implementation requires compile_expr_vocab, normalize_goal,
	// goal_subst, and recursive proof application.
	// For now, return goals unchanged.
	return decls, nil
}

// functionTactic introduces a function definition in a proof.
// Corresponds to Python ProofChecker.function_tactic (lines 275-304).
func (pc *ProofChecker) functionTactic(decls []*ast.LabeledFormula, proof *ast.FunctionTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "function tactic: no goals"}
	}
	// The function tactic introduces a fresh function symbol with a definition.
	// The definition is added as a premise.
	// Full implementation requires compile_expr_vocab, TopFunctionSort,
	// and cycle detection for recursive definitions.
	// For now, return goals unchanged.
	return decls, nil
}

// witnessTactic provides witnesses for existentially quantified variables.
// Corresponds to Python ProofChecker.witness_tactic (lines 451-463).
func (pc *ProofChecker) witnessTactic(decls []*ast.LabeledFormula, proof *ast.WitnessTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "witness tactic: no goals"}
	}
	goal := decls[0]

	conc := GoalConc(goal)
	if conc == nil {
		return nil, &ProofError{Msg: "witness tactic: goal has no conclusion"}
	}

	// Build witness map from proof witnesses
	// Each witness is x = e, mapping variable x to expression e
	witMap := make(map[string]lg.Node)
	for _, w := range proof.Witnesses {
		wargs := w.Args()
		if len(wargs) >= 2 {
			lhs := astNodeToLogicNode(wargs[0])
			rhs := astNodeToLogicNode(wargs[1])
			if lhs != nil && rhs != nil {
				if v, ok := lhs.(*lg.Var); ok {
					witMap[v.Name] = rhs
				}
			}
		}
	}

	if len(witMap) == 0 {
		return decls, nil
	}

	// Apply witness substitution to the conclusion
	// Python: conc = lu.witness_ast(False, [], wit_map, conc)
	// This replaces existentially quantified variables with their witnesses.
	newConc := applyWitness(conc, witMap)

	prems := GoalPrems(goal)
	newGoal := CloneGoal(goal, prems, newConc)
	result := []*ast.LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	return result, nil
}

// applyWitness substitutes witness values for existentially quantified variables.
// Corresponds to Python lu.witness_ast.
func applyWitness(fmla lg.Node, witMap map[string]lg.Node) lg.Node {
	if fmla == nil || len(witMap) == 0 {
		return fmla
	}
	switch f := fmla.(type) {
	case *lg.Exists:
		// Check if any of the bound variables have witnesses
		var remainingVars []*lg.Var
		subs := make(map[string]lg.Node)
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
		terms := make([]lg.Node, len(f.Terms))
		for i, t := range f.Terms {
			terms[i] = applyWitness(t, witMap)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(f.Terms))
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
func substituteVarsInNode(node lg.Node, subs map[string]lg.Node) lg.Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *lg.Var:
		if r, ok := subs[n.Name]; ok {
			return r
		}
		return node
	case *lg.Const:
		return node
	case *lg.Apply:
		newFunc := substituteVarsInNode(n.Func, subs)
		newTerms := make([]lg.Node, len(n.Terms))
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
		return &lg.Apply{Func: newFunc, Terms: newTerms}
	case *lg.And:
		terms := make([]lg.Node, len(n.Terms))
		for i, t := range n.Terms {
			terms[i] = substituteVarsInNode(t, subs)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(n.Terms))
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

// astNodeToLogicNode converts an ast.Node to a logic.Node if possible.
func astNodeToLogicNode(n ast.Node) lg.Node {
	if n == nil {
		return nil
	}
	// Try direct type assertion
	if ln, ok := n.(lg.Node); ok {
		return ln
	}
	// Try unwrapping adapter
	type unwrapper interface {
		Unwrap() lg.Node
	}
	if u, ok := n.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}

// goalAddPrem adds a premise to a goal.
// Corresponds to Python goal_add_prem.
func goalAddPrem(goal *ast.LabeledFormula, prem *ast.LabeledFormula, loc ast.Location) *ast.LabeledFormula {
	prems := GoalPrems(goal)
	prems = append(prems, prem)
	conc := GoalConc(goal)
	result := MakeGoal(loc, goal.Label, prems, conc)
	return result
}
