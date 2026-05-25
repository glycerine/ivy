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
