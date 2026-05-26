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
//
// OPEN 061.1: hash-thunk-backed symbols also get a parallel
// __thunk_<sym> closure slot (default nil). emitAssignLarge installs
// the thunk at write-time; per-symbol get<Sym>(k) routes reads
// through map → thunk → zero.
func (g *Generator) emitStateStruct(w *goWriter) {
	w.linef("// %s holds the runtime state of the Ivy module.", g.StateTypeName)
	w.open(fmt.Sprintf("type %s struct {", g.StateTypeName))
	for _, sym := range g.stateSymbols() {
		w.linef("%s %s", goExportedName(sym.Name), g.goType(sym.Sort))
		if g.symbolNeedsThunkSlot(sym.Sort) {
			fs := sym.Sort.(*goivy.LogicFunctionSort)
			keyT, valT := g.thunkSlotTypes(fs)
			w.linef("__thunk_%s func(%s) %s", goExportedName(sym.Name), keyT, valT)
		}
	}
	// Progress counters — one int / int-array / map[K]int field per
	// `progress P(X) <-> cond` declaration. Updated by Tick(); checked
	// against rely bounds also by Tick().
	for _, f := range g.progressCounterFields() {
		w.line(f)
	}
	w.close("")
	w.blank()
	// Emit per-symbol getters that route through the thunk slot.
	g.emitStateGetters(w)
}

// emitStateGetters writes a get<Sym>(k) helper for every hash-thunk
// state symbol so reads can transparently fall back through map →
// thunk → zero. Mirrors the read-side wiring ivy2cpp does inline
// for hash_thunk<K, V>.
func (g *Generator) emitStateGetters(w *goWriter) {
	for _, sym := range g.stateSymbols() {
		if !g.symbolNeedsThunkSlot(sym.Sort) {
			continue
		}
		fs := sym.Sort.(*goivy.LogicFunctionSort)
		keyT, valT := g.thunkSlotTypes(fs)
		exported := goExportedName(sym.Name)
		w.linef("// get%s returns %s[k] with thunk fallback. Mirrors the", exported, exported)
		w.line("// read-side semantics of ivy2cpp's hash_thunk<K,V>::operator[].")
		w.linef("func (s *%s) get%s(k %s) %s {", g.StateTypeName, exported, keyT, valT)
		w.linef("\tif v, ok := s.%s[k]; ok { return v }", exported)
		w.linef("\tif s.__thunk_%s != nil { return s.__thunk_%s(k) }", exported, exported)
		w.linef("\tvar z %s", valT)
		w.line("\treturn z")
		w.line("}")
		w.blank()
	}
}

// symbolNeedsThunkSlot reports whether the symbol's storage is the
// hash-thunk (map[K]V) form. Only those symbols need the parallel
// thunk-closure field.
func (g *Generator) symbolNeedsThunkSlot(s goivy.Sort) bool {
	fs, ok := s.(*goivy.LogicFunctionSort)
	if !ok {
		return false
	}
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	return st.Kind == goStorageHashThunk
}

// thunkSlotTypes returns the (key, val) Go types for a hash-thunk
// symbol — used both for the State field type and the get<Sym>
// helper signature.
func (g *Generator) thunkSlotTypes(fs *goivy.LogicFunctionSort) (string, string) {
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	return st.KeyType, st.RangeType
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
// field, in deterministic (alphabetical) order. Mirrors ivy2cpp/
// generator.go stateSymbols (line 1067): excludes definitions, sort
// constructors, destructor sort names, and sig.Constructors (enum
// constants and similar) — they are NOT mutable state, so emitting
// them as State fields breaks the solver-driven test harness (the
// runtime fact `right = state.Right` was contradicting Z3's enum
// encoding and forcing every action's precondition UNSAT after the
// first iteration).
func (g *Generator) stateSymbols() []stateSymbol {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	syms := make([]stateSymbol, 0)
	names := make([]string, 0)
	seen := map[string]bool{}
	for name := range g.Mod.Sig.Symbols.All() {
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
		if g.isNonStateSignatureSymbol(name) {
			continue
		}
		syms = append(syms, stateSymbol{Name: name, Sort: entry.Sort})
	}
	return syms
}

// cardinalitySortNames mirrors ivy2cpp/generator.go cardinalitySortNames.
// Returns the names of sorts that need a runtime `__CARD__<name>`
// — either interpreted sorts or plain variant subtypes — sorted
// for deterministic emission.
func (g *Generator) cardinalitySortNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	add := func(name string) {
		if name == "" || name == "bool" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}
	for name := range g.Mod.Sig.Interp {
		add(name)
	}
	for _, name := range g.Mod.SortOrder {
		if g.isPlainVariantSubtypeName(name) {
			add(name)
		}
	}
	sort.Strings(names)
	return names
}

