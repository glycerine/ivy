package goivy

import (
	"strings"
	"testing"
)

func l2sFiniteSortTestVar(t *testing.T, name string, sort Sort) *LogicVariable {
	t.Helper()
	v, err := NewVariable(name, sort)
	if err != nil {
		t.Fatalf("NewVariable(%s): %v", name, err)
	}
	return v
}

func l2sFiniteSortTestApp(t *testing.T, name string, sort Sort, args ...Expr) *Apply {
	t.Helper()
	sorts := make([]Sort, len(args)+1)
	for i, arg := range args {
		sorts[i] = arg.NodeSort()
	}
	sorts[len(sorts)-1] = sort
	fn := NewConst(name, &LogicFunctionSort{Sorts: sorts})
	app, err := NewApply(fn, args...)
	if err != nil {
		t.Fatalf("NewApply(%s): %v", name, err)
	}
	return app
}

func l2sFiniteSortTestForallDef(t *testing.T, cfg *AstConfig, label string, lhs *Apply, rhs Expr, vars ...*LogicVariable) *LabeledFormula {
	t.Helper()
	eq := &Eq{T1: lhs, T2: rhs}
	var formula Expr = eq
	if len(vars) > 0 {
		fa, err := NewForAll(vars, eq)
		if err != nil {
			t.Fatalf("NewForAll(%s): %v", label, err)
		}
		formula = fa
	}
	lf := cfg.NewLabeledFormula(cfg.NewAtom(label), formula)
	lf.IsDefinition = true
	return lf
}

func l2sFiniteSortTestBoolDef(cfg *AstConfig, label, name string, rhs Expr) *LabeledFormula {
	lf := cfg.NewLabeledFormula(cfg.NewAtom(label), &Eq{T1: NewConst(name, Boolean), T2: rhs})
	lf.IsDefinition = true
	return lf
}

func l2sFiniteSortTestGoal(cfg *AstConfig, prems ...Node) *LabeledFormula {
	elems := append([]Node{}, prems...)
	elems = append(elems, &LogicGlobally{Environ: strPtr(""), Body: True})
	return cfg.NewLabeledFormula(cfg.NewAtom("goal"), cfg.NewSchemaBody(elems...))
}

func l2sFiniteSortFindLF(t *testing.T, invars []*LabeledFormula, name string) *LabeledFormula {
	t.Helper()
	for _, inv := range invars {
		if lfName(inv) == name {
			return inv
		}
	}
	t.Fatalf("missing invariant %q", name)
	return nil
}

func l2sFiniteSortCanonHas(c Canonical, name string) bool {
	return strings.Contains(string(c), `name:`+name)
}

func TestL2SFiniteSortLookupUsesDeclaredEnumName(t *testing.T) {
	ltask := &LogicEnumeratedSort{Name: "ltask", Extension: []string{"ready_finish", "o3_finish"}}
	finiteSorts := map[string]bool{"ltask": true}

	if ltask.String() == SortName(ltask) {
		t.Fatalf("test requires enum String() and SortName() to differ, got %q", ltask.String())
	}
	if !l2sSortIsFinite(finiteSorts, ltask) {
		t.Fatal("enum sort should be finite when finiteSorts is keyed by declared SortName")
	}
}

func TestL2SAuto5EnumWorkNeededDoesNotRequireDOrA(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	ltask := &LogicEnumeratedSort{Name: "ltask", Extension: []string{"ready_finish", "o3_finish"}}
	tvar := l2sFiniteSortTestVar(t, "T", ltask)
	readyFinish := NewConst("ready_finish", ltask)
	neededBody := &Eq{T1: tvar, T2: readyFinish}

	workCreated := l2sFiniteSortTestForallDef(t, cfg, "def_created",
		l2sFiniteSortTestApp(t, "work_created", Boolean, tvar), True, tvar)
	workNeeded := l2sFiniteSortTestForallDef(t, cfg, "def_needed",
		l2sFiniteSortTestApp(t, "work_needed", Boolean, tvar), neededBody, tvar)
	workProgress := l2sFiniteSortTestBoolDef(cfg, "def_progress", "work_progress", True)
	workHelpful := l2sFiniteSortTestBoolDef(cfg, "def_helpful", "work_helpful", True)
	goal := l2sFiniteSortTestGoal(cfg, workCreated, workNeeded, workProgress, workHelpful)

	invars, _, _, err := l2sAutoInvariants(
		"l2s_auto5",
		goal,
		nil,
		"",
		&LogicGlobally{Environ: strPtr(""), Body: True},
		map[string]bool{"ltask": true},
		nil,
		mod,
		Location{Filename: "test.ivy", Line: 1},
	)
	if err != nil {
		t.Fatalf("l2sAutoInvariants: %v", err)
	}

	neededWhenStart := l2sFiniteSortFindLF(t, invars, "l2s_needed_when_start")
	if l2sFiniteSortCanonHas(neededWhenStart.Canon(), "l2s_d") {
		t.Fatalf("enum work_needed produced l2s_d guard:\n%s", neededWhenStart.Canon())
	}

	neededAreFrozen := l2sFiniteSortFindLF(t, invars, "l2s_needed_are_frozen")
	if l2sFiniteSortCanonHas(neededAreFrozen.Canon(), "l2s_a") {
		t.Fatalf("enum work_done/work_needed produced l2s_a guard:\n%s", neededAreFrozen.Canon())
	}
}

func TestRankingEnumWorkCreatedDoesNotRequireD(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	ltask := &LogicEnumeratedSort{Name: "ltask", Extension: []string{"ready_finish", "o3_finish"}}
	tvar := l2sFiniteSortTestVar(t, "T", ltask)
	readyFinish := NewConst("ready_finish", ltask)
	neededBody := &Eq{T1: tvar, T2: readyFinish}

	workCreated := l2sFiniteSortTestForallDef(t, cfg, "def_created",
		l2sFiniteSortTestApp(t, "work_created", Boolean, tvar), True, tvar)
	workNeeded := l2sFiniteSortTestForallDef(t, cfg, "def_needed",
		l2sFiniteSortTestApp(t, "work_needed", Boolean, tvar), neededBody, tvar)
	workProgress := l2sFiniteSortTestBoolDef(cfg, "def_progress", "work_progress", True)
	workInvar := l2sFiniteSortTestBoolDef(cfg, "def_invar", "work_invar", True)
	workHelpful := l2sFiniteSortTestBoolDef(cfg, "def_helpful", "work_helpful", True)
	goal := l2sFiniteSortTestGoal(cfg, workCreated, workNeeded, workProgress, workInvar, workHelpful)

	invars, _, _, _, err := rankingInvariants(
		goal,
		nil,
		"",
		&LogicGlobally{Environ: strPtr(""), Body: True},
		map[string]bool{"ltask": true},
		nil,
		mod,
		GoalPrems(goal),
	)
	if err != nil {
		t.Fatalf("rankingInvariants: %v", err)
	}

	created := l2sFiniteSortFindLF(t, invars, "l2s_created")
	if l2sFiniteSortCanonHas(created.Canon(), "l2s_d") {
		t.Fatalf("enum work_created produced l2s_d guard:\n%s", created.Canon())
	}
}
