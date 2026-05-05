// ivy_tactics.go — Mechanical port of Python ivy_tactics.py.
// Defines proof tactics (vcgen, skolemize, skolemizenp, tempind,
// tempcase, sorry) and their helper functions.
package goivy

import (
	"fmt"
)

// pcAstCfg safely extracts the AstConfig from a ProofCheckerInterface, returning a
// default config if pc is nil or pc.GetAstCfg() is nil.
func pcAstCfg(pc ProofCheckerInterface) *AstConfig {
	if pc != nil && pc.GetAstCfg() != nil {
		return pc.GetAstCfg()
	}
	return NewAstConfig()
}

// ---------- helper functions ----------

// VcToGoal converts a verification condition to a goal.
// Corresponds to Python: vc_to_goal (ivy_tactics.py lines 33-35).
//
//	pr.make_goal(lineno, name, [], lg.Not(lu.clauses_to_formula(vc)),
//	    annot=(action, vc.annot))
func VcToGoal(cfg *AstConfig, loc Location, name string, vc *Clauses, action ActionsAction) *LabeledFormula {
	fmla := ClausesToFormula(vc)
	negFmla, err := NewNot(fmla)
	if err != nil {
		// Fallback: use the formula directly if negation fails (shouldn't happen)
		negFmla = &Not{Body: fmla}
	}
	label := cfg.NewAtom(name)
	goal := MakeGoal(cfg, loc, label, nil, negFmla)
	goal.Annot = &ActionAnnotation{Action: action, Annot: vc.Annot}
	return goal
}

// TripleToGoal converts a Hoare triple (precondition, action, postcondition) to a goal.
// Corresponds to Python: triple_to_goal (ivy_tactics.py lines 37-39).
//
//	vc = tr.make_vc(action, precond, postcond)
//	return vc_to_goal(lineno, name, vc, action)
func TripleToGoal(cfg *AstConfig, loc Location, name string, action ActionsAction,
	precond []*LabeledFormula, postcond []*LabeledFormula) *LabeledFormula {
	// Convert labeled formulas to clauses for MakeVC
	var preClauses []*Clauses
	for _, lf := range precond {
		if f, ok := lf.Formula.(Expr); ok {
			preClauses = append(preClauses, NewClauses([]Expr{f}, nil, nil))
		}
	}
	var postClauses []*Clauses
	for _, lf := range postcond {
		if f, ok := lf.Formula.(Expr); ok {
			postClauses = append(postClauses, NewClauses([]Expr{f}, nil, nil))
		}
	}
	vc := MakeVC(action, preClauses, postClauses, false)
	return VcToGoal(cfg, loc, name, vc, action)
}

