// tactics.go implements the 7 missing proof tactics, faithfully ported from
// Python ivy_proof.py.
//
// Each tactic manipulates the goal stack (a list of LabeledFormula) according
// to the proof rule it implements.
package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
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
	rawConc := GoalConc(prem)
	if len(witness) > 0 {
		if concExpr, ok := rawConc.(lg.Expr); ok {
			newConc, werr := module.WitnessAst(true, nil, witness, concExpr)
			if werr == nil {
				rawConc = newConc
			}
		}
	}
	prem = CloneGoal(pc.astCfg(), prem, GoalPrems(prem), rawConc)

	// Python: prem = apply_match_goal(pmatch, prem, apply_match_alt)
	prem = ApplyMatchGoalNode(pc.astCfg(), pmatch, prem)

	// Python: prem = drop_supplied_prems(prem, decl, proof.match())
	prem = DropSuppliedPrems(pc.astCfg(), prem, decl, proof.Matches)

	// Python lines 374-377: label handling.
	// When label is NoneAST, keep the schema's own label (Python creates
	// Atom(proof.label.rep, prem.label.args) which effectively preserves
	// the schema label name). When label is explicit, use it.
	if !isNoneAST(proof.TLabel) {
		prem = prem.Clone([]ast.Node{proof.TLabel, prem.Formula}).(*ast.LabeledFormula)
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
// Corresponds to Python ProofChecker.unfold_tactic (ivy_proof.py:384-398).
//
// Python:
//
//	for unfspec in proof.unfspecs:
//	    defn = self.lookup_schema(defname, decl, proof)   # schemata first
//	    rdefs = [rename_goal(defn, rn) for rn in unfspec.renamings]
//	    rdefs.append(defn)
//	    defns.append(rdefs)
//	if proof.has_premise:
//	    decl = goal_apply_to_prem(decl, premname, lambda g: unfold_goal(g, defns))
//	else:
//	    decl = goal_apply_to_conc(decl, lambda fmla: unfold_fmla(fmla, defns))
//	return [decl] + decls[1:]
func (pc *ProofChecker) unfoldTactic(decls []*ast.LabeledFormula, proof *ast.UnfoldTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "unfold tactic: no goals"}
	}
	goal := decls[0]

	// Python: for unfspec in proof.unfspecs:
	var defns [][]*ast.LabeledFormula
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

		// Python: defn = self.lookup_schema(defname, decl, proof)
		// lookup_schema checks schemata first, then definitions
		defn, err := pc.LookupSchema(defName, goal, proof, false)
		if err != nil {
			return nil, &ProofError{Msg: fmt.Sprintf("unfold tactic: %s not found", defName)}
		}

		// Python: rdefs = [rename_goal(defn, rn) for rn in unfspec.renamings]
		// rdefs.append(defn)
		var rdefs []*ast.LabeledFormula
		for _, rn := range unfspec.Renamings {
			renamed, rerr := RenameGoal(pc.astCfg(), defn, rn)
			if rerr != nil {
				return nil, rerr
			}
			rdefs = append(rdefs, renamed)
		}
		rdefs = append(rdefs, defn)

		// Python: defns.append(rdefs)
		defns = append(defns, rdefs)
	}

	if len(defns) == 0 {
		return decls, nil
	}

	// Python: if proof.has_premise: ... else: ...
	if proof.HasPremise() {
		// Python: premname = proof.premname
		// Python: decl = goal_apply_to_prem(decl, premname, lambda g: unfold_goal(g, defns))
		premName := fmt.Sprint(proof.Premise)
		newGoal := GoalApplyToPrem(pc.astCfg(), goal, premName, func(prem *ast.LabeledFormula) *ast.LabeledFormula {
			return UnfoldGoal(pc.astCfg(), prem, defns)
		})
		if newGoal == nil {
			return nil, &ProofError{Msg: fmt.Sprintf("unfold tactic: no premise %s found", premName), Node: proof}
		}
		goal = newGoal
	} else {
		// Python: decl = goal_apply_to_conc(decl, lambda fmla: unfold_fmla(fmla, defns))
		goal = GoalApplyToConc(pc.astCfg(), goal, func(node ast.Node) ast.Node {
			if fmla, ok := node.(lg.Expr); ok {
				return UnfoldFmla(fmla, defns)
			}
			return node
		})
	}

	// Python: return [decl] + decls[1:]
	return append([]*ast.LabeledFormula{goal}, decls[1:]...), nil
}

// attribGoals sets lineno on all goals from proof's lineno.
// Python: ivy_proof.py:33-36 attrib_goals
func attribGoals(proof ast.Node, goals []*ast.LabeledFormula) []*ast.LabeledFormula {
	if proof == nil {
		return goals
	}
	ln := proof.GetLineno()
	if ln.Line > 0 {
		for _, g := range goals {
			g.SetLineno(ln)
		}
	}
	return goals
}

