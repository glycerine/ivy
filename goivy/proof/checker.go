package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// ProofChecker is Ivy's built-in proof checker.
type ProofChecker struct {
	// Cfg is the per-session proof configuration (tactic registry).
	Cfg *module.ProofConfig
	// AstCfg is the per-session AST configuration (constructor state).
	AstCfg *ast.AstConfig
	// Mod is the current module (for compilation during matching).
	Mod *module.Module
	// Axioms is the list of available axioms.
	Axioms []*ast.LabeledFormula
	// Definitions maps symbol names to their definitions.
	Definitions map[string]*ast.LabeledFormula
	// Schemata maps names to proof schemata.
	// Python: self.schemata = dict() at ivy_proof.py:55. Python 3.7+ dicts are
	// insertion-ordered; we use InsMap to match that guarantee so dumps/snapshots
	// are deterministic on both sides.
	Schemata *iu.InsMap[string, *ast.LabeledFormula]
	// Stale is the set of symbols that have been referenced and are not fresh.
	Stale map[string]bool
}

// NewProofChecker creates a new ProofChecker.
//
// axioms and definitions are lists of LabeledFormula.
// schemata is an optional map from string names to LabeledFormula.
func NewProofChecker(cfg *module.ProofConfig, mod *module.Module, axioms, definitions []*ast.LabeledFormula, schemata *iu.InsMap[string, *ast.LabeledFormula], astCfgs ...*ast.AstConfig) *ProofChecker {
	if cfg == nil {
		cfg = module.TacticNewConfig()
	}
	// Use the caller's AstConfig if provided (shares LF counter with module),
	// otherwise create a private one.
	// Python uses a single global lf_counter, so all ProofChecker normalization
	// advances the same counter as the compiler. We match this by sharing AstCfg.
	var acfg *ast.AstConfig
	if len(astCfgs) > 0 && astCfgs[0] != nil {
		acfg = astCfgs[0]
	} else {
		acfg = ast.NewAstConfig()
	}
	pc := &ProofChecker{
		Cfg:         cfg,
		AstCfg:      acfg,
		Mod:         mod,
		Definitions: make(map[string]*ast.LabeledFormula),
		Schemata:    iu.NewInsMap[string, *ast.LabeledFormula](),
		Stale:       make(map[string]bool),
	}

	// Python: self.axioms = [normalize_goal(ax) for ax in axioms]
	for _, ax := range axioms {
		pc.Axioms = append(pc.Axioms, NormalizeGoal(pc.AstCfg, ax))
	}

	// Python: self.definitions = dict((d.formula.defines().name, normalize_goal(d)) for d in definitions)
	for _, d := range definitions {
		norm := NormalizeGoal(pc.AstCfg, d)
		name := ""
		if def, ok := d.Formula.(*lg.Definition); ok {
			if sym, ok := def.Defines().(*lg.Const); ok {
				name = sym.Name
			}
		}
		if name == "" {
			name = d.LabelName() // fallback
		}
		pc.Definitions[name] = norm
	}

	// Python: if schemata is not None:
	//             for _skey, _sval in schemata.items():
	//                 self.schemata[_skey] = normalize_goal(_sval)
	if schemata != nil {
		for name, s := range schemata.All() {
			norm := NormalizeGoal(pc.AstCfg, s)
			xtracer.Trace("proof.ProofChecker.__init__.fromMod schemata.insert key='%s' value=%s", name, norm.Canon())
			pc.Schemata.Set(name, norm)
		}
	}

	// Python: for ax in axioms:
	//             if ax.label is not None:
	//                 self.schemata[ax.name] = ax
	for _, ax := range axioms {
		if ax.Label != nil {
			xtracer.Trace("proof.ProofChecker.__init__.axiom schemata.insert key='%s' value=%s", ax.LabelName(), ax.Canon())
			pc.Schemata.Set(ax.LabelName(), ax)
		}
	}

	// Mark stale symbols: collect used symbols from axioms and definitions.
	// Corresponds to Python's:
	//   self.stale = set()
	//   for lf in axioms + definitions:
	//       self.stale.update(lu.used_symbols_ast(lf.formula))
	// Python walks the ENTIRE lf.formula (including premises in SchemaBody),
	// not just the conclusion. We use collectStaleSymbols which walks via Args().
	for _, lf := range axioms {
		collectStaleSymbols(lf.Formula, pc.Stale)
	}
	for _, lf := range definitions {
		collectStaleSymbols(lf.Formula, pc.Stale)
	}
	// Also mark stale from schemata vocabularies.
	if schemata != nil {
		for _, s := range schemata.All() {
			vocab := GoalVocab(s)
			for _, sym := range vocab.Symbols {
				pc.Stale[sym.Name] = true
			}
		}
	}

	return pc
}