// allStateSymbols mirrors ivy2cpp/generator.go:1006. Broader than
// stateSymbols: it walks every signature symbol PLUS relations and
// functions without applying stateSymbols' definition/destructor
// filters. Callers that need the universe of mutable state (e.g.,
// extensional-relation analysis) use this; emitters that need a
// fielded subset use stateSymbols.
func (g *Generator) allStateSymbols() []stateSymbol {
	if g == nil || g.Mod == nil {
		return nil
	}
	seen := map[string]bool{}
	knownSig := map[string]bool{}
	var out []stateSymbol
	add := func(name string, s goivy.Sort) {
		if name == "" || seen[name] {
			return
		}
		if g.Mod.Sig != nil && g.Mod.Sig.Constructors[name] {
			return
		}
		if g.isSortConstructorName(name) {
			return
		}
		seen[name] = true
		out = append(out, stateSymbol{Name: name, Sort: s})
	}
	if g.Mod.Sig != nil {
		for _, sym := range g.Mod.Sig.AllSymbols() {
			name := sym.Name
			if name != "" {
				knownSig[name] = true
			}
			if name == "" || seen[name] {
				continue
			}
			if g.Mod.Sig.Constructors[name] || g.isSortConstructorName(name) {
				continue
			}
			n, err := goivy.SolverName(sym, g.Mod.Sig, nil)
			if err == nil && n == "" {
				continue
			}
			add(name, sym.CSort)
		}
	}
	if g.Mod.Relations != nil {
		for key, s := range g.Mod.Relations.All() {
			name := goivy.SymbolNameFromKey(key)
			if !knownSig[name] {
				add(name, s)
			}
		}
	}
	if g.Mod.Functions != nil {
		for key, s := range g.Mod.Functions.All() {
			name := goivy.SymbolNameFromKey(key)
			if !knownSig[name] {
				add(name, s)
			}
		}
	}
	return out
}

// isSortConstructorName mirrors ivy2cpp/generator.go:1098.
func (g *Generator) isSortConstructorName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	if g.Mod.ConstructorSorts != nil {
		if _, ok := g.Mod.ConstructorSorts[name]; ok {
			return true
		}
	}
	if g.Mod.SortConstructors != nil {
		for _, conss := range g.Mod.SortConstructors.All() {
			for _, cons := range conss {
				if cons != nil && cons.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// isNonStateSignatureSymbol mirrors the per-symbol filter inside
// ivy2cpp/generator.go stateSymbols (line 1067). It returns true for
// any signature symbol that is NOT mutable state — enum constants,
// constructors, destructor record names, and pure definitions.
func (g *Generator) isNonStateSignatureSymbol(name string) bool {
	if g == nil || g.Mod == nil {
		return false
	}
	if g.Mod.Sig != nil && g.Mod.Sig.Constructors != nil && g.Mod.Sig.Constructors[name] {
		return true
	}
	if g.Mod.DestructorSorts != nil {
		if _, ok := g.Mod.DestructorSorts[name]; ok {
			return true
		}
	}
	if g.isSortConstructorName(name) {
		return true
	}
	if g.isDefinitionName(name) {
		return true
	}
	return false
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
