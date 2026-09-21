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
		if g.isVariantSuperName(st.Name) {
			return goName(st.Name)
		}
		if g.destructorStructFields(st.Name) != nil {
			return goName(st.Name)
		}
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
	Large     bool
}

func (g *Generator) goFunctionStorageFor(domain []goivy.Sort, rng goivy.Sort) goFunctionStorage {
	rangeType := g.goScalarType(rng)
	st := goFunctionStorage{Domain: domain, Range: rng, RangeType: rangeType, Type: rangeType}
	if len(domain) == 0 {
		return st
	}
	st.KeyType = g.goMapKeyType(domain)
	if !g.canEnumerateDomain(domain) {
		st.Large = true
		st.Type = fmt.Sprintf("ivyThunkMap[%s, %s]", st.KeyType, rangeType)
		return st
	}
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
	key := g.goMapKeyValue(fs.Domain(), args)
	if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		return fmt.Sprintf("%s.Get(%s)", base, key)
	}
	return fmt.Sprintf("%s[%s]", base, key)
}

func (g *Generator) goStorageSet(lhs goivy.Expr, rhs string) (string, bool, error) {
	app, ok := lhs.(*goivy.Apply)
	if !ok {
		return "", false, nil
	}
	name := goivy.ExprName(app.Func)
	if name == "" {
		return "", false, nil
	}
	sort, owner := g.storageSortAndOwner(name)
	if sort == nil {
		return "", false, nil
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || !g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		return "", false, nil
	}
	args := make([]string, len(app.Terms))
	for i, term := range app.Terms {
		code, err := g.emitExpr(term)
		if err != nil {
			return "", true, err
		}
		args[i] = code
	}
	base := goName(name)
	if owner != "" {
		base = owner + "." + base
	}
	return fmt.Sprintf("%s.Set(%s, %s)", base, g.goMapKeyValue(fs.Domain(), args), rhs), true, nil
}

func (g *Generator) storageSortAndOwner(name string) (goivy.Sort, string) {
	if sort, ok := g.isStateSymbolName(name); ok {
		return sort, "ivy"
	}
	if sort, ok := g.localSort(name); ok {
		return sort, ""
	}
	return nil, ""
}

func (g *Generator) goFunctionStorageInit(domain []goivy.Sort, rng goivy.Sort) string {
	st := g.goFunctionStorageFor(domain, rng)
	if st.Large {
		return fmt.Sprintf("newIvyThunkMap[%s, %s](%s)", st.KeyType, st.RangeType, g.goZeroValue(rng))
	}
	return fmt.Sprintf("make(%s)", st.Type)
}

func (g *Generator) goStorageRangeExpr(name string, sort goivy.Sort, obj string) string {
	base := goName(name)
	if obj != "" {
		base = obj + "." + base
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		return base
	}
	if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		return base + ".overrides"
	}
	return base
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
		if g.isVariantSuperName(st.Name) {
			return goName(st.Name) + "{}"
		}
		if g.destructorStructFields(st.Name) != nil {
			return goName(st.Name) + "{}"
		}
		if g.hasStringInterp(st) {
			return `""`
		}
		return "0"
	default:
		return "0"
	}
}

func (g *Generator) goRandomValueExpr(s goivy.Sort, name string, id int64) (string, error) {
	return g.goRandomValueExprWithChooser(s, name, id, "___ivy_choose")
}

func (g *Generator) goActionParamRandomValueExpr(s goivy.Sort, name string, id int64) (string, error) {
	return g.goRandomValueExprWithChooser(s, name, id, "___ivy_randomize")
}

func (g *Generator) goRandomValueExprWithChooser(s goivy.Sort, name string, id int64, chooser string) (string, error) {
	return g.goRandomValueExprWithChooserSeen(s, name, id, chooser, map[string]bool{})
}

func (g *Generator) goRandomValueExprWithChooserSeen(s goivy.Sort, name string, id int64, chooser string, seen map[string]bool) (string, error) {
	call := func(rng string) string {
		return fmt.Sprintf("ivy.%s(%s, %q, %d)", chooser, rng, name, id)
	}
	callInt := func(rng int) string {
		return call(strconv.Itoa(rng))
	}
	if expr, ok, err := g.goRandomVariantValueExprWithChooserSeen(s, name, id, chooser, seen); ok || err != nil {
		return expr, err
	}
	if fields := g.destructorStructFields(sortName(s)); len(fields) > 0 {
		key := "struct:" + sortName(s)
		if seen[key] {
			return g.goZeroValue(s), nil
		}
		seen[key] = true
		defer delete(seen, key)
		inits := make([]string, 0, len(fields))
		for i, d := range fields {
			fs, ok := d.CSort.(*goivy.LogicFunctionSort)
			if !ok || len(fs.Domain()) != 1 {
				continue
			}
			field := goName(memName(d.Name))
			expr, err := g.goRandomValueExprWithChooserSeen(fs.Range(), name+"."+memName(d.Name), int64(i), chooser, seen)
			if err != nil {
				return "", err
			}
			inits = append(inits, field+": "+expr)
		}
		if len(inits) == len(fields) {
			return fmt.Sprintf("%s{%s}", g.goScalarType(s), strings.Join(inits, ", ")), nil
		}
	}
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return fmt.Sprintf("%s != 0", callInt(2)), nil
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) == 0 {
			return g.goZeroValue(st), nil
		}
		if isNumericEnum(st) {
			return fmt.Sprintf("[]int{%s}[%s]", strings.Join(st.Extension, ", "), callInt(len(st.Extension))), nil
		}
		return fmt.Sprintf("%s(%s)", goName(st.Name), callInt(len(st.Extension))), nil
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if ok {
				width := hi - lo + 1
				if width <= 0 {
					return "", fmt.Errorf("ivy2golang: invalid range %s", sortName(s))
				}
				return fmt.Sprintf("(%d + %s)", lo, callInt(width)), nil
			}
			loExpr, hiExpr, ok, err := g.goSymbolicRangeLoopBounds(rs)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", fmt.Errorf("ivy2golang: cannot randomize symbolic range %s", sortName(s))
			}
			return fmt.Sprintf("(%s + %s)", loExpr, call(goRangeRandomWidthExpr(loExpr, hiExpr))), nil
		}
		if card := g.sortCard(s); card > 0 {
			return callInt(card), nil
		}
		g.warnOnce(fmt.Sprintf("ivy2golang: using zero value for non-enumerable sort %s", sortName(s)))
		return g.goZeroValue(s), nil
	}
}

