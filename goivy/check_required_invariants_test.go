package goivy

import (
	"os"
	"strings"
	"testing"
)

func TestRequiredInvariantProvider(t *testing.T) {
	const fixture = "test_vectors/required_invariants.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	mod := New()
	if err := SourceString(fixture, string(source), mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		isolate string
		assumed bool
		checked bool
	}{
		{"consumer", true, false},
		{"provider", false, true},
		{"this", true, false},
	} {
		t.Run(tc.isolate, func(t *testing.T) {
			selected := mod.Copy()
			if err := CreateIsolate(tc.isolate, selected); err != nil {
				t.Fatal(err)
			}
			containsReady := func(invariants []*LabeledFormula) bool {
				for _, lf := range invariants {
					if lfLabelName(lf) == "consumer.ready" {
						return true
					}
				}
				return false
			}
			if got := containsReady(selected.AssumedInvs); got != tc.assumed {
				t.Errorf("assumes consumer.ready = %t, want %t", got, tc.assumed)
			}
			if got := containsReady(selected.LabeledConjs); got != tc.checked {
				t.Errorf("checks consumer.ready = %t, want %t", got, tc.checked)
			}
		})
	}
}

func TestRequiredInvariantDiagnostics(t *testing.T) {
	const fixture = "test_vectors/required_invariants.ivy"
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, tc := range []struct{ name, source, want string }{
		{"verified provider", source, ""},
		{"missing provider", strings.Replace(source, "provide consumer.ready", "", 1), "invariant consumer.ready is required but not provided"},
		{"unknown invariant", strings.Replace(source, "provide consumer.ready", "provide consumer.unknown_label", 1), "invariant consumer.unknown_label is provided but not defined"},
		{"ordinary invariant", strings.Replace(source, "provide consumer.ready", "provide consumer.main", 1), "invariant consumer.main is provided but not required"},
		{"duplicate provider", source + "\nprovide consumer.ready\n", "multiple providers for invariant consumer.ready"},
		{"provider fails", strings.Replace(source, "invariant [ready] true", "invariant [ready] false", 1), "failed checks: 1"},
		{"root assumption", "#lang ivy1.8\nrequire { invariant [external] false }\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := StartSourceWithConfig(fixture, tc.source, NewConfig())
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestInvariantExportsIncludeOnlyRelevantEffects(t *testing.T) {
	const source = `#lang ivy1.8
isolate component = {
    relation state_bit
    export action pure = { assume true }
    export action setter = { state_bit := true }
    export action checks = { assert true }
}
`
	mod := New()
	if err := SourceString("pure_exports.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	iso := mod.Isolates.Get("component")
	if iso == nil {
		t.Fatal("component isolate is missing")
	}
	obligations := GetIsolateExports(mod, mod.CallGraph(), iso)
	if obligations.Get("component.pure") || !obligations.Get("component.setter") || !obligations.Get("component.checks") {
		t.Fatalf("invariant obligations = %v, want only the writing and asserting actions", obligations.AsMap())
	}
	if err := CreateIsolate("component", mod); err != nil {
		t.Fatal(err)
	}
	if !mod.PublicActions.Get("ext:component.pure") {
		t.Fatal("the pure action is inaccessible to callers")
	}
}

func TestAssumedInvariantRetainsItsWireDefinition(t *testing.T) {
	const source = `#lang ivy1.8
isolate consumer = {
    relation driving
    wire helper : bool = driving
    require { invariant [external] helper }
    relation flag
    after init { flag := true }
    export action step = { flag := driving }
    invariant [safe] flag
}
isolate provider = { provide consumer.external }
`
	cfg := NewConfig()
	cfg.Isolate = "consumer"
	if err := StartSourceWithConfig("assumed_wire.ivy", source, cfg); err != nil {
		t.Fatal(err)
	}
}
