package ivy2cpp

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

func memName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}
