// Package ivylogic provides a higher-level logic IR wrapping the core logic
// package. It adds signature management, polymorphic symbols, formula
// classification, and various utility functions.
//
// This corresponds to Python's ivy_logic.py.
package ivylogic

import (
	"fmt"
	"strings"

	lg "github.com/glycerine/goivy/logic"
	iu "github.com/glycerine/goivy/ivyutils"
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
	Sorts              map[string]lg.Sort
	Symbols            map[string]*SymbolEntry
	Constructors       map[string]bool
	Interp             map[string]interface{} // sort name → interpretation
	DefaultSort        lg.Sort                // nil means unset
	DefaultNumericSort lg.Sort
}

// SymbolEntry holds a symbol name and its sort (which may be a UnionSort
// for polymorphic symbols).
type SymbolEntry struct {
	Name string
	Sort lg.Sort
	// Union is non-nil when the symbol is polymorphic (has multiple sorts).
	Union *UnionSort
}

// NewSig creates a fresh signature with default settings.
func NewSig() *Sig {
	s := &Sig{
		Sorts:              make(map[string]lg.Sort),
		Symbols:            make(map[string]*SymbolEntry),
		Constructors:       make(map[string]bool),
		Interp:             make(map[string]interface{}),
		DefaultNumericSort: &lg.UninterpretedSort{Name: "int"},
	}
	// bool is always present
	s.Sorts["bool"] = lg.Boolean
	return s
}