// GetModule returns the current module. Implements module.ProofCheckerInterface.
func (pc *ProofChecker) GetModule() *module.Module {
	if pc == nil {
		return nil
	}
	return pc.Mod
}

// GetAstCfg returns the AST configuration. Implements module.ProofCheckerInterface.
func (pc *ProofChecker) GetAstCfg() *ast.AstConfig { return pc.AstCfg }

// GetAxioms returns the list of available axioms. Implements module.ProofCheckerInterface.
func (pc *ProofChecker) GetAxioms() []*ast.LabeledFormula { return pc.Axioms }

// astCfg returns the AstConfig for this proof checker, preferring
// the module's config when the module is available.
func (pc *ProofChecker) astCfg() *ast.AstConfig {
	if pc.Mod != nil && pc.Mod.Cfg != nil && pc.Mod.Cfg.AstCfg != nil {
		return pc.Mod.Cfg.AstCfg
	}
	if pc.AstCfg != nil {
		return pc.AstCfg
	}
	return ast.NewAstConfig()
}

// AdmitAxiom adds an axiom to the checker.
func (pc *ProofChecker) AdmitAxiom(ax *ast.LabeledFormula) {
	norm := NormalizeGoal(pc.AstCfg, ax)
	pc.Axioms = append(pc.Axioms, norm)
	if ax.Label != nil {
		xtracer.Trace("proof.ProofChecker.admit_axiom schemata.insert key='%s' value=%s", ax.LabelName(), ax.Canon())
		pc.Schemata.Set(ax.LabelName(), ax)
	}
}

// LookupSchema looks up a schema by name in the checker's schemata,
// definitions, or goal premises.
//
// When looking up a definition, converts it to a constraint formula via
// DefinitionToConstraint. If close is true, the constraint is universally
// closed over free variables.
//
// Python: ivy_proof.py:306-322
func (pc *ProofChecker) LookupSchema(name string, goal *ast.LabeledFormula, errNode interface{}, close bool) (*ast.LabeledFormula, error) {
	xtracer.Trace("proof.LookupSchema ENTER schemaName=%s close=%v", name, close)
	if s, ok := pc.Schemata.Get2(name); ok {
		xtracer.Trace("proof.ProofChecker.LookupSchema schemata.lookup key='%s' found=true value=%s", name, s.Canon())
		if err := CheckSchemaCapture(s, goal); err != nil {
			return nil, err
		}
		xtracer.Trace("proof.LookupSchema EXIT HASH canon=%s", s.Canon())
		return s, nil
	}
	xtracer.Trace("proof.ProofChecker.LookupSchema schemata.lookup key='%s' found=false", name)
	if d, ok := pc.Definitions[name]; ok {
		// Convert definition to constraint — Python: goal_conc(schema).to_constraint()
		conc := GoalConc(d)
		if def, ok := conc.(*lg.Definition); ok {
			fmla := il.DefinitionToConstraint(def)
			if close {
				fmla = il.CloseFormula(fmla)
			}
			schema := CloneGoal(pc.astCfg(), d, GoalPrems(d), fmla)
			if err := CheckSchemaCapture(schema, goal); err != nil {
				return nil, err
			}
			xtracer.Trace("proof.LookupSchema EXIT HASH canon=%s", schema.Canon())
			return schema, nil
		}
		// Not a *lg.Definition — return as-is
		if err := CheckSchemaCapture(d, goal); err != nil {
			return nil, err
		}
		xtracer.Trace("proof.LookupSchema EXIT HASH canon=%s", d.Canon())
		return d, nil
	}
	// Check goal premises
	for _, pg := range GoalPremGoals(goal) {
		if pg.LabelName() == name {
			xtracer.Trace("proof.LookupSchema EXIT HASH canon=%s", pg.Canon())
			return pg, nil
		}
	}
	xtracer.Trace("proof.LookupSchema EXIT err=notFound schemaName=%s", name)
	return nil, &ProofError{Node: errNode, Msg: "No property " + name + " exists in the current context"}
}

