package proof

import (
	"github.com/glycerine/goivy/ast"
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

	// Normalize definitions
	for _, d := range definitions {
		norm := NormalizeGoal(d)
		// Use the label name as key (Python uses d.formula.defines().name)
		name := d.LabelName()
		pc.Definitions[name] = norm
	}

	// Normalize schemata
	if schemata != nil {
		for name, s := range schemata {
			pc.Schemata[name] = NormalizeGoal(s)
		}
	}

	// Mark stale symbols
	// TODO: collect used symbols from axioms, definitions, and schemata
	// once lu.UsedSymbolsAST is ported. For now, stale tracking is a stub.

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
// Returns nil and an error if the proof fails.
//
// TODO: This is a stub. The full implementation requires porting all
// tactic types from ivy_ast (SchemaInstantiation, LetTactic,
// ComposeTactics, AssumeTactic, etc.).
func (pc *ProofChecker) ApplyProof(goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, nil
	}
	// TODO: dispatch on proof type (SchemaInstantiation, LetTactic, etc.)
	return nil, &ProofError{Msg: "ApplyProof not yet fully implemented"}
}

// MatchSchema attempts to match a goal to a schema.
//
// TODO: Stub -- depends on match_problem, setup_matching, etc.
func (pc *ProofChecker) MatchSchema(goal *ast.LabeledFormula, schemaName string) ([]*ast.LabeledFormula, error) {
	schema, err := pc.LookupSchema(schemaName, goal)
	if err != nil {
		return nil, err
	}
	_ = schema
	return nil, &NoMatch{Msg: "MatchSchema not yet fully implemented"}
}

// InstSchema instantiates a schema against a goal using the given match.
//
// TODO: Stub -- requires full apply_match_goal implementation.
func InstSchema(checker *ProofChecker, schema, goal *ast.LabeledFormula, match map[string]string) ([]*ast.LabeledFormula, error) {
	return nil, &ProofError{Msg: "InstSchema not yet fully implemented"}
}

// CheckSchema checks whether a goal matches a schema.
//
// TODO: Stub -- requires full matching pipeline.
func CheckSchema(checker *ProofChecker, goal, schema *ast.LabeledFormula) ([]*ast.LabeledFormula, error) {
	return nil, &ProofError{Msg: "CheckSchema not yet fully implemented"}
}
