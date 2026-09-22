package ivy2golang

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func FuzzGoNameDoesNotPanic(f *testing.F) {
	for _, seed := range []string{"", "x", "ext:step", "type", "0bad", "a.b[c]", "__fml:x"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := goName(s)
		if got == "" {
			t.Fatalf("goName(%q) returned empty string", s)
		}
		if strings.HasPrefix(s, "\"") {
			return
		}
		if got[0] >= '0' && got[0] <= '9' {
			t.Fatalf("goName(%q) starts with a digit: %q", s, got)
		}
		if goKeywords[got] {
			t.Fatalf("goName(%q) returned keyword %q", s, got)
		}
	})
}

func FuzzParseArgsDoesNotPanic(f *testing.F) {
	for _, seed := range []string{"target=test", "classname=X", "bad", "=bad", "trace=true"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _ = ParseArgs([]string{s, "x.ivy"})
		_, _, _ = ParseArgs([]string{s})
	})
}

func FuzzMergeParamsDriverSurfaceInvariants(f *testing.F) {
	for _, seed := range []struct {
		key   string
		value string
	}{
		{"target", "test"},
		{"target", "class"},
		{"compiler", "g++"},
		{"compiler", "cl"},
		{"build", "true"},
		{"trace", "yes"},
		{"stdafx", "1"},
		{"classname", "My.Ivy"},
		{"main", "ivy-main"},
		{"outdir", "out"},
		{"test_iters", "7"},
		{"test_runs", "3"},
		{"isolate", "iso"},
		{"unknown", "value"},
	} {
		f.Add(seed.key, seed.value)
	}
	f.Fuzz(func(t *testing.T, key, value string) {
		cfg, ivyParams, err := mergeParams(map[string]string{key: value}, Config{})
		if err != nil {
			return
		}
		switch cfg.RequestedTarget {
		case "impl", "repl", "test", "gen", "class":
		default:
			t.Fatalf("accepted target normalized to unsupported RequestedTarget %q from %q=%q", cfg.RequestedTarget, key, value)
		}
		if cfg.RequestedTarget == "class" {
			if cfg.Target != "repl" || cfg.EmitMain {
				t.Fatalf("class target normalized to target=%q EmitMain=%v", cfg.Target, cfg.EmitMain)
			}
		} else if cfg.Target != cfg.RequestedTarget || !cfg.EmitMain {
			t.Fatalf("target %q normalized to target=%q EmitMain=%v", cfg.RequestedTarget, cfg.Target, cfg.EmitMain)
		}
		switch cfg.Compiler {
		case "default", "g++", "cl":
		default:
			t.Fatalf("accepted compiler normalized to unsupported value %q from %q=%q", cfg.Compiler, key, value)
		}
		if cfg.MainName == "" || cfg.TestIters == "" || cfg.TestRuns == "" {
			t.Fatalf("accepted config has empty main/test defaults: %+v", cfg)
		}
		if key == "isolate" {
			if ivyParams["isolate"] != value {
				t.Fatalf("isolate param not preserved: %q -> %#v", value, ivyParams)
			}
		} else if len(ivyParams) != 0 {
			t.Fatalf("non-isolate param leaked Ivy params: %q=%q -> %#v", key, value, ivyParams)
		}
	})
}

func FuzzGenerateEmptyModuleDoesNotPanic(f *testing.F) {
	for _, seed := range []struct {
		target    string
		className string
		mainName  string
	}{
		{target: "", className: "", mainName: ""},
		{target: "test", className: "FuzzTest", mainName: "main"},
		{target: "gen", className: "FuzzGen", mainName: "ivy_main"},
		{target: "class", className: "FuzzClass", mainName: ""},
		{target: "bad", className: "type", mainName: "func"},
	} {
		f.Add(seed.target, seed.className, seed.mainName)
	}
	f.Fuzz(func(t *testing.T, target, className, mainName string) {
		mod := goivy.New()
		mod.Name = "fuzz"
		out, err := Generate(mod, Config{Target: target, ClassName: className, MainName: mainName})
		if err == nil && (out == nil || out.Source == "") {
			t.Fatalf("Generate(%q, %q, %q) returned empty output without error", target, className, mainName)
		}
	})
}
