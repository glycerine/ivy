package ivy2go

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// largeThresh mirrors ivy2cpp/cpp_types.go largeThresh. Domain
// products above this threshold lower to a map instead of a flat array.
const largeThresh = 1024

// goStorageKind mirrors ivy2cpp/cpp_types.go cppStorageKind. The
// scalar / array / hash-thunk decision is preserved; the lowering to
// Go differs (see ARCHITECTURE_TODO.md §3.5.1).
type goStorageKind int

const (
	goStorageScalar goStorageKind = iota
	goStorageArray
	goStorageHashThunk
)

// goFunctionStorage mirrors ivy2cpp/cpp_types.go cppFunctionStorage.
type goFunctionStorage struct {
	Kind      goStorageKind
	Domain    []goivy.Sort
	Range     goivy.Sort
	RangeType string
	Dims      []int
	KeyType   string
	Type      string
}

// goType returns a Go type expression for s. Mirrors ivy2cpp/types.go
// cppType. Use only when no Generator context is available; otherwise
// call Generator.goType so destructor/variant/interp/native lookups
// succeed.
func goType(s goivy.Sort) string {
	return goTypeWith(nil, s)
}

func (g *Generator) goType(s goivy.Sort) string {
	return goTypeWith(g, s)
}

func goTypeWith(g *Generator, s goivy.Sort) string {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
		switch st.Kind {
		case goStorageArray:
			// Go array prefix order is the same as C++ array suffix
			// when expressed as `[d0][d1]…RangeType`.
			return goArrayPrefix(st.Dims) + st.RangeType
		case goStorageHashThunk:
			// M2 still represents large function storage as a map;
			// M8 will replace with a thunk struct for the
			// hash-thunk path. Until then, map[K]V is correct and
			// gofmt-clean.
			return fmt.Sprintf("map[%s]%s", st.KeyType, st.RangeType)
		default:
			return st.RangeType
		}
	}
	return goScalarTypeWith(g, s)
}

// goScalarType returns the Go type for a non-function (scalar) sort.
// Mirrors ivy2cpp cppScalarTypeWith.
func (g *Generator) goScalarType(s goivy.Sort) string {
	return goScalarTypeWith(g, s)
}

func goScalarTypeWith(g *Generator, s goivy.Sort) string {
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
		return goExportedName(st.Name)
	case *goivy.RangeSort:
		if st.Name != "" {
			return goExportedName(st.Name)
		}
		return goCardinalType(goSortCard(g, st))
	case *goivy.UninterpretedSort:
		if g != nil {
			if typeName, ok := g.nativeTypeName(st); ok {
				return typeName
			}
			if name, ok := g.destructorStructName(st); ok {
				return goExportedName(name)
			}
			if name, ok := g.variantSuperName(st); ok {
				return goExportedName(name)
			}
			if name, ok := g.variantSubtypeName(st); ok {
				if g.isPlainVariantSubtypeName(name) {
					return "int"
				}
				return goExportedName(name)
			}
			if typeName, ok := g.goInterpTypeName(st); ok {
				return typeName
			}
			if g.hasStringInterp(st) {
				return "string"
			}
			if g.hasNatInterp(st) {
				return "uint64"
			}
			// Range-interpreted UninterpretedSort (e.g. `type idx
			// = {0..7}`): emitSortDecls declared `type Idx int`,
			// so use that named type rather than falling back to
			// goCardinalType. This keeps fields, params, and map
			// keys spellable via the declared name.
			if _, ok := g.rangeSortFor(st); ok && st.Name != "" {
				return goExportedName(st.Name)
			}
		}
		if st.Name != "" {
			return goExportedName(st.Name)
		}
		card := goSortCard(g, st)
		if card > 0 {
			return goCardinalType(card)
		}
		return "int"
	default:
		return "int64"
	}
}

