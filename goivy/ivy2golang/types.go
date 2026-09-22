package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

const goLargeThresh = 1024

type stateSymbol struct {
	Name string
	Sort goivy.Sort
}

type goDestructorField struct {
	Const     *goivy.Const
	Sort      *goivy.LogicFunctionSort
	FieldName string
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
		if g.hasStringValuedInterp(st) {
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
	if call, ok, err := g.goDestructorFieldSet(app, name, rhs); ok || err != nil {
		return call, ok, err
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
		if g.hasStringValuedInterp(st) {
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
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.goRandomFunctionValueExprWithChooserSeen(fs, name, id, chooser, seen)
	}
	if expr, ok, err := g.goRandomVariantValueExprWithChooserSeen(s, name, id, chooser, seen); ok || err != nil {
		return expr, err
	}
	if fields := g.destructorStructFieldInfos(sortName(s)); len(fields) > 0 {
		key := "struct:" + sortName(s)
		if seen[key] {
			return g.goZeroValue(s), nil
		}
		seen[key] = true
		defer delete(seen, key)
		if g.destructorStructNeedsInitFunc(fields) {
			expr, err := g.goRandomDestructorStructValueExprWithChooserSeen(s, fields, name, id, chooser, seen)
			if err != nil {
				return "", err
			}
			return expr, nil
		}
		inits := make([]string, 0, len(fields))
		for i, field := range fields {
			expr, err := g.goRandomValueExprWithChooserSeen(field.Sort.Range(), name+"."+memName(field.Const.Name), int64(i), chooser, seen)
			if err != nil {
				return "", err
			}
			inits = append(inits, field.FieldName+": "+expr)
		}
		return fmt.Sprintf("%s{%s}", g.goScalarType(s), strings.Join(inits, ", ")), nil
	}
	if g.isVariantSubtypeName(sortName(s)) {
		return g.goZeroValue(s), nil
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
			if g.hasStrBVInterp(s) {
				return fmt.Sprintf("strconv.Itoa(%s)", callInt(card)), nil
			}
			return callInt(card), nil
		}
		if g.hasIntOrNatInterp(s) {
			return callInt(5), nil
		}
		if err := g.testGeneratorZeroFallbackError(s); err != nil {
			return "", err
		}
		if g.warnsOnNonEnumerableZeroFallback() {
			g.warnOnce(fmt.Sprintf("ivy2golang: using zero value for non-enumerable sort %s", sortName(s)))
		}
		return g.goZeroValue(s), nil
	}
}

func (g *Generator) warnsOnNonEnumerableZeroFallback() bool {
	if g == nil {
		return false
	}
	return g.Config.Target == "test" || g.Config.Target == "gen"
}

func (g *Generator) destructorStructNeedsInitFunc(fields []goDestructorField) bool {
	for _, field := range fields {
		if len(field.Sort.Domain()) > 1 {
			return true
		}
	}
	return false
}

func (g *Generator) goRandomDestructorStructValueExprWithChooserSeen(s goivy.Sort, fields []goDestructorField, name string, id int64, chooser string, seen map[string]bool) (string, error) {
	tmp := g.nextTemp("__ivy_rec")
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", g.goScalarType(s)))
	w.indent++
	w.linef("var %s %s", tmp, g.goScalarType(s))
	for i, field := range fields {
		domain := field.Sort.Domain()
		if g.destructorFieldIsLarge(field) {
			extra := domain[1:]
			tmpThunk := g.nextTemp("__ivy_thunk")
			w.linef("%s := %s", tmpThunk, g.goFunctionStorageInit(extra, field.Sort.Range()))
			g.emitRandomThunkBaseWithChooserSeen(&w, tmpThunk, extra, field.Sort.Range(), name+"."+memName(field.Const.Name), int64(i), chooser, seen)
			w.linef("%s.%s = &%s", tmp, field.FieldName, tmpThunk)
			continue
		}
		if len(domain) == 1 {
			expr, err := g.goRandomValueExprWithChooserSeen(field.Sort.Range(), name+"."+memName(field.Const.Name), int64(i), chooser, seen)
			if err != nil {
				return "", err
			}
			w.linef("%s.%s = %s", tmp, field.FieldName, expr)
			continue
		}
		extra := domain[1:]
		g.emitDomainLoops(&w, extra, func(args []string) {
			expr, err := g.goRandomValueExprWithChooserSeen(field.Sort.Range(), name+"."+memName(field.Const.Name), int64(i), chooser, seen)
			if err != nil {
				g.unsupported(&w, "unsupported destructor field random value: %s", err.Error())
				return
			}
			w.linef("%s.%s%s = %s", tmp, field.FieldName, goIndexSuffix(args), expr)
		})
	}
	w.linef("return %s", tmp)
	w.indent--
	w.raw("}()")
	return w.String(), nil
}

func (g *Generator) goRandomFunctionValueExprWithChooserSeen(fs *goivy.LogicFunctionSort, name string, id int64, chooser string, seen map[string]bool) (string, error) {
	domain := fs.Domain()
	rng := fs.Range()
	if len(domain) == 0 {
		return g.goRandomValueExprWithChooserSeen(rng, name, id, chooser, seen)
	}
	st := g.goFunctionStorageFor(domain, rng)
	tmp := g.nextTemp("__ivy_fn")
	var w goWriter
	w.raw(fmt.Sprintf("func() %s {\n", st.Type))
	w.indent++
	w.linef("%s := %s", tmp, g.goFunctionStorageInit(domain, rng))
	if st.Large {
		g.emitRandomThunkBaseWithChooserSeen(&w, tmp, domain, rng, name, id, chooser, seen)
	} else {
		g.emitDomainLoops(&w, domain, func(args []string) {
			expr, err := g.goRandomValueExprWithChooserSeen(rng, name, id, chooser, seen)
			if err != nil {
				g.unsupported(&w, "unsupported function-sorted random value range: %s", err.Error())
				return
			}
			w.linef("%s[%s] = %s", tmp, g.goMapKeyValue(domain, args), expr)
		})
	}
	w.linef("return %s", tmp)
	w.indent--
	w.raw("}()")
	return w.String(), nil
}

func (g *Generator) emitRandomThunkBaseWithChooserSeen(w *goWriter, base string, domain []goivy.Sort, rng goivy.Sort, label string, id int64, chooser string, seen map[string]bool) {
	st := g.goFunctionStorageFor(domain, rng)
	if !st.Large {
		return
	}
	expr, err := g.goRandomValueExprWithChooserSeen(rng, label, id, chooser, seen)
	if err != nil {
		g.unsupported(w, "unsupported nondet thunk range: %s", err.Error())
		return
	}
	w.open(fmt.Sprintf("%s.base = func(__ivy_key %s) %s {", base, st.KeyType, st.RangeType))
	w.line("_ = __ivy_key")
	w.linef("return %s", expr)
	w.close("")
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
		if g.hasStrBVInterp(s) {
			card := g.sortCard(s)
			if card <= 0 {
				return nil, false
			}
			vals := make([]string, card)
			for i := range vals {
				vals[i] = strconv.Quote(strconv.Itoa(i))
			}
			return vals, true
		}
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
	product := 1
	for _, s := range domain {
		card, ok := g.eagerDomainCardinality(s)
		if !ok || card <= 0 {
			return false
		}
		if product <= goLargeThresh {
			product *= card
		}
	}
	return product <= goLargeThresh
}

func (g *Generator) eagerDomainCardinality(s goivy.Sort) (int, bool) {
	if vals, ok := goLiteralFiniteValueExprs(s); ok {
		return len(vals), len(vals) > 0
	}
	if dim, ok := g.goFiniteIndexDimension(s); ok {
		return dim, dim > 0
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok || hi < lo {
			return 0, false
		}
		return hi - lo + 1, true
	}
	vals, ok := g.finiteValueExprs(s)
	return len(vals), ok && len(vals) > 0
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

func (g *Generator) hasStrBVInterp(s goivy.Sort) bool {
	it, ok := g.goInterpType(s)
	return ok && it.Kind == goInterpStrBV
}

func (g *Generator) hasStringValuedInterp(s goivy.Sort) bool {
	return g.hasStringInterp(s) || g.hasStrBVInterp(s)
}

func (g *Generator) hasNatInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "nat"
}

func (g *Generator) hasIntInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "int"
}

