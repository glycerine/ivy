package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

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

// emitDestructorStruct emits a Go struct for an Ivy record sort whose
// fields are declared via `destructor` symbols.
//
// For a sort `type Point = struct { x: bool, y: bool }` ivy stores
// two destructor functions x : Point -> Bool and y : Point -> Bool.
// We emit:
//
//	type Point struct {
//	    X bool
//	    Y bool
//	}
//
//	func (s Point) Equal(o Point) bool { ... field-by-field == ... }
//
// For destructor functions with extra arguments (e.g. an indexed
// destructor `data : Point * Idx -> Bool`), the field type is the
// goFunctionStorageFor lowering of the arg → range function, so it
// gets either an array or map per its cardinality.
//
// Hash + Less are deferred; they're rarely needed since destructor
// structs aren't used as Go map keys in our emission (we use
// tup__T1__T2 keys for multi-arg map indexing).
func (g *Generator) emitDestructorStruct(w *goWriter, name string) {
	destrs := g.Mod.SortDestructors.Get(name)
	typeName := goExportedName(name)
	w.linef("// %s is the Go record for the Ivy destructor sort %q.", typeName, name)
	w.open(fmt.Sprintf("type %s struct {", typeName))
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			// First domain slot is the receiver record; drop it.
			domain = domain[1:]
		}
		fieldName := goExportedName(memName(d.Name))
		fieldType := g.destructorFieldType(domain, fs.Range())
		w.linef("%s %s", fieldName, fieldType)
	}
	w.close("")
	w.blank()

	// Equal method: field-by-field equality. Map/array fields use
	// helper functions (mapEqual / arrayEqual emitted in runtime).
	g.emitDestructorEqual(w, typeName, destrs)
	g.emitDestructorHash(w, typeName, destrs)
	g.emitDestructorLess(w, typeName, destrs)
}

// emitDestructorHash writes a Hash() uint64 method. We use FNV-1a
// over each field's primitive bytes; map fields fall back to a
// commutative XOR of per-entry hashes so iteration-order doesn't
// affect the result. Mirrors ivy2cpp/destructor.go emitDestructorStructHash.
func (g *Generator) emitDestructorHash(w *goWriter, typeName string, destrs []*goivy.Const) {
	w.open(fmt.Sprintf("func (a %s) Hash() uint64 {", typeName))
	w.line("var h uint64 = 1469598103934665603 // FNV-1a offset basis")
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		field := goExportedName(memName(d.Name))
		if len(domain) == 0 {
			w.linef("h = mixHash(h, a.%s)", field)
			continue
		}
		st := goFunctionStorageFor(g, domain, fs.Range())
		switch st.Kind {
		case goStorageArray:
			// Loop over every cell.
			loops, closes := arrayIndexLoops(st.Dims)
			for _, l := range loops {
				w.line(l)
			}
			w.linef("\th = mixHash(h, a.%s%s)", field, arrayIndexSuffix(st.Dims))
			for _, c := range closes {
				w.line(c)
			}
		case goStorageHashThunk:
			// Commutative aggregate so map iteration order doesn't matter.
			w.linef("var sub uint64")
			w.linef("for k, v := range a.%s {", field)
			w.line("\tsub ^= mixHash(mixHash(0, k), v)")
			w.line("}")
			w.line("h = mixHash(h, sub)")
		}
	}
	w.line("return h")
	w.close("")
	w.blank()
	// Request the mixHash runtime helper.
	g.Ctx.OnceGlobals["__need_mixhash"] = true
}

