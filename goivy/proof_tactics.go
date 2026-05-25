// tactics.go implements the 7 missing proof tactics, faithfully ported from
// Python ivy_proof.py.
//
// Each tactic manipulates the goal stack (a list of LabeledFormula) according
// to the proof rule it implements.
package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// letTactic introduces local definitions in a proof.
// Corresponds to Python ProofChecker.let_tactic (ivy_proof.py:213-223).
//
// Python:
//
//	vocab = goal_vocab(goal)
//	defs = [compile_expr_vocab(ia.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.args]
//	cond = il.And(*[il.Equals(a.args[0], a.args[1]) for a in defs])
//	goal = ia.LabeledFormula(goal.label, il.Implies(cond, goal.formula))
//	return [goal] + decls[1:]
func (pc *ProofChecker) letTactic(decls []*LabeledFormula, proof *LetTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.letTactic ENTER ndecls=%d nDefs=%d", len(decls), len(proof.Defs))
	if len(decls) == 0 {
		xtracer.Trace("proof.letTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "let tactic: no goals"}
	}
	goal := decls[0]

	vocab := GoalVocab(goal)

	// Compile each equality definition with goal vocabulary.
	// Python: defs = [compile_expr_vocab(ia.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.args]
	// Then: cond = il.And(*[il.Equals(a.args[0], a.args[1]) for a in defs])
	var eqs []Expr
	for _, def := range proof.Defs {
		defArgs := def.Args()
		equality := def
		if len(defArgs) >= 2 {
			equality = pc.astCfg().NewAtom("=", defArgs[0], defArgs[1])
		}
		compiled := CompileExprVocab(equality, vocab, pc.Mod)
		if compiled == nil {
			// Fallback: try direct extraction without compilation
			if len(defArgs) >= 2 {
				lhs := astNodeToLogicNode(defArgs[0])
				rhs := astNodeToLogicNode(defArgs[1])
				if lhs != nil && rhs != nil {
					eqs = append(eqs, &Eq{T1: lhs, T2: rhs})
				}
			}
			continue
		}
		// Extract LHS and RHS from compiled equality (il.Equals(a.args[0], a.args[1]))
		if eq, ok := compiled.(*Eq); ok {
			eqs = append(eqs, eq)
		} else if app, ok := compiled.(*Apply); ok {
			// Some compilers produce Apply(=, args...) — extract
			if len(app.Terms) >= 2 {
				eqs = append(eqs, &Eq{T1: app.Terms[0], T2: app.Terms[1]})
			}
		}
	}

	if len(eqs) == 0 {
		xtracer.Trace("proof.letTactic EXIT passthrough ndecls=%d", len(decls))
		return decls, nil
	}

	cond := &LogicAnd{Terms: eqs}

	// Python ivy_proof.py:226:
	//   subgoal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
	// Wrap the ENTIRE goal.Formula (SchemaBody or TemporalModels included) — no descent.
	subgoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), cond, goal.Formula))
	subgoal.SetLineno(goal.GetLineno())

	result := []*LabeledFormula{subgoal}
	result = append(result, decls[1:]...)
	xtracer.Trace("proof.letTactic EXIT HASH canon=%v", subgoal.Canon())
	return result, nil
}