func (g *Generator) hasIntOrNatInterp(s goivy.Sort) bool {
	return g.hasIntInterp(s) || g.hasNatInterp(s)
}

func (g *Generator) isNativeTypeSort(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || g.Mod.NativeTypes == nil {
		return false
	}
	_, ok := g.Mod.NativeTypes[sortName(s)]
	return ok
}

func (g *Generator) nativeTypeSortTranslatesAsGoInt(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || g.Mod.NativeTypes == nil {
		return false
	}
	name := sortName(s)
	if name == "" {
		return false
	}
	return g.nativeTypeNameTranslatesAsGoInt(name, g.Mod.NativeTypes[name])
}

func (g *Generator) nativeTypeNameTranslatesAsGoInt(name string, nt *goivy.NativeType) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	return nativeTypeIsPlainInt(nt)
}

func nativeTypeIsPlainInt(nt *goivy.NativeType) bool {
	if nt == nil || len(nt.Elems) != 1 {
		return false
	}
	switch n := nt.Elems[0].(type) {
	case *goivy.NativeCode:
		return strings.TrimSpace(n.Code) == "int"
	case *goivy.Atom:
		return strings.TrimSpace(n.Rep) == "int"
	case *goivy.Symbol:
		return strings.TrimSpace(n.Rep) == "int"
	case *goivy.Const:
		return strings.TrimSpace(n.Name) == "int"
	default:
		return false
	}
}

