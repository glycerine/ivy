package goivy

import (
	"os"
	"strings"
	"testing"
)

func TestDerivedInvariants(t *testing.T) {
	const fixture = "test_vectors/derived_invariants.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, errorText string
	}{
		{"implied", string(source), ""},
		{"not implied", strings.Replace(string(source), "x & y with lemma", "x & ~y with lemma", 1), "failed checks: 1"},
		{"normal depends on derived", strings.Replace(string(source), "invariant [extra] y", "invariant [extra] y with lemma", 1), "normal invariant has a zero-delay dependency on derived invariant model.lemma"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := StartSourceWithConfig(fixture, tc.source, NewConfig())
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
