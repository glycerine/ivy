package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
)

// Tactic is a function that applies a proof tactic to a goal,
// producing subgoals or an error.
type Tactic func(checker *ProofChecker, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)

// RegisteredTactics is the global registry of named tactics.
var RegisteredTactics = map[string]Tactic{}

// RegisterTactic registers a named tactic.
func RegisterTactic(name string, tactic Tactic) {
	RegisteredTactics[name] = tactic
}

// ProofChecker is Ivy's built-in proof checker.
type ProofChecker struct {
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
func NewProofChecker(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) *ProofChecker {
	pc := &ProofChecker{
		Definitions: make(map[string]*ast.LabeledFormula),
		Schemata:    make(map[string]*ast.LabeledFormula),
		Stale:       make(map[string]bool),
	}

	// Normalize axioms
	for _, ax := range axioms {
		norm := NormalizeGoal(ax)
		pc.Axioms = append(pc.Axioms, norm)
		if ax.Label != nil {
			pc.Schemata[ax.LabelName()] = ax
		}
	}

	// Normalize definitions — key by defines().name per Python ivy_proof.py:53
	for _, d := range definitions {
		norm := NormalizeGoal(d)
		name := ""
		if def, ok := d.Formula.(*lg.Definition); ok {
			if sym, ok := def.Defines().(*lg.Symbol); ok {
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
			pc.Schemata[name] = NormalizeGoal(s)
		}
	}

	// Mark stale symbols: collect used symbols from axioms and definitions.
	// Corresponds to Python's:
	//   self.stale = set()
	//   for lf in axioms + definitions:
	//       self.stale.update(lu.used_symbols_ast(lf.formula))
	for _, lf := range axioms {
		conc := GoalConc(lf)
		if conc != nil {
			for _, sym := range clauseops.UsedSymbolsAST(conc) {
				if c, ok := sym.(*lg.Symbol); ok {
					pc.Stale[c.Name] = true
				}
			}
		}
	}
	for _, lf := range definitions {
		conc := GoalConc(lf)
		if conc != nil {
			for _, sym := range clauseops.UsedSymbolsAST(conc) {
				if c, ok := sym.(*lg.Symbol); ok {
					pc.Stale[c.Name] = true
				}
			}
		}
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

// AdmitAxiom adds an axiom to the checker.
func (pc *ProofChecker) AdmitAxiom(ax *ast.LabeledFormula) {
	norm := NormalizeGoal(ax)
	pc.Axioms = append(pc.Axioms, norm)
	if ax.Label != nil {
		pc.Schemata[ax.LabelName()] = ax
	}
}

// LookupSchema looks up a schema by name in the checker's schemata,
// definitions, or goal premises.
func (pc *ProofChecker) LookupSchema(name string, goal *ast.LabeledFormula) (*ast.LabeledFormula, error) {
	if s, ok := pc.Schemata[name]; ok {
		return s, nil
	}
	if d, ok := pc.Definitions[name]; ok {
		return d, nil
	}
	// Check goal premises
	for _, pg := range GoalPremGoals(goal) {
		if pg.LabelName() == name {
			return pg, nil
		}
	}
	return nil, &ProofError{Msg: "No property " + name + " exists in the current context"}
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

	// Dispatch on proof type.
	switch p := proof.(type) {
	case *ast.SchemaInstantiation:
		sname := nodeToString(p.SchemaName)
		m, err := pc.MatchSchema(goals[0], sname)
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

// MatchSchema attempts to match a goal to a schema.
// Corresponds to Python's ProofChecker.match_schema which:
//  1. Looks up the schema by name.
//  2. Sets up matching (setup_matching) to build a MatchProblem.
//  3. Applies first-order match (fo_match), then second-order match.
//  4. If successful, returns goal_subgoals(schema, decl, lineno).
//
// The full matching pipeline (setup_matching, transform_defn_schema,
// MatchSchema matches a schema to a goal using the full matching pipeline:
// setup_matching → fo_match → match → apply_match_to_problem → detect_nonce_symbols.
//
// Python: ivy_proof.py:412-449
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, schemaName string) ([]*ast.LabeledFormula, error) {
	// Step 1: Build match problem
	prob, pmatch, err := pc.SetupMatching(goal, schemaName)
	if err != nil {
		return nil, err
	}

	goalConc := GoalConc(goal)
	if goalConc == nil {
		return nil, &NoMatch{Msg: "goal has no conclusion"}
	}

	// Step 2: Apply initial proof match (from proof AST bindings) to problem
	if len(pmatch) > 0 {
		ApplyMatchToProblem(pmatch, prob)
	}

	// Step 3: First-order match
	fomatch := FOMatch(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
	if fomatch != nil && len(fomatch) > 0 {
		ApplyMatchToProblem(fomatch, prob)
	}

	// Step 4: Second-order match (full match with lambda extraction)
	somatch := Match(prob.Pat, prob.Inst, prob.FreeSyms, prob.Constants)
	if somatch == nil {
		return nil, &NoMatch{Msg: "goal does not match the given schema"}
	}
	if len(somatch) > 0 {
		ApplyMatchToProblem(somatch, prob)
	}

	// Step 5: Detect nonce symbol clashes
	if err := DetectNonceSymbols(prob); err != nil {
		return nil, err
	}

	// Step 6: Extract subgoals from matched schema
	if prob.SchemaLF == nil {
		return nil, &NoMatch{Msg: "schema is not a labeled formula after matching"}
	}
	return GoalSubgoalsFromSchema(prob.SchemaLF, goal), nil
}

// InstSchema instantiates a schema against a goal using the given match.
// Corresponds to Python's apply_match_goal pipeline. The match maps
// schema-side symbol names to goal-side symbol names. Until the full
// apply_match_goal machinery is ported, this delegates to MatchSchema.
func InstSchema(checker *ProofChecker, schema, goal *ast.LabeledFormula, match map[string]string) ([]*ast.LabeledFormula, error) {
	schemaName := schema.LabelName()
	if schemaName == "" {
		return nil, &ProofError{Msg: "schema has no label"}
	}
	return checker.MatchSchema(goal, schemaName)
}

// CheckSchema checks whether a goal matches a schema.
// Returns the resulting subgoals on success, or an error if
// the match fails. Delegates to MatchSchema.
func CheckSchema(checker *ProofChecker, goal, schema *ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	schemaName := schema.LabelName()
	if schemaName == "" {
		return nil, &ProofError{Msg: "schema has no label"}
	}
	return checker.MatchSchema(goal, schemaName)
}

// AdmitDefinition admits a definition if it is non-recursive or matches a definition schema.
// If a proof is given it is used to match the definition to a schema, else
// default heuristic matching is used.
// Corresponds to Python's ProofChecker.admit_definition (ivy_proof.py:70-96).
func (pc *ProofChecker) AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	defn = NormalizeGoal(defn)
	// Extract the defined symbol
	def, ok := defn.Formula.(*lg.Definition)
	if !ok {
		return nil, &ProofError{Msg: "admit_definition: formula is not a Definition"}
	}
	sym := def.Defines()
	symSym, ok := sym.(*lg.Symbol)
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
	deps := clauseops.SymbolsAST(def.Rhs)
	for _, d := range deps {
		pc.Stale[d.Name] = true
	}
	// Check if recursive (sym in deps)
	recursive := false
	for _, d := range deps {
		if d.Name == symSym.Name {
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
		if subgoals == nil {
			return nil, &NoMatch{Node: defn, Msg: "recursive definition does not match the given schema"}
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
	prop = NormalizeGoal(prop)
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
	if subgoals == nil {
		return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
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
	prop = NormalizeGoal(prop)
	// Python: assert not isinstance(prop.formula, il.Definition)
	if _, isDef := prop.Formula.(*lg.Definition); isDef {
		return nil, &ProofError{Msg: "GetSubgoals: prop may not be a definition"}
	}
	subgoals, err := pc.ApplyProof([]*ast.LabeledFormula{prop}, proof)
	if err != nil {
		return nil, err
	}
	if subgoals == nil {
		return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
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
	var err error
	for _, proof := range proofs {
		decls, err = pc.ApplyProof(decls, proof)
		if err != nil {
			return nil, err
		}
		if decls == nil || len(decls) == 0 {
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
	newGoal := CloneGoal(decl, kept, GoalConc(decl))
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
	tactic, ok := RegisteredTactics[tn]
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
func ApplyMatchGoal(match map[string]string, goal *ast.LabeledFormula) *ast.LabeledFormula {
	if len(match) == 0 || goal == nil {
		return goal
	}
	// Build substitution map: string name → AST node
	subs := make(map[string]ast.Node)
	for k, v := range match {
		subs[k] = ast.NewSymbol(v, nil)
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