// goArrayPrefix builds the `[d0][d1]...` prefix for a Go array type
// of the given dimensions. Inverse of ivy2cpp/cpp_types.go
// cppArraySuffix (which builds `[d0][d1]...` as a *suffix* because
// C++ puts type before name).
func goArrayPrefix(dims []int) string {
	var b strings.Builder
	for _, d := range dims {
		b.WriteString("[")
		b.WriteString(strconv.Itoa(d))
		b.WriteString("]")
	}
	return b.String()
}

// goCardinalType mirrors ivy2cpp/types.go cppCardinalType.
//   - Cardinality fits in uint32      → "uint32"
//   - Otherwise                        → "uint64"
func goCardinalType(card int) string {
	if card > 0 && card <= int(^uint32(0)) {
		return "uint32"
	}
	return "uint64"
}

// isNumericEnum mirrors ivy2cpp/types.go isNumericEnum.
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

// goFunctionStorageFor mirrors ivy2cpp/types.go cppFunctionStorageFor.
// Decides between scalar / array / hash-thunk storage for a function
// sort, matching the C++ decision tree so future cross-language audits
// stay tractable.
func goFunctionStorageFor(g *Generator, domain []goivy.Sort, rng goivy.Sort) goFunctionStorage {
	rangeType := goScalarTypeWith(g, rng)
	st := goFunctionStorage{Kind: goStorageScalar, Domain: domain, Range: rng, RangeType: rangeType, Type: rangeType}
	if len(domain) == 0 {
		return st
	}
	dims := make([]int, len(domain))
	product := 1
	allCards := true
	allIntegerLike := true
	for i, d := range domain {
		card := goArrayDim(g, d)
		if card <= 0 {
			allCards = false
		} else {
			dims[i] = card
			if product <= largeThresh {
				product *= card
			}
		}
		if !goIsAnyIntegerType(g, d) {
			allIntegerLike = false
		}
	}
	if allCards && allIntegerLike && product <= largeThresh {
		st.Kind = goStorageArray
		st.Dims = dims
		st.Type = goArrayPrefix(dims) + rangeType
		return st
	}
	st.Kind = goStorageHashThunk
	st.KeyType = goCTupleNameWith(g, domain)
	st.Type = fmt.Sprintf("map[%s]%s", st.KeyType, rangeType)
	return st
}

// goIsAnyIntegerType mirrors ivy2cpp/types.go cppIsAnyIntegerType.
// Returns true when the sort lowers to a Go type usable as an array
// index (so we can pack it into a flat array).
//
// Named integer-backed types (e.g. `type Idx int` for a range sort)
// count too: although their declared name isn't in the primitive
// switch below, the underlying repr is integer.
func goIsAnyIntegerType(g *Generator, s goivy.Sort) bool {
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return true
	}
	if _, ok := s.(*goivy.RangeSort); ok {
		return true
	}
	if g != nil {
		if it, ok := g.goInterpType(s); ok && it.Kind == goInterpBV {
			return true
		}
		if _, ok := g.rangeSortFor(s); ok {
			return true
		}
	}
	switch goScalarTypeWith(g, s) {
	case "bool", "int", "int64",
		"uint8", "uint16", "uint32", "uint64", "Uint128":
		return true
	}
	return false
}

// goSortCard mirrors ivy2cpp/types.go cppSortCard.
func goSortCard(g *Generator, s goivy.Sort) int {
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
			if it, ok := g.goInterpType(st); ok {
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

// goArrayDim mirrors ivy2cpp/types.go cppArrayDim.
func goArrayDim(g *Generator, s goivy.Sort) int {
	return goSortCard(g, s)
}

// goZeroValue mirrors ivy2cpp/types.go cppZeroValue (the
// no-generator variant). M5 will refine with the destructor/variant
// branches.
func goZeroValue(s goivy.Sort) string {
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
			return goExportedName(st.Extension[0])
		}
		return "0"
	default:
		return "0"
	}
}