// assumeTactic introduces an assumption from a schema or premise.
// Faithful port of Python ProofChecker.assume_tactic (ivy_proof.py:350-382).
//
// isGlobal distinguishes AssumeGlobalTactic (from "assume" keyword) from
// AssumeTactic (from "instantiate" keyword). Python uses isinstance() check.
func (pc *ProofChecker) assumeTactic(decls []*LabeledFormula, proof *AssumeTactic, isGlobal bool) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.assumeTactic ENTER ndecls=%d isGlobal=%v", len(decls), isGlobal)
	if len(decls) == 0 {
		xtracer.Trace("proof.assumeTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "assume tactic: no goals"}
	}
	decl := decls[0]

	// Python: schemaname = proof.schemaname()
	schemaName := ""
	if proof.SchemaName != nil {
		schemaName = fmt.Sprint(proof.SchemaName)
	}
	if schemaName == "" {
		xtracer.Trace("proof.assumeTactic EXIT err=noSchemaName")
		return nil, &ProofError{Msg: "assume tactic: no schema name"}
	}

	// Python: premmap = dict((x.name,x) for x in goal_prem_goals(decl))
	premMap := make(map[string]*LabeledFormula)
	for _, pg := range GoalPremGoals(decl) {
		premMap[pg.LabelName()] = pg
	}

	// Python lines 354-359: AssumeGlobalTactic vs AssumeTactic distinction.
	// AssumeTactic looks in premises first; AssumeGlobalTactic skips to global.
	var schema *LabeledFormula
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
			xtracer.Trace("proof.assumeTactic EXIT err=noSchema schema=%s", schemaName)
			return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
				"No property %s exists in the current context", schemaName)}
		}
	}

	// Python: schema = remove_explicit(schema)
	schema = RemoveExplicit(schema)

	// Python: prob, pmatch = self.setup_schema_matching(decl, proof, schema, allow_witness=True)
	prob, pmatch, err := pc.SetupSchemaMatchingRaw(decl, proof.Ren, proof.Matches, schema, true)
	if err != nil {
		xtracer.Trace("proof.assumeTactic EXIT err=%v schema=%s", err, schemaName)
		return nil, err
	}

	// Python lines 365-367: extract witnesses (variables not in freesyms).
	// iswit = lambda x: isinstance(x, il.Variable) and x not in prob.freesyms
	// pmatch is an InsMap — iteration is insertion-order, matching
	// CPython 3.7+ dict iteration so our isWitVar traces line up with
	// Python's witness-arg order.
	witness := NewInsMap[NodeKey, Expr]()
	pmatchClean := NewInsMap[NodeKey, Expr]()
	for k, v := range pmatch.All() {
		if isWitVar(k, v, prob) {
			witness.Set(k, v)
		} else {
			pmatchClean.Set(k, v)
		}
	}
	pmatch = pmatchClean
	xtracer.Trace("proof.assumeTactic witnessSplit schema=%s nWitness=%d nPmatch=%d", schemaName, witness.Len(), pmatch.Len())

	// Python: prem = prob.schema
	prem := prob.SchemaLF

	// Downstream functions take plain map; convert InsMap lookup-views.
	pmatchMap := insMapToMap(pmatch)
	witnessMap := insMapToMap(witness)

	// Python: if schemaname not in premmap: prem = close_unmatched(prem, pmatch)
	if _, inPrems := premMap[schemaName]; !inPrems {
		prem = CloseUnmatched(pc.astCfg(), prem, pmatchMap)
	}

	// Python: conc = goal_conc(prem)
	//         conc = lu.witness_ast(True, [], witness, conc)
	//         prem = clone_goal(prem, goal_prems(prem), conc)
	// Python calls witness_ast unconditionally — empty witness map is
	// fine and still emits the ENTER/EXIT traces. Match that here.
	rawConc := GoalConc(prem)
	newConc, werr := WitnessAst(true, nil, witnessMap, rawConc)
	if werr != nil {
		xtracer.Trace("proof.assumeTactic EXIT err=witnessSubst schema=%s err=%v", schemaName, werr)
		return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
			"assume tactic witness substitution: %v", werr)}
	}
	rawConc = newConc
	if concExpr, ok := rawConc.(Expr); ok {
		xtracer.Trace("proof.assumeTactic postWitnessAst schema=%s HASH canon=%v", schemaName, concExpr.Canon())
	} else {
		xtracer.Trace("proof.assumeTactic postWitnessAst schema=%s concType=%s", schemaName, TypeName(rawConc))
	}
	prem = CloneGoal(pc.astCfg(), prem, GoalPrems(prem), rawConc)

	// Python: prem = apply_match_goal(pmatch, prem, apply_match_alt)
	prem = ApplyMatchGoalNode(pc.astCfg(), pmatchMap, prem)

	// Python: prem = drop_supplied_prems(prem, decl, proof.match())
	prem = DropSuppliedPrems(pc.astCfg(), prem, decl, proof.Matches)

	// Python lines 374-377: label handling.
	// When label is NoneAST, keep the schema's own label (Python creates
	// Atom(proof.label.rep, prem.label.args) which effectively preserves
	// the schema label name). When label is explicit, use it.
	if !isNoneAST(proof.TLabel) {
		prem = prem.Clone([]Node{proof.TLabel, prem.Formula}).(*LabeledFormula)
	}

	// Python lines 378-381: clash detection.
	// AssumeGlobalTactic renames to avoid clash; AssumeTactic errors.
	for _, pg := range GoalPremGoals(decl) {
		if pg.LabelName() == prem.LabelName() {
			if isGlobal {
				prem = RenamePremNoClash(prem, decl)
			} else {
				xtracer.Trace("proof.assumeTactic EXIT err=clash prem=%s", prem.LabelName())
				return nil, &ProofError{Node: proof, Msg: fmt.Sprintf(
					"instance name %s clashes with context", prem.LabelName())}
			}
			break
		}
	}

	// Python: return [goal_add_prem(decl, prem, proof.lineno)] + decls[1:]
	newGoal := GoalAddPrem(pc.astCfg(), decl, prem, proof.GetLineno())
	result := []*LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	xtracer.Trace("proof.assumeTactic EXIT schema=%s HASH canon=%v", schemaName, newGoal.Canon())
	return result, nil
}

