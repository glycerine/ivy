package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// SetupMatching creates a MatchProblem for matching a schema to a declaration.
// This is the first stage of the proof matching pipeline.
//
// Python: ivy_proof.py:324-340
func (pc *ProofChecker) SetupMatching(decl *ast.LabeledFormula, schemaName string) (*MatchProblem, map[lg.NodeKey]lg.Node, error) {
	schema, err := pc.LookupSchema(schemaName, decl)
	if err != nil {
		return nil, nil, err
	}
	return pc.SetupSchemaMatching(decl, schema)
}

// SetupSchemaMatching builds a MatchProblem from a schema and goal.
//
// Python: ivy_proof.py:329-340
func (pc *ProofChecker) SetupSchemaMatching(decl *ast.LabeledFormula, schema *ast.LabeledFormula) (*MatchProblem, map[lg.NodeKey]lg.Node, error) {
	// Build match problem
	prob := buildMatchProblem(schema, decl)
	if prob == nil {
		return nil, nil, &NoMatch{Msg: "cannot build match problem from schema and goal"}
	}

	// Transform definition schemas if needed
	prob = transformDefnMatch(prob)
	if prob == nil {
		return nil, nil, &NoMatch{Msg: "definition does not match the given schema"}
	}

	// The compiled match starts empty — in the full Python version,
	// proof.match() from the AST would provide initial bindings.
	pmatch := make(map[lg.NodeKey]lg.Node)
	return prob, pmatch, nil
}

// buildMatchProblem creates a MatchProblem from a schema and declaration.
//
// Python: ivy_proof.py:771-776
func buildMatchProblem(schema, decl *ast.LabeledFormula) *MatchProblem {
	vocab := GoalVocab(schema)

	freesyms := make(map[lg.NodeKey]lg.Node)
	for _, sym := range vocab.Symbols {
		freesyms[lg.Key(sym)] = sym
	}
	for _, s := range vocab.Sorts {
		freesyms[lg.Key(s)] = s
	}
	for _, v := range vocab.Variables {
		freesyms[lg.Key(v)] = v
	}

	constants := make(map[lg.NodeKey]lg.Node)
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

// transformDefnMatch transforms a definition matching problem.
// If both the schema and goal conclusions are definitions, ensures they can be matched.
//
// Python: ivy_proof.py:790-834
func transformDefnMatch(prob *MatchProblem) *MatchProblem {
	// Check if both are definitions — if not, no transformation needed
	_, patDef := prob.Pat.(*il.Definition)
	_, instDef := prob.Inst.(*il.Definition)
	if !patDef || !instDef {
		return prob
	}

	// For definitions: match using equality representation
	patD := prob.Pat.(*il.Definition)
	instD := prob.Inst.(*il.Definition)

	// Convert definitions to equality for matching
	prob.Pat = &lg.Eq{T1: patD.Lhs, T2: patD.Rhs}
	prob.Inst = &lg.Eq{T1: instD.Lhs, T2: instD.Rhs}

	return prob
}

// ApplyMatchToProblem applies a match map to a MatchProblem, updating
// the schema, pattern, and free symbols.
//
// Python: ivy_proof.py:994-999
func ApplyMatchToProblem(match map[lg.NodeKey]lg.Node, prob *MatchProblem) {
	if len(match) == 0 {
		return
	}

	// Apply match to schema
	if prob.SchemaLF != nil {
		newConc := ApplyMatch(match, GoalConc(prob.SchemaLF))
		if newConc != nil {
			newPrems := GoalPrems(prob.SchemaLF)
			prob.SchemaLF = CloneGoal(prob.SchemaLF, newPrems, newConc)
		}
	}

	// Apply match to pattern
	prob.Pat = ApplyMatch(match, prob.Pat)

	// Update free symbols
	prob.FreeSyms = ApplyMatchFreesyms(match, prob.FreeSyms)

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
func GoalSubgoalsFromSchema(schema *ast.LabeledFormula, goal *ast.LabeledFormula) []*ast.LabeledFormula {
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