func (g *Generator) isRuntimeHandleSort(s goivy.Sort) bool {
	return g.isRuntimeSocketSort(s)
}

func (g *Generator) isRuntimeSocketSort(s goivy.Sort) bool {
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return false
	}
	name := sortName(s)
	if name == "" {
		return false
	}
	if name == "socket" {
		return g.hasNativeSocketFactory("", name)
	}
	if !strings.HasSuffix(name, ".socket") {
		return false
	}
	return g.hasNativeSocketFactory(strings.TrimSuffix(name, ".socket"), name)
}

func (g *Generator) hasNativeSocketFactory(prefix, socketSortName string) bool {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return false
	}
	names := []string{"open", "connect"}
	if prefix != "" {
		names = []string{prefix + ".open", prefix + ".connect"}
	}
	for _, baseName := range names {
		for _, name := range []string{baseName, "ext:" + baseName} {
			act, ok := g.Mod.Actions.Get2(name)
			if !ok || act == nil {
				continue
			}
			if actionReturnsSortName(act, socketSortName) && actionContainsNativeAction(act) {
				return true
			}
		}
	}
	if prefix == "" {
		for name, act := range g.Mod.Actions.All() {
			if act == nil {
				continue
			}
			actionName := strings.TrimPrefix(name, "ext:")
			if !strings.HasSuffix(actionName, ".open") && !strings.HasSuffix(actionName, ".connect") {
				continue
			}
			if actionReturnsSortName(act, socketSortName) && actionContainsNativeAction(act) {
				return true
			}
		}
	}
	return false
}

func actionReturnsSortName(act goivy.Action, sortNameText string) bool {
	if act == nil {
		return false
	}
	for _, ret := range act.GetFormalReturns() {
		if ret != nil && sortName(ret.CSort) == sortNameText {
			return true
		}
	}
	return false
}

func actionContainsNativeAction(act goivy.Action) bool {
	if act == nil {
		return false
	}
	if _, ok := act.(*goivy.LogicNativeAction); ok {
		return true
	}
	iter, ok := act.(interface {
		IterSubactions() []goivy.ActionsAction
	})
	if !ok {
		return false
	}
	for _, sub := range iter.IterSubactions() {
		if _, ok := sub.(*goivy.LogicNativeAction); ok {
			return true
		}
	}
	return false
}

func (g *Generator) uninterpretedTestGeneratorError(s goivy.Sort) error {
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return nil
	}
	if _, ok := g.rangeSortFor(s); ok {
		return nil
	}
	if _, ok := g.goInterpType(s); ok {
		return nil
	}
	if g.hasIntOrNatInterp(s) || g.hasStringInterp(s) {
		return nil
	}
	if g.isVariantSuperName(sortName(s)) || g.isVariantSubtypeName(sortName(s)) || g.destructorStructFields(sortName(s)) != nil {
		return nil
	}
	if g.isNativeTypeSort(s) {
		return nil
	}
	if g.isRuntimeHandleSort(s) {
		return nil
	}
	return fmt.Errorf("ivy2golang: cannot create test generator because type %s is uninterpreted", sortName(s))
}

