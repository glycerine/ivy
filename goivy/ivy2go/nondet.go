package ivy2go

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// nondet.go is the literal port of ivy2cpp/nondet.go. The cpp version
// emits `___ivy_choose(0, "label", uniqueID)` calls that the runtime
// resolves either to a deterministic 0 (impl/repl) or to a solver-
// driven value (gen/test). ivy2go's emitted runtime exposes
// `ivyChoose(N)` returning an integer in [0, N) — labels and unique
// IDs aren't threaded through to the runtime because the test target
// drives input synthesis through goivy.Solver.GetModelClauses rather
// than per-choose-point solver hooks. The generator-side signatures
// keep the label / uniqueID arguments for cpp parity even though the
// emitted runtime calls drop them.

// mkNondet emits a single nondeterministic assignment to an existing
// variable `varExpr` of sort `sort`. Mirrors ivy2cpp/nondet.go:32.
// When sort is nil the assignment is `<varExpr> = ivyChoose(0)` (an
// `int` fallback matching cpp's `(int)___ivy_choose(0, …)` shape).
func (g *Generator) mkNondet(w *goWriter, varExpr string, _rng int, label string, uniqueID int64, sort goivy.Sort) {
	_ = label
	_ = uniqueID
	if sort == nil {
		w.linef("%s = ivyChoose(0)", varExpr)
		return
	}
	g.mkNondetWithGoType(w, varExpr, label, uniqueID, sort)
}

// mkNondetWithGoType is the Go counterpart of ivy2cpp/nondet.go:40
// mkNondetWithCType. Emits a typed ivyChoose call sized to the sort's
// cardinality, falling back to leaving the var at its zero value when
// the sort has no computable cardinality.
func (g *Generator) mkNondetWithGoType(w *goWriter, varExpr string, label string, uniqueID int64, sort goivy.Sort) {
	_ = label
	_ = uniqueID
	if _, isBool := sort.(*goivy.BooleanSort); isBool {
		w.linef("%s = ivyChoose(2) == 1", varExpr)
		return
	}
	if expr, ok := g.rangeChoiceExpr(sort, fmt.Sprintf("ivyChoose(%d)", goSortCard(g, sort))); ok {
		w.linef("%s = %s", varExpr, expr)
		return
	}
	card := goSortCard(g, sort)
	typeName := g.goType(sort)
	if card > 0 {
		w.linef("%s = %s(ivyChoose(%d))", varExpr, typeName, card)
		return
	}
	// No computable cardinality — leave at Go zero value.
	w.linef("// nondet skipped for %s (no cardinality)", typeName)
}

// mkNondetSym mirrors ivy2cpp/nondet.go:54. Emits nondet init for
// `local`, dispatching on whether the symbol is scalar, bounded-
// array function, hash-thunk function, or sort with destructors.
func (g *Generator) mkNondetSym(w *goWriter, local goivy.Expr, name string, uniqueID int64) {
	if local == nil {
		return
	}
	lhsName := goivy.ExprName(local)
	if lhsName == "" {
		lhsName = name
	}
	sort := local.NodeSort()

	skipSort := sort
	if fs, ok := sort.(*goivy.LogicFunctionSort); ok {
		skipSort = fs.Range()
	}
	if g.nondetSkipSort(skipSort) {
		return
	}
	g.mkNondetSymbolStorage(w, goIdent(lhsName), sort, name, uniqueID)
}

func (g *Generator) mkNondetSymbolStorage(w *goWriter, lhs string, sort goivy.Sort, name string, uniqueID int64) {
	fs, isFunc := sort.(*goivy.LogicFunctionSort)
	if !isFunc {
		g.mkNondetValue(w, lhs, sort, name, uniqueID)
		return
	}
	dom := fs.Domain()
	if len(dom) == 0 {
		g.mkNondetValueScoped(w, lhs, fs.Range(), name, uniqueID, "")
		return
	}
	st := goFunctionStorageFor(g, dom, fs.Range())
	if st.Kind == goStorageHashThunk {
		// Sparse maps default to nil; treat that as the empty function.
		// Solver-driven cell-by-cell init lands with the bigger M9 ports.
		return
	}
	// Bounded-array storage: nested loops over each domain dimension.
	indices := make([]string, len(dom))
	for i := range dom {
		indices[i] = fmt.Sprintf("__nd_i%d", i)
	}
	opened := 0
	for i, d := range dom {
		card := goSortCard(g, d)
		if card <= 0 {
			for j := 0; j < opened; j++ {
				w.line("}")
			}
			return
		}
		w.linef("for %s := 0; %s < %d; %s++ {", indices[i], indices[i], card, indices[i])
		opened++
	}
	cell := lhs
	for _, idx := range indices {
		cell += "[" + idx + "]"
	}
	g.mkNondetValue(w, cell, fs.Range(), name, uniqueID)
	for j := 0; j < opened; j++ {
		w.line("}")
	}
}

// mkNondetValue mirrors ivy2cpp/nondet.go:122.
func (g *Generator) mkNondetValue(w *goWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64) {
	g.mkNondetValueScoped(w, lhsExpr, sort, name, uniqueID, "")
}