// isNoneAST checks if a node is a NoneAST (or nil).
func isNoneAST(n Node) bool {
	if n == nil {
		return true
	}
	_, ok := n.(*NoneAST)
	return ok
}

// defargNameSort extracts the name and sort-top-ness from a defarg parameter
// node. The grammar's defnlhs produces an *ast.Atom whose Terms are either
// *ast.App (from lparam: SYMBOLx:atype) or *ast.Variable (from var).
// Returns (name, sortName, isTopSort).
func defargNameSort(n Node) (string, string, bool) {
	switch a := n.(type) {
	case *App:
		name := a.Relname()
		if a.ASort == nil {
			return name, "", true
		}
		sn := fmt.Sprint(a.ASort)
		return name, sn, false
	case *Variable:
		if a.VSort == "" || a.VSort == "S" {
			return a.Rep, a.VSort, true
		}
		return a.Rep, a.VSort, false
	case *Atom:
		if a.ASort == nil {
			return a.Rep, "", true
		}
		sn := fmt.Sprint(a.ASort)
		return a.Rep, sn, false
	}
	return fmt.Sprint(n), "", true
}

// isWitVar checks if a match entry is a witness variable.
// Python: iswit = lambda x: isinstance(x, il.Variable) and x not in prob.freesyms
//
// x is the match KEY (the LHS), not the value. A key is a witness-eligible
// Variable iff it is a Variable used in the schema conclusion (bound or free)
// AND it is not in prob.FreeSyms. This mirrors Python's CompileMatchFull
// behavior: when allow_witness=True, the local freesyms is augmented with
// used_variables(conc); the keys added beyond prob.FreeSyms are exactly the
// bound variables we want to treat as witnesses.
//
// Earlier Go port incorrectly checked val.(*lg.Variable). For
// `instantiate ifabric_rw_fair_ax with P=_P`, _P is a skolemized Const, so
// the old check rejected it, leaving {P: _P} in pmatch. Downstream
// applyMatchAltRec then preserved the vacuous `ForAll([P], body_with__P)`
// wrapping — the divergence observed at log.golden.2hr index 2411549.
func isWitVar(key NodeKey, val Expr, prob *MatchProblem) bool {
	if _, inFree := prob.FreeSyms[key]; inFree {
		xtracer.Trace("proof.isWitVar key=%s result=%v reason=inFreeSyms", string(key), false)
		return false
	}
	if prob.SchemaLF == nil {
		xtracer.Trace("proof.isWitVar key=%s result=%v reason=schemaLFNil", string(key), false)
		return false
	}
	conc := ConcAsExpr(GoalConc(prob.SchemaLF))
	if conc == nil {
		xtracer.Trace("proof.isWitVar key=%s result=%v reason=concNil", string(key), false)
		return false
	}
	_, isUsedVar := UsedVariables(conc)[key]
	xtracer.Trace("proof.isWitVar key=%s result=%v", string(key), isUsedVar)
	return isUsedVar
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
func (pc *ProofChecker) unfoldTactic(decls []*LabeledFormula, proof *UnfoldTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.unfoldTactic ENTER ndecls=%d nUnfSpecs=%d", len(decls), len(proof.UnfSpecs))
	if len(decls) == 0 {
		xtracer.Trace("proof.unfoldTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "unfold tactic: no goals"}
	}
	goal := decls[0]

	// Python: for unfspec in proof.unfspecs:
	var defns [][]*LabeledFormula
	for _, unfspecNode := range proof.UnfSpecs {
		unfspec, ok := unfspecNode.(*UnfoldSpec)
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
			xtracer.Trace("proof.unfoldTactic EXIT err=notFound defName=%s", defName)
			return nil, &ProofError{Msg: fmt.Sprintf("unfold tactic: %s not found", defName)}
		}

		// Python: rdefs = [rename_goal(defn, rn) for rn in unfspec.renamings]
		// rdefs.append(defn)
		var rdefs []*LabeledFormula
		for _, rn := range unfspec.Renamings {
			renamed, rerr := RenameGoal(pc.astCfg(), defn, rn)
			if rerr != nil {
				xtracer.Trace("proof.unfoldTactic EXIT err=rename defName=%s err=%v", defName, rerr)
				return nil, rerr
			}
			rdefs = append(rdefs, renamed)
		}
		rdefs = append(rdefs, defn)

		// Python: defns.append(rdefs)
		defns = append(defns, rdefs)
	}

	if len(defns) == 0 {
		xtracer.Trace("proof.unfoldTactic EXIT passthrough ndecls=%d", len(decls))
		return decls, nil
	}

	// Python: if proof.has_premise: ... else: ...
	if proof.HasPremise() {
		// Python: premname = proof.premname
		// Python: decl = goal_apply_to_prem(decl, premname, lambda g: unfold_goal(g, defns))
		premName := fmt.Sprint(proof.Premise)
		newGoal := GoalApplyToPrem(pc.astCfg(), goal, premName, func(prem *LabeledFormula) *LabeledFormula {
			return UnfoldGoal(pc.astCfg(), prem, defns)
		})
		if newGoal == nil {
			xtracer.Trace("proof.unfoldTactic EXIT err=noPremise premName=%s", premName)
			return nil, &ProofError{Msg: fmt.Sprintf("unfold tactic: no premise %s found", premName), Node: proof}
		}
		goal = newGoal
	} else {
		// Python: decl = goal_apply_to_conc(decl, lambda fmla: unfold_fmla(fmla, defns))
		// UnfoldFmla takes ast.Node and handles *ast.TemporalModels
		// internally (unwrap/recurse/rewrap), matching Python's duck-typed
		// recursion through the wrapper.
		goal = GoalApplyToConc(pc.astCfg(), goal, func(node Node) Node {
			return UnfoldFmla(node, defns)
		})
	}

	// Python: return [decl] + decls[1:]
	xtracer.Trace("proof.unfoldTactic EXIT HASH canon=%v", goal.Canon())
	return append([]*LabeledFormula{goal}, decls[1:]...), nil
}

// attribGoals sets lineno on all goals from proof's lineno.
// Python: ivy_proof.py:33-36 attrib_goals
func attribGoals(proof Node, goals []*LabeledFormula) []*LabeledFormula {
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
//
//	cond = proof.args[0]
//	true_goal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
//	true_goal.lineno = decls[0].lineno
//	false_goal = ia.LabeledFormula(decls[0].label, il.Implies(il.Not(cond), decls[0].formula))
//	false_goal.lineno = decls[0].lineno
//	return (attrib_goals(proof.args[1], apply_proof([true_goal], proof.args[1])) +
//	        attrib_goals(proof.args[2], apply_proof([false_goal], proof.args[2])) +
//	        decls[1:])
func (pc *ProofChecker) ifTactic(decls []*LabeledFormula, proof *IfTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.ifTactic ENTER ndecls=%d", len(decls))
	if len(decls) == 0 {
		xtracer.Trace("proof.ifTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "if tactic: no goals"}
	}
	goal := decls[0]

	// Get condition as logic node
	// Python: cond = proof.args[0]
	cond := astNodeToLogicNode(proof.Cond)
	if cond == nil {
		xtracer.Trace("proof.ifTactic EXIT err=condConversion")
		return nil, &ProofError{Msg: "if tactic: could not convert condition to logic node"}
	}

	// Python ivy_proof.py:414-416:
	//   true_goal  = ia.LabeledFormula(decls[0].label, il.Implies(cond,      decls[0].formula))
	//   false_goal = ia.LabeledFormula(decls[0].label, il.Implies(Not(cond), decls[0].formula))
	// Wrap the ENTIRE goal.Formula (SchemaBody or TemporalModels included) — no descent.
	notCond := &LogicNot{Body: cond}
	trueGoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), cond, goal.Formula))
	falseGoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), notCond, goal.Formula))
	trueGoal.SetLineno(goal.GetLineno())
	falseGoal.SetLineno(goal.GetLineno())

	// Apply proof branches with attrib_goals lineno attribution
	// Python: attrib_goals(proof.args[1], self.apply_proof([true_goal], proof.args[1]))
	var result []*LabeledFormula

	if proof.Then != nil {
		trueResult, err := pc.ApplyProof([]*LabeledFormula{trueGoal}, proof.Then)
		if err != nil {
			xtracer.Trace("proof.ifTactic EXIT err=%v branch=true", err)
			return nil, err
		}
		result = append(result, attribGoals(proof.Then, trueResult)...)
	} else {
		result = append(result, trueGoal)
	}

	// Python: attrib_goals(proof.args[2], self.apply_proof([false_goal], proof.args[2]))
	if proof.Else != nil {
		falseResult, err := pc.ApplyProof([]*LabeledFormula{falseGoal}, proof.Else)
		if err != nil {
			xtracer.Trace("proof.ifTactic EXIT err=%v branch=false", err)
			return nil, err
		}
		result = append(result, attribGoals(proof.Else, falseResult)...)
	} else {
		result = append(result, falseGoal)
	}

	result = append(result, decls[1:]...)
	xtracer.Trace("proof.ifTactic EXIT nresult=%d", len(result))
	return result, nil
}

