package ivy2go

import (
	"fmt"
	"strings"
)

// specialNames mirrors ivy2cpp/names.go specialNames. Ivy uses textual
// operators for comparisons that aren't valid identifiers; we map them
// to the same __lt/__le/__gt/__ge prefix ivy2cpp does so name-mangled
// symbols remain comparable across packages.
var specialNames = map[string]string{
	"<":  "__lt",
	"<=": "__le",
	">":  "__gt",
	">=": "__ge",
}

// varName ports ivy2cpp/names.go varName. The output character set is
// already a subset of valid Go identifier characters (alphanumerics
// plus underscore), so we apply the same regex chain Python uses and
// rely on no additional Go-side escaping. Reserved-word collisions are
// handled by goIdent (below).
func varName(name any) string {
	s := fmt.Sprint(name)
	if n, ok := name.(interface{ GetName() string }); ok {
		s = n.GetName()
	}
	if v, ok := specialNames[s]; ok {
		return v
	}
	if strings.HasPrefix(s, "\"") {
		return s
	}
	repls := []struct{ old, new string }{
		{"loc:", "loc__"},
		{"ext:", "ext__"},
		{"___branch:", "__branch__"},
		{"__prm:", "prm__"},
		{"prm:", "prm__"},
		{"__fml:", ""},
		{"fml:", ""},
		{"ret:", ""},
	}
	for _, r := range repls {
		s = strings.ReplaceAll(s, r.old, r.new)
	}
	s = strings.NewReplacer(".", "__", "[", "__", "]", "__").Replace(s)
	s = strings.ReplaceAll(s, "@@", ".")
	s = strings.ReplaceAll(s, ":", "__COLON__")
	return s
}

// funName ports ivy2cpp/names.go funName. Mangles function names that
// begin with digits or minus signs, and rejects quoted strings.
func funName(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "__num" + name, nil
	}
	if name[0] == '-' {
		return "__negnum" + name, nil
	}
	if name[0] == '"' {
		return "", fmt.Errorf("cannot compile a function whose name is a quoted string: %s", name)
	}
	return varName(name), nil
}

// memName ports ivy2cpp/names.go memName. Returns the last dotted
// component of a qualified Ivy name.
func memName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// goReservedWords is the set of Go keywords plus predeclared identifiers
// we should avoid colliding with when generating Go code. Ivy programs
// can reasonably name things "type" or "select" or "len".
var goReservedWords = map[string]bool{
	// keywords (Go spec):
	"break": true, "case": true, "chan": true, "const": true,
	"continue": true, "default": true, "defer": true, "else": true,
	"fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true,
	"map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true,
	"var": true,
	// predeclared types / constants / functions worth avoiding:
	"any": true, "bool": true, "byte": true, "comparable": true,
	"complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "rune": true,
	"string": true, "uint": true, "uint8": true, "uint16": true,
	"uint32": true, "uint64": true, "uintptr": true,
	"true": true, "false": true, "iota": true, "nil": true,
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true,
	"recover": true,
}

// goIdent returns a valid Go identifier derived from name. It applies
// varName mangling, then suffixes "_" if the result collides with a
// Go reserved word or predeclared identifier. Use this for variable
// names, parameter names, and unexported field names in emitted code.
func goIdent(name string) string {
	id := varName(name)
	if goReservedWords[id] {
		return id + "_"
	}
	return id
}

// goExportedName returns an exported Go identifier (PascalCase) derived
// from name. snake_case input ("set_flag") becomes PascalCase ("SetFlag")
// so emitted methods read naturally. If the lowered identifier would
// start with a digit, it's prefixed with "X" so the Go compiler accepts
// it.
func goExportedName(name string) string {
	id := varName(name)
	if id == "" {
		return "X"
	}
	if id[0] >= '0' && id[0] <= '9' {
		id = "X" + id
	}
	// snake_case → PascalCase: split on '_', capitalise each part,
	// rejoin. Empty parts (e.g. "foo__bar" → ["foo","","bar"]) get
	// dropped so we don't introduce stutter.
	parts := strings.Split(id, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(upperFirst(p))
	}
	out := b.String()
	if out == "" {
		return "X"
	}
	return out
}

// goPackageName returns a valid Go package name derived from base.
// Package names by convention are short, all-lowercase, no underscores
// where avoidable. If base starts with a digit, prefix with "pkg_".
func goPackageName(base string) string {
	s := strings.ToLower(varName(base))
	if s == "" {
		return "ivygen"
	}
	if s[0] >= '0' && s[0] <= '9' {
		s = "pkg_" + s
	}
	if goReservedWords[s] {
		s = s + "_"
	}
	return s
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
