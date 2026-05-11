// Package ivylogic provides a higher-level logic IR wrapping the core logic
// package. It adds signature management, polymorphic symbols, formula
// classification, and various utility functions.
//
// This corresponds to Python's ivy_logic.py.
package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// UnionSort holds multiple sorts for a polymorphic symbol.
type UnionSort struct {
	Sorts []Sort
}

func (u *UnionSort) String() string {
	parts := make([]string, len(u.Sorts))
	for i, s := range u.Sorts {
		parts[i] = s.String()
	}
	return "UnionSort(" + strings.Join(parts, ",") + ")"
}

// Sig represents a first-order signature — the collection of all declared
// sorts, symbols (constants/functions/relations), constructors, and
// interpretations.
type Sig struct {
	IuCfg              *IvyUtilsConfig // per-session config for version flags
	Sorts              *InsMap[string, Sort]
	Symbols            *InsMap[string, *SymbolEntry]
	Constructors       map[string]bool
	Interp             map[string]interface{} // sort name → interpretation
	DefaultSort        Sort                   // nil means unset
	DefaultNumericSort Sort
	AllowUnsorted      bool // when true, unknown sorts are accepted
}

// SymbolEntry holds a symbol name and its sort (which may be a UnionSort
// for polymorphic symbols).
type SymbolEntry struct {
	Name string
	Sort Sort
	// Union is non-nil when the symbol is polymorphic (has multiple sorts).
	Union *UnionSort
}

// NewSigOn creates a Sig with the given per-session config. Production code should use this.
func NewSigOn(iuCfg *IvyUtilsConfig) *Sig {
	s := &Sig{
		IuCfg:              iuCfg,
		Sorts:              NewInsMap[string, Sort](),
		Symbols:            NewInsMap[string, *SymbolEntry](),
		Constructors:       make(map[string]bool),
		Interp:             make(map[string]interface{}),
		DefaultNumericSort: &UninterpretedSort{Name: "int"},
	}
	// bool is always present
	s.Sorts.Set("bool", Boolean)
	return s
}

// NewSig creates a fresh signature with a fresh default config. For tests.
func NewSig() *Sig {
	return NewSigOn(NewIvyUtilsConfig())
}

// Copy returns a shallow copy of the signature.
func (s *Sig) Copy() *Sig {
	res := &Sig{
		IuCfg:              s.IuCfg,
		Sorts:              NewInsMap[string, Sort](),
		Symbols:            NewInsMap[string, *SymbolEntry](),
		Constructors:       make(map[string]bool, len(s.Constructors)),
		Interp:             make(map[string]interface{}, len(s.Interp)),
		DefaultSort:        s.DefaultSort,
		DefaultNumericSort: s.DefaultNumericSort,
		AllowUnsorted:      s.AllowUnsorted,
	}
	for k, v := range s.Sorts.All() {
		res.Sorts.Set(k, v)
	}
	for k, v := range s.Symbols.All() {
		res.Symbols.Set(k, v)
	}
	for k, v := range s.Constructors {
		res.Constructors[k] = v
	}
	for k, v := range s.Interp {
		res.Interp[k] = v
	}
	return res
}

// AddSymbol adds a symbol with the given name and sort to the signature.
// For polymorphic symbols, it accumulates sorts in a UnionSort.
// Returns the resulting Const.
func (s *Sig) AddSymbol(name string, sort Sort) (*Const, error) {
	if s.IuCfg != nil && s.IuCfg.HavePolymorphism && IsPolymorphicName(name) {
		entry, exists := s.Symbols.Get2(name)
		if !exists {
			u := &UnionSort{Sorts: []Sort{sort}}
			s.Symbols.Set(name, &SymbolEntry{Name: name, Sort: sort, Union: u})
			return NewConst(name, sort), nil
		}
		if entry.Union == nil {
			// Convert to union
			u := &UnionSort{Sorts: []Sort{entry.Sort, sort}}
			entry.Union = u
		} else {
			// Check if sort already present
			found := false
			for _, existing := range entry.Union.Sorts {
				if SortEqual(existing, sort) {
					found = true
					break
				}
			}
			if !found {
				entry.Union.Sorts = append(entry.Union.Sorts, sort)
			}
		}
		return NewConst(name, sort), nil
	}

	if entry, exists := s.Symbols.Get2(name); exists {
		if !SortEqual(sort, entry.Sort) {
			return nil, &IvyError{Msg: fmt.Sprintf("redefining symbol: %s", name)}
		}
		return NewConst(name, entry.Sort), nil
	}

	s.Symbols.Set(name, &SymbolEntry{Name: name, Sort: sort})
	return NewConst(name, sort), nil
}