// mkNondetValueScoped mirrors ivy2cpp/nondet.go:126. Recurses into
// destructor records, dispatches on variant supertype, and falls
// back to a typed ivyChoose for leaf sorts.
func (g *Generator) mkNondetValueScoped(w *goWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64, className string) {
	_ = className
	if sort == nil {
		return
	}
	if g.trackNondetSort(sort) {
		sn := sortName(sort)
		if g.enterNondetSort(sn) {
			if g.isVariantSuperName(sn) && g.mkNondetVariantBaseScoped(w, lhsExpr, sort, name, uniqueID, className) {
				return
			}
			w.linef("// nondet skipped for recursive sort %s", g.goType(sort))
			return
		}
		defer g.leaveNondetSort(sn)
	}
	if g.Mod != nil && g.Mod.SortDestructors != nil {
		if _, ok := g.Mod.SortDestructors.Get2(sortName(sort)); ok {
			g.mkNondetStructFieldsScoped(w, lhsExpr, sort, name, uniqueID, className)
			return
		}
	}
	if sn := sortName(sort); sn != "" && g.isVariantSuperName(sn) {
		g.mkNondetVariantScoped(w, lhsExpr, sort, name, uniqueID, className)
		return
	}
	if g.nondetSkipSort(sort) {
		return
	}
	g.mkNondetWithGoType(w, lhsExpr, name, uniqueID, sort)
}

func (g *Generator) trackNondetSort(sort goivy.Sort) bool {
	if g == nil || sort == nil {
		return false
	}
	sn := sortName(sort)
	if sn == "" {
		return false
	}
	if g.isVariantSuperName(sn) {
		return true
	}
	return g.Mod != nil && g.Mod.SortDestructors != nil && g.Mod.SortDestructors.Get(sn) != nil
}

func (g *Generator) enterNondetSort(name string) bool {
	if g.nondetSortStack == nil {
		g.nondetSortStack = map[string]int{}
	}
	if g.nondetSortStack[name] > 0 {
		return true
	}
	g.nondetSortStack[name] = 1
	return false
}

func (g *Generator) leaveNondetSort(name string) {
	if g == nil || g.nondetSortStack == nil {
		return
	}
	if g.nondetSortStack[name] <= 1 {
		delete(g.nondetSortStack, name)
		return
	}
	g.nondetSortStack[name]--
}

func (g *Generator) nextNondetTemp(prefix string) string {
	if prefix == "" {
		prefix = "tmp"
	}
	n := 0
	if g != nil {
		n = g.nondetTempCtr
		g.nondetTempCtr++
	}
	return fmt.Sprintf("__nd_%s_%d", prefix, n)
}

// mkNondetVariant mirrors ivy2cpp/nondet.go:145.
func (g *Generator) mkNondetVariant(w *goWriter, lhsExpr string, super goivy.Sort, name string, uniqueID int64) {
	g.mkNondetVariantScoped(w, lhsExpr, super, name, uniqueID, "")
}

// mkNondetVariantScoped emits tag-pick + per-leaf nondet construction
// for a variant supertype. Mirrors ivy2cpp/nondet.go:149 but uses Go's
// switch-on-tag pattern and the leaf constructors emitted by
// emitVariantSuperStruct (variant.go:139).
func (g *Generator) mkNondetVariantScoped(w *goWriter, lhsExpr string, super goivy.Sort, name string, uniqueID int64, className string) {
	_ = className
	superName := sortName(super)
	variants := g.Mod.Variants[superName]
	if len(variants) == 0 {
		g.mkNondetWithGoType(w, lhsExpr, name, uniqueID, super)
		return
	}
	choice := fmt.Sprintf("__nd_tag_%d", uniqueID)
	w.linef("%s := ivyChoose(%d)", choice, len(variants))
	for i, sub := range variants {
		prefix := "if"
		if i > 0 {
			prefix = "} else if"
		}
		w.linef("%s %s == %d {", prefix, choice, i)
		subName := sortName(sub)
		ctor := "New" + goExportedName(superName) + goExportedName(subName)
		if g.isPlainVariantSubtypeName(subName) {
			w.linef("%s = %s()", lhsExpr, ctor)
			continue
		}
		tmp := fmt.Sprintf("__nd_v_%d_%d", uniqueID, i)
		w.linef("var %s %s", tmp, g.goType(sub))
		g.mkNondetValueScoped(w, tmp, sub, name, uniqueID, className)
		w.linef("%s = %s(%s)", lhsExpr, ctor, tmp)
	}
	w.line("}")
}

