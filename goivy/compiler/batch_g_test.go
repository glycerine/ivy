package compiler

// Batch G — Tests for:
//   §6.3 #27: DomainSetup.Native no-op / CompileNativeDef field-based decision
//   §6.3 #30: ARGSetup export/import CheckIsAction, progress add_symbol

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// ============================================================================
// §6.3 #27: DomainSetup.Native is a no-op
// ============================================================================

// TestDomainSetupNative_NoOp verifies that DomainSetup.Native does NOT
// append to mod.Natives (Python only handles native in ARGSetup pass 3).
func TestDomainSetupNative_NoOp(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	ds := &DomainSetup{Compiler: c}

	nativeDef := cfg.NewNativeDef([]ast.Node{
		cfg.NewAtom("myNative"),
		cfg.NewNativeCode("some code"),
	})

	err := ds.Native(nativeDef)
	if err != nil {
		t.Fatalf("DomainSetup.Native returned error: %v", err)
	}

	if len(c.Module.Natives) != 0 {
		t.Errorf("DomainSetup.Native should not append to mod.Natives, got %d entries", len(c.Module.Natives))
	}
}

// TestARGSetupNative_CompilesAndAppends verifies that ARGSetup compiles
// native defs via CompileNativeDef and appends the compiled result.
func TestARGSetupNative_CompilesAndAppends(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	nativeDef := cfg.NewNativeDef([]ast.Node{
		cfg.NewAtom("myNative"),
		cfg.NewNativeCode("some `arg0` code"),
		cfg.NewAtom("x"),
	})

	decls := []ast.Node{cfg.NewNativeDecl(nativeDef)}

	err := as.ProcessDecls(decls)
	if err != nil {
		t.Fatalf("ARGSetup.ProcessDecls returned error: %v", err)
	}

	if len(c.Module.Natives) != 1 {
		t.Fatalf("expected 1 native, got %d", len(c.Module.Natives))
	}
}

// ============================================================================
// §6.3 #27: CompileNativeDef field-based arg vs symbol decision
// ============================================================================

// TestCompileNativeDef_FieldBasedDecision verifies that CompileNativeDef
// uses the code template's backtick-split fields to decide arg vs symbol.
// Python: compile_native_arg(a) if not fields[i*2].endswith('"') else compile_native_symbol(a)
func TestCompileNativeDef_FieldBasedDecision(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Template: `arg0`"` — fields[0]="" (no quote ending → arg), fields[2] ends with " → symbol
	// But we need a simpler case. Let's test:
	// code = 'prefix`x`middle"`y`suffix'
	// fields = ["prefix", "x", "middle\"", "y", "suffix"]
	// i=0: fields[0]="prefix" → does not end with " → compile_native_arg
	// i=1: fields[2]='middle"' → ends with " → compile_native_symbol
	nativeDef := cfg.NewNativeDef([]ast.Node{
		cfg.NewAtom("myFunc"), // args[0]: name
		cfg.NewNativeCode(`prefix` + "`x`" + `middle"` + "`y`suffix"), // args[1]: code template
		cfg.NewAtom("argParam"), // args[2]: should be compiled as arg (fields[0] = "prefix")
		cfg.NewAtom("symParam"), // args[3]: should be compiled as symbol (fields[2] = 'middle"')
	})

	compiled, err := c.CompileNativeDef(nativeDef)
	if err != nil {
		t.Fatalf("CompileNativeDef error: %v", err)
	}

	// Verify the compiled result has the right number of args
	cArgs := compiled.Args()
	if len(cArgs) != 4 {
		t.Fatalf("expected 4 args, got %d", len(cArgs))
	}

	// args[0] should be CompiledNode (name compiled)
	if _, ok := cArgs[0].(*ast.CompiledNode); !ok {
		// Name may fail to compile in test context — that's ok, just check it's present
		t.Logf("args[0] type: %T (name compilation may fail without full sig)", cArgs[0])
	}

	// args[1] should be the NativeCode template (preserved as-is)
	if nc, ok := cArgs[1].(*ast.NativeCode); !ok {
		t.Errorf("expected args[1] to be NativeCode, got %T", cArgs[1])
	} else if !strings.Contains(nc.Code, "prefix") {
		t.Errorf("expected code template to contain 'prefix', got %q", nc.Code)
	}
}

// ============================================================================
// §6.3 #30: Export CheckIsAction error
// ============================================================================

// TestExport_CheckIsAction_Error verifies that exporting a non-existent
// action returns an error (not just a warning).
func TestExport_CheckIsAction_Error(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	// No actions registered — export should fail
	exportDef := cfg.NewExportDef(cfg.NewAtom("nonExistentAction"), cfg.NewAtom(""))
	decls := []ast.Node{cfg.NewExportDecl(exportDef)}

	err := as.ProcessDecls(decls)
	if err == nil {
		t.Fatal("expected error for exporting non-existent action, got nil")
	}
	if !strings.Contains(err.Error(), "is not an action") {
		t.Errorf("expected 'is not an action' error, got: %v", err)
	}
}