// goZeroValue (Generator method) mirrors ivy2cpp/types.go
// (g *Generator).cppZeroValue. M2 implements just the basic path —
// destructor/variant/interp branches are added in M5/M8/M3.
func (g *Generator) goZeroValue(s goivy.Sort) string {
	if g != nil {
		if typeName, ok := g.goInterpTypeName(s); ok {
			switch typeName {
			case "uint32", "uint64", "uint8", "uint16", "int", "int64", "bool":
				return "0"
			}
			if typeName == "Uint128" {
				return "Uint128{}"
			}
			if typeName == "*big.Int" {
				return "new(big.Int)"
			}
			return typeName + "{}"
		}
		if g.hasStringInterp(s) {
			return `""`
		}
		if g.hasNatInterp(s) {
			return "0"
		}
	}
	return goZeroValue(s)
}

// --- emitSortDecls ---------------------------------------------------

// emitSortDecls mirrors ivy2cpp/generator.go emitSortDecls. Walks
// SortOrder and emits the Go equivalent of each sort. Variant /
// destructor / native paths are stubbed via the empty hooks in
// variant.go / destructor.go / native.go and become functional in
// later milestones.
func (g *Generator) emitSortDecls(w *goWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return
	}
	emitted := map[string]bool{}
	visiting := map[string]bool{}

	var emitOne func(name string, force bool)
	emitOne = func(name string, force bool) {
		if name == "" || name == "bool" || emitted[name] {
			return
		}
		if visiting[name] {
			return
		}
		visiting[name] = true
		for _, dep := range g.sortDeclDependencyNames(name) {
			emitOne(dep, true)
		}
		visiting[name] = false

		// Variant super / destructor / native / variant-leaf paths
		// are no-ops for M2 (their hooks return false); M8 fills
		// them. Plain-variant-subtypes are emitted nothing too.
		if g.isVariantSuperName(name) {
			emitted[name] = true
			g.emitVariantSuperStruct(w, name)
			return
		}
		if nt, ok := g.nativeTypeForSort(name); ok {
			emitted[name] = true
			g.emitNativeTypeDecl(w, name, nt)
			return
		}
		if g.Mod.SortDestructors != nil {
			if _, ok := g.Mod.SortDestructors.Get2(name); ok {
				emitted[name] = true
				g.emitDestructorStruct(w, name)
				return
			}
		}
		if g.isPlainVariantSubtypeName(name) {
			emitted[name] = true
			return
		}
		if g.isVariantSubtypeName(name) {
			emitted[name] = true
			g.emitVariantLeafStruct(w, name)
			return
		}
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			emitted[name] = true
			return
		}
		if !force && !g.sortNeededForGeneratedDecl(name) {
			return
		}
		emitted[name] = true
		if it, ok := g.goInterpType(s); ok {
			// Plain bv[N] with a Go primitive (uint32 / uint64 /
			// Uint128) doesn't need its own declared type — the
			// interp text steers consumers to the primitive
			// directly. Wider BVs need helper struct decls; that
			// emission is M3's job.
			if it.Kind == goInterpBV && it.primitiveType() != "" {
				return
			}
			g.emitGoTypeDecl(w, s, it)
			return
		}
		// Range-interpreted sorts (e.g. `type idx = {0..7}` →
		// UninterpretedSort named "idx" with a RangeSort interp)
		// don't produce a typedef in ivy2cpp. In Go we DO emit a
		// named type alias so consumers can reference `Idx` in
		// field types and parameter signatures — a deliberate
		// ivy2go divergence, documented in ARCHITECTURE_TODO.md
		// §3.5.1.
		if rs, ok := g.rangeSortFor(s); ok {
			if name := sortName(s); name != "" {
				g.emitRangeDeclNamed(w, name, rs)
			}
			return
		}
		if _, interpreted := g.Mod.Sig.Interp[name]; interpreted {
			// e.g., strlit / nat — these have a fixed mapping handled
			// by goScalarType. Nothing to declare.
			return
		}
		switch st := s.(type) {
		case *goivy.LogicEnumeratedSort:
			if isNumericEnum(st) {
				return
			}
			g.emitEnumDecl(w, st)
		case *goivy.RangeSort:
			if st.Name != "" {
				g.emitRangeDecl(w, st)
			}
		case *goivy.UninterpretedSort:
			if st.Name != "" {
				w.linef("type %s int", goExportedName(st.Name))
				w.blank()
			}
		}
	}

	for _, name := range g.Mod.SortOrder {
		emitOne(name, false)
	}
	g.emitCTupleDecls(w)
}