// TempindFmla recursively transforms a formula for temporal induction.
// Corresponds to Python: tempind_fmla (ivy_tactics.py lines 57-71).
//
// Handles three cases:
//  1. ForAll: accumulates variables, recurses into body
//  2. Implies: splits variables by premise/conclusion usage
//  3. Globally (when vs+params non-empty): creates Or(G(body), WhenOperator("next", body, whencond))
func TempindFmla(fmla Expr, cond Expr, params []Expr, vs []*Variable) Expr {
	// Case 1: ForAll — accumulate quantified variables
	if fa, ok := fmla.(*ForAll); ok {
		newVs := make([]*Variable, len(vs))
		copy(newVs, vs)
		newVs = append(newVs, fa.Variables...)
		return TempindFmla(fa.Body, cond, params, newVs)
	}

	// Case 2: Implies — split variables
	if imp, ok := fmla.(*Implies); ok {
		// Python: prem_vars = set(ilu.variables_ast(fmla.args[0]))
		premVarMap := UsedVariables(imp.T1)
		premVarNames := make(map[string]bool)
		for _, node := range premVarMap {
			if v, ok := node.(*Variable); ok {
				premVarNames[v.Name] = true
			}
		}
		// Python: conc = tempind_fmla(fmla.args[1], cond, params, [v for v in vs if v.name not in prem_vars])
		var concVs []*Variable
		for _, v := range vs {
			if !premVarNames[v.Name] {
				concVs = append(concVs, v)
			}
		}
		conc := TempindFmla(imp.T2, cond, params, concVs)
		// Python: res = lg.Implies(fmla.args[0], conc)
		res, _ := NewImplies(imp.T1, conc)
		// Python: uvs = [v for v in vs if v.name in prem_vars]
		var uvs []*Variable
		for _, v := range vs {
			if premVarNames[v.Name] {
				uvs = append(uvs, v)
			}
		}
		// Python: return lg.ForAll(uvs, res) if uvs else res
		if len(uvs) > 0 {
			return IvyForAll(uvs, res)
		}
		return res
	}

	// Case 3: Globally (when vs+params is non-empty)
	if gb, ok := fmla.(*Globally); ok && (len(vs) > 0 || len(params) > 0) {
		// Python: body = lg.Implies(cond, fmla.body) if params else fmla.body
		var body Expr
		if len(params) > 0 {
			body, _ = NewImplies(cond, gb.Body)
		} else {
			body = gb.Body
		}
		// Python: gbly = fmla.clone([body])
		gbly := gb.Clone([]Node{body}).(Expr)
		// Python: whencond = lg.Not(lg.ForAll(vs, fmla.body))
		forallBody := IvyForAll(vs, gb.Body)
		whencond, _ := NewNot(forallBody)
		// Python: return lg.ForAll(vs+params, lg.Or(gbly, lg.WhenOperator("next", body, whencond)))
		whenOp, _ := NewWhenOperator("next", body, whencond)
		orExpr, _ := NewOr(gbly, whenOp)
		allVs := make([]*Variable, 0, len(vs)+len(params))
		allVs = append(allVs, vs...)
		for _, p := range params {
			if v, ok := p.(*Variable); ok {
				allVs = append(allVs, v)
			}
		}
		return IvyForAll(allVs, orExpr)
	}

	// Default: wrap in ForAll if vs is non-empty
	// Python: return lg.ForAll(vs, fmla) if vs else fmla
	if len(vs) > 0 {
		return IvyForAll(vs, fmla)
	}
	return fmla
}

// compileTacticLets extracts and compiles let-bindings from a TacticTactic proof node.
// Returns (compiled definitions as Eq nodes, condition, params).
// Shared between ApplyTempind and ApplyTempcase.
func compileTacticLets(mod *Module, cfg *AstConfig, goal *LabeledFormula, proofNode Node) ([]Expr, Expr, []Expr, error) {
	tt, ok := proofNode.(*TacticTactic)
	if !ok {
		return nil, nil, nil, fmt.Errorf("expected TacticTactic, got %T", proofNode)
	}

	// Python: if proof.tactic_decls: raise IvyError(proof, 'tactic does not take declarations')
	if decls := tt.TacticDeclsList(); len(decls) > 0 {
		return nil, nil, nil, fmt.Errorf("tactic does not take declarations")
	}

	lets := tt.TacticLetsList()
	if len(lets) == 0 {
		// No lets — empty condition
		return nil, nil, nil, nil
	}

	// Python: vocab = pr.goal_vocab(goal, bound=True)
	vocab := GoalVocabBound(goal)

	// Python: defs = [pr.compile_expr_vocab(ivy_ast.Atom('=', x.args[0], x.args[1]), vocab) for x in proof.tactic_lets]
	var defs []Expr
	for _, letNode := range lets {
		args := letNode.Args()
		if len(args) < 2 {
			return nil, nil, nil, fmt.Errorf("let binding must have two arguments")
		}
		atom := cfg.NewAtom("=", args[0], args[1])
		compiled := CompileExprVocab(atom, vocab, mod)
		if compiled == nil {
			return nil, nil, nil, fmt.Errorf("could not compile let binding: %v", letNode)
		}
		defs = append(defs, compiled)
	}

	// Python: conds = [lg.Equals(a.args[0], a.args[1]) for a in defs]
	var conds []Expr
	for _, d := range defs {
		// The compiled expression should be an Eq node (from Atom("=", lhs, rhs))
		if eq, ok := d.(*Eq); ok {
			conds = append(conds, eq)
		} else {
			// Fallback: use the compiled expression as-is
			conds = append(conds, d)
		}
	}

	// Python: cond = conds[0] if len(conds) == 1 else lg.normalized_and(*conds)
	var cond Expr
	if len(conds) == 1 {
		cond = conds[0]
	} else {
		cond = NormalizedAnd(conds...)
	}

	// Python: params = list(a.args[0] for a in defs)
	var params []Expr
	for _, d := range defs {
		if eq, ok := d.(*Eq); ok {
			params = append(params, eq.T1)
		}
	}

	return defs, cond, params, nil
}

