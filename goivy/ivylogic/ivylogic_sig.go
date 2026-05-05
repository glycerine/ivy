// Package ivylogic provides a higher-level logic IR wrapping the core logic
// package. It adds signature management, polymorphic symbols, formula
// classification, and various utility functions.
//
// This corresponds to Python's ivy_logic.py.
package ivylogic

import (
	"fmt"
	"sort"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// UnionSort holds multiple sorts for a polymorphic symbol.
type UnionSort struct {
	Sorts []lg.Sort
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
	IuCfg              *iu.IvyUtilsConfig // per-session config for version flags
	Sorts              *iu.InsMap[string, lg.Sort]
	Symbols            *iu.InsMap[string, *SymbolEntry]
	Constructors       map[string]bool
	Interp             map[string]interface{} // sort name → interpretation
	DefaultSort        lg.Sort                // nil means unset
	DefaultNumericSort lg.Sort
	AllowUnsorted      bool // when true, unknown sorts are accepted
}

// SymbolEntry holds a symbol name and its sort (which may be a UnionSort
// for polymorphic symbols).
type SymbolEntry struct {
	Name string
	Sort lg.Sort
	// Union is non-nil when the symbol is polymorphic (has multiple sorts).
	Union *UnionSort
}

// NewSigOn creates a Sig with the given per-session config. Production code should use this.
func NewSigOn(iuCfg *iu.IvyUtilsConfig) *Sig {
	s := &Sig{
		IuCfg:              iuCfg,
		Sorts:              iu.NewInsMap[string, lg.Sort](),
		Symbols:            iu.NewInsMap[string, *SymbolEntry](),
		Constructors:       make(map[string]bool),
		Interp:             make(map[string]interface{}),
		DefaultNumericSort: &lg.UninterpretedSort{Name: "int"},
	}
	// bool is always present
	s.Sorts.Set("bool", lg.Boolean)
	return s
}

// NewSig creates a fresh signature with a fresh default config. For tests.
func NewSig() *Sig {
	return NewSigOn(iu.NewIvyUtilsConfig())
}

// Copy returns a shallow copy of the signature.
func (s *Sig) Copy() *Sig {
	res := &Sig{
		IuCfg:              s.IuCfg,
		Sorts:              iu.NewInsMap[string, lg.Sort](),
		Symbols:            iu.NewInsMap[string, *SymbolEntry](),
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
func (s *Sig) AddSymbol(name string, sort lg.Sort) (*lg.Const, error) {
	if s.IuCfg != nil && s.IuCfg.HavePolymorphism && IsPolymorphicName(name) {
		entry, exists := s.Symbols.Get2(name)
		if !exists {
			u := &UnionSort{Sorts: []lg.Sort{sort}}
			s.Symbols.Set(name, &SymbolEntry{Name: name, Sort: sort, Union: u})
			return lg.NewConst(name, sort), nil
		}
		if entry.Union == nil {
			// Convert to union
			u := &UnionSort{Sorts: []lg.Sort{entry.Sort, sort}}
			entry.Union = u
		} else {
			// Check if sort already present
			found := false
			for _, existing := range entry.Union.Sorts {
				if lg.SortEqual(existing, sort) {
					found = true
					break
				}
			}
			if !found {
				entry.Union.Sorts = append(entry.Union.Sorts, sort)
			}
		}
		return lg.NewConst(name, sort), nil
	}

	if entry, exists := s.Symbols.Get2(name); exists {
		if !lg.SortEqual(sort, entry.Sort) {
			return nil, &lg.IvyError{Msg: fmt.Sprintf("redefining symbol: %s", name)}
		}
		return lg.NewConst(name, entry.Sort), nil
	}

	s.Symbols.Set(name, &SymbolEntry{Name: name, Sort: sort})
	return lg.NewConst(name, sort), nil
}

// RemoveSymbol removes a symbol from the signature. For union sorts,
// removes only the specific sort variant.
func (s *Sig) RemoveSymbol(name string, sort lg.Sort) {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return
	}
	if entry.Union != nil {
		for i, us := range entry.Union.Sorts {
			if lg.SortEqual(us, sort) {
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
func (s *Sig) ContainsSymbol(name string, sort lg.Sort) bool {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return false
	}
	if entry.Union != nil {
		for _, us := range entry.Union.Sorts {
			if lg.SortEqual(us, sort) {
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
	if sym, ok := sortOrSymbol.(*lg.Const); ok {
		return s.ContainsSymbol(sym.Name, sym.CSort)
	}
	if sort, ok := sortOrSymbol.(lg.Sort); ok {
		name := IvySortName(sort)
		existing, found := s.Sorts.Get2(name)
		if !found {
			return false
		}
		return lg.SortEqual(existing, sort)
	}
	return false
}

// AllSymbols returns all symbols in the signature, expanding union sorts.
func (s *Sig) AllSymbols() []*lg.Const {
	var result []*lg.Const
	for name, entry := range s.Symbols.All() {
		if entry.Union != nil {
			for _, sort := range entry.Union.Sorts {
				result = append(result, lg.NewConst(name, sort))
			}
		} else {
			result = append(result, lg.NewConst(name, entry.Sort))
		}
	}
	return result
}

// AllSymbolsNamed returns all sort variants for a given symbol name.
func (s *Sig) AllSymbolsNamed(name string) []*lg.Const {
	entry, exists := s.Symbols.Get2(name)
	if !exists {
		return nil
	}
	if entry.Union != nil {
		result := make([]*lg.Const, len(entry.Union.Sorts))
		for i, sort := range entry.Union.Sorts {
			result[i] = lg.NewConst(name, sort)
		}
		return result
	}
	return []*lg.Const{lg.NewConst(name, entry.Sort)}
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
		if _, ok := sort.(*lg.UninterpretedSort); !ok {
			b.WriteString(" = ")
			b.WriteString(sort.String())
		}
		b.WriteByte('\n')
	}
	for name, entry := range s.Symbols.All() {
		var sorts []lg.Sort
		if entry.Union != nil {
			sorts = entry.Union.Sorts
		} else {
			sorts = []lg.Sort{entry.Sort}
		}
		for _, sort := range sorts {
			if IsRelationalSort(sort) {
				b.WriteString("relation ")
			} else if fs, ok := sort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
				b.WriteString("function ")
			} else {
				b.WriteString("individual ")
			}
			b.WriteString(name)
			if fs, ok := sort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
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
func (s *Sig) FindSort(name string, allowUnsorted bool) (lg.Sort, error) {
	if allowUnsorted {
		if name == "S" {
			xtracer.Trace("compiler.FindSort allowUnsorted name=S HASH canon=%s", s.Canon())
			return lg.TopS, nil
		}
		xtracer.Trace("compiler.FindSort allowUnsorted name=%s\n  returning UninterpretedSort", name)
		return &lg.UninterpretedSort{Name: name}, nil
	}
	sort, ok := s.Sorts.Get2(name)
	if ok {
		if name == "S" {
			xtracer.Trace("compiler.FindSort name=S FOUND HASH canon=%s", s.Canon())
		}
		return sort, nil
	}
	if name == "S" {
		xtracer.Trace("compiler.FindSort name=S NOT_FOUND calling_GetDefaultSort HASH canon=%s", s.Canon())
		return GetDefaultSort(s)
	}
	//vv("about to return unknown type: '%v', call stack is:\n%v\n", name, stack())
	return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown type: %s", name)}
}

// SortListNames returns sort names from a slice (for diagnostics).
func SortListNames(sorts []lg.Sort) []string {
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
func (s *Sig) Canon() iu.Canonical {
	sortNames := s.SortNames()
	sort.Strings(sortNames)
	symParts := make([]string, 0, s.Symbols.Len())
	for name, entry := range s.Symbols.All() {
		symParts = append(symParts, name+":"+IvySortName(entry.Sort))
	}
	sort.Strings(symParts)
	return iu.Canonical(fmt.Sprintf("(sig sorts:[%s] symbols:[%s])",
		strings.Join(sortNames, " "),
		strings.Join(symParts, " ")))
}

// AddSort adds a sort to the signature.
// Silently overwrites on redefinition, matching Python ivy_logic.py:333-336
// where IvyError is created but never raised.
func (s *Sig) AddSort(sort lg.Sort) error {
	name := IvySortName(sort)
	s.Sorts.Set(name, sort)
	return nil
}

// FindSymbol looks up a symbol by name.
func (s *Sig) FindSymbol(name string, allowUnsorted bool) (*lg.Const, error) {
	if allowUnsorted {
		return lg.NewConst(name, lg.TopS), nil
	}
	entry, ok := s.Symbols.Get2(name)
	if ok {
		return lg.NewConst(name, entry.Sort), nil
	}
	if name == "=" {
		return IvyEquals, nil
	}
	if _, isSorted := s.Sorts.Get2(name); isSorted {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("type %s used where a function or individual symbol is expected", name)}
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown symbol: %s", name)}
}

// WithSymbols temporarily adds symbols to a signature. Use Enter/Exit
// or defer pattern.
type WithSymbols struct {
	sig     *Sig
	symbols []*lg.Const
	saved   []savedSymbol
}

type savedSymbol struct {
	name  string
	entry *SymbolEntry
}

// NewWithSymbols creates a context for temporarily adding symbols.
func NewWithSymbols(sig *Sig, symbols []*lg.Const) *WithSymbols {
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
	sorts []lg.Sort
	saved []savedSort
}

type savedSort struct {
	name string
	sort lg.Sort
}

// NewWithSorts creates a context for temporarily adding sorts.
func NewWithSorts(sig *Sig, sorts []lg.Sort) *WithSorts {
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
