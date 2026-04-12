package ivyutils

import "reflect"

// TypeName returns the bare struct name of v, stripping pointer
// and package prefixes. Matches Python's type(x).__name__.
// Example: (*ast.ComposeTactics) → "ComposeTactics"
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

// ShortTypeName is like TypeName but also maps Go type names
// to their Python equivalents where they differ:
//
//	Variable → Var
//	App      → Apply
func ShortTypeName(v interface{}) string {
	s := TypeName(v)
	switch s {
	case "Variable":
		return "Var"
	case "App":
		return "Apply"
	}
	return s
}

func BoolPythonStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