// ApplyProof applies a proof to a list of goals, producing subgoals.
// Returns nil and an error if the proof fails. Corresponds to Python's
// ProofChecker.apply_proof which dispatches on the type of proof:
//   - SchemaInstantiation -> match_schema
//   - LetTactic -> let_tactic
//   - ComposeTactics -> compose_proofs
//   - AssumeTactic -> assume_tactic
//   - UnfoldTactic -> unfold_tactic
//   - ForgetTactic -> forget_tactic
//   - ShowGoalsTactic -> show_goals_tactic
//   - DeferGoalTactic -> defer_goal_tactic
//   - IfTactic -> if_tactic
//   - NullTactic -> returns decls unchanged
//   - PropertyTactic -> property_tactic
//   - FunctionTactic -> function_tactic
//   - TacticTactic -> tactic_tactic (dispatches to registered tactics)
//   - ProofTactic -> proof_tactic
//   - WitnessTactic -> witness_tactic
func (pc *ProofChecker) ApplyProof(goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, nil
	}
	if proof == nil {
		return nil, &ProofError{Msg: "nil proof supplied"}
	}

	if len(goals) > 0 && goals[0] != nil {
		xtracer.Trace("proof.ApplyProof ENTER proofType=%s goal[0].Formula type=%s", iu.TypeName(proof), iu.TypeName(goals[0].Formula))
	}

	// Dispatch on proof type.
	switch p := proof.(type) {
	case *ast.SchemaInstantiation:
		xtracer.Trace("proof.ApplyProof dispatch name=SchemaInstantiation")
		m, err := pc.MatchSchema(goals[0], p)
		if err != nil {
			xtracer.Trace("proof.ApplyProof EXIT proofType=SchemaInstantiation err=%v", err)
			return nil, err
		}
		if m == nil {
			xtracer.Trace("proof.ApplyProof EXIT proofType=SchemaInstantiation err=NoMatch")
			return nil, &NoMatch{Msg: "goal does not match the given schema"}
		}
		xtracer.Trace("proof.ApplyProof EXIT proofType=SchemaInstantiation ngoals=%d", len(m)+len(goals)-1)
		return append(m, goals[1:]...), nil

	case *ast.ComposeTactics:
		xtracer.Trace("proof.ApplyProof dispatch name=ComposeTactics")
		res, err := pc.composeProofs(goals, p.Tactics)
		xtracer.Trace("proof.ApplyProof EXIT proofType=ComposeTactics ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.NullTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=NullTactic")
		xtracer.Trace("proof.ApplyProof EXIT proofType=NullTactic ngoals=%d", len(goals))
		return goals, nil

	case *ast.ShowGoalsTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=ShowGoalsTactic")
		res := pc.showGoalsTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=ShowGoalsTactic ngoals=%d", len(res))
		return res, nil

	case *ast.DeferGoalTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=DeferGoalTactic")
		res := pc.deferGoalTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=DeferGoalTactic ngoals=%d", len(res))
		return res, nil

	case *ast.ForgetTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=ForgetTactic")
		res, err := pc.forgetTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=ForgetTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.ProofTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=ProofTactic")
		res, err := pc.proofTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=ProofTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.TacticTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=TacticTactic")
		res, err := pc.tacticTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=TacticTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.LetTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=LetTactic")
		res, err := pc.letTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=LetTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.AssumeGlobalTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=AssumeGlobalTactic")
		res, err := pc.assumeTactic(goals, &p.AssumeTactic, true)
		xtracer.Trace("proof.ApplyProof EXIT proofType=AssumeGlobalTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.AssumeTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=AssumeTactic")
		res, err := pc.assumeTactic(goals, p, false)
		xtracer.Trace("proof.ApplyProof EXIT proofType=AssumeTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.UnfoldTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=UnfoldTactic")
		res, err := pc.unfoldTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=UnfoldTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.IfTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=IfTactic")
		res, err := pc.ifTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=IfTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.PropertyTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=PropertyTactic")
		res, err := pc.propertyTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=PropertyTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.FunctionTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=FunctionTactic")
		res, err := pc.functionTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=FunctionTactic ngoals=%d err=%v", len(res), err)
		return res, err

	case *ast.WitnessTactic:
		xtracer.Trace("proof.ApplyProof dispatch name=WitnessTactic")
		res, err := pc.witnessTactic(goals, p)
		xtracer.Trace("proof.ApplyProof EXIT proofType=WitnessTactic ngoals=%d err=%v", len(res), err)
		return res, err
	}

	// Fallback: unrecognised proof type.
	xtracer.Trace("proof.ApplyProof EXIT proofType=%s err=unknown", iu.TypeName(proof))
	return nil, &ProofError{Msg: fmt.Sprintf("unknown proof type %T", proof)}
}

