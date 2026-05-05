package goivy

// impure_test.go — tests for NativeAction "impure" keyword parsing
// during compilation (Python: NativeAction.__init__ in ivy_actions.py:1270-1277).

import (
	"testing"
)

func TestCompileNativeAction_ImpureFlag(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Create a NativeAction AST node with "impure" as first line of code
	code := cfg.NewNativeCode("impure\nactual_code_here")
	nativeAST := cfg.NewNativeAction(code)

	result, err := c.CompileNativeAction(nativeAST)
	if err != nil {
		t.Fatalf("CompileNativeAction: %v", err)
	}

	na, ok := result.(*NativeAction)
	if !ok {
		t.Fatalf("expected *actions.NativeAction, got %T", result)
	}

	if !na.Impure {
		t.Error("NativeAction.Impure should be true when code starts with 'impure'")
	}

	// Verify "impure" line was stripped from code
	codeConst, ok := na.Code.(*Const)
	if !ok {
		t.Fatalf("expected Code to be *lg.Const, got %T", na.Code)
	}
	if codeConst.Name != "actual_code_here" {
		t.Errorf("code after stripping = %q, want %q", codeConst.Name, "actual_code_here")
	}
}

func TestCompileNativeAction_NotImpure(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Create a NativeAction AST node WITHOUT "impure" keyword
	code := cfg.NewNativeCode("normal_code\nmore_code")
	nativeAST := cfg.NewNativeAction(code)

	result, err := c.CompileNativeAction(nativeAST)
	if err != nil {
		t.Fatalf("CompileNativeAction: %v", err)
	}

	na, ok := result.(*NativeAction)
	if !ok {
		t.Fatalf("expected *actions.NativeAction, got %T", result)
	}

	if na.Impure {
		t.Error("NativeAction.Impure should be false for normal code")
	}

	// Verify code is preserved unchanged
	codeConst, ok := na.Code.(*Const)
	if !ok {
		t.Fatalf("expected Code to be *lg.Const, got %T", na.Code)
	}
	if codeConst.Name != "normal_code\nmore_code" {
		t.Errorf("code = %q, want %q", codeConst.Name, "normal_code\nmore_code")
	}
}

func TestCompileNativeAction_ImpureWithWhitespace(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Python: args[0].code.split('\n')[0].strip() == "impure"
	// So "  impure  \ncode" should match
	code := cfg.NewNativeCode("  impure  \ncode_body")
	nativeAST := cfg.NewNativeAction(code)

	result, err := c.CompileNativeAction(nativeAST)
	if err != nil {
		t.Fatalf("CompileNativeAction: %v", err)
	}

	na, ok := result.(*NativeAction)
	if !ok {
		t.Fatalf("expected *actions.NativeAction, got %T", result)
	}

	if !na.Impure {
		t.Error("NativeAction.Impure should be true for '  impure  ' (with whitespace)")
	}

	codeConst, ok := na.Code.(*Const)
	if !ok {
		t.Fatalf("expected Code to be *lg.Const, got %T", na.Code)
	}
	if codeConst.Name != "code_body" {
		t.Errorf("code after stripping = %q, want %q", codeConst.Name, "code_body")
	}
}

func TestCompileNativeAction_ImpureOnly(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Edge case: code is just "impure" with no following code
	code := cfg.NewNativeCode("impure")
	nativeAST := cfg.NewNativeAction(code)

	result, err := c.CompileNativeAction(nativeAST)
	if err != nil {
		t.Fatalf("CompileNativeAction: %v", err)
	}

	na, ok := result.(*NativeAction)
	if !ok {
		t.Fatalf("expected *actions.NativeAction, got %T", result)
	}

	if !na.Impure {
		t.Error("NativeAction.Impure should be true when code is just 'impure'")
	}

	codeConst, ok := na.Code.(*Const)
	if !ok {
		t.Fatalf("expected Code to be *lg.Const, got %T", na.Code)
	}
	if codeConst.Name != "" {
		t.Errorf("code after stripping = %q, want empty string", codeConst.Name)
	}
}
