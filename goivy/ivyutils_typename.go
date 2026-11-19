package goivy

import (
	"reflect"
	"strings"
)

// TypeName returns the concrete Go type name for v, stripped of pointer and
// package prefixes.
func TypeName(v interface{}) string {
	if v == nil || isNil(v) {
		return "nil"
	}
	if _, ok := v.(*LogicVariable); ok {
		return "Var"
	}
	if _, ok := v.(*nativeAtomExpr); ok {
		return "Atom"
	}
	if sc, ok := v.(*SomeCondition); ok {
		// Python class names of ivy_ast.Some / SomeMin / SomeMax.
		switch sc.Kind {
		case "some_min":
			return "SomeMin"
		case "some_max":
			return "SomeMax"
		}
		return "Some"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	nm := t.Name()
	if strings.HasPrefix(nm, "Logic") {
		return nm[5:]
	}
	return nm
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