func goRangeRandomWidthExpr(lo, hi string) string {
	if strings.TrimSpace(lo) == "0" {
		return "(" + hi + " + 1)"
	}
	return "((" + hi + ") - (" + lo + ") + 1)"
}

func (g *Generator) goRandomVariantValueExprWithChooserSeen(s goivy.Sort, name string, id int64, chooser string, seen map[string]bool) (string, bool, error) {
	sortText := sortName(s)
	if sortText == "" || !g.isVariantSuperName(sortText) {
		return "", false, nil
	}
	key := "variant:" + sortText
	if seen[key] {
		return g.goZeroValue(s), true, nil
	}
	seen[key] = true
	defer delete(seen, key)
	variants := g.Mod.Variants[sortText]
	if len(variants) == 0 {
		return g.goZeroValue(s), true, nil
	}
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", goName(sortText)))
	w.indent++
	w.linef("switch ivy.%s(%d, %q, %d) {", chooser, len(variants), name, id)
	w.indent++
	for i, sub := range variants {
		expr, err := g.goRandomValueExprWithChooserSeen(sub, name+"."+sortName(sub), int64(i), chooser, seen)
		if err != nil {
			return "", true, err
		}
		w.linef("case %d:", i)
		w.indent++
		w.linef("return %s", g.variantUpcastExpr(s, sub, expr))
		w.indent--
	}
	w.indent--
	w.line("}")
	w.linef("return %s", g.goZeroValue(s))
	w.indent--
	w.raw("}()")
	return w.String(), true, nil
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

func (g *Generator) canEnumerateDomain(domain []goivy.Sort) bool {
	for _, s := range domain {
		if _, ok := g.finiteValueExprs(s); !ok {
			return false
		}
	}
	return true
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

func (g *Generator) hasNatInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "nat"
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

func (g *Generator) isVariantSuperName(name string) bool {
	if g == nil || g.Mod == nil || len(g.Mod.Variants) == 0 || name == "" {
		return false
	}
	return len(g.Mod.Variants[name]) > 0
}

func (g *Generator) isVariantSubtypeName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	if g.Mod.Supertypes != nil {
		if _, ok := g.Mod.Supertypes.Get2(name); ok {
			return true
		}
	}
	for _, variants := range g.Mod.Variants {
		for _, s := range variants {
			if sortName(s) == name {
				return true
			}
		}
	}
	return false
}

func (g *Generator) destructorStructFields(name string) []*goivy.Const {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil || name == "" {
		return nil
	}
	fields, ok := g.Mod.SortDestructors.Get2(name)
	if !ok || len(fields) == 0 {
		return nil
	}
	out := make([]*goivy.Const, 0, len(fields))
	for _, d := range fields {
		if d == nil {
			return nil
		}
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) != 1 || sortName(fs.Domain()[0]) != name {
			return nil
		}
		out = append(out, d)
	}
	return out
}

func (g *Generator) destructorRecordFields(name string) []*goivy.Const {
	if g.isVariantSuperName(name) || g.isVariantSubtypeName(name) {
		return nil
	}
	return g.destructorStructFields(name)
}

func (g *Generator) destructorFieldName(name string) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil || name == "" {
		return "", false
	}
	sort, ok := g.Mod.DestructorSorts[name]
	if !ok {
		return "", false
	}
	if len(g.destructorStructFields(sortName(sort))) == 0 {
		return "", false
	}
	return goName(memName(name)), true
}

func (g *Generator) variantUpcastExpr(super, sub goivy.Sort, expr string) string {
	idx := -1
	if g != nil && g.Mod != nil {
		idx = g.Mod.VariantIndex(super, sub)
	}
	return fmt.Sprintf("%s{tag: %d, value: %s, valid: true}", g.goScalarType(super), idx, expr)
}

func (g *Generator) variantDowncastExpr(superExpr string, sub goivy.Sort) string {
	return fmt.Sprintf("%s.value.(%s)", superExpr, g.goScalarType(sub))
}

func (g *Generator) maybeVariantUpcastExpr(target, value goivy.Sort, expr string) string {
	if g != nil && g.Mod != nil && g.Mod.IsVariant(target, value) {
		return g.variantUpcastExpr(target, value, expr)
	}
	return expr
}

func (g *Generator) goLessExpr(left, right string, s goivy.Sort) string {
	if _, ok := s.(*goivy.BooleanSort); ok {
		return fmt.Sprintf("ivyBoolOrd(%s) < ivyBoolOrd(%s)", left, right)
	}
	return fmt.Sprintf("%s < %s", left, right)
}