// MatchSchema attempts to match a goal to a schema using the full matching
// pipeline from Python ivy_proof.py:412-449 (match_schema).
//
// Pipeline:
//  1. setup_matching (rename_goal, transform_defn_schema, match_problem,
//     transform_defn_match, add_prem_match, compile_match)
//  2. apply initial compiled match
//  3. fo_match + match (with Tuple handling for premise matches)
//  4. detect_nonce_symbols
//  5. extract subgoals
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, proof *ast.SchemaInstantiation) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.MatchSchema ENTER goalLabel=%s HASH canon=%v", goal.LabelForTrace(), goal.Canon())
	goalConc := GoalConc(goal)
	if goalConc == nil {
		xtracer.Trace("proof.MatchSchema EXIT err=noConclusion")
		return nil, &NoMatch{Msg: "goal has no conclusion"}
	}
	// Schema matching does not apply to *ast.TemporalModels goals.
	// Mirror Python ivy_proof.py:429 which raises NoMatch in this case.
	if _, isTM := goalConc.(*ast.TemporalModels); isTM {
		xtracer.Trace("proof.MatchSchema EXIT err=temporalModels")
		return nil, &NoMatch{Msg: "schema matching does not apply to temporal-models goals"}
	}

	// Step 1: Build match problem (full pipeline)
	prob, pmatchIns, err := pc.SetupMatching(goal, proof, pc.Mod)
	if err != nil {
		return nil, err
	}
	pmatch := insMapToMap(pmatchIns)

	// Step 2: Apply initial proof match (from compile_match) to problem
	if len(pmatch) > 0 {
		ApplyMatchToProblem(pc.astCfg(), pmatch, prob)
	}

	// Step 3+4: Match (with Tuple handling for premise matches)
	// Python: if isinstance(prob.pat, ia.Tuple): ...
	if prob.TuplePats != nil {
		// Tuple matching: match each (pat, inst) pair sequentially
		for i := range prob.TuplePats {
			pat := prob.TuplePats[i]
			inst := prob.TupleInsts[i]

			fomatch := FOMatch(pat, inst, prob.FreeSyms, prob.Constants)
			if fomatch != nil && len(fomatch) > 0 {
				ApplyMatchToProblem(pc.astCfg(), fomatch, prob)
				// Update remaining tuple patterns with this match
				for j := i + 1; j < len(prob.TuplePats); j++ {
					prob.TuplePats[j] = ApplyMatch(fomatch, prob.TuplePats[j])
					prob.TupleInsts[j] = ApplyMatch(fomatch, prob.TupleInsts[j])
				}
			}

			somatch := Match(pat, inst, prob.FreeSyms, prob.Constants)
			if somatch == nil {
				return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
			}
			if len(somatch) > 0 {
				ApplyMatchToProblem(pc.astCfg(), somatch, prob)
				for j := i + 1; j < len(prob.TuplePats); j++ {
					prob.TuplePats[j] = ApplyMatchAlt(somatch, prob.TuplePats[j], nil)
					prob.TupleInsts[j] = ApplyMatchAlt(somatch, prob.TupleInsts[j], nil)
				}
			}
		}
	} else {
		// Non-tuple: single pattern matching
		fomatch := FOMatch(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
		if fomatch != nil && len(fomatch) > 0 {
			// Python: apply_match_to_problem(fomatch, prob, apply_match) — non-alt
			ApplyMatchToProblemNonAlt(pc.astCfg(), fomatch, prob)
		}

		somatch := Match(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
		if somatch == nil {
			xtracer.Trace("proof.MatchSchema EXIT err=matchFailed")
			return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
		}
		if len(somatch) > 0 {
			ApplyMatchToProblem(pc.astCfg(), somatch, prob)
		}
	}

	// Step 5: Detect nonce symbol clashes
	if err := DetectNonceSymbols(prob); err != nil {
		xtracer.Trace("proof.MatchSchema EXIT err=%v", err)
		return nil, err
	}

	// Step 6: Extract subgoals from matched schema
	if prob.SchemaLF == nil {
		xtracer.Trace("proof.MatchSchema EXIT err=schemaLFNil")
		return nil, &NoMatch{Msg: "schema is not a labeled formula after matching"}
	}
	result, err := GoalSubgoalsFromSchema(pc.astCfg(), prob.SchemaLF, goal)
	xtracer.Trace("proof.MatchSchema EXIT nsubgoals=%d err=%v", len(result), err)
	return result, err
}

// InstSchema instantiates a schema against a goal using the given match.
// Constructs a synthetic SchemaInstantiation and delegates to MatchSchema.
func InstSchema(checker *ProofChecker, schema, goal *ast.LabeledFormula, match map[string]string) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.InstSchema ENTER schemaLabel=%s goalLabel=%s", schema.LabelName(), goal.LabelName())
	schemaName := schema.LabelName()
	if schemaName == "" {
		xtracer.Trace("proof.InstSchema EXIT err=noLabel")
		return nil, &ProofError{Msg: "schema has no label"}
	}
	// Build a synthetic SchemaInstantiation with no renaming and no matches
	proof := checker.astCfg().NewSchemaInstantiation(checker.astCfg().NewAtom(schemaName), nil)
	result, err := checker.MatchSchema(goal, proof)
	xtracer.Trace("proof.InstSchema EXIT nsubgoals=%d err=%v", len(result), err)
	return result, err
}

