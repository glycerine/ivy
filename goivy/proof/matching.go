package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
)

// SetupMatching creates a MatchProblem for matching a schema to a declaration.
// This is the first stage of the proof matching pipeline.
//
// Python: ivy_proof.py:324-327
func (pc *ProofChecker) SetupMatching(decl *ast.LabeledFormula, proof *ast.SchemaInstantiation, mod *module.Module) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {
	schemaName := nodeToString(proof.SchemaName)
	schema, err := pc.LookupSchema(schemaName, decl, proof, false)
	if err != nil {
		return nil, nil, err
	}
	return pc.SetupSchemaMatching(decl, proof, schema, false, mod)
}

// SetupSchemaMatchingRaw implements the complete Python pipeline from
// ivy_proof.py:337-348 (setup_schema_matching), taking raw ren/matches
// fields so it can be used by both MatchSchema (SchemaInstantiation)
// and assumeTactic (AssumeTactic).
//
// Pipeline steps:
//  1. rename_goal(schema, proof.renaming())
//  2. transform_defn_schema(schema, decl)
//  3. match_problem(schema, decl)
//  4. transform_defn_match(prob)
//  5. add_prem_match(proof.match(), prob, decl, self)
//  6. compile_match(proof_match, prob, decl, allow_witness)
func (pc *ProofChecker) SetupSchemaMatchingRaw(
	decl *ast.LabeledFormula,
	ren ast.Node,
	matches []ast.Node,
	schema *ast.LabeledFormula,
	allowWitness bool,
) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {

	// Step 1: Rename schema using proof renaming
	// Python: schema = rename_goal(schema, proof.renaming())
	if ren != nil {
		var err error
		schema, err = RenameGoal(pc.astCfg(), schema, ren)
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 2: Transform definition schema for parameter arity matching
	// Python: schema = transform_defn_schema(schema, decl)
	schema = TransformDefnSchema(pc.astCfg(), schema, decl)

	// Step 3: Build match problem
	// Python: prob = match_problem(schema, decl)
	prob := buildMatchProblem(schema, decl)
	if prob == nil {
		return nil, nil, &NoMatch{Msg: "cannot build match problem from schema and goal"}
	}

	// Step 4: Transform definition match (full version from phase5)
	// Python: prob = transform_defn_match(prob)
	prob = TransformDefnMatch(pc.astCfg(), prob)
	if prob == nil {
		return nil, nil, &NoMatch{Msg: "definition does not match the given schema"}
	}

	// Step 5: Process premise matches
	// Python: proof_match, prob = add_prem_match(proof.match(), prob, decl, self)
	proofMatches := matches
	proofMatches, prob = AddPremMatch(proofMatches, prob, decl, pc)

	// Step 6: Compile symbolic matches
	// Python: pmatch = compile_match(proof_match, prob, decl, allow_witness)
	pmatch := CompileMatchFull(proofMatches, prob, decl, allowWitness, pc.Mod)
	if pmatch == nil && len(proofMatches) > 0 {
		return nil, nil, &ProofError{Msg: "Match is inconsistent"}
	}
	if pmatch == nil {
		pmatch = make(map[lg.NodeKey]lg.Expr)
	}

	return prob, pmatch, nil
}

// SetupSchemaMatching is the SchemaInstantiation-specific wrapper around
// SetupSchemaMatchingRaw. Extracts ren/matches from the proof node.
func (pc *ProofChecker) SetupSchemaMatching(
	decl *ast.LabeledFormula,
	proof *ast.SchemaInstantiation,
	schema *ast.LabeledFormula,
	allowWitness bool,
	mod *module.Module,
) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {
	var ren ast.Node
	var matches []ast.Node
	if proof != nil {
		ren = proof.Ren
		matches = proof.Matches
	}
	return pc.SetupSchemaMatchingRaw(decl, ren, matches, schema, allowWitness)
}

// buildMatchProblem creates a MatchProblem from a schema and declaration.
//
// Python: ivy_proof.py:771-776
func buildMatchProblem(schema, decl *ast.LabeledFormula) *MatchProblem {
	vocab := GoalVocab(schema)

	freesyms := make(map[lg.NodeKey]lg.Expr)
	for _, sym := range vocab.Symbols {
		freesyms[lg.Key(sym)] = sym
	}
	for _, s := range vocab.Sorts {
		freesyms[lg.Key(s)] = s
	}
	for _, v := range vocab.Variables {
		freesyms[lg.Key(v)] = v
	}

	constants := make(map[lg.NodeKey]lg.Expr)
	freeVars := GoalFreeVars(decl)
	for _, v := range freeVars {
		constants[lg.Key(v)] = v
	}

	// Python: MatchProblem(schema, goal_conc(schema), goal_conc(decl), freesyms, constants)
	// Python is duck-typed and passes conclusions as-is. Go needs lg.Expr
	// for Pat/Inst; when the conclusion is *ast.TemporalModels (or other
	// non-lg.Expr), Pat/Inst will be nil. This is fine — assume_tactic
	// uses the prob primarily for SchemaLF/FreeSyms, and the matching
	// pipeline tolerates nil Pat/Inst when there are no proof matches.
	// Note: the TemporalModels rejection for match_schema lives in
	// MatchSchema (checker.go:307), NOT here. Python's match_problem
	// (ivy_proof.py:779-784) has no such rejection.
	schemaConc, _ := GoalConc(schema).(lg.Expr)
	declConc, _ := GoalConc(decl).(lg.Expr)

	// Store the schema LabeledFormula reference in the problem
	prob := NewMatchProblem(nil, schemaConc, declConc, freesyms, constants)
	prob.SchemaLF = schema
	return prob
}

// ApplyMatchToProblem applies a match map to a MatchProblem, updating
// the schema, pattern, and free symbols.
// Uses apply_match_alt (capture-checking) for schema and pattern.
// Used for somatch applications.
//
// Python: ivy_proof.py:1002-1007 (apply_match_to_problem with apply_match_alt)
func ApplyMatchToProblem(cfg *ast.AstConfig, match map[lg.NodeKey]lg.Expr, prob *MatchProblem) {
	// Python's apply_match_to_problem has no early return for empty match.

	// Avoid capture before applying — Python: avoid_capture_problem(prob, match)
	AvoidCaptureProblem(cfg, prob, match)

	// Apply match to schema — use ApplyMatchGoalNode (processes premises + conclusion)
	if prob.SchemaLF != nil {
		prob.SchemaLF = ApplyMatchGoalNode(cfg, match, prob.SchemaLF)
	}

	// Apply match to pattern — use ApplyMatchAlt for capture safety
	prob.Pat = ApplyMatchAlt(match, prob.Pat, nil)

	// Update free symbols — Python uses non-alt apply_match_freesyms here.
	prob.FreeSyms = ApplyMatchFreesyms(match, prob.FreeSyms)

	// Remove matched symbols from revmap
	for k := range prob.RevMap {
		if _, matched := match[k]; matched {
			delete(prob.RevMap, k)
		}
	}
}

// ApplyMatchToProblemNonAlt applies a match map using the non-alt (non-capture-checking)
// apply function. Used for fomatch applications.
//
// Python: ivy_proof.py:1002-1007 (apply_match_to_problem with apply_match)
func ApplyMatchToProblemNonAlt(cfg *ast.AstConfig, match map[lg.NodeKey]lg.Expr, prob *MatchProblem) {
	// Python's apply_match_to_problem has no early return for empty match.
	AvoidCaptureProblem(cfg, prob, match)
	if prob.SchemaLF != nil {
		prob.SchemaLF = ApplyMatchGoalNodeNonAlt(cfg, match, prob.SchemaLF)
	}
	// Non-alt: use ApplyMatch (no capture detection)
	prob.Pat = ApplyMatch(match, prob.Pat)
	prob.FreeSyms = ApplyMatchFreesyms(match, prob.FreeSyms)
	for k := range prob.RevMap {
		if _, matched := match[k]; matched {
			delete(prob.RevMap, k)
		}
	}
}

// DetectNonceSymbols checks that no nonce symbols from avoid_capture remain
// free after matching. If any remain, it means the matching created
// symbol clashes between the schema and the goal.
//
// Python: ivy_proof.py:1020-1028
func DetectNonceSymbols(prob *MatchProblem) error {
	for _, sym := range prob.RevMap {
		return &CaptureError{
			Msg: fmt.Sprintf("Symbol %s in schema clashes with corresponding symbol in goal. Suggest renaming or instantiating it.", sym),
		}
	}
	return nil
}

// GoalSubgoalsFromSchema computes the subgoals remaining after matching a schema to a goal.
//
// Python: ivy_proof.py:634-643
func GoalSubgoalsFromSchema(cfg *ast.AstConfig, schema *ast.LabeledFormula, goal *ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	goalPremGoals := GoalPremGoals(goal)
	goalPremNames := make(map[string]bool, len(goalPremGoals))
	for _, pg := range goalPremGoals {
		goalPremNames[pg.LabelName()] = true
	}

	var subgoals []*ast.LabeledFormula
	for _, pg := range GoalPremGoals(schema) {
		if goalPremNames[pg.LabelName()] {
			continue
		}
		if TrivialGoal(pg) {
			continue
		}
		sub, err := GoalSubst(cfg, goal, pg, goal.GetLineno())
		if err != nil {
			return nil, err
		}
		subgoals = append(subgoals, sub)
	}
	return subgoals, nil
}

// GoalFreeVars returns the free variables of a goal's conclusion.
// Unwraps *ast.TemporalModels via ConcAsExpr — mirrors Python's
// duck-typed access to free variables across temporal goals.
func GoalFreeVars(g *ast.LabeledFormula) []*lg.Variable {
	conc := GoalConcUnwrap(g)
	if conc == nil {
		return nil
	}
	return lu.FreeVariablesList(conc)
}
