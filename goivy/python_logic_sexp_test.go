package goivy

import (
	"strings"
	"testing"
)

func TestPythonDefinitionSexpHandlesAstNativeExpr(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import canon_ast, ivy_ast, ivy_logic, logic, logic_sexp
canon_ast.install()
logic_sexp.install()
d = ivy_logic.Definition(
    logic.Const('x', logic.TopS),
    ivy_ast.NativeExpr(ivy_ast.NativeCode('foo')),
)
print(d.sexp())
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python Definition.sexp with AST NativeExpr failed: %v\n%s", err, out)
	}
	got := string(out)
	if strings.Contains(got, "AttributeError") {
		t.Fatalf("NativeExpr should have a sexp helper; got:\n%s", got)
	}
	if !strings.Contains(got, "rhs:(nativeExpr)") {
		t.Fatalf("NativeExpr should use Go-compatible AST canon shape; got:\n%s", got)
	}
}