// goCTuples mirrors ivy2cpp/types.go cppCTuples, with the Go-specific
// addition that action parameters/returns/locals also need declarations:
// their map-backed function storage can mention the same tuple key
// types even though they are not State fields.
func (g *Generator) goCTuples() [][]goivy.Sort {
	if g == nil || g.Mod == nil {
		return nil
	}
	seen := map[string]bool{}
	var out [][]goivy.Sort
	addSort := func(s goivy.Sort) {
		fs, ok := s.(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) <= 1 {
			return
		}
		st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
		if st.Kind != goStorageHashThunk {
			return
		}
		key := goCTupleNameWith(g, fs.Domain())
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, fs.Domain())
	}
	for _, sym := range g.stateSymbols() {
		addSort(sym.Sort)
	}
	if g.Mod.Actions != nil {
		names := make([]string, 0)
		for name := range g.Mod.Actions.All() {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			act, _ := g.Mod.Actions.Get2(name)
			g.collectActionFunctionSorts(act, addSort)
		}
	}
	return out
}

func (g *Generator) collectActionFunctionSorts(act goivy.Action, add func(goivy.Sort)) {
	if act == nil || add == nil {
		return
	}
	for _, p := range act.GetFormalParams() {
		if p != nil {
			add(p.CSort)
		}
	}
	for _, r := range act.GetFormalReturns() {
		if r != nil {
			add(r.CSort)
		}
	}
	if local, ok := act.(*goivy.LogicLocalAction); ok {
		for _, loc := range local.Locals {
			if loc != nil {
				add(loc.NodeSort())
			}
		}
	}
	for _, arg := range act.ActionArgs() {
		child, ok := goivy.ToAction(arg)
		if !ok || child == act {
			continue
		}
		g.collectActionFunctionSorts(child, add)
	}
}

// emitCTupleDecls emits a Go struct declaration for each composite
// key shape returned by goCTuples. Used as the map-key type of
// hash-thunk-backed function symbols (e.g. `bvxor : byte * byte ->
// byte` lowers to `map[tup__uint32__uint32]uint32`).
//
// Mirrors ivy2cpp/generator.go emitCTupleDecls (line 625). Go's
// built-in map key comparability removes the need for a hash method
// — for the all-comparable-fields case at least; nested map/slice
// fields are not supported by the current storage classifier.
func (g *Generator) emitCTupleDecls(w *goWriter) {
	for _, dom := range g.goCTuples() {
		name := goCTupleNameWith(g, dom)
		w.linef("// %s is a composite map-key for a multi-arg function symbol.", name)
		w.open(fmt.Sprintf("type %s struct {", name))
		for i, s := range dom {
			w.linef("Arg%d %s", i, g.goType(s))
		}
		w.close("")
		w.blank()
	}
}