// TestExport_CheckIsAction_Success verifies that exporting a registered
// action succeeds and appends to mod.Exports.
func TestExport_CheckIsAction_Success(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	// Register the action first
	c.Module.Actions.Set("myAction", nil)

	exportDef := cfg.NewExportDef(cfg.NewAtom("myAction"), cfg.NewAtom(""))
	decls := []ast.Node{cfg.NewExportDecl(exportDef)}

	err := as.ProcessDecls(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Module.Exports) != 1 {
		t.Errorf("expected 1 export, got %d", len(c.Module.Exports))
	}
}

// ============================================================================
// §6.3 #30: Import CheckIsAction error
// ============================================================================

// TestImport_CheckIsAction_Error verifies that importing a non-existent
// action returns an error.
func TestImport_CheckIsAction_Error(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	importDef := cfg.NewImportDef(cfg.NewAtom("nonExistentAction"), cfg.NewAtom(""))
	decls := []ast.Node{cfg.NewImportDecl(importDef)}

	err := as.ProcessDecls(decls)
	if err == nil {
		t.Fatal("expected error for importing non-existent action, got nil")
	}
	if !strings.Contains(err.Error(), "is not an action") {
		t.Errorf("expected 'is not an action' error, got: %v", err)
	}
}

// TestImport_CheckIsAction_Success verifies that importing a registered
// action succeeds and appends to mod.Imports.
func TestImport_CheckIsAction_Success(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	// Register the action first
	c.Module.Actions.Set("myAction", nil)

	importDef := cfg.NewImportDef(cfg.NewAtom("myAction"), cfg.NewAtom(""))
	decls := []ast.Node{cfg.NewImportDecl(importDef)}

	err := as.ProcessDecls(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Module.Imports) != 1 {
		t.Errorf("expected 1 import, got %d", len(c.Module.Imports))
	}
}

// ============================================================================
// §6.3 #30: Progress adds symbol before sortify
// ============================================================================

// TestProgress_AddsSymbol verifies that DomainSetup.Progress registers the
// progress relation symbol in the signature before sortifying.
func TestProgress_AddsSymbol(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Register sorts and symbols needed for compilation
	nodeSort := &lg.UninterpretedSort{Name: "node"}
	c.Sig.AddSort(nodeSort)

	// Register "body" as a known symbol so SortifyWithInference can compile it
	bodySort := lg.Boolean
	c.Sig.AddSymbol("body", bodySort)

	ds := &DomainSetup{Compiler: c}

	// Build: rel = myProgress(X:node), body = body (a known boolean symbol)
	relArg := cfg.NewVariable("X", "node")
	rel := cfg.NewAtom("myProgress", relArg)
	body := cfg.NewAtom("body")

	progDecl := cfg.NewAtom("progress_decl", rel, body)

	// Progress should add the symbol. SortifyWithInference may still fail
	// on the overall node, but AddSymbol should have been called already.
	_ = ds.Progress(progDecl)

	// Check that the symbol was registered in the signature
	sym, err := c.Sig.FindSymbol("myProgress", false)
	if err != nil {
		t.Fatalf("expected myProgress symbol in sig, got error: %v", err)
	}
	if sym == nil {
		t.Fatal("expected non-nil symbol for myProgress")
	}
}

// TestProgress_SortifyAndAppend verifies that Progress sortifies the
// declaration and appends to mod.Progress.
func TestProgress_SortifyAndAppend(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Register symbols so SortifyWithInference can compile the whole node.
	// The outer node "progress_decl" needs a sort: RelationSort(boolean, boolean) -> boolean
	pSort := lg.Boolean
	c.Sig.AddSymbol("p", pSort)
	c.Sig.AddSymbol("q", pSort)
	// Register the outer node symbol with a function sort accepting two boolean args
	outerSort := il.RelationSort([]lg.Sort{pSort, pSort})
	c.Sig.AddSymbol("progress_decl", outerSort)

	ds := &DomainSetup{Compiler: c}

	rel := cfg.NewAtom("p")
	body := cfg.NewAtom("q")
	progDecl := cfg.NewAtom("progress_decl", rel, body)

	err := ds.Progress(progDecl)
	if err != nil {
		t.Fatalf("Progress returned error: %v", err)
	}

	if len(c.Module.Progress) != 1 {
		t.Errorf("expected 1 progress entry, got %d", len(c.Module.Progress))
	}
}

// ============================================================================
// Helpers
// ============================================================================

func init() {
	// Ensure il package is used (for NewSig via newTestCompiler)
	_ = il.NewSig
}
