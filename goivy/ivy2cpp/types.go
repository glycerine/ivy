package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

const largeThresh = 1024
const maxUint32Int = int(^uint32(0))

type cppStorageKind int

const (
	cppStorageScalar cppStorageKind = iota
	cppStorageArray
	cppStorageHashThunk
)

type cppFunctionStorage struct {
	Kind      cppStorageKind
	Domain    []goivy.Sort
	Range     goivy.Sort
	RangeType string
	Dims      []int
	KeyType   string
	Type      string
}

func cppType(s goivy.Sort) string {
	return cppTypeWith(nil, s, "")
}

func cppTypeWith(g *Generator, s goivy.Sort, className string) string {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), className)
		if st.Kind == cppStorageArray {
			return st.RangeType + cppArraySuffix(st.Dims)
		}
		return st.Type
	}
	return cppScalarTypeWith(g, s, className)
}

func cppScalarTypeWith(g *Generator, s goivy.Sort, className string) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			return "int"
		}
		if st.Name == "" {
			return "int"
		}
		if className != "" {
			return className + "::" + varName(st.Name)
		}
		return varName(st.Name)
	case *goivy.RangeSort:
		return cppCardinalType(cppSortCard(g, st))
	case *goivy.UninterpretedSort:
		if g != nil {
			if typeName, ok := g.nativeTypeName(st, className); ok {
				return typeName
			}
			if name, ok := g.destructorStructName(st); ok {
				if className != "" {
					return className + "::" + varName(name)
				}
				return varName(name)
			}
			if name, ok := g.variantSuperName(st); ok {
				if className != "" {
					return className + "::" + varName(name)
				}
				return varName(name)
			}
			if name, ok := g.variantSubtypeName(st); ok {
				if className != "" {
					return className + "::" + varName(name)
				}
				return varName(name)
			}
			if typeName, ok := g.cppInterpTypeName(st, className); ok {
				return typeName
			}
			if g.hasStringInterp(st) {
				return "__strlit"
			}
			if g.hasNatInterp(st) {
				return "unsigned long long"
			}
		}
		card := cppSortCard(g, st)
		if card > 0 {
			return cppCardinalType(card)
		}
		return "int"
	default:
		return "long long"
	}
}

func cppFunctionType(s *goivy.LogicFunctionSort) string {
	return cppType(s)
}

func cppQualifiedType(s goivy.Sort, className string) string {
	if className == "" {
		return cppType(s)
	}
	return cppTypeWith(nil, s, className)
}

func cppQualifiedFunctionType(s *goivy.LogicFunctionSort, className string) string {
	return cppQualifiedType(s, className)
}

func (g *Generator) cppType(s goivy.Sort) string {
	return cppTypeWith(g, s, "")
}

func (g *Generator) cppQualifiedType(s goivy.Sort, className string) string {
	return cppTypeWith(g, s, className)
}

func cppCardinalType(card int) string {
	if card > 0 && card <= maxUint32Int {
		return "unsigned"
	}
	return "unsigned long long"
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

func cppFunctionStorageFor(g *Generator, domain []goivy.Sort, rng goivy.Sort, className string) cppFunctionStorage {
	rangeType := cppScalarTypeWith(g, rng, className)
	st := cppFunctionStorage{Kind: cppStorageScalar, Domain: domain, Range: rng, RangeType: rangeType, Type: rangeType}
	if len(domain) == 0 {
		return st
	}
	dims := make([]int, len(domain))
	product := 1
	allCards := true
	allIntegerLike := true
	for i, d := range domain {
		card := cppArrayDim(g, d)
		if card <= 0 {
			allCards = false
		} else {
			dims[i] = card
			if product <= largeThresh {
				product *= card
			}
		}
		if !cppIsAnyIntegerType(g, d) {
			allIntegerLike = false
		}
	}
	if allCards && allIntegerLike && product <= largeThresh {
		st.Kind = cppStorageArray
		st.Dims = dims
		st.Type = rangeType + cppArraySuffix(dims)
		return st
	}
	st.Kind = cppStorageHashThunk
	st.KeyType = cppCTupleNameWith(g, domain, className)
	st.Type = fmt.Sprintf("hash_thunk<%s,%s>", st.KeyType, rangeType)
	return st
}

func cppArraySuffix(dims []int) string {
	var b strings.Builder
	for _, d := range dims {
		b.WriteString("[")
		b.WriteString(strconv.Itoa(d))
		b.WriteString("]")
	}
	return b.String()
}

func cppCTupleName(domain []goivy.Sort, className string) string {
	return cppCTupleNameWith(nil, domain, className)
}

func cppCTupleNameWith(g *Generator, domain []goivy.Sort, className string) string {
	if len(domain) == 1 {
		return cppScalarTypeWith(g, domain[0], className)
	}
	parts := make([]string, len(domain))
	for i, s := range domain {
		part := cppScalarTypeWith(g, s, "")
		part = strings.ReplaceAll(part, " ", "_")
		if idx := strings.LastIndex(part, "::"); idx >= 0 {
			part = part[idx+2:]
		}
		parts[i] = part
	}
	name := "__tup__" + strings.Join(parts, "__")
	if className != "" {
		return className + "::" + name
	}
	return name
}

func cppCTupleLocalName(domain []goivy.Sort) string {
	return cppCTupleLocalNameWith(nil, domain)
}

func cppCTupleLocalNameWith(g *Generator, domain []goivy.Sort) string {
	if len(domain) == 1 {
		return cppScalarTypeWith(g, domain[0], "")
	}
	parts := make([]string, len(domain))
	for i, s := range domain {
		part := cppScalarTypeWith(g, s, "")
		part = strings.ReplaceAll(part, " ", "_")
		parts[i] = part
	}
	return "__tup__" + strings.Join(parts, "__")
}

func cppHashType(g *Generator, s goivy.Sort) string {
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return "int"
	}
	return cppScalarTypeWith(g, s, "")
}

