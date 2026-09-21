package ivy2golang

import (
	"fmt"
	"strings"
)

var specialNames = map[string]string{
	"<":  "__lt",
	"<=": "__le",
	">":  "__gt",
	">=": "__ge",
}

var goKeywords = map[string]bool{
	"break":       true,
	"default":     true,
	"func":        true,
	"interface":   true,
	"select":      true,
	"case":        true,
	"defer":       true,
	"go":          true,
	"map":         true,
	"struct":      true,
	"chan":        true,
	"else":        true,
	"goto":        true,
	"package":     true,
	"switch":      true,
	"const":       true,
	"fallthrough": true,
	"if":          true,
	"range":       true,
	"type":        true,
	"continue":    true,
	"for":         true,
	"import":      true,
	"return":      true,
	"var":         true,
}

func goName(name any) string {
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
	s = strings.NewReplacer(".", "__", "[", "__", "]", "__", "@@", "__").Replace(s)
	s = strings.ReplaceAll(s, ":", "__COLON__")
	if s == "" {
		s = "_"
	}
	var b strings.Builder
	for i, r := range s {
		ok := r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		if i > 0 {
			ok = ok || r >= '0' && r <= '9'
		}
		if ok {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	s = b.String()
	if s == "" {
		s = "_"
	}
	if s[0] >= '0' && s[0] <= '9' {
		s = "__num" + s
	}
	if goKeywords[s] {
		s = s + "_"
	}
	return s
}

func funName(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "__num" + name, nil
	}
	if name[0] == '-' {
		return "__negnum" + name[1:], nil
	}
	if name[0] == '"' {
		return "", fmt.Errorf("cannot compile an action or function whose name is a quoted string: %s", name)
	}
	return goName(name), nil
}

func memName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func exportedishName(name string) string {
	name = goName(name)
	if name == "" {
		return "Ivy"
	}
	if name[0] >= 'a' && name[0] <= 'z' {
		return string(name[0]-('a'-'A')) + name[1:]
	}
	if name[0] == '_' {
		return "Ivy" + name
	}
	return name
}
