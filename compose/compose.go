// Package compose implements the compose tactic for liveness proofs.
// This is a port of Python's ivy_compose.py.
//
// The compose tactic supports compositional liveness reasoning by
// decomposing a liveness proof into work items with ranking functions.
// Each work item must define:
//   - work_created: when new work is created
//   - work_needed: when work still needs to be done
//   - work_progress: what constitutes progress
//   - work_invar: invariant during work
//   - work_helpful: when work is helpful
package compose

import (
	"fmt"

	lg "github.com/glycerine/goivy/logic"
	mod "github.com/glycerine/goivy/module"
)

// Debug controls compose tactic debug output.
var Debug bool

// RankingDef holds the definitions for a single ranking function (work item).
type RankingDef struct {
	Suffix       string
	WorkCreated  lg.Node
	WorkNeeded   lg.Node
	WorkProgress lg.Node
	WorkInvar    lg.Node
	WorkHelpful  lg.Node
	WorkStart    lg.Node // optional trigger
	WorkWitness  lg.Node // optional witness
}

// ComposeTactic is the main entry point for the "ranking" proof tactic.
// It decomposes a liveness proof into work items with ranking functions.
//
// NOTE: This is a skeleton implementation. The full implementation requires:
// - proof infrastructure (goal_vocab, goal_conc, goal_prems)
// - compiler integration
// - temporal logic normalization
func ComposeTactic(m *mod.Module, goals []interface{}, proof interface{}) error {
	return composeTacticInt(m, goals, proof, "ranking")
}

func composeTacticInt(m *mod.Module, goals []interface{}, proof interface{}, tacticName string) error {
	if len(goals) == 0 {
		return fmt.Errorf("compose: no proof goals")
	}

	// The full compose tactic involves:
	// 1. Extract temporal formula from proof goal
	// 2. Parse work item definitions from tactic declarations
	// 3. For each work item:
	//    a. Validate that all required predicates are defined
	//    b. Infer work_start if not provided
	// 4. Create ranking function from work items
	// 5. Generate subgoals for each work item
	// 6. Add invariant strengthening from work_helpful

	return fmt.Errorf("compose: not yet fully implemented (requires proof infrastructure)")
}

// CreateRankingDefn creates a ranking function definition from work item predicates.
func CreateRankingDefn(rd *RankingDef) lg.Node {
	// The ranking function is: not all done yet AND work is needed
	// With decreasing measure from the work items
	if rd.WorkCreated == nil || rd.WorkNeeded == nil {
		return nil
	}
	// ranking = work_needed & ~work_created
	return &lg.And{Terms: []lg.Node{
		rd.WorkNeeded,
		&lg.Not{Body: rd.WorkCreated},
	}}
}

// ValidateRankingDef checks that a ranking definition has all required fields.
func ValidateRankingDef(rd *RankingDef) error {
	if rd.WorkCreated == nil {
		return fmt.Errorf("tactic requires a definition of work_created%s", rd.Suffix)
	}
	if rd.WorkNeeded == nil {
		return fmt.Errorf("tactic requires a definition of work_needed%s", rd.Suffix)
	}
	if rd.WorkProgress == nil {
		return fmt.Errorf("tactic requires a definition of work_progress%s", rd.Suffix)
	}
	if rd.WorkHelpful == nil {
		return fmt.Errorf("tactic requires a definition of work_helpful%s", rd.Suffix)
	}
	return nil
}