// ApplyTempind applies temporal induction transformation to a proof goal.
// Corresponds to Python: apply_tempind (ivy_tactics.py lines 73-89).
func ApplyTempind(mod *Module, cfg *AstConfig, goal *LabeledFormula, proofNode Node) (*LabeledFormula, error) {
	_, cond, params, err := compileTacticLets(mod, cfg, goal, proofNode)
	if err != nil {
		return nil, err
	}

	// Python: conc = pr.goal_conc(goal)
	// GoalConc now returns ast.Node so it can carry *ast.TemporalModels.
	// Check the formula directly for TemporalModels first; otherwise use
	// GoalConcExpr to get the lg.Expr conclusion.
	fmlaNode := goal.Formula

	// Python: if not (goal.temporal or isinstance(conc, ivy_ast.TemporalModels)):
	tm, isTM := fmlaNode.(*AstTemporalModels)
	if !goal.IsTemporal() && !isTM {
		return nil, fmt.Errorf("tactics/ivy_tactics: [3]proof goal is not temporal")
	}

	var newFmla Node
	if isTM {
		// Python: fmla = tempind_fmla(conc.fmla, cond, params)
		innerFmla, ok := tm.Fmla.(Expr)
		if !ok {
			return nil, fmt.Errorf("TemporalModels.Fmla is not lg.Expr: %T", tm.Fmla)
		}
		transformed := TempindFmla(innerFmla, cond, params, nil)
		// Python: fmla = conc.clone([fmla])
		newFmla = tm.Clone([]Node{transformed})
	} else {
		// Python: fmla = tempind_fmla(conc, cond, params)
		concExpr := GoalConcExpr(goal)
		if concExpr == nil {
			return nil, fmt.Errorf("tempind: goal conclusion is not an lg.Expr")
		}
		newFmla = TempindFmla(concExpr, cond, params, nil)
	}

	// Python: return pr.clone_goal(goal, pr.goal_prems(goal), fmla)
	newGoal := CloneGoal(cfg, goal, GoalPrems(goal), newFmla)
	newGoal.Temporal = goal.Temporal
	return newGoal, nil
}

// TempcaseFmla recursively transforms a formula for temporal case analysis.
// Corresponds to Python: tempcase_fmla (ivy_tactics.py lines 98-108).
func TempcaseFmla(fmla Expr, cond Expr, vs []Expr, proofNode Node) (Expr, error) {
	// Case 1: ForAll — check for variable capture, recurse into body
	if fa, ok := fmla.(*ForAll); ok {
		// Python: for v in fmla.variables: if v in vs: raise IvyError(...)
		for _, v := range fa.Variables {
			for _, vsElem := range vs {
				if vv, ok := vsElem.(*Variable); ok {
					if v.Name == vv.Name {
						return nil, fmt.Errorf("variable %s would be captured by quantifier", v.Name)
					}
				}
			}
		}
		// Python: return fmla.clone([tempcase_fmla(fmla.body, cond, vs, proof)])
		newBody, err := TempcaseFmla(fa.Body, cond, vs, proofNode)
		if err != nil {
			return nil, err
		}
		return fa.Clone([]Node{newBody}).(Expr), nil
	}

	// Case 2: Implies — recurse into conclusion
	if imp, ok := fmla.(*Implies); ok {
		// Python: return fmla.clone([fmla.args[0], tempcase_fmla(fmla.args[1], cond, vs, proof)])
		newConc, err := TempcaseFmla(imp.T2, cond, vs, proofNode)
		if err != nil {
			return nil, err
		}
		return imp.Clone([]Node{imp.T1, newConc}).(Expr), nil
	}

	// Case 3: Globally — wrap in ForAll with implied condition
	if gb, ok := fmla.(*Globally); ok {
		// Python: return lg.forall(vs, fmla.clone([lg.Implies(cond, fmla.body)]))
		implBody, _ := NewImplies(cond, gb.Body)
		cloned := gb.Clone([]Node{implBody}).(Expr)
		var varList []*Variable
		for _, v := range vs {
			if vv, ok := v.(*Variable); ok {
				varList = append(varList, vv)
			}
		}
		return IvyForAll(varList, cloned), nil
	}

	// Default: return formula unchanged
	return fmla, nil
}