// CheckSchema checks whether a goal matches a schema.
// Returns the resulting subgoals on success, or an error if
// the match fails. Delegates to MatchSchema.
func CheckSchema(checker *ProofChecker, goal, schema *ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	schemaName := schema.LabelName()
	if schemaName == "" {
		return nil, &ProofError{Msg: "schema has no label"}
	}
	proof := checker.astCfg().NewSchemaInstantiation(checker.astCfg().NewAtom(schemaName), nil)
	return checker.MatchSchema(goal, proof)
}

// AdmitDefinition admits a definition if it is non-recursive or matches a definition schema.
// If a proof is given it is used to match the definition to a schema, else
// default heuristic matching is used.
// Corresponds to Python's ProofChecker.admit_definition (ivy_proof.py:70-96).
func (pc *ProofChecker) AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.AdmitDefinition ENTER defnLabel=%s hasProof=%v", defn.LabelName(), proof != nil)
	defn = NormalizeGoal(pc.astCfg(), defn)
	// Extract the defined symbol
	def, ok := defn.Formula.(*lg.Definition)
	if !ok {
		return nil, &ProofError{Msg: "admit_definition: formula is not a Definition"}
	}
	sym := def.Defines()
	symSym, ok := sym.(*lg.Const)
	if !ok {
		return nil, &ProofError{Msg: "admit_definition: defines() did not return a Symbol"}
	}
	if _, exists := pc.Definitions[symSym.Name]; exists {
		return nil, &Redefinition{Node: defn, Msg: fmt.Sprintf("redefinition of %s", symSym.Name)}
	}
	if pc.Stale[symSym.Name] {
		return nil, &Circular{Node: defn, Msg: fmt.Sprintf("symbol %s defined after reference", symSym.Name)}
	}
	// Get dependencies from RHS
	deps := module.SymbolsAST(def.Rhs)
	for _, d := range deps {
		pc.Stale[lg.ExprName(d)] = true
	}
	// Check if recursive (sym in deps)
	recursive := false
	for _, d := range deps {
		if lg.ExprName(d) == symSym.Name {
			recursive = true
			break
		}
	}
	var subgoals []*ast.LabeledFormula
	if recursive {
		if proof == nil {
			xtracer.Trace("proof.AdmitDefinition EXIT err=noProof")
			return nil, &NoMatch{Node: defn, Msg: "no proof given for recursive definition"}
		}
		var err error
		subgoals, err = pc.ApplyProof([]*ast.LabeledFormula{defn}, proof)
		if err != nil {
			xtracer.Trace("proof.AdmitDefinition EXIT err=%v", err)
			return nil, err
		}
	}
	pc.Definitions[symSym.Name] = defn
	xtracer.Trace("proof.AdmitDefinition EXIT nsubgoals=%d sym=%s", len(subgoals), symSym.Name)
	return subgoals, nil
}

