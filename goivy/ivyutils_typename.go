package goivy

import "reflect"

// TypeName returns the bare struct name of v, stripping pointer
// and package prefixes. Matches Python's type(x).__name__.
// Example: (*ast.ComposeTactics) → "ComposeTactics".
//
// Maps Go's logic.Variable → "Var" to match Python lg.Var's class
// name (logic.py:117 defines `class Var`; Go uses the more
// descriptive Go-idiomatic "Variable"). ast.AstVariable stays
// "Variable" (corresponds to Python ivy_ast.py:385 `class Variable`).
func TypeName(v interface{}) string {
	if v == nil || isNil(v) {
		return "nil"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	s := t.Name()
	if s == "Variable" && isLogicPkg(t.PkgPath()) {
		return "Var"
	}
	return s
}

// isLogicPkg reports whether the given Go package path refers to
// ~/ivy/goivy/logic (which contains the logic.Variable that maps
// to Python's lg.Var). ast.AstVariable lives in a different package
// and should not be remapped.
func isLogicPkg(pkgPath string) bool {
	return pkgPath == "github.com/glycerine/ivy/goivy/logic"
}

// ShortTypeName is an alias for TypeName kept for callers that
// intentionally want the Python-mapped short name.
func ShortTypeName(v interface{}) string {
	return TypeName(v)
}

func BoolPythonStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
