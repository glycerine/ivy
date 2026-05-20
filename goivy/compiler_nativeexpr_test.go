package goivy

import "testing"

func TestCompileNativeExprCompilesNativeCodeChild(t *testing.T) {
	mod := New()
	acfg := mod.Cfg.AstCfg
	native := acfg.NewNativeExpr([]Node{
		acfg.NewNativeCode("1"),
	})

	compiled, err := NewFromModule(mod).Thing(native)
	if err != nil {
		t.Fatalf("compile NativeExpr: %v", err)
	}
	nativeExpr, ok := compiled.(*LogicNativeExpr)
	if !ok {
		t.Fatalf("compiled NativeExpr = %T, want *LogicNativeExpr", compiled)
	}
	if len(nativeExpr.CompiledChildren) != 1 {
		t.Fatalf("compiled NativeExpr has %d children, want 1", len(nativeExpr.CompiledChildren))
	}
	if _, ok := nativeExpr.CompiledChildren[0].(*NativeCode); !ok {
		t.Fatalf("compiled NativeExpr child = %T, want *NativeCode", nativeExpr.CompiledChildren[0])
	}
}

func TestCompileNativeExprConvertsLegacyAtomCodeChild(t *testing.T) {
	mod := New()
	acfg := mod.Cfg.AstCfg
	native := acfg.NewNativeExpr([]Node{
		acfg.NewAtom(" `0` < `1` "),
		acfg.NewAtom("x"),
		acfg.NewAtom("y"),
	})
	mod.Sig.AddSymbol("x", &UninterpretedSort{Name: "idx"})
	mod.Sig.AddSymbol("y", &UninterpretedSort{Name: "idx"})

	compiled, err := NewFromModule(mod).Thing(native)
	if err != nil {
		t.Fatalf("compile NativeExpr with legacy Atom code child: %v", err)
	}
	nativeExpr, ok := compiled.(*LogicNativeExpr)
	if !ok {
		t.Fatalf("compiled NativeExpr = %T, want *LogicNativeExpr", compiled)
	}
	if len(nativeExpr.CompiledChildren) != 3 {
		t.Fatalf("compiled NativeExpr has %d children, want 3", len(nativeExpr.CompiledChildren))
	}
	code, ok := nativeExpr.CompiledChildren[0].(*NativeCode)
	if !ok {
		t.Fatalf("compiled NativeExpr first child = %T, want *NativeCode", nativeExpr.CompiledChildren[0])
	}
	if code.Code != " `0` < `1` " {
		t.Fatalf("compiled NativeCode = %q, want %q", code.Code, " `0` < `1` ")
	}
}
