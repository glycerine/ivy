package goivy

import "testing"

func TestTypeNameConcreteVariableNamesAfterUniPackageMerge(t *testing.T) {
	cfg := NewAstConfig()
	astVar := cfg.NewVariable("X", "node")
	if got := TypeName(astVar); got != "Variable" {
		t.Fatalf("TypeName(Variable) = %q, want Variable", got)
	}

	logicVar, err := NewVariable("X", TopS)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	if got := TypeName(logicVar); got != "Variable" {
		t.Fatalf("TypeName(Variable) = %q, want Variable", got)
	}
}