func cppSortCard(g *Generator, s goivy.Sort) int {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return 2
	case *goivy.LogicEnumeratedSort:
		return len(st.Extension)
	case *goivy.RangeSort:
		if _, hi, ok := numericRangeBoundsInt(st); ok {
			return hi + 1
		}
	case *goivy.UninterpretedSort:
		if g != nil && g.Mod != nil {
			if it, ok := g.cppInterpType(st); ok {
				return it.card()
			}
			if rs, ok := g.rangeSortFor(st); ok {
				if _, hi, ok := numericRangeBoundsInt(rs); ok {
					return hi + 1
				}
			}
			if g.Mod.Cfg != nil {
				if card := g.Mod.SortCard(st); card > 0 {
					return card
				}
			}
			if g.Mod.Sig != nil {
				if card := goivy.SortCard(st, g.Mod.Sig); card > 0 {
					return card
				}
			}
		}
	}
	if g != nil && g.Mod != nil && g.Mod.Sig != nil {
		if card := goivy.SortCard(s, g.Mod.Sig); card > 0 {
			return card
		}
	}
	return -1
}

func cppArrayDim(g *Generator, s goivy.Sort) int {
	return cppSortCard(g, s)
}

func cppIsAnyIntegerType(g *Generator, s goivy.Sort) bool {
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return true
	}
	switch cppScalarTypeWith(g, s, "") {
	case "bool", "int", "long long", "unsigned", "unsigned long long":
		return true
	default:
		return false
	}
}

func numericRangeBoundsInt(rs *goivy.RangeSort) (int, int, bool) {
	loText, hiText, ok := numericRangeBounds(rs)
	if !ok {
		return 0, 0, false
	}
	lo, err := strconv.Atoi(loText)
	if err != nil {
		return 0, 0, false
	}
	hi, err := strconv.Atoi(hiText)
	if err != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

func cppZeroValue(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "false"
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			if len(st.Extension) > 0 {
				return st.Extension[0]
			}
			return "0"
		}
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
		if isNumericEnum(st) {
			if len(st.Extension) > 0 {
				return st.Extension[0]
			}
			return "0"
		}
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
		if typeName, ok := g.nativeTypeName(s, ""); ok {
			return typeName + "()"
		}
		if typeName, ok := g.cppInterpTypeName(s, ""); ok {
			if strings.Contains(typeName, " ") {
				return "0"
			}
			return typeName + "()"
		}
		if g.hasStringInterp(s) {
			return "__strlit()"
		}
		if name, ok := g.destructorStructName(s); ok {
			return varName(name) + "()"
		}
		if name, ok := g.variantSuperName(s); ok {
			return varName(name) + "()"
		}
		if name, ok := g.variantSubtypeName(s); ok {
			return varName(name) + "()"
		}
	}
	return cppZeroValue(s)
}