// AdmitProposition admits a proposition with proof.
// If a proof is given it is used to match the proposition to a schema,
// else default heuristic matching is used.
// Corresponds to Python's ProofChecker.admit_proposition (ivy_proof.py:98-121).
func (pc *ProofChecker) AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.AdmitProposition ENTER propLabel=%s hasProof=%v nExistingSubgoals=%d", prop.LabelName(), proof != nil, len(existingSubgoals))
	prop = NormalizeGoal(pc.astCfg(), prop)
	if _, isDef := prop.Formula.(*lg.Definition); isDef {
		xtracer.Trace("proof.AdmitProposition delegateToDefinition")
		return pc.AdmitDefinition(prop, proof)
	}
	if proof == nil {
		xtracer.Trace("proof.AdmitProposition EXIT err=noProof")
		return nil, &NoMatch{Node: prop, Msg: "no proof given for property"}
	}
	// Python: subgoals = subgoals or [prop]
	subgoals := existingSubgoals
	if len(subgoals) == 0 {
		subgoals = []*ast.LabeledFormula{prop}
	}
	var err error
	subgoals, err = pc.ApplyProof(subgoals, proof)
	if err != nil {
		xtracer.Trace("proof.AdmitProposition EXIT err=%v", err)
		return nil, err
	}
	pc.Axioms = append(pc.Axioms, prop)
	xtracer.Trace("proof.ProofChecker.admit_proposition schemata.insert key='%s' value=%s", prop.LabelName(), prop.Canon())
	pc.Schemata.Set(prop.LabelName(), prop)
	vocab := GoalVocab(prop)
	for _, sym := range vocab.Symbols {
		pc.Stale[sym.Name] = true
	}
	xtracer.Trace("proof.AdmitProposition EXIT nsubgoals=%d", len(subgoals))
	return subgoals, nil
}

// GetSubgoals returns the subgoals that result from applying proof to property
// prop, but does not admit prop in the context. Note, prop may not be a definition.
// Corresponds to Python's ProofChecker.get_subgoals (ivy_proof.py:123-134).
func (pc *ProofChecker) GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.GetSubgoals ENTER propLabel=%s", prop.LabelName())
	// Python: assert not isinstance(prop.formula, il.Definition) — checked BEFORE normalize
	if _, isDef := prop.Formula.(*lg.Definition); isDef {
		xtracer.Trace("proof.GetSubgoals EXIT err=isDefinition")
		return nil, &ProofError{Msg: "GetSubgoals: prop may not be a definition"}
	}
	prop = NormalizeGoal(pc.astCfg(), prop)
	subgoals, err := pc.ApplyProof([]*ast.LabeledFormula{prop}, proof)
	if err != nil {
		xtracer.Trace("proof.GetSubgoals EXIT err=%v", err)
		return nil, err
	}
	xtracer.Trace("proof.GetSubgoals EXIT nsubgoals=%d", len(subgoals))
	return subgoals, nil
}

// SetLastAxiom updates the last admitted axiom.
// Used by the compiler after named_trans to update the prover's state.
func (pc *ProofChecker) SetLastAxiom(prop *ast.LabeledFormula) {
	if len(pc.Axioms) > 0 {
		pc.Axioms[len(pc.Axioms)-1] = prop
	}
}

// SetSchema updates a schema entry by name.
func (pc *ProofChecker) SetSchema(name string, prop *ast.LabeledFormula) {
	xtracer.Trace("proof.ProofChecker.SetSchema schemata.insert key='%s' value=%s", name, prop.Canon())
	pc.Schemata.Set(name, prop)
}

// --- Helper methods for ApplyProof ---