// ApplyTempcase applies temporal case analysis transformation to a proof goal.
// Corresponds to Python: apply_tempcase (ivy_tactics.py lines 110-125).
func ApplyTempcase(mod *Module, cfg *AstConfig, goal *LabeledFormula, proofNode Node) (*LabeledFormula, error) {
	_, cond, params, err := compileTacticLets(mod, cfg, goal, proofNode)
	if err != nil {
		return nil, err
	}

	// params are the variables (LHS of each let), used as vs in tempcase_fmla
	vs := params

	fmlaNode := goal.Formula

	var newFmla Node
	if tm, ok := fmlaNode.(*AstTemporalModels); ok {
		innerFmla, ok := tm.Fmla.(Expr)
		if !ok {
			return nil, fmt.Errorf("TemporalModels.Fmla is not lg.Expr: %T", tm.Fmla)
		}
		transformed, err := TempcaseFmla(innerFmla, cond, vs, proofNode)
		if err != nil {
			return nil, err
		}
		newFmla = tm.Clone([]Node{transformed})
	} else {
		concExpr := GoalConcExpr(goal)
		if concExpr == nil {
			return nil, fmt.Errorf("tempcase: goal conclusion is not an lg.Expr")
		}
		transformed, err := TempcaseFmla(concExpr, cond, vs, proofNode)
		if err != nil {
			return nil, err
		}
		newFmla = transformed
	}

	// Python: subgoal = pr.clone_goal(goal, pr.goal_prems(goal), fmla)
	// Use proof.CloneGoal to emit matching CloneGoal ENTER/EXIT traces.
	newGoal := CloneGoal(cfg, goal, GoalPrems(goal), newFmla)
	newGoal.Temporal = goal.Temporal
	return newGoal, nil
}

// ---------- tactic entry points ----------
// All follow the ProofTactic signature.

// Vcgen reduces a safety property to initiation and consecution subgoals.
// Corresponds to Python: vcgen (ivy_tactics.py lines 21-31).
func Vcgen(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, fmt.Errorf("vcgen: no goals")
	}
	goal := decls[0]
	conc := GoalConc(goal)

	// Python: if not isinstance(conc, ivy_ast.TemporalModels) or not lg.is_true(conc.fmla):
	//            raise iu.IvyError(self, 'vcgen tactic applies only to safety properties')
	tm, ok := conc.(*AstTemporalModels)
	if !ok {
		return nil, fmt.Errorf("vcgen tactic applies only to safety properties")
	}
	tmFmla, ok := tm.Fmla.(Expr)
	if !ok || !IsTrue(tmFmla) {
		return nil, fmt.Errorf("vcgen tactic applies only to safety properties")
	}

	// Python: model = conc.model
	model, ok := tm.Model.(*NormalProgram)
	if !ok {
		return nil, fmt.Errorf("vcgen: model is not a NormalProgram: %T", tm.Model)
	}

	loc := proofNode.GetLineno()

	// Python: goal1 = triple_to_goal(proof.lineno, 'initiation', model.init, postcond=model.invars)
	goal1 := TripleToGoal(pcAstCfg(pc), loc, "initiation", model.Init, nil, model.Invars)

	// Python: goal2 = triple_to_goal(proof.lineno, 'consecution', tm.env_action(model.bindings),
	//                                precond=model.invars+model.asms, postcond=model.invars)
	envAct := TemporalEnvAction(pc.GetModule().Cfg.ActCfg, model.Bindings)
	preconds := make([]*LabeledFormula, 0, len(model.Invars)+len(model.Asms))
	preconds = append(preconds, model.Invars...)
	preconds = append(preconds, model.Asms...)
	goal2 := TripleToGoal(pcAstCfg(pc), loc, "consecution", envAct, preconds, model.Invars)

	// Python: return [goal1, goal2] + decls[1:]
	// Note: Python has a bug here: decls[1:] instead of decls (which was already decls[1:])
	// The original Python does: decls = decls[1:] on line 24, then returns [goal1, goal2] + decls[1:]
	// which skips decls[1]. We follow Python literally.
	rest := decls[1:]
	if len(rest) > 0 {
		rest = rest[1:]
	}
	result := make([]*LabeledFormula, 0, 2+len(rest))
	result = append(result, goal1, goal2)
	result = append(result, rest...)
	return result, nil
}

