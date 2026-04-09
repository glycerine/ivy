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
	Schemata map[string]*ast.LabeledFormula
	// Stale is the set of symbols that have been referenced and are not fresh.
	Stale map[string]bool
}

// NewProofChecker creates a new ProofChecker.
//
// axioms and definitions are lists of LabeledFormula.
// schemata is an optional map from string names to LabeledFormula.
func NewProofChecker(cfg *module.ProofConfig, mod *module.Module, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula, astCfgs ...*ast.AstConfig) *ProofChecker {
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
		Schemata:    make(map[string]*ast.LabeledFormula),
		Stale:       make(map[string]bool),
	}

	// Normalize axioms
	for _, ax := range axioms {
		norm := NormalizeGoal(pc.AstCfg, ax)
		pc.Axioms = append(pc.Axioms, norm)
		if ax.Label != nil {
			pc.Schemata[ax.LabelName()] = ax
		}
	}

	// Normalize definitions — key by defines().name per Python ivy_proof.py:53
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

	// Normalize schemata
	if schemata != nil {
		for name, s := range schemata {
			pc.Schemata[name] = NormalizeGoal(pc.AstCfg, s)
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
		for _, s := range schemata {
			vocab := GoalVocab(s)
			for _, sym := range vocab.Symbols {
				pc.Stale[sym.Name] = true
			}
		}
	}

	return pc
}

// GetModule returns the current module. Implements module.ProofCheckerInterface.
func (pc *ProofChecker) GetModule() *module.Module { return pc.Mod }

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
		pc.Schemata[ax.LabelName()] = ax
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
	if s, ok := pc.Schemata[name]; ok {
		if err := CheckSchemaCapture(s, goal); err != nil {
			return nil, err
		}
		return s, nil
	}
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
			return schema, nil
		}
		// Not a *lg.Definition — return as-is
		if err := CheckSchemaCapture(d, goal); err != nil {
			return nil, err
		}
		return d, nil
	}
	// Check goal premises
	for _, pg := range GoalPremGoals(goal) {
		if pg.LabelName() == name {
			return pg, nil
		}
	}
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
		m, err := pc.MatchSchema(goals[0], p)
		if err != nil {
			return nil, err
		}
		if m == nil {
			return nil, &NoMatch{Msg: "goal does not match the given schema"}
		}
		return append(m, goals[1:]...), nil

	case *ast.ComposeTactics:
		return pc.composeProofs(goals, p.Tactics)

	case *ast.NullTactic:
		return goals, nil

	case *ast.ShowGoalsTactic:
		fmt.Println()
		loc := p.GetLineno()
		fmt.Printf("line %d: Proof goals:\n", loc.Line)
		for _, decl := range goals {
			fmt.Println()
			fmt.Println("theorem " + decl.String())
			fmt.Println()
		}
		return goals, nil

	case *ast.DeferGoalTactic:
		if len(goals) <= 1 {
			return goals, nil
		}
		return append(goals[1:], goals[0]), nil

	case *ast.ForgetTactic:
		return pc.forgetTactic(goals, p)

	case *ast.ProofTactic:
		return pc.proofTactic(goals, p)

	case *ast.TacticTactic:
		return pc.tacticTactic(goals, p)

	case *ast.LetTactic:
		return pc.letTactic(goals, p)

	case *ast.AssumeTactic:
		return pc.assumeTactic(goals, p)

	case *ast.UnfoldTactic:
		return pc.unfoldTactic(goals, p)

	case *ast.IfTactic:
		return pc.ifTactic(goals, p)

	case *ast.PropertyTactic:
		return pc.propertyTactic(goals, p)

	case *ast.FunctionTactic:
		return pc.functionTactic(goals, p)

	case *ast.WitnessTactic:
		return pc.witnessTactic(goals, p)
	}

	// Fallback: unrecognised proof type.
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
	goalConc := GoalConc(goal)
	if goalConc == nil {
		return nil, &NoMatch{Msg: "goal has no conclusion"}
	}

	// Step 1: Build match problem (full pipeline)
	prob, pmatch, err := pc.SetupMatching(goal, proof, pc.Mod)
	if err != nil {
		return nil, err
	}

	// Step 2: Apply initial proof match (from compile_match) to problem
	if len(pmatch) > 0 {
		ApplyMatchToProblem(pc.astCfg(),pmatch, prob)
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
				ApplyMatchToProblem(pc.astCfg(),fomatch, prob)
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
				ApplyMatchToProblem(pc.astCfg(),somatch, prob)
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
			ApplyMatchToProblem(pc.astCfg(),fomatch, prob)
		}

		somatch := Match(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
		if somatch == nil {
			return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
		}
		if len(somatch) > 0 {
			ApplyMatchToProblem(pc.astCfg(),somatch, prob)
		}
	}

	// Step 5: Detect nonce symbol clashes
	if err := DetectNonceSymbols(prob); err != nil {
		return nil, err
	}

	// Step 6: Extract subgoals from matched schema
	if prob.SchemaLF == nil {
		return nil, &NoMatch{Msg: "schema is not a labeled formula after matching"}
	}
	return GoalSubgoalsFromSchema(pc.astCfg(), prob.SchemaLF, goal), nil
}

