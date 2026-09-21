package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type stateSymbol struct {
	Name string
	Sort goivy.Sort
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

func isNumericEnum(s *goivy.LogicEnumeratedSort) bool {
	if s == nil || len(s.Extension) == 0 {
		return false
	}
	x := s.Extension[0]
	if x == "" {
		return false
	}
	if x[0] >= '0' && x[0] <= '9' {
		return true
	}
	return x[0] == '-' && len(x) > 1 && x[1] >= '0' && x[1] <= '9'
}

func (g *Generator) goType(s goivy.Sort) string {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.goFunctionStorageFor(fs.Domain(), fs.Range()).Type
	}
	return g.goScalarType(s)
}

func (g *Generator) goScalarType(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name != "" && !isNumericEnum(st) {
			return goName(st.Name)
		}
		return "int"
	case *goivy.RangeSort:
		return "int"
	case *goivy.UninterpretedSort:
		if g.hasStringInterp(st) {
			return "string"
		}
		return "int"
	default:
		return "int"
	}
}

type goFunctionStorage struct {
	Domain    []goivy.Sort
	Range     goivy.Sort
	RangeType string
	KeyType   string
	Type      string
}

func (g *Generator) goFunctionStorageFor(domain []goivy.Sort, rng goivy.Sort) goFunctionStorage {
	rangeType := g.goScalarType(rng)
	st := goFunctionStorage{Domain: domain, Range: rng, RangeType: rangeType, Type: rangeType}
	if len(domain) == 0 {
		return st
	}
	st.KeyType = g.goMapKeyType(domain)
	st.Type = fmt.Sprintf("map[%s]%s", st.KeyType, rangeType)
	return st
}

func (g *Generator) goMapKeyType(domain []goivy.Sort) string {
	if len(domain) == 1 {
		return g.goScalarType(domain[0])
	}
	fields := make([]string, len(domain))
	for i, s := range domain {
		fields[i] = fmt.Sprintf("A%d %s", i, g.goScalarType(s))
	}
	return "struct{ " + strings.Join(fields, "; ") + " }"
}

func (g *Generator) goMapKeyValue(domain []goivy.Sort, args []string) string {
	if len(domain) == 1 {
		return args[0]
	}
	return g.goMapKeyType(domain) + "{" + strings.Join(args, ", ") + "}"
}

func (g *Generator) goStorageDecl(name string, s goivy.Sort) string {
	return fmt.Sprintf("%s %s", goName(name), g.goType(s))
}

func (g *Generator) goStorageAccess(name string, sort goivy.Sort, args []string, obj string) string {
	base := goName(name)
	if obj != "" {
		base = obj + "." + base
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		return base
	}
	return fmt.Sprintf("%s[%s]", base, g.goMapKeyValue(fs.Domain(), args))
}

func (g *Generator) goZeroValue(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "false"
	case *goivy.LogicEnumeratedSort:
		if st.Name != "" && !isNumericEnum(st) && len(st.Extension) > 0 {
			return goName(st.Extension[0])
		}
		return "0"
	case *goivy.UninterpretedSort:
		if g.hasStringInterp(st) {
			return `""`
		}
		return "0"
	default:
		return "0"
	}
}

func (g *Generator) goRandomValueExpr(s goivy.Sort, name string, id int64) (string, error) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return fmt.Sprintf("ivy.___ivy_choose(2, %q, %d) != 0", name, id), nil
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) == 0 {
			return g.goZeroValue(st), nil
		}
		if isNumericEnum(st) {
			return fmt.Sprintf("[]int{%s}[ivy.___ivy_choose(%d, %q, %d)]", strings.Join(st.Extension, ", "), len(st.Extension), name, id), nil
		}
		return fmt.Sprintf("%s(ivy.___ivy_choose(%d, %q, %d))", goName(st.Name), len(st.Extension), name, id), nil
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok {
				return "", fmt.Errorf("ivy2golang: cannot randomize symbolic range %s", sortName(s))
			}
			width := hi - lo + 1
			if width <= 0 {
				return "", fmt.Errorf("ivy2golang: invalid range %s", sortName(s))
			}
			return fmt.Sprintf("(%d + ivy.___ivy_choose(%d, %q, %d))", lo, width, name, id), nil
		}
		if card := g.sortCard(s); card > 0 {
			return fmt.Sprintf("ivy.___ivy_choose(%d, %q, %d)", card, name, id), nil
		}
		return "", fmt.Errorf("ivy2golang: cannot randomize sort %s", sortName(s))
	}
}

func (g *Generator) finiteValueExprs(s goivy.Sort) ([]string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return []string{"false", "true"}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			if isNumericEnum(st) {
				vals[i] = v
			} else {
				vals[i] = goName(v)
			}
		}
		return vals, true
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok || hi < lo {
				return nil, false
			}
			vals := make([]string, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				vals = append(vals, strconv.Itoa(i))
			}
			return vals, true
		}
		if card := g.sortCard(s); card > 0 {
			vals := make([]string, card)
			for i := range vals {
				vals[i] = strconv.Itoa(i)
			}
			return vals, true
		}
		return nil, false
	}
}

func (g *Generator) rangeSortFor(s goivy.Sort) (*goivy.RangeSort, bool) {
	switch st := s.(type) {
	case *goivy.RangeSort:
		return st, true
	case *goivy.UninterpretedSort:
		if g != nil && g.Mod != nil && g.Mod.Sig != nil {
			if rs, ok := g.Mod.Sig.Interp[st.Name].(*goivy.RangeSort); ok {
				return rs, true
			}
		}
	}
	return nil, false
}

func numericRangeBounds(rs *goivy.RangeSort) (int, int, bool) {
	if rs == nil || rs.Lb == nil || rs.Ub == nil || !rs.Lb.IsNumeral() || !rs.Ub.IsNumeral() {
		return 0, 0, false
	}
	lo, err := strconv.Atoi(rs.LbString())
	if err != nil {
		return 0, 0, false
	}
	hi, err := strconv.Atoi(rs.UbString())
	if err != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

func (g *Generator) sortInterpString(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return "", false
	}
	name := sortName(s)
	if name == "" {
		return "", false
	}
	text, ok := g.Mod.Sig.Interp[name].(string)
	return text, ok
}

func (g *Generator) hasStringInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "strlit"
}

func (g *Generator) sortCard(s goivy.Sort) int {
	if g != nil && g.Mod != nil {
		if card := g.Mod.SortCard(s); card > 0 {
			return card
		}
		if g.Mod.Sig != nil {
			if card := goivy.SortCard(s, g.Mod.Sig); card > 0 {
				return card
			}
		}
	}
	return goivy.SortCardDefault(s)
}

func (g *Generator) sortDependenciesReferenceName(s goivy.Sort, name string) bool {
	if s == nil || name == "" {
		return false
	}
	if sortName(s) == name {
		return true
	}
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		for _, d := range fs.Domain() {
			if g.sortDependenciesReferenceName(d, name) {
				return true
			}
		}
		return g.sortDependenciesReferenceName(fs.Range(), name)
	}
	return false
}
