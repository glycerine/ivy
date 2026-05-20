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

func cppZeroValue(s goivy.Sort) string {
	switch s.(type) {
	case *goivy.BooleanSort:
		return "false"
	default:
		return "0"
	}
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