func (g *Generator) testGeneratorZeroFallbackError(s goivy.Sort) error {
	if g == nil || !(g.Config.Build && (g.Config.Target == "test" || g.Config.Target == "gen")) {
		return nil
	}
	if err := g.uninterpretedTestGeneratorError(s); err != nil {
		return err
	}
	if g.isNativeTypeSort(s) || g.isRuntimeHandleSort(s) {
		return nil
	}
	return fmt.Errorf("ivy2golang: cannot create test generator because type %s is non-enumerable", sortName(s))
}

func (g *Generator) sortCard(s goivy.Sort) int {
	if it, ok := g.goInterpType(s); ok {
		if card := it.card(); card > 0 {
			return card
		}
	}
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
	infos := g.destructorStructFieldInfos(name)
	if len(infos) == 0 {
		return nil
	}
	out := make([]*goivy.Const, 0, len(infos))
	for _, info := range infos {
		out = append(out, info.Const)
	}
	return out
}

func (g *Generator) destructorStructFieldInfos(name string) []goDestructorField {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil || name == "" {
		return nil
	}
	fields, ok := g.Mod.SortDestructors.Get2(name)
	if !ok || len(fields) == 0 {
		return nil
	}
	out := make([]goDestructorField, 0, len(fields))
	for _, d := range fields {
		if d == nil {
			return nil
		}
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) == 0 || sortName(fs.Domain()[0]) != name {
			return nil
		}
		if !g.destructorFieldExtraDomainsRepresentable(fs) {
			continue
		}
		out = append(out, goDestructorField{
			Const:     d,
			Sort:      fs,
			FieldName: goName(memName(d.Name)),
		})
	}
	return out
}

func (g *Generator) destructorFieldExtraDomainsRepresentable(fs *goivy.LogicFunctionSort) bool {
	if fs == nil || len(fs.Domain()) == 0 {
		return false
	}
	domain := fs.Domain()[1:]
	if len(domain) == 0 {
		return true
	}
	if !g.canEnumerateDomain(domain) {
		return g.goFunctionStorageFor(domain, fs.Range()).Large
	}
	for _, s := range domain {
		if _, ok := g.goFiniteIndexDimension(s); !ok {
			return false
		}
	}
	return true
}

func (g *Generator) goDestructorFieldType(fs *goivy.LogicFunctionSort) (string, bool) {
	if fs == nil || len(fs.Domain()) == 0 {
		return "", false
	}
	typ := g.goScalarType(fs.Range())
	domain := fs.Domain()[1:]
	if len(domain) > 0 && !g.canEnumerateDomain(domain) {
		st := g.goFunctionStorageFor(domain, fs.Range())
		if !st.Large {
			return "", false
		}
		return "*" + st.Type, true
	}
	for i := len(domain) - 1; i >= 0; i-- {
		dim, ok := g.goFiniteIndexDimension(domain[i])
		if !ok {
			return "", false
		}
		typ = fmt.Sprintf("[%d]%s", dim, typ)
	}
	return typ, true
}

func (g *Generator) destructorFieldIsLarge(field goDestructorField) bool {
	if field.Sort == nil {
		return false
	}
	domain := field.Sort.Domain()
	if len(domain) <= 1 {
		return false
	}
	return !g.canEnumerateDomain(domain[1:])
}

func (g *Generator) destructorStructNeedsEqualMethod(fields []goDestructorField) bool {
	for _, field := range fields {
		if field.Sort == nil {
			continue
		}
		if g.destructorFieldIsLarge(field) {
			return true
		}
		if g.sortNeedsCustomEquality(field.Sort.Range()) {
			return true
		}
	}
	return false
}

func (g *Generator) sortNeedsCustomEquality(s goivy.Sort) bool {
	return g.sortNeedsCustomEqualitySeen(s, map[string]bool{})
}

func (g *Generator) sortNeedsCustomEqualitySeen(s goivy.Sort, seen map[string]bool) bool {
	if s == nil {
		return false
	}
	name := sortName(s)
	if name == "" || seen[name] {
		return false
	}
	if g.isVariantSuperName(name) {
		return true
	}
	fields := g.destructorStructFieldInfos(name)
	if len(fields) == 0 {
		return false
	}
	seen[name] = true
	for _, field := range fields {
		if field.Sort == nil {
			continue
		}
		if g.destructorFieldIsLarge(field) {
			return true
		}
		if g.sortNeedsCustomEqualitySeen(field.Sort.Range(), seen) {
			return true
		}
	}
	return false
}