func (g *Generator) cppZeroValueInScope(s goivy.Sort) string {
	if g != nil {
		if typeName, ok := g.nativeTypeName(s, g.ClassName); ok {
			return typeName + "()"
		}
		if typeName, ok := g.cppInterpTypeName(s, g.ClassName); ok {
			if strings.Contains(typeName, " ") {
				return "0"
			}
			return typeName + "()"
		}
		if g.hasStringInterp(s) {
			return "__strlit()"
		}
		if name, ok := g.destructorStructName(s); ok {
			typeName := varName(name)
			if g.ClassName != "" {
				typeName = g.ClassName + "::" + typeName
			}
			return typeName + "()"
		}
		if name, ok := g.variantSuperName(s); ok {
			typeName := varName(name)
			if g.ClassName != "" {
				typeName = g.ClassName + "::" + typeName
			}
			return typeName + "()"
		}
		if name, ok := g.variantSubtypeName(s); ok {
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

func (g *Generator) nativeTypeName(s goivy.Sort, className string) (string, bool) {
	if g == nil || g.Mod == nil {
		return "", false
	}
	name := sortName(s)
	if name == "" {
		return "", false
	}
	if _, ok := g.nativeTypeForSort(name); !ok {
		return "", false
	}
	typeName := varName(name)
	if className != "" {
		typeName = className + "::" + typeName
	}
	return typeName, true
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

func (g *Generator) variantSuperName(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil {
		return "", false
	}
	name := sortName(s)
	if name == "" || !g.isVariantSuperName(name) {
		return "", false
	}
	return name, true
}

func (g *Generator) variantSubtypeName(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil {
		return "", false
	}
	name := sortName(s)
	if name == "" || !g.isVariantSubtypeName(name) {
		return "", false
	}
	return name, true
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

func (g *Generator) cppStorageDecl(name string, s goivy.Sort, className string) string {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.cppFunctionStorageDecl(name, fs.Domain(), fs.Range(), className)
	}
	return fmt.Sprintf("%s %s", cppScalarTypeWith(g, s, className), varName(name))
}

func (g *Generator) cppFunctionStorageDecl(name string, domain []goivy.Sort, rng goivy.Sort, className string) string {
	st := cppFunctionStorageFor(g, domain, rng, className)
	switch st.Kind {
	case cppStorageArray:
		return fmt.Sprintf("%s %s%s", st.RangeType, varName(name), cppArraySuffix(st.Dims))
	default:
		return fmt.Sprintf("%s %s", st.Type, varName(name))
	}
}

func (g *Generator) cppStorageParamDecl(c *goivy.Const, className string) string {
	if c == nil {
		return ""
	}
	return g.cppStorageDecl(c.Name, c.CSort, className)
}

// cppDestructorFieldAccess is the sibling of cppStorageAccess that takes
// the destructor field's domain (excluding the implicit struct receiver)
// and range directly. Mirrors Python emit_app destructor branch at
// ivy_to_cpp.py:3187-3221.
func (g *Generator) cppDestructorFieldAccess(field string, dom []goivy.Sort, rng goivy.Sort, args []string, obj string) string {
	base := varName(field)
	if obj != "" {
		base = obj + "." + base
	}
	if len(dom) == 0 {
		return base
	}
	st := cppFunctionStorageFor(g, dom, rng, "")
	switch st.Kind {
	case cppStorageArray:
		return base + cppIndexSuffix(args)
	case cppStorageHashThunk:
		if len(args) == 1 {
			return fmt.Sprintf("%s[%s]", base, args[0])
		}
		return fmt.Sprintf("%s[%s(%s)]", base, cppCTupleLocalNameWith(g, dom), strings.Join(args, ", "))
	default:
		return base
	}
}

func (g *Generator) cppStorageAccess(name string, sort goivy.Sort, args []string, obj string) string {
	base := varName(name)
	if obj != "" {
		base = obj + "." + base
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		return base
	}
	st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
	switch st.Kind {
	case cppStorageArray:
		return base + cppIndexSuffix(args)
	case cppStorageHashThunk:
		if len(args) == 1 {
			return fmt.Sprintf("%s[%s]", base, args[0])
		}
		return fmt.Sprintf("%s[%s(%s)]", base, cppCTupleLocalNameWith(g, fs.Domain()), strings.Join(args, ", "))
	default:
		return base
	}
}

func cppIndexSuffix(args []string) string {
	var b strings.Builder
	for _, a := range args {
		b.WriteString("[")
		b.WriteString(a)
		b.WriteString("]")
	}
	return b.String()
}

func (g *Generator) cppCTuples() [][]goivy.Sort {
	seen := map[string]bool{}
	var out [][]goivy.Sort
	addSort := func(s goivy.Sort) {
		fs, ok := s.(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) <= 1 {
			return
		}
		st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
		if st.Kind != cppStorageHashThunk {
			return
		}
		key := cppCTupleLocalNameWith(g, fs.Domain())
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, fs.Domain())
	}
	for _, sym := range g.stateSymbols() {
		addSort(sym.Sort)
	}
	for _, p := range g.progressDecls() {
		if len(p.Vars) <= 1 {
			continue
		}
		domain := make([]goivy.Sort, len(p.Vars))
		for i, v := range p.Vars {
			domain[i] = v.VSort
		}
		st := progressCounterStorage(g, domain)
		if st.Kind == cppStorageHashThunk {
			key := cppCTupleLocalNameWith(g, domain)
			if !seen[key] {
				seen[key] = true
				out = append(out, domain)
			}
		}
	}
	for _, p := range g.Mod.Params {
		addSort(p.CSort)
	}
	if g.Mod.Actions != nil {
		for _, act := range g.Mod.Actions.All() {
			for _, p := range act.GetFormalParams() {
				addSort(p.CSort)
			}
			for _, r := range act.GetFormalReturns() {
				addSort(r.CSort)
			}
		}
	}
	return out
}