// propertyTactic introduces a property (cut) in a proof.
// Corresponds to Python ProofChecker.property_tactic (ivy_proof.py:232-281).
//
// Python:
//
//	vocab = goal_vocab(goal)
//	cut = compile_expr_vocab(proof.args[0], vocab)
//	cut = normalize_goal(cut)
//	subgoal = goal_subst(goal, cut, cut.lineno)
//	[handle Skolem if proof.args[1] not NoneAST]
//	subgoals = [subgoal]
//	if proof.args[2] not NoneAST: subgoals = apply_proof(subgoals, proof.args[2])
//	return [goal_add_prem(goal, cut, cut.lineno)] + decls[1:] + subgoals
func (pc *ProofChecker) propertyTactic(decls []*LabeledFormula, proof *PropertyTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.propertyTactic ENTER ndecls=%d", len(decls))
	if len(decls) == 0 {
		xtracer.Trace("proof.propertyTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "property tactic: no goals"}
	}
	goal := decls[0]

	// Python: vocab = goal_vocab(goal)
	vocab := GoalVocab(goal)

	// Python: cut = compile_expr_vocab(proof.args[0], vocab)
	// proof.Prop is a LabeledFormula from the grammar
	var cut *LabeledFormula
	if propLF, ok := proof.Prop.(*LabeledFormula); ok {
		var err error
		cut, err = compileExprVocabLF(propLF, vocab, pc.Mod)
		if err != nil {
			xtracer.Trace("proof.propertyTactic EXIT err=compileFailed err=%v", err)
			return nil, err
		}
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
		xtracer.Trace("proof.propertyTactic EXIT err=compileFailed")
		return nil, &ProofError{Msg: "property tactic: could not compile cut formula"}
	}

	// Python: cut = normalize_goal(cut)
	cut = NormalizeGoal(pc.astCfg(), cut)

	// Python: subgoal = goal_subst(goal, cut, cut.lineno)
	subgoal, err := GoalSubst(pc.astCfg(), goal, cut, cut.GetLineno())
	if err != nil {
		xtracer.Trace("proof.propertyTactic EXIT err=goalSubst err=%v", err)
		return nil, err
	}

	// Python: lhs = proof.args[1]; if not isinstance(lhs, ia.NoneAST): [Skolem handling]
	if !isNoneAST(proof.PName) {
		lhs, ok := proof.PName.(*Atom)
		if !ok {
			return nil, &ProofError{Msg: "property tactic: optskolem must be an Atom"}
		}

		// Python: fmla = il.drop_universals(cut.formula)
		cutExpr := GoalConcUnwrap(cut)
		if cutExpr == nil {
			return nil, &ProofError{Msg: "property tactic: cut formula is not a logic expression"}
		}
		fmla := IvyDropUniversals(cutExpr)

		// Python: if not il.is_exists(fmla) or len(fmla.variables) != 1:
		if !IsExists(fmla) || len(BinderVars(fmla)) != 1 {
			xtracer.Trace("proof.propertyTactic EXIT err=notExistential")
			return nil, &ProofError{Msg: "property is not existential"}
		}
		evar := BinderVars(fmla)[0]
		rng := evar.VSort

		// Python: vmap = dict((x.name, x) for x in lu.variables_ast(fmla))
		varsList := VariablesAstList(fmla)
		vmap := make(map[string]*LogicVariable, len(varsList))
		for _, v := range varsList {
			vmap[v.Name] = v
		}

		// Python: used = set(); args = lhs.args; targs = []
		used := make(map[string]bool)
		args := lhs.Terms
		targs := make([]Expr, 0, len(args))

		for _, a := range args {
			aName, aSortName, aIsTop := defargNameSort(a)

			// Python: if a.name in used: raise IvyError(lhs,'repeat parameter: ...')
			if used[aName] {
				return nil, &ProofError{Msg: fmt.Sprintf("repeat parameter: %s", aName)}
			}
			used[aName] = true

			// Python: if a.name in vmap:
			if v, inVmap := vmap[aName]; inVmap {
				targs = append(targs, v)
				// Python: if not (il.is_topsort(a.sort) or a.sort != v.sort):
				vSortName := SortName(v.VSort)
				if !aIsTop && aSortName == vSortName {
					return nil, &ProofError{Msg: fmt.Sprintf("bad sort for %s", aName)}
				}
			} else {
				// Python: if il.is_topsort(a.sort): raise IvyError(...)
				if aIsTop {
					return nil, &ProofError{Msg: fmt.Sprintf("cannot infer sort for %s", aName)}
				}
				paramVar := &LogicVariable{Name: aName, VSort: &UninterpretedSort{Name: aSortName}}
				targs = append(targs, paramVar)
			}
		}

		// Python: for x in vmap: if x not in used: raise IvyError(...)
		for x := range vmap {
			if !used[x] {
				return nil, &ProofError{Msg: fmt.Sprintf("%s must be a parameter of %s", x, lhs.Rep)}
			}
		}

		// Python: dom = [x.sort for x in targs]
		domSorts := make([]Sort, 0, len(targs)+1)
		for _, t := range targs {
			domSorts = append(domSorts, t.NodeSort())
		}
		domSorts = append(domSorts, rng)
		symSort := FuncConstSort(domSorts...)

		// Python: sym = il.Symbol(lhs.rep, il.FuncConstSort(*(dom+[rng])))
		sym := NewConst(lhs.Rep, symSort)

		// Python: if sym in self.stale or sym in goal_defns(goal):
		goalDefns := GoalDefns(goal)
		if pc.Stale[sym.Name] {
			return nil, &ProofError{Msg: fmt.Sprintf("%s is not fresh", sym.Name)}
		}
		if _, inDefns := goalDefns[Key(sym)]; inDefns {
			return nil, &ProofError{Msg: fmt.Sprintf("%s is not fresh", sym.Name)}
		}

		// Python: term = sym(*targs) if targs else sym
		var term Expr
		if len(targs) > 0 {
			term = &Apply{Func: sym, Terms: targs}
		} else {
			term = sym
		}

		// Python: fmla = lu.substitute_ast(fmla.body, {evar.name: term})
		body := BinderBody(fmla)
		substituted := SubstituteAstByName(body, map[string]Expr{evar.Name: term})

		// Python: cut = clone_goal(cut, [], fmla)
		cut = CloneGoal(pc.astCfg(), cut, nil, substituted)

		// Python: goal = goal_add_prem(goal, ia.ConstantDecl(sym), goal.lineno)
		goal = GoalAddPrem(pc.astCfg(), goal, pc.astCfg().NewConstantDecl(sym), goal.GetLineno())

		xtracer.Trace("proof.propertyTactic skolem sym=%s nTargs=%d", sym.Name, len(targs))
	}

	// Python: subgoals = [subgoal]
	subgoals := []*LabeledFormula{subgoal}

	// Python: pf = proof.args[2]; if not isinstance(pf, ia.NoneAST): subgoals = apply_proof(subgoals, pf)
	if !isNoneAST(proof.Proof) {
		applied, err := pc.ApplyProof(subgoals, proof.Proof)
		if err != nil {
			xtracer.Trace("proof.propertyTactic EXIT err=applyProof err=%v", err)
			return nil, err
		}
		if applied == nil {
			xtracer.Trace("proof.propertyTactic EXIT applied=nil")
			return nil, nil
		}
		subgoals = applied
	}

	// Python: return [goal_add_prem(goal, cut, cut.lineno)] + decls[1:] + subgoals
	modifiedGoal := GoalAddPrem(pc.astCfg(), goal, cut, cut.GetLineno())
	result := []*LabeledFormula{modifiedGoal}
	result = append(result, decls[1:]...)
	result = append(result, subgoals...)
	xtracer.Trace("proof.propertyTactic EXIT nresult=%d nsubgoals=%d", len(result), len(subgoals))
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
func (pc *ProofChecker) functionTactic(decls []*LabeledFormula, proof *FunctionTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.functionTactic ENTER ndecls=%d nElems=%d", len(decls), len(proof.Elems))
	if len(decls) == 0 {
		xtracer.Trace("proof.functionTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "function tactic: no goals"}
	}
	goal := decls[0]

	// Python: for df in proof.args:
	for _, df := range proof.Elems {
		// Python: if isinstance(df, ia.ConstantDecl): assert False
		if _, isCD := df.(*ConstantDecl); isCD {
			xtracer.Trace("proof.functionTactic EXIT err=unexpectedConstantDecl")
			return nil, &ProofError{Msg: "function tactic: unexpected ConstantDecl element"}
		}
		// Python else: compile definition and add ConstantDecl + LF as premises.
		// CompileDefinitionGoalVocab implements Python compile_definition_goal_vocab
		// (ivy_proof.py:1477-1506), which is the factored-out version of the
		// inline code in function_tactic.
		var err error
		goal, err = CompileDefinitionGoalVocab(pc.astCfg(), df, goal, pc.Mod)
		if err != nil {
			xtracer.Trace("proof.functionTactic EXIT err=%v", err)
			return nil, err
		}
	}

	// Python: return [goal] + decls[1:]
	xtracer.Trace("proof.functionTactic EXIT HASH canon=%v", goal.Canon())
	return append([]*LabeledFormula{goal}, decls[1:]...), nil
}

// witnessTactic provides witnesses for existentially quantified variables.
// Corresponds to Python ProofChecker.witness_tactic (ivy_proof.py:459-471).
func (pc *ProofChecker) witnessTactic(decls []*LabeledFormula, proof *WitnessTactic) ([]*LabeledFormula, error) {
	xtracer.Trace("proof.witnessTactic ENTER ndecls=%d nArgs=%d", len(decls), len(proof.Witnesses))
	if len(decls) == 0 {
		xtracer.Trace("proof.witnessTactic EXIT err=noGoals")
		return nil, &ProofError{Msg: "witness tactic: no goals"}
	}
	goal := decls[0]

	if GoalConc(goal) == nil {
		xtracer.Trace("proof.witnessTactic EXIT err=noConclusion")
		return nil, &ProofError{Msg: "witness tactic: goal has no conclusion"}
	}
	if gc := GoalConc(goal); gc != nil {
		xtracer.Trace("proof.witnessTactic preConc HASH canon=%v", gc.Canon())
	}

	// Python: if ia.has_temporal(proof) and not goal_is_temporal(goal): raise error
	// (Python has `goal` here but means `decl`/`goal` = decls[0])
	if HasTemporal(proof) && !GoalIsTemporal(goal) {
		xtracer.Trace("proof.witnessTactic EXIT err=temporalInNonTemporal")
		return nil, &ProofError{Msg: "temporal operator not allowed in instantiation", Node: proof}
	}

	// Python: wits = compile_witness_list(proof, decls[0])
	wits := CompileWitnessList(proof, goal, pc.Mod)
	xtracer.Trace("proof.witnessTactic compiled nwits=%d", len(wits))

	// Python: for wit in wits: if not il.is_variable(wit.args[0]): raise error
	// Python: wit_map = dict((x.args[0], x.args[1]) for x in wits)
	witness := make(map[NodeKey]Expr)
	for _, w := range wits {
		defn, ok := w.(*LogicDefinition)
		if !ok {
			xtracer.Trace("proof.witnessTactic witnessSkip nonDefn type=%s", TypeName(w))
			continue
		}
		v, ok := defn.Lhs.(*LogicVariable)
		if !ok {
			xtracer.Trace("proof.witnessTactic EXIT err=lhsNotVariable type=%s", TypeName(defn.Lhs))
			return nil, &ProofError{Msg: "left-hand side of witness must be a variable"}
		}
		witness[Key(v)] = defn.Rhs
		xtracer.Trace("proof.witnessTactic witnessPair key=%s rhs HASH canon=%v", string(Key(v)), defn.Rhs.Canon())
	}

	if len(witness) == 0 {
		xtracer.Trace("proof.witnessTactic EXIT passthrough nwitness=0")
		return decls, nil
	}

	// Python: conc = lu.witness_ast(False, [], wit_map, conc)
	// module.WitnessAst takes ast.Node and handles *ast.TemporalModels
	// internally (unwrap/recurse/rewrap), matching Python's duck-typed
	// recursion at ivy_logic_utils.py:1704.
	newConc, err := WitnessAst(false, nil, witness, GoalConc(goal))
	if err != nil {
		xtracer.Trace("proof.witnessTactic EXIT err=witnessAst err=%v", err)
		return nil, &ProofError{Msg: fmt.Sprintf("witness tactic: %v", err)}
	}
	if newConc != nil {
		xtracer.Trace("proof.witnessTactic postWitnessAst HASH canon=%v", newConc.Canon())
	}

	// Python: prems = goal_prems(decl); return [clone_goal(decl,prems,conc)] + decls[1:]
	prems := GoalPrems(goal)
	newGoal := CloneGoal(pc.astCfg(), goal, prems, newConc)
	xtracer.Trace("proof.witnessTactic EXIT HASH canon=%v", newGoal.Canon())
	return append([]*LabeledFormula{newGoal}, decls[1:]...), nil
}

// astNodeToLogicNode converts an ast.Node to a logic.Expr if possible.
func astNodeToLogicNode(n Node) Expr {
	if n == nil {
		return nil
	}
	// Try direct type assertion
	if ln, ok := n.(Expr); ok {
		return ln
	}
	// Try unwrapping adapter
	type unwrapper interface {
		Unwrap() Expr
	}
	if u, ok := n.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