// composeProofs applies a sequence of proofs one after another.
// Corresponds to Python's compose_proofs.
func (pc *ProofChecker) composeProofs(decls []*ast.LabeledFormula, proofs []ast.Node) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.composeProofs ENTER nproofs=%d ndecls=%d", len(proofs), len(decls))
	var err error
	for i, proof := range proofs {
		if len(decls) > 0 && decls[0] != nil {
			xtracer.Trace("proof.composeProofs step=%d/%d proofType=%s goal[0].Formula type=%s", i, len(proofs), iu.TypeName(proof), iu.TypeName(decls[0].Formula))
		}
		decls, err = pc.ApplyProof(decls, proof)
		if err != nil {
			xtracer.Trace("proof.composeProofs EXIT err=%v step=%d", err, i)
			return nil, err
		}
		if len(decls) == 0 {
			xtracer.Trace("proof.composeProofs EXIT ndecls=0 step=%d", i)
			return decls, nil
		}
	}
	xtracer.Trace("proof.composeProofs EXIT ndecls=%d", len(decls))
	return decls, nil
}

func (pc *ProofChecker) deferGoalTactic(decls []*ast.LabeledFormula, proof *ast.DeferGoalTactic) []*ast.LabeledFormula {
	xtracer.Trace("proof.deferGoalTactic ENTER ndecls=%d", len(decls))
	result := append(decls[1:], decls[0:1]...)
	xtracer.Trace("proof.deferGoalTactic EXIT ndecls=%d", len(result))
	return result
}

func (pc *ProofChecker) showGoalsTactic(decls []*ast.LabeledFormula, proof *ast.ShowGoalsTactic) []*ast.LabeledFormula {
	xtracer.Trace("proof.showGoalsTactic ENTER ndecls=%d", len(decls))
	fmt.Println()
	loc := proof.GetLineno()
	fmt.Printf("line %d: Proof goals:\n", loc.Line)
	for _, decl := range decls {
		fmt.Println()
		fmt.Println("theorem " + decl.String())
		fmt.Println()
	}
	xtracer.Trace("proof.showGoalsTactic EXIT ndecls=%d", len(decls))
	return decls
}

func (pc *ProofChecker) forgetTactic(decls []*ast.LabeledFormula, proof *ast.ForgetTactic) ([]*ast.LabeledFormula, error) {
	xtracer.Trace("proof.forgetTactic ENTER ndecls=%d nNames=%d", len(decls), len(proof.Names))
	decl := decls[0]
	forgetNames := make(map[string]bool)
	for _, n := range proof.Names {
		forgetNames[nodeToString(n)] = true
	}
	prems := GoalPrems(decl)
	var kept []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			if forgetNames[lf.LabelName()] {
				continue
			}
		}
		kept = append(kept, p)
	}
	newGoal := CloneGoal(pc.astCfg(), decl, kept, GoalConc(decl))
	result := []*ast.LabeledFormula{newGoal}
	result = append(result, decls[1:]...)
	xtracer.Trace("proof.forgetTactic EXIT nkept=%d", len(kept))
	return result, nil
}

// proofTactic applies a proof to a specific labeled goal.
// Corresponds to Python's proof_tactic.
func (pc *ProofChecker) proofTactic(decls []*ast.LabeledFormula, proof *ast.ProofTactic) ([]*ast.LabeledFormula, error) {
	labelStr := nodeToString(proof.TLabel)
	xtracer.Trace("proof.proofTactic ENTER label=%s ndecls=%d", labelStr, len(decls))
	for idx, decl := range decls {
		if nodeToString(decl.Label) == labelStr {
			subgoals, err := pc.ApplyProof([]*ast.LabeledFormula{decl}, proof.Proof)
			if err != nil {
				xtracer.Trace("proof.proofTactic EXIT err=%v", err)
				return nil, err
			}
			// Remove the matched goal and append subgoals at the end.
			rest := make([]*ast.LabeledFormula, 0, len(decls)-1+len(subgoals))
			rest = append(rest, decls[:idx]...)
			rest = append(rest, decls[idx+1:]...)
			rest = append(rest, subgoals...)
			xtracer.Trace("proof.proofTactic EXIT nsubgoals=%d ndecls=%d", len(subgoals), len(rest))
			return rest, nil
		}
	}
	xtracer.Trace("proof.proofTactic EXIT err=noLabel label=%s", labelStr)
	return nil, &ProofError{Msg: fmt.Sprintf("no goal with label %s", labelStr)}
}

