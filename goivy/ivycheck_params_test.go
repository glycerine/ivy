package goivy

import "testing"

func TestParseIvyCheckParamsRejectsMultipleEqualsLikePython(t *testing.T) {
	_, _, err := ParseIvyCheckParams([]string{"diagnose=true=false", "foo.ivy"})
	if err == nil {
		t.Fatal("ParseIvyCheckParams accepted parameter with multiple '=' characters; Python ivy_init.read_params rejects it")
	}
}

func TestReadParamsRejectsMultipleEqualsLikePython(t *testing.T) {
	reg := NewParameterRegistry()
	NewParameterOn(reg, "diagnose", "")
	_, err := ReadParams([]string{"diagnose=true=false", "foo.ivy"}, reg)
	if err == nil {
		t.Fatal("ReadParams accepted parameter with multiple '=' characters; Python ivy_init.read_params rejects it")
	}
}

func TestApplyIvyCheckParamsMarksEmptyPrioritizePresentLikePython(t *testing.T) {
	params, _, err := ParseIvyCheckParams([]string{"prioritize=", "foo.ivy"})
	if err != nil {
		t.Fatalf("ParseIvyCheckParams returned error: %v", err)
	}

	cfg := NewConfig()
	if err := ApplyIvyCheckParams(cfg, params); err != nil {
		t.Fatalf("ApplyIvyCheckParams returned error: %v", err)
	}
	if !cfg.PriorityActionsSet {
		t.Fatal("expected prioritize= to be recorded as explicitly present")
	}
	got := GetPrioritizedActions(cfg)
	if len(got) != 1 || got[0] != "ext:" {
		t.Fatalf("expected prioritize= to produce [ext:] like Python, got %v", got)
	}
}

