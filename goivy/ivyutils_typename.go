package goivy

import "reflect"

// TypeName returns the Python-facing type name for v. It strips pointer and
// package prefixes, and maps Go type names that had to change during the
// uni-package merge back to their Python class names.
func TypeName(v interface{}) string {
	if v == nil || isNil(v) {
		return "nil"
	}
	switch v.(type) {
	case *AstVariable, AstVariable:
		return "Variable" // Python ivy_ast.Variable
	case *Variable, Variable:
		return "Var" // Python logic.Var
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
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
