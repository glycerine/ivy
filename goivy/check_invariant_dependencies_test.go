package goivy

import (
	"os"
	"strings"
	"testing"
)

func TestInvariantDependencies(t *testing.T) {
	const fixture = "test_vectors/invariant_dependencies.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, errorText string
	}{
		{"scoped hypotheses", string(source), ""},
		{"without hypotheses", strings.Replace(string(source), " with basis, auxiliary", "", 1), "failed checks"},
		{"missing hypothesis", strings.Replace(string(source), "basis, auxiliary", "basis, outside.auxiliary", 1), "zero-delay dependency outside.auxiliary is not visible"},
		{"circular hypotheses", strings.Replace(string(source), "invariant [basis] x", "invariant [basis] x with dependent", 1), "zero-delay invariant dependency cycle: scoped.basis, scoped.dependent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig()
			if err := ApplyIvyCheckParams(cfg, map[string]string{"check": "scoped.dependent"}); err != nil {
				t.Fatal(err)
			}
			err := StartSourceWithConfig(fixture, tc.source, cfg)
			if tc.errorText == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.errorText) {
				t.Fatalf("got %v, want an error containing %q", err, tc.errorText)
			}
		})
	}
}
