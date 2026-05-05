package goivy

import "reflect"

// TypeName returns the concrete Go type name for v, stripped of pointer and
// package prefixes.
func TypeName(v interface{}) string {
	if v == nil || isNil(v) {
		return "nil"
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