// ifTactic splits the goal into two subgoals based on a condition.
// Corresponds to Python ProofChecker.if_tactic (ivy_proof.py:410-418).
//
// Python:
//   cond = proof.args[0]
//   true_goal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
//   true_goal.lineno = decls[0].lineno
//   false_goal = ia.LabeledFormula(decls[0].label, il.Implies(il.Not(cond), decls[0].formula))
//   false_goal.lineno = decls[0].lineno
//   return (attrib_goals(proof.args[1], apply_proof([true_goal], proof.args[1])) +
//           attrib_goals(proof.args[2], apply_proof([false_goal], proof.args[2])) +
//           decls[1:])
func (pc *ProofChecker) ifTactic(decls []*ast.LabeledFormula, proof *ast.IfTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "if tactic: no goals"}
	}
	goal := decls[0]

	// Get condition as logic node
	// Python: cond = proof.args[0]
	cond := astNodeToLogicNode(proof.Cond)
	if cond == nil {
		return nil, &ProofError{Msg: "if tactic: could not convert condition to logic node"}
	}

	// Python: true_goal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
	// Uses decls[0].formula directly (not goal_conc). If formula is lg.Expr, wrap it.
	// For SchemaBody (rare), fall back to wrapping just the conclusion.
	var trueGoal, falseGoal *ast.LabeledFormula
	if goalExpr, ok := goal.Formula.(lg.Expr); ok {
		trueGoal = pc.astCfg().NewLabeledFormula(goal.Label,
			lg.Expr(&lg.Implies{T1: cond, T2: goalExpr}))
		falseGoal = pc.astCfg().NewLabeledFormula(goal.Label,
			lg.Expr(&lg.Implies{T1: &lg.Not{Body: cond}, T2: goalExpr}))
	} else {
		// SchemaBody or other non-expr: wrap the conclusion (best approximation)
		trueConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
			return &lg.Implies{T1: cond, T2: c}
		})
		trueGoal = CloneGoal(pc.astCfg(), goal, GoalPrems(goal), trueConc)
		falseConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
			return &lg.Implies{T1: &lg.Not{Body: cond}, T2: c}
		})
		falseGoal = CloneGoal(pc.astCfg(), goal, GoalPrems(goal), falseConc)
	}
	trueGoal.SetLineno(goal.GetLineno())
	falseGoal.SetLineno(goal.GetLineno())

	// Apply proof branches with attrib_goals lineno attribution
	// Python: attrib_goals(proof.args[1], self.apply_proof([true_goal], proof.args[1]))
	var result []*ast.LabeledFormula

	if proof.Then != nil {
		trueResult, err := pc.ApplyProof([]*ast.LabeledFormula{trueGoal}, proof.Then)
		if err != nil {
			return nil, err
		}
		result = append(result, attribGoals(proof.Then, trueResult)...)
	} else {
		result = append(result, trueGoal)
	}

	// Python: attrib_goals(proof.args[2], self.apply_proof([false_goal], proof.args[2]))
	if proof.Else != nil {
		falseResult, err := pc.ApplyProof([]*ast.LabeledFormula{falseGoal}, proof.Else)
		if err != nil {
			return nil, err
		}
		result = append(result, attribGoals(proof.Else, falseResult)...)
	} else {
		result = append(result, falseGoal)
	}

	result = append(result, decls[1:]...)
	return result, nil
}

