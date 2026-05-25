package goivy

import (
	"strings"
	"testing"
)

func TestSetupSchemaMatchingTraceOrderMatchesPython(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("decl"), True)
		if _, _, err := pc.SetupSchemaMatchingRaw(decl, nil, nil, schema, false); err != nil {
			t.Fatalf("SetupSchemaMatchingRaw failed: %v", err)
		}
	})

	wrapperIdx := strings.Index(out, "XTRACE: proof.SetupSchemaMatching ENTER schemaLabel=schema declLabel=decl allowWitness=False\n")
	if wrapperIdx < 0 {
		t.Fatalf("missing SetupSchemaMatching wrapper trace:\n%s", out)
	}
	rawIdx := strings.Index(out, "XTRACE: proof.SetupSchemaMatchingRaw ENTER schemaLabel=schema declLabel=decl nmatches=0 allowWitness=False\n")
	if rawIdx < 0 {
		t.Fatalf("missing SetupSchemaMatchingRaw trace:\n%s", out)
	}
	if rawIdx < wrapperIdx {
		t.Fatalf("raw trace preceded wrapper trace; want Python order:\n%s", out)
	}
}

func TestSetupMatchingSuccessTraceOmitsNilErrLikePython(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("decl"), True)
		pc.Schemata.Set("schema", schema)
		proof := cfg.NewSchemaInstantiation(cfg.NewAtom("schema"), nil)
		if _, _, err := pc.SetupMatching(decl, proof, mod); err != nil {
			t.Fatalf("SetupMatching failed: %v", err)
		}
	})

	if !strings.Contains(out, "XTRACE: proof.SetupMatching EXIT npmatch=0\n") {
		t.Fatalf("missing Python-shaped SetupMatching success trace:\n%s", out)
	}
	if strings.Contains(out, "XTRACE: proof.SetupMatching EXIT err=<nil>") {
		t.Fatalf("SetupMatching success trace included Go-only nil err:\n%s", out)
	}
}

func TestMatchSchemaAppliesEmptyCompiledMatchBeforeFOMatch(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("introE"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("goal"), True)
		pc.Schemata.Set("introE", schema)
		proof := cfg.NewSchemaInstantiation(cfg.NewAtom("introE"), nil)
		if _, err := pc.MatchSchema(decl, proof); err != nil {
			t.Fatalf("MatchSchema failed: %v", err)
		}
	})

	setupIdx := strings.Index(out, "XTRACE: proof.SetupMatching EXIT npmatch=0\n")
	if setupIdx < 0 {
		t.Fatalf("missing SetupMatching exit trace:\n%s", out)
	}
	afterSetup := out[setupIdx:]
	goalFreeRel := strings.Index(afterSetup, "XTRACE: proof.GoalFree ENTER label=introE\n")
	if goalFreeRel < 0 {
		t.Fatalf("empty compiled match did not run Python's capture setup path:\n%s", out)
	}
	foRel := strings.Index(afterSetup, "XTRACE: proof.FOMatch ENTER")
	if foRel < 0 {
		t.Fatalf("missing FOMatch trace:\n%s", out)
	}
	if foRel < goalFreeRel {
		t.Fatalf("FOMatch preceded empty-match capture setup; want Python order:\n%s", out)
	}
}

func TestApplyMatchGoalNodeDoesNotTraceRecursiveWrapper(t *testing.T) {
	cfg := NewAstConfig()
	prem := cfg.NewLabeledFormula(cfg.NewAtom("prem"), True)
	schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), cfg.NewSchemaBody(prem, True))

	out := captureActionUpdateStdout(t, func() {
		ApplyMatchGoalNode(cfg, map[NodeKey]Expr{}, schema)
	})

	if !strings.Contains(out, "XTRACE: proof.ApplyMatchGoalNode ENTER label=schema nmatch=0\n") {
		t.Fatalf("missing top-level ApplyMatchGoalNode trace:\n%s", out)
	}
	if strings.Contains(out, "XTRACE: proof.ApplyMatchGoalNode ENTER label=prem nmatch=0\n") {
		t.Fatalf("recursive ApplyMatchGoalNode wrapper trace should be suppressed like Python:\n%s", out)
	}
	if !strings.Contains(out, "XTRACE: proof.CloneGoalPreserveID ENTER label=prem") {
		t.Fatalf("recursive premise was not still transformed and cloned:\n%s", out)
	}
}

func TestApplyMatchToProblemRunsEmptyAvoidCaptureRename(t *testing.T) {
	cfg := NewAstConfig()
	schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
	prob := NewMatchProblem(nil, True, True, map[NodeKey]Expr{}, map[NodeKey]Expr{})
	prob.SchemaLF = schema

	out := captureActionUpdateStdout(t, func() {
		ApplyMatchToProblem(cfg, map[NodeKey]Expr{}, prob)
	})

	if got := strings.Count(out, "XTRACE: proof.ApplyMatchGoalNode ENTER label=schema nmatch=0\n"); got != 2 {
		t.Fatalf("empty apply_match_to_problem should run rename_problem and main schema apply; got %d:\n%s", got, out)
	}
}

func TestApplyMatchGoalNodeNonAltUsesPythonWrapperTrace(t *testing.T) {
	cfg := NewAstConfig()
	schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)

	out := captureActionUpdateStdout(t, func() {
		ApplyMatchGoalNodeNonAlt(cfg, map[NodeKey]Expr{}, schema)
	})

	if !strings.Contains(out, "XTRACE: proof.ApplyMatchGoalNode ENTER label=schema nmatch=0\n") {
		t.Fatalf("missing Python wrapper trace for non-alt goal match:\n%s", out)
	}
	if strings.Contains(out, "XTRACE: proof.ApplyMatchGoalNodeNonAlt ENTER") {
		t.Fatalf("non-alt helper used Go-only wrapper trace name:\n%s", out)
	}
	if !strings.Contains(out, "XTRACE: proof.ApplyMatch ENTER nmatch=0") {
		t.Fatalf("non-alt goal match did not use ApplyMatch internally:\n%s", out)
	}
}

func TestMatchSchemaUsesGoalSubgoalsWrapper(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	out := captureActionUpdateStdout(t, func() {
		schema := cfg.NewLabeledFormula(cfg.NewAtom("schema"), True)
		decl := cfg.NewLabeledFormula(cfg.NewAtom("goal"), True)
		pc.Schemata.Set("schema", schema)
		proof := cfg.NewSchemaInstantiation(cfg.NewAtom("schema"), nil)
		if _, err := pc.MatchSchema(decl, proof); err != nil {
			t.Fatalf("MatchSchema failed: %v", err)
		}
	})

	subgoalsIdx := strings.Index(out, "XTRACE: proof.GoalSubgoals ENTER schemaLabel=schema goalLabel=goal\n")
	if subgoalsIdx < 0 {
		t.Fatalf("MatchSchema did not use Python-shaped GoalSubgoals helper:\n%s", out)
	}
	substIdx := strings.Index(out[subgoalsIdx:], "XTRACE: proof.GoalSubst ENTER g1Label=goal")
	if substIdx < 0 {
		t.Fatalf("GoalSubgoals did not perform the expected internal GoalSubst:\n%s", out)
	}
}