func TestApplyIvyCheckParamsRejectsBadBooleanLikePython(t *testing.T) {
	cfg := NewConfig()
	err := ApplyIvyCheckParams(cfg, map[string]string{"diagnose": "yes"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted diagnose=yes; Python BooleanParameter only accepts true/false")
	}

	cfg = NewConfig()
	err = ApplyIvyCheckParams(cfg, map[string]string{"diagnose": "1"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted diagnose=1; Python BooleanParameter only accepts true/false")
	}
}

func TestApplyIvyCheckParamsParserParameterReturnsErrorLikePython(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ApplyIvyCheckParams panicked for parser=lalr; Python reports a parameter error: %v", r)
		}
	}()

	err := ApplyIvyCheckParams(NewConfig(), map[string]string{"parser": "lalr"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted parser=lalr; Python reports an undefined parameter")
	}
}

func TestApplyIvyCheckParamsAssertParameterNormalizesLocationLikePython(t *testing.T) {
	cfg := NewConfig()
	err := ApplyIvyCheckParams(cfg, map[string]string{"assert": "foo:17"})
	if err != nil {
		t.Fatalf("ApplyIvyCheckParams returned error: %v", err)
	}
	if cfg.CheckLineno != "foo.ivy:17" {
		t.Fatalf("expected assert=foo:17 to become foo.ivy:17 like Python, got %q", cfg.CheckLineno)
	}
}

func TestApplyIvyCheckParamsRejectsBadAssertLocationLikePython(t *testing.T) {
	for _, val := range []string{"foo", "foo:bar", "foo:17:extra"} {
		cfg := NewConfig()
		err := ApplyIvyCheckParams(cfg, map[string]string{"assert": val})
		if err == nil {
			t.Fatalf("ApplyIvyCheckParams accepted assert=%s; Python rejects malformed checked_assert locations", val)
		}
	}
}

func TestApplyIvyCheckParamsCheckedAssertParameterReturnsErrorLikePython(t *testing.T) {
	err := ApplyIvyCheckParams(NewConfig(), map[string]string{"checked_assert": "foo:17"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted checked_assert=foo:17; Python registers this checker parameter as assert")
	}
}

func TestApplyIvyCheckParamsImportedPythonParameterSurface(t *testing.T) {
	cfg := NewConfig()
	params := map[string]string{
		"seed":                    "7",
		"incremental":             "false",
		"show_vcs":                "true",
		"detailed":                "false",
		"use_numerals":            "false",
		"new_ui":                  "true",
		"catch":                   "false",
		"ui":                      "art",
		"debug":                   "true",
		"mode":                    "induction",
		"fullqi":                  "true",
		"show_compiled":           "true",
		"coi":                     "false",
		"filter_symbols":          "false",
		"create_imports":          "true",
		"enforce_axioms":          "true",
		"interference":            "false",
		"pedantic":                "true",
		"prefer_impls":            "true",
		"keep_destructors":        "true",
		"isolate_mode":            "compile",
		"compile_with_invariants": "true",
		"assume_invariants":       "false",
		"ext":                     "external",
		"mutax":                   "true",
		"l2s_debug":               "false",
		"ranking_debug":           "true",
		"abs_init":                "true",
	}
	if err := ApplyIvyCheckParams(cfg, params); err != nil {
		t.Fatalf("ApplyIvyCheckParams returned error for Python ivy_check imported parameter: %v", err)
	}

	if cfg.SolverOpts.Seed != 7 {
		t.Fatalf("seed = %d, want 7", cfg.SolverOpts.Seed)
	}
	if !cfg.SolverOpts.SeedSet {
		t.Fatal("seed parameter did not mark SolverOpts.SeedSet")
	}
	if cfg.SolverOpts.Incremental {
		t.Fatal("incremental = true, want false")
	}
	if !cfg.SolverOpts.ShowVCs {
		t.Fatal("show_vcs = false, want true")
	}
	if cfg.TraceDetailed {
		t.Fatal("detailed = true, want false")
	}
	if cfg.IuCfg.UseNumerals {
		t.Fatal("use_numerals = true, want false")
	}
	if !cfg.IuCfg.UseNewUI {
		t.Fatal("new_ui = false, want true")
	}
	if cfg.IuCfg.Catch {
		t.Fatal("catch = true, want false")
	}
	if cfg.IuCfg.DefaultUI != "art" {
		t.Fatalf("ui = %q, want art", cfg.IuCfg.DefaultUI)
	}
	if !cfg.IuCfg.EnableDebug {
		t.Fatal("debug = false, want true")
	}
	if cfg.IuCfg.DefaultMode != "induction" {
		t.Fatalf("mode = %q, want induction", cfg.IuCfg.DefaultMode)
	}
	if !cfg.FullQI {
		t.Fatal("fullqi = false, want true")
	}
	iso := cfg.IsolateCfg
	if !iso.ShowCompiled || iso.ConeOfInfluence || iso.FilterSymbols || !iso.CreateImports ||
		!iso.EnforceAxioms || iso.DoCheckInterference || !iso.Pedantic ||
		!iso.PreferImpls || !iso.KeepDestructors || !iso.CompileWithInvariants ||
		iso.AssumeInvariants {
		t.Fatalf("isolate imported parameters not applied: %+v", *iso)
	}
	if iso.IsolateMode != "compile" {
		t.Fatalf("isolate_mode = %q, want compile", iso.IsolateMode)
	}
	if cfg.ExtAction != "external" || iso.ExtAction != "external" {
		t.Fatalf("ext did not set both Config and IsolateConfig fields: cfg=%q iso=%q", cfg.ExtAction, iso.ExtAction)
	}
	if !cfg.OptMutax {
		t.Fatal("mutax = false, want true")
	}
	if cfg.L2SDebug {
		t.Fatal("l2s_debug = true, want false")
	}
	if !cfg.RankingDebug {
		t.Fatal("ranking_debug = false, want true")
	}
	if !cfg.OptionAbsInit {
		t.Fatal("abs_init = false, want true")
	}
}

func TestApplyIvyCheckParamsRejectsBadSeedLikePython(t *testing.T) {
	err := ApplyIvyCheckParams(NewConfig(), map[string]string{"seed": "abc"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted seed=abc; Python Parameter(process=int) rejects it")
	}
}

func TestApplyIvyCheckParamsRejectsBadModeLikePython(t *testing.T) {
	err := ApplyIvyCheckParams(NewConfig(), map[string]string{"mode": "sideways"})
	if err == nil {
		t.Fatal("ApplyIvyCheckParams accepted mode=sideways; Python ivy_ui.default_mode rejects it")
	}
}

func TestApplyIvyCheckParamsValidatesCompleteLogicsLikePython(t *testing.T) {
	cfg := NewConfig()
	if err := ApplyIvyCheckParams(cfg, map[string]string{"complete": "epr,qf,fo"}); err != nil {
		t.Fatalf("ApplyIvyCheckParams rejected valid complete logics: %v", err)
	}
	if cfg.CompleteLogic != "epr,qf,fo" {
		t.Fatalf("complete logic = %q, want epr,qf,fo", cfg.CompleteLogic)
	}

	for _, val := range []string{"", "bogus", "epr,bogus"} {
		err := ApplyIvyCheckParams(NewConfig(), map[string]string{"complete": val})
		if err == nil {
			t.Fatalf("ApplyIvyCheckParams accepted complete=%q; Python validates each logic name", val)
		}
	}
}
