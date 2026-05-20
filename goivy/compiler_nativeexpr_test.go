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
