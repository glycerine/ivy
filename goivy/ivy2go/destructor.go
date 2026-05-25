package ivy2go

import "github.com/glycerine/ivy/goivy"

// destructor.go mirrors ivy2cpp/destructor.go. Full emission of
// destructor record structs (with Equal / Hash / Less methods) is
// scheduled for M8. M2 only provides the hook methods that
// types.go / generator.go need to call so the emission pipeline runs
// end to end; they return "no destructor here" for now.

// destructorStructName returns the struct name for a sort that has
// destructor declarations. Mirrors ivy2cpp destructorStructName.
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

// destructorSortNames mirrors ivy2cpp/destructor.go destructorSortNames.
// M2 returns the names from SortDestructors in declaration order so
// emitSortDecls's destructor branch can iterate.
func (g *Generator) destructorSortNames() []string {
	if g == nil || g.Mod == nil || g.Mod.SortDestructors == nil {
		return nil
	}
	var names []string
	for name := range g.Mod.SortDestructors.All() {
		names = append(names, name)
	}
	return names
}

// emitDestructorStruct is the M2 stub. M8 replaces this with the
// full struct + Equal / Hash / Less / String emission.
func (g *Generator) emitDestructorStruct(w *goWriter, name string) {
	w.linef("// destructor struct %s: deferred to M8.", goExportedName(name))
	w.linef("type %s struct{}", goExportedName(name))
	w.blank()
}