// propertyTactic introduces a property (cut) in a proof.
// Corresponds to Python ProofChecker.property_tactic (ivy_proof.py:232-281).
//
// Python:
//   vocab = goal_vocab(goal)
//   cut = compile_expr_vocab(proof.args[0], vocab)
//   cut = normalize_goal(cut)
//   subgoal = goal_subst(goal, cut, cut.lineno)
//   [handle Skolem if proof.args[1] not NoneAST]
//   subgoals = [subgoal]
//   if proof.args[2] not NoneAST: subgoals = apply_proof(subgoals, proof.args[2])
//   return [goal_add_prem(goal, cut, cut.lineno)] + decls[1:] + subgoals
func (pc *ProofChecker) propertyTactic(decls []*ast.LabeledFormula, proof *ast.PropertyTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "property tactic: no goals"}
	}
	goal := decls[0]

	// Python: vocab = goal_vocab(goal)
	vocab := GoalVocab(goal)

	// Python: cut = compile_expr_vocab(proof.args[0], vocab)
	// proof.Prop is a LabeledFormula from the grammar
	var cut *ast.LabeledFormula
	if propLF, ok := proof.Prop.(*ast.LabeledFormula); ok {
		cut = CompileExprVocabExtLF(propLF, vocab, pc.Mod)
	}
	if cut == nil {
		// Fallback: compile as expression, wrap in LabeledFormula
		compiled := CompileExprVocab(proof.Prop, vocab, pc.Mod)
		if compiled != nil {
			cut = pc.astCfg().NewLabeledFormula(nil, compiled)
			cut.SetLineno(proof.Prop.GetLineno())
		}
	}
	if cut == nil {
		return nil, &ProofError{Msg: "property tactic: could not compile cut formula"}
	}

	// Python: cut = normalize_goal(cut)
	cut = NormalizeGoal(pc.astCfg(), cut)

	// Python: subgoal = goal_subst(goal, cut, cut.lineno)
	subgoal, err := GoalSubst(pc.astCfg(), goal, cut, cut.GetLineno())
	if err != nil {
		return nil, err
	}

	// Python: lhs = proof.args[1]; if not isinstance(lhs, ia.NoneAST): [Skolem handling]
	// proof.PName is always NoneAST from current Go grammar (optskolem not yet parsed)
	if !isNoneAST(proof.PName) {
		return nil, &ProofError{Msg: "property tactic: Skolem function witness not implemented"}
	}

	// Python: subgoals = [subgoal]
	subgoals := []*ast.LabeledFormula{subgoal}

	// Python: pf = proof.args[2]; if not isinstance(pf, ia.NoneAST): subgoals = apply_proof(subgoals, pf)
	if !isNoneAST(proof.Proof) {
		applied, err := pc.ApplyProof(subgoals, proof.Proof)
		if err != nil {
			return nil, err
		}
		if applied == nil {
			return nil, nil
		}
		subgoals = applied
	}

	// Python: return [goal_add_prem(goal, cut, cut.lineno)] + decls[1:] + subgoals
	modifiedGoal := GoalAddPrem(pc.astCfg(), goal, cut, cut.GetLineno())
	result := []*ast.LabeledFormula{modifiedGoal}
	result = append(result, decls[1:]...)
	result = append(result, subgoals...)
	return result, nil
}

// functionTactic introduces a function definition in a proof.
// Corresponds to Python ProofChecker.function_tactic (ivy_proof.py:282-312).
//
// Python:
//
//	for df in proof.args:
//	    if isinstance(df, ia.ConstantDecl): assert False
//	    else: compile definition, add ConstantDecl + definition LF as premises
//	return [goal] + decls[1:]
func (pc *ProofChecker) functionTactic(decls []*ast.LabeledFormula, proof *ast.FunctionTactic) ([]*ast.LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, &ProofError{Msg: "function tactic: no goals"}
	}
	goal := decls[0]

	// Python: for df in proof.args:
	for _, df := range proof.Elems {
		// Python: if isinstance(df, ia.ConstantDecl): assert False
		if _, isCD := df.(*ast.ConstantDecl); isCD {
			return nil, &ProofError{Msg: "function tactic: unexpected ConstantDecl element"}
		}
		// Python else: compile definition and add ConstantDecl + LF as premises.
		// CompileDefinitionGoalVocab implements Python compile_definition_goal_vocab
		// (ivy_proof.py:1477-1506), which is the factored-out version of the
		// inline code in function_tactic.
		var err error
		goal, err = CompileDefinitionGoalVocab(pc.astCfg(), df, goal, pc.Mod)
		if err != nil {
			return nil, err
		}
	}

	// Python: return [goal] + decls[1:]
	return append([]*ast.LabeledFormula{goal}, decls[1:]...), nil
}

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

	// Python: wits = compile_witness_list(proof, decls[0])
	wits := CompileWitnessList(proof, goal, pc.Mod)

	// Python: for wit in wits: if not il.is_variable(wit.args[0]): raise error
	// Python: wit_map = dict((x.args[0], x.args[1]) for x in wits)
	witness := make(map[lg.NodeKey]lg.Expr)
	for _, w := range wits {
		defn, ok := w.(*lg.Definition)
		if !ok {
			continue
		}
		v, ok := defn.Lhs.(*lg.Variable)
		if !ok {
			return nil, &ProofError{Msg: "left-hand side of witness must be a variable"}
		}
		witness[lg.Key(v)] = defn.Rhs
	}

	if len(witness) == 0 {
		return decls, nil
	}

	// Python: conc = lu.witness_ast(False, [], wit_map, conc)
	rawConc := GoalConc(goal)
	var newConc ast.Node
	if concExpr, ok := rawConc.(lg.Expr); ok {
		witnessed, werr := module.WitnessAst(false, nil, witness, concExpr)
		if werr != nil {
			return nil, &ProofError{Msg: fmt.Sprintf("witness tactic: %v", werr)}
		}
		newConc = witnessed
	} else {
		newConc = rawConc
	}

	// Python: prems = goal_prems(decl); return [clone_goal(decl,prems,conc)] + decls[1:]
	prems := GoalPrems(goal)
	newGoal := CloneGoal(pc.astCfg(), goal, prems, newConc)
	return append([]*ast.LabeledFormula{newGoal}, decls[1:]...), nil
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
