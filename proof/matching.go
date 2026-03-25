package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/module"
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

// SetupSchemaMatching implements the complete Python pipeline from ivy_proof.py:329-340:
//  1. rename_goal(schema, proof.renaming())
//  2. transform_defn_schema(schema, decl)
//  3. match_problem(schema, decl)
//  4. transform_defn_match(prob)
//  5. add_prem_match(proof.match(), prob, decl, self)
//  6. compile_match(proof_match, prob, decl, allow_witness)
func (pc *ProofChecker) SetupSchemaMatching(
	decl *ast.LabeledFormula,
	proof *ast.SchemaInstantiation,
	schema *ast.LabeledFormula,
	allowWitness bool,
	mod *module.Module,
) (*MatchProblem, map[lg.NodeKey]lg.Expr, error) {

	// Step 1: Rename schema using proof renaming
	// Python: schema = rename_goal(schema, proof.renaming())
	if proof != nil && proof.Ren != nil {
		var err error
		schema, err = RenameGoal(schema, proof.Ren)
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 2: Transform definition schema for parameter arity matching
	// Python: schema = transform_defn_schema(schema, decl)
	schema = TransformDefnSchema(schema, decl)

	// Step 3: Build match problem
	// Python: prob = match_problem(schema, decl)
	prob := buildMatchProblem(schema, decl)
	if prob == nil {
		return nil, nil, &NoMatch{Msg: "cannot build match problem from schema and goal"}
	}

	// Step 4: Transform definition match (full version from phase5)
	// Python: prob = transform_defn_match(prob)
	prob = TransformDefnMatch(prob)
	if prob == nil {
		return nil, nil, &NoMatch{Node: proof, Msg: "definition does not match the given schema"}
	}

	// Step 5: Process premise matches
	// Python: proof_match, prob = add_prem_match(proof.match(), prob, decl, self)
	var proofMatches []ast.Node
	if proof != nil {
		proofMatches = proof.Matches
	}
	proofMatches, prob = AddPremMatch(proofMatches, prob, decl, pc)

	// Step 6: Compile symbolic matches
	// Python: pmatch = compile_match(proof_match, prob, decl, allow_witness)
	pmatch := CompileMatchFull(proofMatches, prob, decl, allowWitness, mod)
	if pmatch == nil && len(proofMatches) > 0 {
		return nil, nil, &ProofError{Node: proof, Msg: "Match is inconsistent"}
	}
	if pmatch == nil {
		pmatch = make(map[lg.NodeKey]lg.Expr)
	}

	return prob, pmatch, nil
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

	schemaConc := GoalConc(schema)
	declConc := GoalConc(decl)
	if schemaConc == nil || declConc == nil {
		return nil
	}

	// Store the schema LabeledFormula reference in the problem
	prob := NewMatchProblem(nil, schemaConc, declConc, freesyms, constants)
	prob.SchemaLF = schema
	return prob
}

// ApplyMatchToProblem applies a match map to a MatchProblem, updating
// the schema, pattern, and free symbols.
// Includes capture avoidance and uses ApplyMatchGoalNode for full schema
// processing (premises + conclusion).
//
// Python: ivy_proof.py:994-999 (apply_match_to_problem)
func ApplyMatchToProblem(match map[lg.NodeKey]lg.Expr, prob *MatchProblem) {
	if len(match) == 0 {
		return
	}

	// Avoid capture before applying — Python: avoid_capture_problem(prob, match)
	AvoidCaptureProblem(prob, match)

	// Apply match to schema — use ApplyMatchGoalNode (processes premises + conclusion)
	if prob.SchemaLF != nil {
		prob.SchemaLF = ApplyMatchGoalNode(match, prob.SchemaLF)
	}

	// Apply match to pattern — use ApplyMatchAlt for capture safety
	prob.Pat = ApplyMatchAlt(match, prob.Pat, nil)

	// Update free symbols
	prob.FreeSyms = ApplyMatchFreesymsAlt(match, prob.FreeSyms)

	// Remove matched symbols from revmap
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
func GoalSubgoalsFromSchema(cfg *ast.AstConfig, schema *ast.LabeledFormula, goal *ast.LabeledFormula) []*ast.LabeledFormula {
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
		sub := GoalSubst(goal, pg, goal.GetLineno())
		subgoals = append(subgoals, sub)
	}
	return subgoals
}

// GoalFreeVars returns the free variables of a goal's conclusion.
func GoalFreeVars(g *ast.LabeledFormula) []*lg.Variable {
	conc := GoalConc(g)
	if conc == nil {
		return nil
	}
	return lu.FreeVariablesList(conc)
}
