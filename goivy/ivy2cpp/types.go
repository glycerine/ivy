package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func cppType(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return "int"
		}
		return varName(st.Name)
	case *goivy.RangeSort:
		return "long long"
	case *goivy.UninterpretedSort:
		return varName(st.Name)
	case *goivy.LogicFunctionSort:
		return cppFunctionType(st)
	default:
		return "long long"
	}
}

func cppFunctionType(s *goivy.LogicFunctionSort) string {
	dom := s.Domain()
	rng := cppType(s.Range())
	if len(dom) == 0 {
		return rng
	}
	if len(dom) == 1 {
		return fmt.Sprintf("std::map<%s,%s>", cppType(dom[0]), rng)
	}
	parts := make([]string, len(dom))
	for i, d := range dom {
		parts[i] = cppType(d)
	}
	return fmt.Sprintf("std::map<std::tuple<%s>,%s>", strings.Join(parts, ","), rng)
}

func cppQualifiedType(s goivy.Sort, className string) string {
	if className == "" {
		return cppType(s)
	}
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return "int"
		}
		return className + "::" + varName(st.Name)
	case *goivy.RangeSort:
		return "long long"
	case *goivy.UninterpretedSort:
		return className + "::" + varName(st.Name)
	case *goivy.LogicFunctionSort:
		return cppQualifiedFunctionType(st, className)
	default:
		return cppType(s)
	}
}

func cppQualifiedFunctionType(s *goivy.LogicFunctionSort, className string) string {
	dom := s.Domain()
	rng := cppQualifiedType(s.Range(), className)
	if len(dom) == 0 {
		return rng
	}
	if len(dom) == 1 {
		return fmt.Sprintf("std::map<%s,%s>", cppQualifiedType(dom[0], className), rng)
	}
	parts := make([]string, len(dom))
	for i, d := range dom {
		parts[i] = cppQualifiedType(d, className)
	}
	return fmt.Sprintf("std::map<std::tuple<%s>,%s>", strings.Join(parts, ","), rng)
}

func cppZeroValue(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "false"
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) > 0 {
			return varName(st.Extension[0])
		}
		return "0"
	default:
		return "0"
	}
}

func cppZeroValueInScope(s goivy.Sort, className string) string {
	switch st := s.(type) {
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) > 0 {
			name := varName(st.Extension[0])
			if className != "" {
				return className + "::" + name
			}
			return name
		}
	}
	return cppZeroValue(s)
}

func (g *Generator) cppZeroValue(s goivy.Sort) string {
	if g != nil {
		if name, ok := g.destructorStructName(s); ok {
			return varName(name) + "()"
		}
	}
	return cppZeroValue(s)
}

func (g *Generator) cppZeroValueInScope(s goivy.Sort) string {
	if g != nil {
		if name, ok := g.destructorStructName(s); ok {
			typeName := varName(name)
			if g.ClassName != "" {
				typeName = g.ClassName + "::" + typeName
			}
			return typeName + "()"
		}
	}
	className := ""
	if g != nil {
		className = g.ClassName
	}
	return cppZeroValueInScope(s, className)
}

func (g *Generator) destructorStructName(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil {
		return "", false
	}
	switch st := s.(type) {
	case *goivy.UninterpretedSort:
		if _, ok := g.Mod.SortDestructors.Get2(st.Name); ok {
			return st.Name, true
		}
	case *goivy.LogicEnumeratedSort:
		if _, ok := g.Mod.SortDestructors.Get2(st.Name); ok {
			return st.Name, true
		}
	case *goivy.RangeSort:
		if _, ok := g.Mod.SortDestructors.Get2(st.Name); ok {
			return st.Name, true
		}
	}
	return "", false
}

func sortName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		return st.Name
	case *goivy.RangeSort:
		return st.Name
	case *goivy.UninterpretedSort:
		return st.Name
	default:
		return fmt.Sprint(s)
	}
}