// emitDestructorLess writes a Less(other) bool method giving a
// lexicographic order: compare field-by-field, first differing pair
// decides. Map and array fields use cell-by-cell comparison. The
// total order isn't semantically meaningful for record types but is
// useful for stable sorting in test harnesses.
func (g *Generator) emitDestructorLess(w *goWriter, typeName string, destrs []*goivy.Const) {
	w.open(fmt.Sprintf("func (a %s) Less(b %s) bool {", typeName, typeName))
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		field := goExportedName(memName(d.Name))
		if len(domain) == 0 {
			w.linef("if a.%s != b.%s { return lessOrd(a.%s, b.%s) }", field, field, field, field)
			continue
		}
		st := goFunctionStorageFor(g, domain, fs.Range())
		switch st.Kind {
		case goStorageArray:
			loops, closes := arrayIndexLoops(st.Dims)
			for _, l := range loops {
				w.line(l)
			}
			suf := arrayIndexSuffix(st.Dims)
			w.linef("\tif a.%s%s != b.%s%s { return lessOrd(a.%s%s, b.%s%s) }", field, suf, field, suf, field, suf, field, suf)
			for _, c := range closes {
				w.line(c)
			}
		case goStorageHashThunk:
			w.linef("if len(a.%s) != len(b.%s) { return len(a.%s) < len(b.%s) }", field, field, field, field)
			// Maps have no natural order; treat equal-len maps as
			// equal for ordering purposes.
		}
	}
	w.line("return false")
	w.close("")
	w.blank()
	g.Ctx.OnceGlobals["__need_lessord"] = true
}

// arrayIndexLoops returns nested for-loops with index variables
// __i0, __i1, ... and the matching close braces.
func arrayIndexLoops(dims []int) (open, close []string) {
	for i, d := range dims {
		open = append(open, fmt.Sprintf("for __i%d := 0; __i%d < %d; __i%d++ {", i, i, d, i))
		_ = i
		close = append([]string{"}"}, close...)
	}
	return open, close
}

// arrayIndexSuffix returns "[__i0][__i1]…" matching arrayIndexLoops.
func arrayIndexSuffix(dims []int) string {
	var b strings.Builder
	for i := range dims {
		b.WriteString(fmt.Sprintf("[__i%d]", i))
	}
	return b.String()
}

// destructorFieldType returns the Go type for a destructor field with
// the given remaining domain and range. Scalar destructors (no extra
// args) get the range type directly; indexed destructors get the
// goFunctionStorageFor lowering (array or map).
func (g *Generator) destructorFieldType(domain []goivy.Sort, rng goivy.Sort) string {
	if len(domain) == 0 {
		return g.goScalarType(rng)
	}
	st := goFunctionStorageFor(g, domain, rng)
	switch st.Kind {
	case goStorageArray:
		return goArrayPrefix(st.Dims) + st.RangeType
	case goStorageHashThunk:
		return fmt.Sprintf("map[%s]%s", st.KeyType, st.RangeType)
	default:
		return st.RangeType
	}
}

// emitDestructorEqual writes a value-receiver Equal method that
// returns true iff every field matches.
func (g *Generator) emitDestructorEqual(w *goWriter, typeName string, destrs []*goivy.Const) {
	w.open(fmt.Sprintf("func (a %s) Equal(b %s) bool {", typeName, typeName))
	if len(destrs) == 0 {
		w.line("return true")
		w.close("")
		w.blank()
		return
	}
	for _, d := range destrs {
		fs, ok := d.CSort.(*goivy.LogicFunctionSort)
		if !ok {
			continue
		}
		domain := fs.Domain()
		if len(domain) > 0 {
			domain = domain[1:]
		}
		field := goExportedName(memName(d.Name))
		if len(domain) == 0 {
			w.linef("if a.%s != b.%s { return false }", field, field)
			continue
		}
		st := goFunctionStorageFor(g, domain, fs.Range())
		switch st.Kind {
		case goStorageArray:
			// Direct == works for fixed-size arrays of comparable
			// types; otherwise fall back to a per-cell loop.
			w.linef("if a.%s != b.%s { return false }", field, field)
		case goStorageHashThunk:
			w.linef("if len(a.%s) != len(b.%s) { return false }", field, field)
			w.linef("for k, v := range a.%s {", field)
			w.linef("\tif w, ok := b.%s[k]; !ok || w != v { return false }", field)
			w.line("}")
		default:
			w.linef("if a.%s != b.%s { return false }", field, field)
		}
	}
	w.line("return true")
	w.close("")
	w.blank()
}
