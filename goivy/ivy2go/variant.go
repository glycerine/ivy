package ivy2go

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// variant.go mirrors ivy2cpp/variant.go.
//
// Variants in Ivy are tagged unions: `type Animal = Cat | Dog` makes
// Animal a super-sort whose values carry a tag indicating which
// subtype they hold. The C++ emission builds a struct with a `tag`
// field and per-subtype payload fields; we mirror that shape in Go.
//
// OPEN 053 ships the super-struct emission. Constructors and the
// `*>` downcast operator integration are still deferred to a
// follow-up.

// variantSuperName returns the variant super name iff s is a sort
// whose name appears in Mod.Variants as a super.
func (g *Generator) variantSuperName(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil {
		return "", false
	}
	us, ok := s.(*goivy.UninterpretedSort)
	if !ok {
		return "", false
	}
	if _, has := g.Mod.Variants[us.Name]; has {
		return us.Name, true
	}
	return "", false
}

// variantIndex mirrors ivy2cpp/variant.go variantIndex. Returns the
// integer tag of `sub` within `super`'s variant list, or -1 if
// `sub` is not a variant of `super`.
func (g *Generator) variantIndex(super, sub goivy.Sort) int {
	if g == nil || g.Mod == nil {
		return -1
	}
	return g.Mod.VariantIndex(super, sub)
}

// variantIsaExpr mirrors ivy2cpp/variant.go variantIsaExpr. Emits a
// Go boolean expression that checks whether `superExpr` currently
// holds the `sub` variant: `(superExpr.Tag == <index>)`.
func (g *Generator) variantIsaExpr(superExpr string, super, sub goivy.Sort) string {
	return fmt.Sprintf("((%s).Tag == %d)", superExpr, g.variantIndex(super, sub))
}

// variantDowncastExpr mirrors ivy2cpp/variant.go variantDowncastExpr.
// Emits a Go expression that extracts the `sub`-typed payload from
// `superExpr` — `(*superExpr.<Sub>)`. Callers should guard with
// variantIsaExpr; dereferencing the wrong-tag payload-pointer panics.
func (g *Generator) variantDowncastExpr(superExpr string, super, sub goivy.Sort) string {
	subName := sortName(sub)
	_ = super
	return fmt.Sprintf("(*(%s).%s)", superExpr, goExportedName(subName))
}

// variantClassName is the Go counterpart of ivy2cpp/variant.go
// variantClassName. cpp uses this to qualify type names with the
// enclosing class; Go uses package scope, so the parameter is
// returned unchanged.
func (g *Generator) variantClassName(className string) string { return className }

// variantSolverRelationName mirrors ivy2cpp/variant.go
// variantSolverRelationName. Returns the solver-side relation
// identifier (`*>:<super>:<sub>`) used to encode variant downcasts
// in SMT-LIB. Free function (matches cpp).
func variantSolverRelationName(super, sub goivy.Sort) string {
	return "*>:" + sortName(super) + ":" + sortName(sub)
}

// variantUpcastExpr wraps an expression of a variant leaf sort in
// the supertype constructor `New<Super><Leaf>(expr)`. Mirrors
// ivy2cpp/variant.go:36 variantUpcastExpr. The Go-side constructor
// shape is defined in emitVariantSuperStruct (variant.go:139).
func (g *Generator) variantUpcastExpr(super, sub goivy.Sort, expr string) string {
	superName, _ := g.variantSuperName(super)
	if superName == "" {
		superName = sortName(super)
	}
	subName := sortName(sub)
	ctor := "New" + goExportedName(superName) + goExportedName(subName)
	return fmt.Sprintf("%s(%s)", ctor, expr)
}

// maybeVariantUpcast returns expr unchanged unless value is a variant
// leaf of target; in that case it wraps expr in the supertype
// constructor. Mirrors ivy2cpp/variant.go:44 maybeVariantUpcast.
// Without this, assignments like `root := l` (l: leaf, root: tree)
// emit raw `s.Root = loc__l` which Go rejects ("cannot use Leaf as
// Tree value in assignment").
func (g *Generator) maybeVariantUpcast(target, value goivy.Sort, expr string) string {
	if g != nil && g.Mod != nil && g.Mod.IsVariant(target, value) {
		return g.variantUpcastExpr(target, value, expr)
	}
	return expr
}

// variantSubtypeName returns the subtype name iff s is a subtype of
// some variant super.
func (g *Generator) variantSubtypeName(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil {
		return "", false
	}
	name := sortName(s)
	for _, subs := range g.Mod.Variants {
		for _, sub := range subs {
			if sortName(sub) == name && name != "" {
				return name, true
			}
		}
	}
	return "", false
}

// isVariantSuperName reports whether name is the super-name of any
// variant family.
func (g *Generator) isVariantSuperName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	_, ok := g.Mod.Variants[name]
	return ok
}

// isVariantSubtypeName reports whether name appears as a subtype of
// any variant family.
func (g *Generator) isVariantSubtypeName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	for _, subs := range g.Mod.Variants {
		for _, sub := range subs {
			if sortName(sub) == name {
				return true
			}
		}
	}
	return false
}

// isPlainVariantSubtypeName reports whether a subtype is "plain":
// just an uninterpreted sort tag with no associated destructors or
// natives. Plain subtypes lower to int (no struct needed).
func (g *Generator) isPlainVariantSubtypeName(name string) bool {
	if !g.isVariantSubtypeName(name) {
		return false
	}
	if g.Mod.SortDestructors != nil {
		if _, ok := g.Mod.SortDestructors.Get2(name); ok {
			return false
		}
	}
	return true
}

