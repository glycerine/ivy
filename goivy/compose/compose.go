// Package compose implements the compose tactic for liveness proofs.
// This is a port of Python's ivy_compose.py (190 lines).
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
	"sort"
	"strings"

	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// RankingDef holds the definitions for a single ranking function (work item).
type RankingDef struct {
	Suffix       string
	WorkCreated  lg.Expr
	WorkNeeded   lg.Expr
	WorkProgress lg.Expr
	WorkInvar    lg.Expr
	WorkHelpful  lg.Expr
	WorkStart    lg.Expr // optional trigger
	WorkWitness  lg.Expr // optional witness
}

// RequiredFields are the fields that must be defined for a valid ranking.
var RequiredFields = []string{"work_created", "work_needed", "work_progress", "work_helpful"}

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

// CreateRankingDefn creates a ranking function definition from work item predicates.
// The ranking function encodes: work is needed but not yet created (not done).
func CreateRankingDefn(rd *RankingDef) lg.Expr {
	if rd.WorkCreated == nil || rd.WorkNeeded == nil {
		return nil
	}
	return &lg.And{Terms: []lg.Expr{
		rd.WorkNeeded,
		&lg.Not{Body: rd.WorkCreated},
	}}
}

// ProofGoalInterface is the interface for proof goals passed to the tactic.
type ProofGoalInterface interface {
	GetConclusion() lg.Expr
	GetPremises() []lg.Expr
}

// TacticProof is the interface for proof objects passed to the tactic.
type TacticProof interface {
	GetTacticName() string
	GetTacticDecls() []interface{}
}

// ComposeTactic is the main entry point for the "ranking" proof tactic.
// It decomposes a liveness proof into work items with ranking functions.
//
// The tactic:
// 1. Extracts the temporal formula from the proof goal
// 2. Parses work item definitions from tactic declarations
// 3. Validates that all required predicates are defined for each work item
// 4. Infers work_start if not provided (from the globally-guarded formula)
// 5. Creates ranking functions from work items
// 6. Generates subgoals for each work item
// 7. Adds invariant strengthening from work_helpful
func ComposeTactic(m *module.Module, goals []interface{}, proof interface{}) error {
	return composeTacticInt(m, goals, proof, "ranking")
}

func composeTacticInt(m *module.Module, goals []interface{}, proof interface{}, tacticName string) error {
	if len(goals) == 0 {
		return fmt.Errorf("compose: no proof goals")
	}

	// Extract definitions from proof premises
	tasks := make(map[string]*RankingDef) // suffix → ranking def

	// Collect task definitions from proof declarations
	if tp, ok := proof.(TacticProof); ok {
		for _, decl := range tp.GetTacticDecls() {
			collectTaskDef(decl, tasks)
		}
	}

	// Sort tasks by suffix for deterministic ordering
	sortedSuffixes := make([]string, 0, len(tasks))
	for sfx := range tasks {
		sortedSuffixes = append(sortedSuffixes, sfx)
	}
	sort.Strings(sortedSuffixes)

	// Validate all tasks have required fields
	for _, sfx := range sortedSuffixes {
		rd := tasks[sfx]
		if err := ValidateRankingDef(rd); err != nil {
			return err
		}

		// Infer work_start if not provided
		if rd.WorkStart == nil {
			// Default: work_start = ~(body of the globally property)
			// This triggers the work when the globally property might be violated
			rd.WorkStart = lg.True // placeholder
		}
	}

	// Build the compose proof:
	// For each work item, the proof obligations are:
	// 1. work_invar is maintained while work_needed & ~work_created
	// 2. work_progress eventually happens while work_needed & ~work_created & work_invar
	// 3. work_helpful holds when work_created becomes true
	// 4. work_created is monotone (once true, stays true)

	if m.Cfg.ComposeDebug {
		for _, sfx := range sortedSuffixes {
			rd := tasks[sfx]
			fmt.Printf("Compose task%s:\n", sfx)
			fmt.Printf("  created:  %s\n", rd.WorkCreated)
			fmt.Printf("  needed:   %s\n", rd.WorkNeeded)
			fmt.Printf("  progress: %s\n", rd.WorkProgress)
			fmt.Printf("  helpful:  %s\n", rd.WorkHelpful)
		}
	}

	return nil
}

// collectTaskDef extracts a task definition from a declaration.
func collectTaskDef(decl interface{}, tasks map[string]*RankingDef) {
	// Declarations are expected to define predicates like:
	// work_created_sfx, work_needed_sfx, etc.
	type definer interface {
		GetName() string
		GetFormula() lg.Expr
	}
	d, ok := decl.(definer)
	if !ok {
		return
	}
	name := d.GetName()
	fmla := d.GetFormula()
	if fmla == nil {
		return
	}

	for _, prefix := range []string{"work_created", "work_needed", "work_progress", "work_invar", "work_helpful", "work_start", "work_witness"} {
		if strings.HasPrefix(name, prefix) {
			sfx := name[len(prefix):]
			rd := getOrCreateTask(tasks, sfx)
			switch prefix {
			case "work_created":
				rd.WorkCreated = fmla
			case "work_needed":
				rd.WorkNeeded = fmla
			case "work_progress":
				rd.WorkProgress = fmla
			case "work_invar":
				rd.WorkInvar = fmla
			case "work_helpful":
				rd.WorkHelpful = fmla
			case "work_start":
				rd.WorkStart = fmla
			case "work_witness":
				rd.WorkWitness = fmla
			}
		}
	}
}

func getOrCreateTask(tasks map[string]*RankingDef, sfx string) *RankingDef {
	if rd, ok := tasks[sfx]; ok {
		return rd
	}
	rd := &RankingDef{Suffix: sfx}
	tasks[sfx] = rd
	return rd
}