// Skolemize skolemizes a goal in prenex form.
// Corresponds to Python: skolemize (ivy_tactics.py lines 43-46).
func Skolemize(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, fmt.Errorf("skolemize: no goals")
	}
	goal := decls[0]
	// Python: goal = pr.skolemize_goal(goal)
	goal = SkolemizeGoal(pcAstCfg(pc), goal, true)
	result := make([]*LabeledFormula, 0, len(decls))
	result = append(result, goal)
	result = append(result, decls[1:]...)
	return result, nil
}

// Skolemizenp skolemizes a goal without requiring prenex normal form.
// Corresponds to Python: skolemizenp (ivy_tactics.py lines 50-53).
func Skolemizenp(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, fmt.Errorf("skolemizenp: no goals")
	}
	goal := decls[0]
	// Python: goal = pr.skolemize_goal(goal, prenex=False)
	goal = SkolemizeGoal(pcAstCfg(pc), goal, false)
	result := make([]*LabeledFormula, 0, len(decls))
	result = append(result, goal)
	result = append(result, decls[1:]...)
	return result, nil
}

// Tempind applies temporal induction to a proof goal.
// Corresponds to Python: tempind (ivy_tactics.py lines 91-94).
func Tempind(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, fmt.Errorf("tempind: no goals")
	}
	goal := decls[0]
	// Python: goal = apply_tempind(goal, proof)
	var mod *Module
	if pc != nil {
		mod = pc.GetModule()
	}
	goal, err := ApplyTempind(mod, pcAstCfg(pc), goal, proofNode)
	if err != nil {
		return nil, err
	}
	result := make([]*LabeledFormula, 0, len(decls))
	result = append(result, goal)
	result = append(result, decls[1:]...)
	return result, nil
}

// Tempcase applies temporal case analysis to a proof goal.
// Corresponds to Python: tempcase (ivy_tactics.py lines 127-130).
func Tempcase(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(decls) == 0 {
		return nil, fmt.Errorf("tempcase: no goals")
	}
	goal := decls[0]
	// Python: goal = apply_tempcase(goal, proof)
	goal, err := ApplyTempcase(pc.GetModule(), pcAstCfg(pc), goal, proofNode)
	if err != nil {
		return nil, err
	}
	result := make([]*LabeledFormula, 0, len(decls))
	result = append(result, goal)
	result = append(result, decls[1:]...)
	return result, nil
}

// Sorry drops the current goal without proof, marking that unsound proofs were used.
// Corresponds to Python: sorry (ivy_tactics.py lines 136-139).
func Sorry(pc ProofCheckerInterface, decls []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	// Python: used_sorry = True; return decls[1:]
	if pc != nil && pc.GetModule() != nil && pc.GetModule().Cfg != nil {
		pc.GetModule().Cfg.UsedSorry = true
	}
	if len(decls) == 0 {
		return nil, nil
	}
	return decls[1:], nil
}

// ---------- registration ----------

// RegisterProofTactics registers all ivy_tactics.py proof tactics on the given proof config.
// Corresponds to Python's module-level register_tactic calls (lines 41, 48, 55, 96, 132, 141).
func RegisterProofTactics(cfg *ProofConfig) {
	cfg.RegisterTactic("vcgen", Vcgen)
	cfg.RegisterTactic("skolemize", Skolemize)
	cfg.RegisterTactic("skolemizenp", Skolemizenp)
	cfg.RegisterTactic("tempind", Tempind)
	cfg.RegisterTactic("tempcase", Tempcase)
	cfg.RegisterTactic("sorry", Sorry)
}