func (g *Generator) mkNondetVariantBaseScoped(w *goWriter, lhsExpr string, super goivy.Sort, name string, uniqueID int64, className string) bool {
	superName := sortName(super)
	variants := g.Mod.Variants[superName]
	if len(variants) == 0 {
		return false
	}
	for _, sub := range variants {
		if g.sortReferencesSort(sub, superName, map[string]bool{}) {
			continue
		}
		g.emitNondetVariantLeafValue(w, lhsExpr, superName, sub, name, uniqueID, className, false)
		return true
	}
	// Degenerate recursive-only variant family. There is no finite value
	// that fully satisfies the sort, so emit the shallowest constructor
	// and leave recursive fields at their Go zero value.
	g.emitNondetVariantLeafValue(w, lhsExpr, superName, variants[0], name, uniqueID, className, true)
	return true
}

func (g *Generator) emitNondetVariantLeafValue(w *goWriter, lhsExpr string, superName string, sub goivy.Sort, name string, uniqueID int64, className string, shallow bool) {
	subName := sortName(sub)
	ctor := "New" + goExportedName(superName) + goExportedName(subName)
	if g.isPlainVariantSubtypeName(subName) {
		w.linef("%s = %s()", lhsExpr, ctor)
		return
	}
	tmp := g.nextNondetTemp("base")
	w.linef("var %s %s", tmp, g.goType(sub))
	if shallow {
		for _, f := range g.destructorScalarFields(subName) {
			if g.sortReferencesSort(f.Sort, superName, map[string]bool{}) {
				continue
			}
			g.mkNondetValueScoped(w, tmp+"."+goExportedName(f.Name), f.Sort, name, uniqueID, className)
		}
	} else {
		g.mkNondetValueScoped(w, tmp, sub, name, uniqueID, className)
	}
	w.linef("%s = %s(%s)", lhsExpr, ctor, tmp)
}

func (g *Generator) sortReferencesSort(s goivy.Sort, target string, seen map[string]bool) bool {
	if g == nil || s == nil || target == "" {
		return false
	}
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return g.sortReferencesSort(fs.Range(), target, seen)
	}
	sn := sortName(s)
	if sn == target {
		return true
	}
	if sn == "" || seen[sn] {
		return false
	}
	seen[sn] = true
	if g.Mod != nil && g.Mod.SortDestructors != nil {
		for _, d := range g.Mod.SortDestructors.Get(sn) {
			fs, ok := d.CSort.(*goivy.LogicFunctionSort)
			if ok && g.sortReferencesSort(fs.Range(), target, seen) {
				return true
			}
		}
	}
	if g.Mod != nil {
		for _, sub := range g.Mod.Variants[sn] {
			if g.sortReferencesSort(sub, target, seen) {
				return true
			}
		}
	}
	return false
}

// mkNondetStructFields mirrors ivy2cpp/nondet.go:186.
func (g *Generator) mkNondetStructFields(w *goWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64) {
	g.mkNondetStructFieldsScoped(w, lhsExpr, sort, name, uniqueID, "")
}

// mkNondetStructFieldsScoped mirrors ivy2cpp/nondet.go:190.
// Recursively initializes each destructor field of a record sort.
func (g *Generator) mkNondetStructFieldsScoped(w *goWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64, className string) {
	_ = className
	if g.Mod == nil || g.Mod.SortDestructors == nil {
		return
	}
	destrs := g.Mod.SortDestructors.Get(sortName(sort))
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			// Drop the receiver-self domain element (per cpp).
			domain = domain[1:]
		}
		rng := fs.Range()
		field := goExportedName(memName(d.Name))
		fieldExpr := lhsExpr + "." + field
		if len(domain) == 0 {
			g.mkNondetValueScoped(w, fieldExpr, rng, name, uniqueID, className)
			continue
		}
		indices := make([]string, len(domain))
		opened := 0
		bail := false
		for i, dSort := range domain {
			indices[i] = fmt.Sprintf("__nd_d%d", i)
			card := goSortCard(g, dSort)
			if card <= 0 {
				bail = true
				break
			}
			w.linef("for %s := 0; %s < %d; %s++ {", indices[i], indices[i], card, indices[i])
			opened++
		}
		if bail {
			for j := 0; j < opened; j++ {
				w.line("}")
			}
			continue
		}
		lhs := fieldExpr
		for _, idx := range indices {
			lhs += "[" + idx + "]"
		}
		g.mkNondetValueScoped(w, lhs, rng, name, uniqueID, className)
		for j := 0; j < opened; j++ {
			w.line("}")
		}
	}
}

// nondetSkipSort mirrors ivy2cpp/nondet.go:216. Reports whether `s`
// is a sort whose locals are reasonably initialized by Go's zero
// value rather than ivyChoose:
//   - native-typed sorts use the user-supplied Go zero value.
//   - string-interpreted sorts default to Go's "".
func (g *Generator) nondetSkipSort(s goivy.Sort) bool {
	if g == nil || s == nil {
		return false
	}
	if _, ok := g.nativeTypeName(s); ok {
		return true
	}
	if g.hasStringInterp(s) {
		return true
	}
	return false
}

// emitNondetDecls retains its no-op shape — no per-program decls are
// needed beyond the ivyChoose runtime helper, which is emitted by
// runtime.go. Kept here for parallel call-site shape with ivy2cpp.
func (g *Generator) emitNondetDecls(w *goWriter) { _ = w }