// emitVariantSuperStruct emits a Go tagged-union struct for an Ivy
// variant super sort plus one constructor per variant.
//
// Shape:
//
//	type Animal struct {
//	    Tag int          // 0..N-1 identifying which variant is active
//	    Cat *Cat         // non-nil when Tag == 0 (omitted for plain leaves)
//	    Dog *Dog         // non-nil when Tag == 1
//	}
//
//	func NewAnimalCat(v Cat) Animal { return Animal{Tag: 0, Cat: &v} }
//	func NewAnimalDog(v Dog) Animal { return Animal{Tag: 1, Dog: &v} }
//
// Plain leaves (subtypes with no destructors) get only a tag-bearing
// constructor: `func NewAnimalCat() Animal { return Animal{Tag: 0} }`.
func (g *Generator) emitVariantSuperStruct(w *goWriter, name string) {
	subs, ok := g.Mod.Variants[name]
	if !ok {
		return
	}
	typeName := goExportedName(name)
	w.linef("// %s is a tagged-union super-sort with %d variants.", typeName, len(subs))
	w.open(fmt.Sprintf("type %s struct {", typeName))
	w.line("Tag int")
	for _, sub := range subs {
		subName := sortName(sub)
		if subName == "" {
			continue
		}
		if g.isPlainVariantSubtypeName(subName) {
			continue
		}
		w.linef("%s *%s", goExportedName(subName), goExportedName(subName))
	}
	w.close("")
	w.blank()

	// Per-leaf constructors.
	for i, sub := range subs {
		subName := sortName(sub)
		if subName == "" {
			continue
		}
		exportedSub := goExportedName(subName)
		ctor := "New" + typeName + exportedSub
		if g.isPlainVariantSubtypeName(subName) {
			w.linef("func %s() %s { return %s{Tag: %d} }", ctor, typeName, typeName, i)
		} else {
			w.linef("func %s(v %s) %s { return %s{Tag: %d, %s: &v} }",
				ctor, exportedSub, typeName, typeName, i, exportedSub)
		}
	}
	w.blank()
	g.emitVariantStreamImpl(w, name)
}

// emitVariantStreamImpl mirrors ivy2cpp/variant.go emitVariantStreamImpl.
// Emits a String() string method on the variant super-type that
// renders as `{<leaf>:<payload>}`, matching cpp's operator<< output.
func (g *Generator) emitVariantStreamImpl(w *goWriter, name string) {
	subs, ok := g.Mod.Variants[name]
	if !ok {
		return
	}
	typeName := goExportedName(name)
	// Only import fmt if at least one leaf carries a payload (plain
	// leaves render as a string literal with no Sprintf).
	needsFmt := false
	for _, sub := range subs {
		if sn := sortName(sub); sn != "" && !g.isPlainVariantSubtypeName(sn) {
			needsFmt = true
			break
		}
	}
	if needsFmt {
		g.Ctx.AddImport("types", "fmt", "")
	}
	w.linef("// String renders as `{<leaf>:<payload>}` — matches ivy2cpp's")
	w.linef("// operator<< output for cross-tool trace parity.")
	w.linef("func (t %s) String() string {", typeName)
	w.line("\tswitch t.Tag {")
	for i, sub := range subs {
		subName := sortName(sub)
		if subName == "" {
			continue
		}
		exportedSub := goExportedName(subName)
		if g.isPlainVariantSubtypeName(subName) {
			w.linef("\tcase %d:", i)
			w.linef("\t\treturn %q", "{"+subName+":}")
			continue
		}
		w.linef("\tcase %d:", i)
		w.linef("\t\treturn fmt.Sprintf(%q, *t.%s)", "{"+subName+":%v}", exportedSub)
	}
	w.line("\t}")
	w.linef(`	return "{}"`)
	w.line("}")
	w.blank()
}

// emitVariantLeafStruct emits any per-leaf struct declarations the
// variant super references. For plain leaves we emit nothing (they
// reuse int); for destructor-backed leaves the destructor emission
// already covers them.
func (g *Generator) emitVariantLeafStruct(w *goWriter, name string) {
	// Destructor-backed leaves land via emitDestructorStruct; plain
	// leaves lower to int and need no declaration.
	_ = w
	_ = name
}

// variantLeafInfo bundles what variant-input synthesis needs about
// one leaf of a variant super sort.
type variantLeafInfo struct {
	Name    string             // leaf sort name (e.g. "leaf_a")
	IsPlain bool               // true → no payload (NewSuperLeafA())
	Fields  []destructorField  // scalar destructor fields (empty when plain)
}

// variantLeaves returns the leaf metadata for a variant super sort
// in declaration order — the order matches the Tag values emitted in
// emitVariantSuperStruct (i=0,1,...).
func (g *Generator) variantLeaves(superName string) []variantLeafInfo {
	if g == nil || g.Mod == nil {
		return nil
	}
	subs, ok := g.Mod.Variants[superName]
	if !ok {
		return nil
	}
	out := make([]variantLeafInfo, 0, len(subs))
	for _, sub := range subs {
		ln := sortName(sub)
		if ln == "" {
			continue
		}
		out = append(out, variantLeafInfo{
			Name:    ln,
			IsPlain: g.isPlainVariantSubtypeName(ln),
			Fields:  g.destructorScalarFields(ln),
		})
	}
	return out
}
