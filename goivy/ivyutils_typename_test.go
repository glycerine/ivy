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
	if got := TypeName(logicVar); got != "Var" {
		t.Fatalf("TypeName(LogicVariable) = %q, want Var", got)
	}

	atom := cfg.NewAtom("byte.random")
	if got := TypeName(newNativeAtomExpr(atom)); got != "Atom" {
		t.Fatalf("TypeName(nativeAtomExpr) = %q, want Atom", got)
	}
}
