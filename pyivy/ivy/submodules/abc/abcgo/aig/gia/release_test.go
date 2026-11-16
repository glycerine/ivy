package gia

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNoLiveCFreeCallsInGia(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(file), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, issue := range giaReleaseSourceIssues(path, source) {
			t.Error(issue)
		}
	}
}

// Inspect compiled Go syntax: a variable called free is not a C free() call.
// Also retain the earlier checks for untranslated reference/string operations.
func giaReleaseSourceIssues(path string, source []byte) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		return []string{err.Error()}
	}
	var issues []string
	report := func(pos token.Pos, message string) {
		issues = append(issues, fmt.Sprintf("%s at %s", message, fset.Position(pos)))
	}
	isNil := func(e ast.Expr) bool { name, ok := ast.Unparen(e).(*ast.Ident); return ok && name.Name == "nil" }
	checkStringNil := func(left, right ast.Expr) {
		if !isNil(right) {
			return
		}
		field, ok := ast.Unparen(left).(*ast.SelectorExpr)
		if ok && (field.Sel.Name == "pName" || field.Sel.Name == "pSpec" || field.Sel.Name == "pCellStr") {
			report(field.Pos(), "string field used with nil")
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			if name, ok := ast.Unparen(n.Fun).(*ast.Ident); ok && (name.Name == "free" || name.Name == "ABC_FREE") {
				report(name.Pos(), "C free call remains")
			}
		case *ast.SelectorExpr:
			if n.Sel.Name == "pRefs" {
				report(n.Sel.Pos(), "legacy manager ref spelling")
			}
		case *ast.BinaryExpr:
			if n.Op == token.EQL || n.Op == token.NEQ {
				checkStringNil(n.X, n.Y)
				checkStringNil(n.Y, n.X)
			}
		case *ast.AssignStmt:
			if len(n.Lhs) == len(n.Rhs) {
				for i := range n.Lhs {
					checkStringNil(n.Lhs[i], n.Rhs[i])
				}
			}
		}
		return true
	})
	return issues
}

func TestReleaseSourceAuditDistinguishesCalls(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"local-free", `func f() { free := uint64(1); _ = free >> 1 }`, 0},
		{"parameter-free", `func f(free []int) []int { return append(free, 1) }`, 0},
		{"field-free", `func f(p *T) { p.free(1) }`, 0},
		{"call-free", `func f(p *T) { free(p) }`, 1},
		{"wrapped-call-free", `func f(p *T) { (free)(p) }`, 1},
		{"call-macro", `func f(p *T) { ABC_FREE(p) }`, 1},
		{"comment", "// free(p)\nfunc f() {}", 0},
		{"string", `func f() { _ = "free(p)" }`, 0},
		{"comment-in-call", `func f(p *T) { free /* detail */ (p) }`, 1},
		{"refs", `func f(p *T) { _ = p.pRefs }`, 1},
		{"string-nil", `func f(p *T) { p.pName = nil }`, 1},
		{"string-nil-comparison", `func f(p *T) bool { return p.pSpec != nil }`, 1},
		{"string-nil-reversed", `func f(p *T) bool { return nil == p.pCellStr }`, 1},
		{"valid-string", `func f(p *T) { p.pName = "" }`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			issues := giaReleaseSourceIssues("fixture.go", []byte("package gia\n"+tc.source))
			if len(issues) != tc.count {
				t.Fatalf("issues %v, want %d", issues, tc.count)
			}
		})
	}
}