// emitEnumDecl emits a Go enum-like type for a non-numeric Ivy enum:
//
//	type Color int
//
//	const (
//		Red Color = iota
//		Green
//		Blue
//	)
//
// Mirrors ivy2cpp/generator.go emitSortDecls' enum branch (which emits
// `enum Color { Red, Green, Blue };`).
func (g *Generator) emitEnumDecl(w *goWriter, st *goivy.LogicEnumeratedSort) {
	typeName := goExportedName(st.Name)
	w.linef("type %s int", typeName)
	w.blank()
	if len(st.Extension) == 0 {
		return
	}
	w.open("const (")
	for i, v := range st.Extension {
		exported := goExportedName(v)
		if i == 0 {
			w.linef("%s %s = iota", exported, typeName)
		} else {
			w.line(exported)
		}
	}
	w.closeParen()
	w.blank()
	// String() — mirrors cpp's operator<< overload which prints
	// enum values as their original ivy symbolic name (e.g. "red"
	// not "0"). Required for byte-equivalent trace output against
	// the cpp-emitted binary.
	w.linef("func (e %s) String() string {", typeName)
	w.line("\tswitch e {")
	for i, v := range st.Extension {
		exported := goExportedName(v)
		w.linef("\tcase %s: return %q", exported, v)
		_ = i
	}
	w.line("\t}")
	w.linef("\treturn fmt.Sprintf(\"%s(%%d)\", int(e))", typeName)
	w.line("}")
	w.blank()
	g.Ctx.AddImport("types", "fmt", "")
}

// emitRangeDecl emits a typed alias for a named range sort. Range
// constraints (lo ≤ x ≤ hi) are enforced at action boundaries via
// ivyAssume; the Go type itself is just `int` with a comment so the
// declared bounds remain visible in source.
func (g *Generator) emitRangeDecl(w *goWriter, st *goivy.RangeSort) {
	g.emitRangeDeclNamed(w, st.Name, st)
}

// emitRangeDeclNamed is the shared helper used both for direct
// *RangeSort declarations and for UninterpretedSorts whose interp is
// a RangeSort. The latter is how `type idx = {0..7}` is represented
// after goivy's frontend.
func (g *Generator) emitRangeDeclNamed(w *goWriter, name string, rs *goivy.RangeSort) {
	if name == "" {
		return
	}
	typeName := goExportedName(name)
	if lo, hi, ok := numericRangeBounds(rs); ok {
		w.linef("// %s = Range[%s..%s]", typeName, lo, hi)
	}
	w.linef("type %s int", typeName)
	w.blank()
}

// emitGoTypeDecl emits the type declaration for an interpreted sort
// that has its own helper struct (strbv, intbv, huge bv). M2 emits a
// placeholder struct {} so the symbol resolves; M3 replaces the body
// with the actual lowering.
func (g *Generator) emitGoTypeDecl(w *goWriter, s goivy.Sort, it goInterpType) {
	name := sortName(s)
	if name == "" {
		return
	}
	typeName := goExportedName(name)
	switch it.Kind {
	case goInterpStrBV:
		w.linef("// %s: strbv[%d] — wide BV stored as bytes. Filled in by M3.", typeName, it.Bits)
		w.linef("type %s struct{}", typeName)
	case goInterpIntBV:
		w.linef("// %s: intbv[%d..%d] over bv[%d]. Filled in by M3.", typeName, it.Lo, it.Hi, it.Bits)
		w.linef("type %s int", typeName)
	case goInterpBV:
		// Wide (≥129) BV. M3 will replace with *big.Int wrapper.
		w.linef("// %s: bv[%d] — wide BV. Filled in by M3.", typeName, it.Bits)
		w.linef("type %s struct{}", typeName)
	}
	w.blank()
}

// sortDeclDependencyNames mirrors ivy2cpp/generator.go
// sortDeclDependencyNames.
func (g *Generator) sortDeclDependencyNames(name string) []string {
	if g == nil || g.Mod == nil || name == "" {
		return nil
	}
	seen := map[string]bool{}
	var deps []string
	for _, dep := range g.Mod.SortDependencies(name, false) {
		if dep == "" || dep == "bool" || dep == name || seen[dep] {
			continue
		}
		seen[dep] = true
		deps = append(deps, dep)
	}
	return deps
}

// sortNeededForGeneratedDecl mirrors ivy2cpp/generator.go same name.
// M2 keeps the permissive default: emit every sort that appears in
// SortOrder. M5+ may prune unused sorts based on isolate/COI analysis.
func (g *Generator) sortNeededForGeneratedDecl(name string) bool {
	_ = name
	return true
}
