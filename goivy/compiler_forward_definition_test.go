package goivy

import (
	"os"
	"strings"
	"testing"
)

func TestForwardDefinitions(t *testing.T) {
	const fixture = "test_vectors/forward_definitions.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := StartSourceWithConfig(fixture, string(source), NewConfig()); err != nil {
		t.Fatal(err)
	}
}

func TestDefinitionUsesLaterDeclaredWireSignature(t *testing.T) {
	const source = `#lang ivy1.8
type index
interpret index -> bv[1]
wire zero_case : bool = true
wire one_case : bool = false
definition selected(X) = zero_case if X = 0 else one_case
wire selected(X:index) : bool
property [points] selected(0) & ~selected(1)
`
	if err := StartSourceWithConfig("forward_wire.ivy", source, NewConfig()); err != nil {
		t.Fatal(err)
	}
}

func TestForwardDefinitionErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "undeclared symbol",
			source: "#lang ivy1.7\nrelation output\ndefinition output = missing\n",
			want:   []string{"line 3", "unknown symbol: missing"},
		},
		{
			name: "cycle",
			source: "#lang ivy1.7\nrelation left_pred\ndefinition left_pred = right_pred\n" +
				"relation right_pred\ndefinition right_pred = left_pred\n",
			want: []string{"dependency cycle", "left_pred", "right_pred"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StartSourceWithConfig("forward.ivy", tt.source, NewConfig())
			if err == nil {
				t.Fatal("expected a compilation error")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}

func TestForwardDefinitionKeepsAttachedProof(t *testing.T) {
	const source = `
relation output
definition output = later
proof { showgoals }
relation later
`
	parsed, err := Parse(source, Version{1, 7})
	if err != nil {
		t.Fatal(err)
	}
	compiler := newTestCompiler()
	if err := NewDomainSetup(compiler).ProcessDecls(parsed.Decls); err != nil {
		t.Fatal(err)
	}
	mod := compiler.Module
	if len(mod.Proofs) != 1 || len(mod.LabeledProps) != 1 {
		t.Fatalf("got %d proofs and %d properties, want one of each", len(mod.Proofs), len(mod.LabeledProps))
	}
	if mod.Proofs[0].Formula != mod.LabeledProps[0] {
		t.Fatal("proof is detached from its definition")
	}
	if _, ok := mod.Proofs[0].Formula.Formula.(*LogicDefinition); !ok {
		t.Fatalf("proof points to %T, want a compiled definition", mod.Proofs[0].Formula.Formula)
	}
}