// RemoveSymbol removes a symbol from the signature. For union sorts,
// removes only the specific sort variant.
func (s *Sig) RemoveSymbol(name string, sort Sort) {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return
	}
	if entry.Union != nil {
		for i, us := range entry.Union.Sorts {
			if SortEqual(us, sort) {
				entry.Union.Sorts = append(entry.Union.Sorts[:i], entry.Union.Sorts[i+1:]...)
				break
			}
		}
		if len(entry.Union.Sorts) == 0 {
			s.Symbols.Delkey(name)
		}
		return
	}
	s.Symbols.Delkey(name)
}

// ContainsSymbol checks if the signature contains a symbol with the given
// name and sort.
func (s *Sig) ContainsSymbol(name string, sort Sort) bool {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return false
	}
	if entry.Union != nil {
		for _, us := range entry.Union.Sorts {
			if SortEqual(us, sort) {
				return true
			}
		}
		return false
	}
	return true
}

// Contains checks if the signature contains the given sort or symbol.
// Python: ivy_logic.py:943-946
func (s *Sig) Contains(sortOrSymbol interface{}) bool {
	if sym, ok := sortOrSymbol.(*Const); ok {
		return s.ContainsSymbol(sym.Name, sym.CSort)
	}
	if sort, ok := sortOrSymbol.(Sort); ok {
		name := IvySortName(sort)
		existing, found := s.Sorts.Get2(name)
		if !found {
			return false
		}
		return SortEqual(existing, sort)
	}
	return false
}

// AllSymbols returns all symbols in the signature, expanding union sorts.
func (s *Sig) AllSymbols() []*Const {
	var result []*Const
	for name, entry := range s.Symbols.All() {
		if entry.Union != nil {
			for _, sort := range entry.Union.Sorts {
				result = append(result, NewConst(name, sort))
			}
		} else {
			result = append(result, NewConst(name, entry.Sort))
		}
	}
	return result
}

// AllSymbolsNamed returns all sort variants for a given symbol name.
func (s *Sig) AllSymbolsNamed(name string) []*Const {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return nil
	}
	if entry.Union != nil {
		result := make([]*Const, len(entry.Union.Sorts))
		for i, sort := range entry.Union.Sorts {
			result[i] = NewConst(name, sort)
		}
		return result
	}
	return []*Const{NewConst(name, entry.Sort)}
}