func (g *Generator) goFiniteIndexDimension(s goivy.Sort) (int, bool) {
	switch st := s.(type) {
	case *goivy.LogicEnumeratedSort:
		if len(st.Extension) == 0 {
			return 0, false
		}
		if !isNumericEnum(st) {
			return len(st.Extension), true
		}
		max := -1
		for _, elem := range st.Extension {
			n, err := strconv.Atoi(elem)
			if err != nil || n < 0 {
				return 0, false
			}
			if n > max {
				max = n
			}
		}
		return max + 1, true
	case *goivy.RangeSort:
		lo, hi, ok := numericRangeBounds(st)
		if !ok || lo < 0 || hi < lo {
			return 0, false
		}
		return hi + 1, true
	case *goivy.UninterpretedSort:
		if rs, ok := g.rangeSortFor(st); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok || lo < 0 || hi < lo {
				return 0, false
			}
			return hi + 1, true
		}
		if !g.hasStringValuedInterp(st) && !g.isVariantSuperName(st.Name) {
			if card := g.sortCard(st); card > 0 {
				return card, true
			}
		}
	}
	return 0, false
}

func (g *Generator) destructorRecordFields(name string) []*goivy.Const {
	if g.isVariantSuperName(name) || g.isVariantSubtypeName(name) {
		return nil
	}
	return g.destructorStructFields(name)
}

func (g *Generator) destructorFieldName(name string) (string, bool) {
	field, ok := g.destructorFieldInfo(name)
	if !ok {
		return "", false
	}
	return field.FieldName, true
}

func (g *Generator) destructorFieldInfo(name string) (goDestructorField, bool) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil || name == "" {
		return goDestructorField{}, false
	}
	sort, ok := g.Mod.DestructorSorts[name]
	if !ok {
		return goDestructorField{}, false
	}
	for _, field := range g.destructorStructFieldInfos(sortName(sort)) {
		if field.Const != nil && field.Const.Name == name {
			return field, true
		}
	}
	return goDestructorField{}, false
}

func (g *Generator) destructorFieldAccess(name string, args []string) (string, bool, error) {
	field, ok := g.destructorFieldInfo(name)
	if !ok {
		return "", false, nil
	}
	if len(args) != len(field.Sort.Domain()) {
		return "", true, fmt.Errorf("ivy2golang: destructor %s expected %d arguments, got %d", name, len(field.Sort.Domain()), len(args))
	}
	expr := args[0] + "." + field.FieldName
	if g.destructorFieldIsLarge(field) {
		key := g.goMapKeyValue(field.Sort.Domain()[1:], args[1:])
		return fmt.Sprintf("ivyThunkGet(%s, %s, %s)", expr, key, g.goZeroValue(field.Sort.Range())), true, nil
	}
	if len(args) > 1 {
		expr += goIndexSuffix(args[1:])
	}
	return expr, true, nil
}

func (g *Generator) goDestructorFieldSet(app *goivy.Apply, name string, rhs string) (string, bool, error) {
	field, ok := g.destructorFieldInfo(name)
	if !ok || !g.destructorFieldIsLarge(field) {
		return "", false, nil
	}
	if len(app.Terms) != len(field.Sort.Domain()) {
		return "", true, fmt.Errorf("ivy2golang: destructor %s expected %d arguments, got %d", name, len(field.Sort.Domain()), len(app.Terms))
	}
	obj, err := g.emitExpr(app.Terms[0])
	if err != nil {
		return "", true, err
	}
	args := make([]string, len(app.Terms)-1)
	for i, term := range app.Terms[1:] {
		code, err := g.emitExpr(term)
		if err != nil {
			return "", true, err
		}
		args[i] = code
	}
	key := g.goMapKeyValue(field.Sort.Domain()[1:], args)
	return fmt.Sprintf("ivyThunkSet(&%s.%s, %s, %s, %s)", obj, field.FieldName, key, rhs, g.goZeroValue(field.Sort.Range())), true, nil
}

func goIndexSuffix(args []string) string {
	var b strings.Builder
	for _, arg := range args {
		b.WriteByte('[')
		b.WriteString(arg)
		b.WriteByte(']')
	}
	return b.String()
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

func (g *Generator) goEqualExpr(left, right string, s goivy.Sort) string {
	if g.sortNeedsCustomEquality(s) {
		return fmt.Sprintf("(%s).Equal(%s)", left, right)
	}
	return fmt.Sprintf("%s == %s", left, right)
}
