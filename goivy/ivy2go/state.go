package ivy2go

import (
	"fmt"
	"sort"

	"github.com/glycerine/ivy/goivy"
)

// state.go: emits the State struct and its NewState constructor.
// Mirrors the per-class member emission in ivy2cpp/generator.go's
// emitHeader (lines around 232-289), split out for clarity in Go.

// emitStateStruct walks the module's signature symbols and emits the
// State struct with one exported field per symbol. Function-sorted
// symbols use the storage decided in goFunctionStorageFor (per
// ARCHITECTURE_TODO.md §3.5.4): scalar / fixed array / map[K]V.
func (g *Generator) emitStateStruct(w *goWriter) {
	w.linef("// %s holds the runtime state of the Ivy module.", g.StateTypeName)
	w.open(fmt.Sprintf("type %s struct {", g.StateTypeName))
	for _, sym := range g.stateSymbols() {
		w.linef("%s %s", goExportedName(sym.Name), g.goType(sym.Sort))
	}
	w.close("")
	w.blank()
}

// emitNewState emits the NewState constructor. Maps are allocated;
// scalar / array fields take their Go zero value.
func (g *Generator) emitNewState(w *goWriter) {
	w.open(fmt.Sprintf("func New%s() *%s {", g.StateTypeName, g.StateTypeName))
	w.linef("s := &%s{}", g.StateTypeName)
	for _, sym := range g.stateSymbols() {
		if g.symbolNeedsMakeMap(sym.Sort) {
			w.linef("s.%s = %s{}", goExportedName(sym.Name), g.goType(sym.Sort))
		}
	}
	w.line("return s")
	w.close("")
	w.blank()
}

// stateSymbol bundles a signature symbol with its sort for stable
// iteration. Used by emitStateStruct / emitNewState / emitInit.
type stateSymbol struct {
	Name string
	Sort goivy.Sort
}

// stateSymbols returns the list of symbols that contribute a State
// field, in deterministic (alphabetical) order. Excludes the
// _generating bookkeeping symbol prepareModuleForGo injects, which
// lives separately in M9.
func (g *Generator) stateSymbols() []stateSymbol {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	syms := make([]stateSymbol, 0)
	names := make([]string, 0)
	seen := map[string]bool{}
	for name := range g.Mod.Sig.Symbols.All() {
		if name == "_generating" {
			continue
		}
		if seen[name] {
			continue
		}
		names = append(names, name)
		seen[name] = true
	}
	sort.Strings(names)
	for _, name := range names {
		entry, ok := g.Mod.Sig.Symbols.Get2(name)
		if !ok || entry == nil {
			continue
		}
		syms = append(syms, stateSymbol{Name: name, Sort: entry.Sort})
	}
	return syms
}

// symbolNeedsMakeMap returns true when the symbol's storage maps to a
// Go map[K]V (hash-thunk storage). Arrays and scalars don't need an
// explicit allocation — Go zero-initialises them.
func (g *Generator) symbolNeedsMakeMap(s goivy.Sort) bool {
	fs, ok := s.(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	return st.Kind == goStorageHashThunk
}