// String returns a human-readable representation of the signature.
func (s *Sig) String() string {
	var b strings.Builder
	for name, sort := range s.Sorts.All() {
		if name == "bool" {
			continue
		}
		b.WriteString("type ")
		b.WriteString(name)
		if _, ok := sort.(*UninterpretedSort); !ok {
			b.WriteString(" = ")
			b.WriteString(sort.String())
		}
		b.WriteByte('\n')
	}
	for name, entry := range s.Symbols.All() {
		var sorts []Sort
		if entry.Union != nil {
			sorts = entry.Union.Sorts
		} else {
			sorts = []Sort{entry.Sort}
		}
		for _, sort := range sorts {
			if IsRelationalSort(sort) {
				b.WriteString("relation ")
			} else if fs, ok := sort.(*LogicFunctionSort); ok && fs.Arity() > 0 {
				b.WriteString("function ")
			} else {
				b.WriteString("individual ")
			}
			b.WriteString(name)
			if fs, ok := sort.(*LogicFunctionSort); ok && fs.Arity() > 0 {
				b.WriteByte('(')
				dom := fs.Domain()
				for i, d := range dom {
					if i > 0 {
						b.WriteByte(',')
					}
					fmt.Fprintf(&b, "V%d:%s", i, d)
				}
				b.WriteByte(')')
			}
			if !IsRelationalSort(sort) {
				rng := SortRange(sort)
				if rng != nil {
					fmt.Fprintf(&b, " : %s", rng)
				}
			}
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// FindSort looks up a sort by name in the signature.
// If allowUnsorted is true, unknown sorts are created as UninterpretedSort.
func (s *Sig) FindSort(name string, allowUnsorted bool) (Sort, error) {
	if allowUnsorted {
		if name == "S" {
			xtraceParts("compiler.FindSort allowUnsorted name=S HASH canon=", canonPart(s))
			return TopS, nil
		}
		xtracer.Trace("compiler.FindSort allowUnsorted name=%s\n  returning UninterpretedSort", name)
		return &UninterpretedSort{Name: name}, nil
	}
	sort, ok := s.Sorts.Get2(name)
	if ok {
		if name == "S" {
			xtraceParts("compiler.FindSort name=S FOUND HASH canon=", canonPart(s))
		}
		return sort, nil
	}
	if name == "S" {
		xtraceParts("compiler.FindSort name=S NOT_FOUND calling_GetDefaultSort HASH canon=", canonPart(s))
		return GetDefaultSort(s)
	}
	//vv("about to return unknown type: '%v', call stack is:\n%v\n", name, stack())
	return nil, &IvyError{Msg: fmt.Sprintf("unknown type: %s", name)}
}

// SortListNames returns sort names from a slice (for diagnostics).
func SortListNames(sorts []Sort) []string {
	names := make([]string, len(sorts))
	for i, s := range sorts {
		names[i] = IvySortName(s)
	}
	return names
}

// SortNames returns the names of all sorts in the signature (for diagnostics).
func (s *Sig) SortNames() []string {
	names := make([]string, 0, s.Sorts.Len())
	for n, _ := range s.Sorts.All() {
		names = append(names, n)
	}
	return names
}

// Canon returns a canonical s-expression representation of the signature.
// Sorts and symbols are alphabetically sorted for deterministic output.
// Used by SigCheck for Merkle-chained conformance auditing with Python.
func (s *Sig) Canon() Canonical {
	sortNames := s.SortNames()
	sort.Strings(sortNames)
	symParts := make([]string, 0, s.Symbols.Len())
	for name, entry := range s.Symbols.All() {
		symParts = append(symParts, name+":"+IvySortName(entry.Sort))
	}
	sort.Strings(symParts)
	return Canonical(fmt.Sprintf("(sig sorts:[%s] symbols:[%s])",
		strings.Join(sortNames, " "),
		strings.Join(symParts, " ")))
}

// AddSort adds a sort to the signature.
// Silently overwrites on redefinition, matching Python ivy_logic.py:333-336
// where IvyError is created but never raised.
func (s *Sig) AddSort(sort Sort) error {
	name := IvySortName(sort)
	s.Sorts.Set(name, sort)
	return nil
}

// FindSymbol looks up a symbol by name.
func (s *Sig) FindSymbol(name string, allowUnsorted bool) (*Const, error) {
	if allowUnsorted {
		return NewConst(name, TopS), nil
	}
	entry, ok := s.Symbols.Get2(name)
	if ok {
		return NewConst(name, entry.Sort), nil
	}
	if name == "=" {
		return IvyEquals, nil
	}
	if _, isSorted := s.Sorts.Get2(name); isSorted {
		return nil, &IvyError{Msg: fmt.Sprintf("type %s used where a function or individual symbol is expected", name)}
	}
	return nil, &IvyError{Msg: fmt.Sprintf("unknown symbol: %s", name)}
}

// WithSymbols temporarily adds symbols to a signature. Use Enter/Exit
// or defer pattern.
type WithSymbols struct {
	sig     *Sig
	symbols []*Const
	saved   []savedSymbol
}

type savedSymbol struct {
	name  string
	entry *SymbolEntry
}

// NewWithSymbols creates a context for temporarily adding symbols.
func NewWithSymbols(sig *Sig, symbols []*Const) *WithSymbols {
	return &WithSymbols{sig: sig, symbols: symbols}
}

// Enter adds the symbols to the signature.
func (ws *WithSymbols) Enter() {
	for _, sym := range ws.symbols {
		if entry, exists := ws.sig.Symbols.Get2(sym.Name); exists {
			ws.saved = append(ws.saved, savedSymbol{sym.Name, entry})
			ws.sig.Symbols.Delkey(sym.Name)
		}
		ws.sig.Symbols.Set(sym.Name, &SymbolEntry{Name: sym.Name, Sort: sym.CSort})
	}
}

// Exit restores the original symbols.
func (ws *WithSymbols) Exit() {
	for _, sym := range ws.symbols {
		ws.sig.Symbols.Delkey(sym.Name)
	}
	for _, s := range ws.saved {
		ws.sig.Symbols.Set(s.name, s.entry)
	}
}

// WithSorts temporarily adds sorts to a signature.
type WithSorts struct {
	sig   *Sig
	sorts []Sort
	saved []savedSort
}

type savedSort struct {
	name string
	sort Sort
}

// NewWithSorts creates a context for temporarily adding sorts.
func NewWithSorts(sig *Sig, sorts []Sort) *WithSorts {
	return &WithSorts{sig: sig, sorts: sorts}
}

// Enter adds the sorts to the signature.
func (ws *WithSorts) Enter() {
	xtracer.Trace("ivylogic.WithSorts.Enter nSorts=%d", len(ws.sorts))
	//pp("adding=%v existing=%v", len(ws.sorts), SortListNames(ws.sorts), ws.sig.SortNames())
	for _, s := range ws.sorts {
		name := IvySortName(s)
		if existing, ok := ws.sig.Sorts.Get2(name); ok {
			ws.saved = append(ws.saved, savedSort{name, existing})
		}
		ws.sig.Sorts.Set(name, s)
	}
}

// Exit restores the original sorts.
func (ws *WithSorts) Exit() {
	for _, s := range ws.sorts {
		name := IvySortName(s)
		ws.sig.Sorts.Delkey(name)
	}
	for _, s := range ws.saved {
		ws.sig.Sorts.Set(s.name, s.sort)
	}
}