// Copy returns a shallow copy of the signature.
func (s *Sig) Copy() *Sig {
	res := &Sig{
		Sorts:              make(map[string]lg.Sort, len(s.Sorts)),
		Symbols:            make(map[string]*SymbolEntry, len(s.Symbols)),
		Constructors:       make(map[string]bool, len(s.Constructors)),
		Interp:             make(map[string]interface{}, len(s.Interp)),
		DefaultSort:        s.DefaultSort,
		DefaultNumericSort: s.DefaultNumericSort,
	}
	for k, v := range s.Sorts {
		res.Sorts[k] = v
	}
	for k, v := range s.Symbols {
		res.Symbols[k] = v
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
func (s *Sig) AddSymbol(name string, sort lg.Sort) (*lg.Symbol, error) {
	if iu.IvyHavePolymorphism && IsPolymorphicName(name) {
		entry, exists := s.Symbols[name]
		if !exists {
			u := &UnionSort{Sorts: []lg.Sort{sort}}
			s.Symbols[name] = &SymbolEntry{Name: name, Sort: sort, Union: u}
			return lg.NewSymbol(name, sort), nil
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
		return lg.NewSymbol(name, sort), nil
	}

	if entry, exists := s.Symbols[name]; exists {
		if !lg.SortEqual(sort, entry.Sort) {
			return nil, &lg.IvyError{Msg: fmt.Sprintf("redefining symbol: %s", name)}
		}
		return lg.NewSymbol(name, entry.Sort), nil
	}

	s.Symbols[name] = &SymbolEntry{Name: name, Sort: sort}
	return lg.NewSymbol(name, sort), nil
}

// RemoveSymbol removes a symbol from the signature. For union sorts,
// removes only the specific sort variant.
func (s *Sig) RemoveSymbol(name string, sort lg.Sort) {
	entry, exists := s.Symbols[name]
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
			delete(s.Symbols, name)
		}
		return
	}
	delete(s.Symbols, name)
}

// ContainsSymbol checks if the signature contains a symbol with the given
// name and sort.
func (s *Sig) ContainsSymbol(name string, sort lg.Sort) bool {
	entry, exists := s.Symbols[name]
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
	if sym, ok := sortOrSymbol.(*lg.Symbol); ok {
		return s.ContainsSymbol(sym.Name, sym.CSort)
	}
	if sort, ok := sortOrSymbol.(lg.Sort); ok {
		name := SortName(sort)
		existing, found := s.Sorts[name]
		if !found {
			return false
		}
		return lg.SortEqual(existing, sort)
	}
	return false
}

// AllSymbols returns all symbols in the signature, expanding union sorts.
func (s *Sig) AllSymbols() []*lg.Symbol {
	var result []*lg.Symbol
	for name, entry := range s.Symbols {
		if entry.Union != nil {
			for _, sort := range entry.Union.Sorts {
				result = append(result, lg.NewSymbol(name, sort))
			}
		} else {
			result = append(result, lg.NewSymbol(name, entry.Sort))
		}
	}
	return result
}

// AllSymbolsNamed returns all sort variants for a given symbol name.
func (s *Sig) AllSymbolsNamed(name string) []*lg.Symbol {
	entry, exists := s.Symbols[name]
	if !exists {
		return nil
	}
	if entry.Union != nil {
		result := make([]*lg.Symbol, len(entry.Union.Sorts))
		for i, sort := range entry.Union.Sorts {
			result[i] = lg.NewSymbol(name, sort)
		}
		return result
	}
	return []*lg.Symbol{lg.NewSymbol(name, entry.Sort)}
}

// String returns a human-readable representation of the signature.
func (s *Sig) String() string {
	var b strings.Builder
	for name, sort := range s.Sorts {
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
	for name, entry := range s.Symbols {
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
			return lg.TopS, nil
		}
		return &lg.UninterpretedSort{Name: name}, nil
	}
	sort, ok := s.Sorts[name]
	if ok {
		return sort, nil
	}
	if name == "S" {
		if s.DefaultSort != nil {
			return s.DefaultSort, nil
		}
		return nil, &lg.IvyError{Msg: "unspecified type"}
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown type: %s", name)}
}

// AddSort adds a sort to the signature.
// Silently overwrites on redefinition, matching Python ivy_logic.py:333-336
// where IvyError is created but never raised.
func (s *Sig) AddSort(sort lg.Sort) error {
	name := SortName(sort)
	s.Sorts[name] = sort
	return nil
}

// FindSymbol looks up a symbol by name.
func (s *Sig) FindSymbol(name string, allowUnsorted bool) (*lg.Symbol, error) {
	if allowUnsorted {
		return lg.NewSymbol(name, lg.TopS), nil
	}
	entry, ok := s.Symbols[name]
	if ok {
		return lg.NewSymbol(name, entry.Sort), nil
	}
	if name == "=" {
		return Equals, nil
	}
	if _, isSorted := s.Sorts[name]; isSorted {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("type %s used where a function or individual symbol is expected", name)}
	}
	return nil, &lg.IvyError{Msg: fmt.Sprintf("unknown symbol: %s", name)}
}

// WithSymbols temporarily adds symbols to a signature. Use Enter/Exit
// or defer pattern.
type WithSymbols struct {
	sig     *Sig
	symbols []*lg.Symbol
	saved   []savedSymbol
}

type savedSymbol struct {
	name  string
	entry *SymbolEntry
}

// NewWithSymbols creates a context for temporarily adding symbols.
func NewWithSymbols(sig *Sig, symbols []*lg.Symbol) *WithSymbols {
	return &WithSymbols{sig: sig, symbols: symbols}
}

// Enter adds the symbols to the signature.
func (ws *WithSymbols) Enter() {
	for _, sym := range ws.symbols {
		if entry, exists := ws.sig.Symbols[sym.Name]; exists {
			ws.saved = append(ws.saved, savedSymbol{sym.Name, entry})
			delete(ws.sig.Symbols, sym.Name)
		}
		ws.sig.Symbols[sym.Name] = &SymbolEntry{Name: sym.Name, Sort: sym.CSort}
	}
}

// Exit restores the original symbols.
func (ws *WithSymbols) Exit() {
	for _, sym := range ws.symbols {
		delete(ws.sig.Symbols, sym.Name)
	}
	for _, s := range ws.saved {
		ws.sig.Symbols[s.name] = s.entry
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
	for _, s := range ws.sorts {
		name := SortName(s)
		if existing, ok := ws.sig.Sorts[name]; ok {
			ws.saved = append(ws.saved, savedSort{name, existing})
		}
		ws.sig.Sorts[name] = s
	}
}

// Exit restores the original sorts.
func (ws *WithSorts) Exit() {
	for _, s := range ws.sorts {
		name := SortName(s)
		delete(ws.sig.Sorts, name)
	}
	for _, s := range ws.saved {
		ws.sig.Sorts[s.name] = s.sort
	}
}