// tacticTactic dispatches to a registered tactic by name.
// Corresponds to Python's tactic_tactic.
func (pc *ProofChecker) tacticTactic(decls []*ast.LabeledFormula, proof *ast.TacticTactic) ([]*ast.LabeledFormula, error) {
	tn := nodeToString(proof.TName)
	if len(decls) > 0 && decls[0] != nil {
		xtracer.Trace("proof.tacticTactic name='%s' goal[0].Formula type=%s", tn, iu.TypeName(decls[0].Formula))
	}
	tactic, ok := pc.Cfg.Tactics[tn]
	if !ok {
		xtracer.Trace("proof.tacticTactic EXIT name=%s err=unknownTactic", tn)
		return nil, &ProofError{Msg: fmt.Sprintf("unknown tactic: %s", tn)}
	}
	result, err := tactic(pc, decls, proof)
	xtracer.Trace("proof.tacticTactic EXIT name=%s nresult=%d err=%v", tn, len(result), err)
	return result, err
}

// ApplyMatchGoal applies a match (symbol substitution map) to a goal.
// Corresponds to Python's apply_match_goal.
// For now this delegates to the matching infrastructure in match.go;
// goals without SchemaBody formulas have the match applied to the
// formula directly.
// ApplyMatchGoal applies a match substitution to a proof goal.
// Matches Python ivy_proof.py apply_match_goal:
//   - For LabeledFormula with SchemaBody: apply match to premises and conclusion
//   - For LabeledFormula with plain formula: apply match to the formula
//   - Uses alpha-renaming to avoid capture by binders
func ApplyMatchGoal(cfg *ast.AstConfig, match map[string]string, goal *ast.LabeledFormula) *ast.LabeledFormula {
	xtracer.Trace("proof.ApplyMatchGoal ENTER nmatch=%d goalNil=%v", len(match), goal == nil)
	if len(match) == 0 || goal == nil {
		xtracer.Trace("proof.ApplyMatchGoal EXIT passthrough")
		return goal
	}
	// Build substitution map: string name → AST node
	subs := make(map[string]ast.Node)
	for k, v := range match {
		subs[k] = cfg.NewSymbol(v, nil)
	}
	// Apply substitution to the formula
	fmla := goal.Formula
	if fmla != nil {
		if _, isSchema := fmla.(*ast.SchemaBody); isSchema {
			// SchemaBody: apply match to premises and conclusion separately.
			// Python: prems + [apply_match(match, fmla.conc(), env)]
			fmla = ast.SubstituteAst(fmla, subs)
		} else {
			fmla = ast.SubstituteAst(fmla, subs)
		}
	}
	result := goal.Clone([]ast.Node{goal.Label, fmla}).(*ast.LabeledFormula)
	xtracer.Trace("proof.ApplyMatchGoal EXIT HASH canon=%v", result.Canon())
	return result
}

// --- Helpers ---

// nodeToString extracts a string name from an AST node.
// For *Atom, returns Relname(); otherwise uses String().
func nodeToString(n ast.Node) string {
	if n == nil {
		return ""
	}
	if a, ok := n.(*ast.Atom); ok {
		return a.Relname()
	}
	return fmt.Sprint(n)
}

// collectStaleSymbols walks an ast.Node tree via Args() and marks all
// lg.Const symbols found as stale. This matches Python's
// lu.used_symbols_ast(lf.formula) which walks the entire formula,
// including premises in SchemaBody — not just the conclusion.
func collectStaleSymbols(n ast.Node, stale map[string]bool) {
	if n == nil {
		return
	}
	// If this node is an lg.Const, mark it.
	if c, ok := n.(*lg.Const); ok {
		stale[c.Name] = true
		return
	}
	// If this node is an lg.Expr, use the existing UsedSymbolsAST
	// which handles logic-level nodes efficiently.
	if expr, ok := n.(lg.Expr); ok {
		for _, c := range module.UsedSymbolsAST(expr) {
			stale[lg.ExprName(c)] = true
		}
		return
	}
	// Otherwise walk children via Args().
	for _, child := range n.Args() {
		collectStaleSymbols(child, stale)
	}
}

// PrettyLineno formats a location-holding node for display.
func PrettyLineno(n ast.Node) string {
	if n == nil {
		return "(internal) "
	}
	loc := n.GetLineno()
	if loc.Line > 0 {
		return fmt.Sprintf("line %d: ", loc.Line)
	}
	return "(internal) "
}