// InstSchema instantiates a schema against a goal using the given match.
// Constructs a synthetic SchemaInstantiation and delegates to MatchSchema.
func InstSchema(checker *ProofChecker, schema, goal *ast.LabeledFormula, match map[string]string) ([]*ast.LabeledFormula, error) {
	schemaName := schema.LabelName()
	if schemaName == "" {
		return nil, &ProofError{Msg: "schema has no label"}
	}
	// Build a synthetic SchemaInstantiation with no renaming and no matches
	proof := checker.astCfg().NewSchemaInstantiation(checker.astCfg().NewAtom(schemaName), nil)
	return checker.MatchSchema(goal, proof)
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
			return nil, &NoMatch{Node: defn, Msg: "no proof given for recursive definition"}
		}
		var err error
		subgoals, err = pc.ApplyProof([]*ast.LabeledFormula{defn}, proof)
		if err != nil {
			return nil, err
		}
	}
	pc.Definitions[symSym.Name] = defn
	return subgoals, nil
}

// AdmitProposition admits a proposition with proof.
// If a proof is given it is used to match the proposition to a schema,
// else default heuristic matching is used.
// Corresponds to Python's ProofChecker.admit_proposition (ivy_proof.py:98-121).
func (pc *ProofChecker) AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	prop = NormalizeGoal(pc.astCfg(), prop)
	if _, isDef := prop.Formula.(*lg.Definition); isDef {
		return pc.AdmitDefinition(prop, proof)
	}
	if proof == nil {
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
		return nil, err
	}
	pc.Axioms = append(pc.Axioms, prop)
	pc.Schemata[prop.LabelName()] = prop
	vocab := GoalVocab(prop)
	for _, sym := range vocab.Symbols {
		pc.Stale[sym.Name] = true
	}
	return subgoals, nil
}

// GetSubgoals returns the subgoals that result from applying proof to property
// prop, but does not admit prop in the context. Note, prop may not be a definition.
// Corresponds to Python's ProofChecker.get_subgoals (ivy_proof.py:123-134).
func (pc *ProofChecker) GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	// Python: assert not isinstance(prop.formula, il.Definition) — checked BEFORE normalize
	if _, isDef := prop.Formula.(*lg.Definition); isDef {
		return nil, &ProofError{Msg: "GetSubgoals: prop may not be a definition"}
	}
	prop = NormalizeGoal(pc.astCfg(), prop)
	subgoals, err := pc.ApplyProof([]*ast.LabeledFormula{prop}, proof)
	if err != nil {
		return nil, err
	}
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
	pc.Schemata[name] = prop
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
			return nil, err
		}
		if len(decls) == 0 {
			return decls, nil
		}
	}
	return decls, nil
}

// forgetTactic removes named premises from the first goal.
// Corresponds to Python's forget_tactic.
func (pc *ProofChecker) forgetTactic(decls []*ast.LabeledFormula, proof *ast.ForgetTactic) ([]*ast.LabeledFormula, error) {
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
	return result, nil
}

// proofTactic applies a proof to a specific labeled goal.
// Corresponds to Python's proof_tactic.
func (pc *ProofChecker) proofTactic(decls []*ast.LabeledFormula, proof *ast.ProofTactic) ([]*ast.LabeledFormula, error) {
	labelStr := nodeToString(proof.TLabel)
	for idx, decl := range decls {
		if nodeToString(decl.Label) == labelStr {
			subgoals, err := pc.ApplyProof([]*ast.LabeledFormula{decl}, proof.Proof)
			if err != nil {
				return nil, err
			}
			// Remove the matched goal and append subgoals at the end.
			rest := make([]*ast.LabeledFormula, 0, len(decls)-1+len(subgoals))
			rest = append(rest, decls[:idx]...)
			rest = append(rest, decls[idx+1:]...)
			rest = append(rest, subgoals...)
			return rest, nil
		}
	}
	return nil, &ProofError{Msg: fmt.Sprintf("no goal with label %s", labelStr)}
}

// tacticTactic dispatches to a registered tactic by name.
// Corresponds to Python's tactic_tactic.
func (pc *ProofChecker) tacticTactic(decls []*ast.LabeledFormula, proof *ast.TacticTactic) ([]*ast.LabeledFormula, error) {
	tn := nodeToString(proof.TName)
	if len(decls) > 0 && decls[0] != nil {
		xtracer.Trace("proof.tacticTactic name=%q goal[0].Formula type=%s", tn, iu.TypeName(decls[0].Formula))
	}
	tactic, ok := pc.Cfg.Tactics[tn]
	if !ok {
		return nil, &ProofError{Msg: fmt.Sprintf("unknown tactic: %s", tn)}
	}
	return tactic(pc, decls, proof)
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
	if len(match) == 0 || goal == nil {
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
	return goal.Clone([]ast.Node{goal.Label, fmla}).(*ast.LabeledFormula)
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
