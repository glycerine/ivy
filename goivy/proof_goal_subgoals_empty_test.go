package goivy

import "testing"

// Python goal_subgoals returns a list ([] when nothing is left to prove),
// and apply_proof tells [] (matched) from None (no match). goivy returned
// a nil slice for no subgoals, which the schema-instantiation step read as
// "goal does not match the given schema" (test/ded3.ivy).
func TestGoalSubgoalsEmptyIsNonNil(t *testing.T) {
	cfg := NewAstConfig()
	schema := mkLF(cfg.NewAtom("s"), True)
	goal := mkLF(cfg.NewAtom("g"), True)
	subs, err := GoalSubgoals(cfg, schema, goal, Location{})
	if err != nil {
		t.Fatal(err)
	}
	if subs == nil || len(subs) != 0 {
		t.Fatalf("subgoals = %#v, want an empty non-nil slice", subs)
	}
}
