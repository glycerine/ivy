package ivylogic

import (
	lg "github.com/glycerine/goivy/logic"
)

// DefaultSort is the global default sort for the current context.
// It may be nil if no default sort has been set.
var DefaultSort lg.Sort

// IsDefaultSort returns true if s is the current default sort.
func IsDefaultSort(s lg.Sort) bool {
	if DefaultSort == nil {
		return false
	}
	return lg.SortEqual(s, DefaultSort)
}

// IsDefaultNumericSort returns true if s is the default numeric sort
// of the given signature.
func IsDefaultNumericSort(sig *Sig, s lg.Sort) bool {
	if sig.DefaultNumericSort == nil {
		return false
	}
	return lg.SortEqual(s, sig.DefaultNumericSort)
}

// interpretedSorts tracks which sort names have an interpretation.
// In the Python code this is sig.interp; here we also consult the Sig.
// Use the Sig.Interp map directly for full functionality.

// IsInterpretedSort returns true if the sort has an interpretation
// in the given signature.
func IsInterpretedSort(sig *Sig, s lg.Sort) bool {
	name := SortName(s)
	_, ok := sig.Interp[name]
	return ok
}

// IsUninterpretedSort returns true if the sort is an uninterpreted sort
// that has no interpretation in the given signature.
func IsUninterpretedSort(sig *Sig, s lg.Sort) bool {
	if !IsUISort(s) {
		return false
	}
	name := SortName(s)
	_, hasInterp := sig.Interp[name]
	return !hasInterp
}

// IsInterpretedSymbol returns true if the symbol is interpreted.
// A symbol is interpreted if it is a numeral with an interpreted sort,
// or if it is a polymorphic symbol with domain in an interpreted sort
// and is not in the uninterpreted polymorphic symbols set.
func IsInterpretedSymbol(sig *Sig, s *lg.Const) bool {
	// Check if it's a numeral with interpreted sort
	if IsNumeralName(s.Name) && IsInterpretedSort(sig, SortRange(s.CSort)) {
		return true
	}
	// Check if it's a polymorphic symbol over an interpreted sort
	if SymbolIsPolymorphic(s.Name) {
		dom := SortDomain(s.CSort)
		if len(dom) > 0 && IsInterpretedSort(sig, dom[0]) {
			if _, uninterp := UninterpretedPolymorphicSymbols[s.Name]; !uninterp {
				return true
			}
		}
	}
	return false
}

// SortInterp returns the interpretation for a sort in the given signature,
// or nil if none exists.
func SortInterp(sig *Sig, s lg.Sort) interface{} {
	name := SortName(s)
	interp, ok := sig.Interp[name]
	if !ok {
		return nil
	}
	return interp
}

// ImplementType sets the interpretation of sort1 to sort2 in the signature.
func ImplementType(sig *Sig, sort1 lg.Sort, sort2 interface{}) {
	sig.Interp[SortName(sort1)] = sort2
}

// BindSymbols provides enter/exit scoping for a set of symbols in a
// map[string]bool environment (emulating Python's set-based context manager).
type BindSymbols struct {
	env     map[string]bool
	symbols []string
	saved   []string // symbols that were already in env
}

// NewBindSymbols creates a new BindSymbols scope.
func NewBindSymbols(env map[string]bool, symbols []string) *BindSymbols {
	return &BindSymbols{env: env, symbols: symbols}
}

// Enter adds the symbols to the environment, saving any that were already present.
func (bs *BindSymbols) Enter() {
	bs.saved = nil
	for _, sym := range bs.symbols {
		if bs.env[sym] {
			bs.saved = append(bs.saved, sym)
			delete(bs.env, sym)
		}
		bs.env[sym] = true
	}
}

// Exit removes the symbols from the environment and restores saved ones.
func (bs *BindSymbols) Exit() {
	for _, sym := range bs.symbols {
		delete(bs.env, sym)
	}
	for _, sym := range bs.saved {
		bs.env[sym] = true
	}
}

// BindSymbolValues provides enter/exit scoping for key-value bindings
// in a map[string]lg.Node environment.
type BindSymbolValues struct {
	env      map[string]lg.Node
	bindings []SymbolBinding
	saved    []SymbolBinding
}

// SymbolBinding is a key-value pair for BindSymbolValues.
type SymbolBinding struct {
	Name  string
	Value lg.Node
}

// NewBindSymbolValues creates a new BindSymbolValues scope.
func NewBindSymbolValues(env map[string]lg.Node, bindings []SymbolBinding) *BindSymbolValues {
	return &BindSymbolValues{env: env, bindings: bindings}
}

// Enter adds the bindings to the environment.
func (bsv *BindSymbolValues) Enter() {
	bsv.saved = nil
	for _, b := range bsv.bindings {
		if old, ok := bsv.env[b.Name]; ok {
			bsv.saved = append(bsv.saved, SymbolBinding{b.Name, old})
			delete(bsv.env, b.Name)
		}
		bsv.env[b.Name] = b.Value
	}
}

// Exit removes the bindings and restores saved ones.
func (bsv *BindSymbolValues) Exit() {
	for _, b := range bsv.bindings {
		delete(bsv.env, b.Name)
	}
	for _, b := range bsv.saved {
		bsv.env[b.Name] = b.Value
	}
}

// UnsortedContext provides enter/exit scoping for allowing unsorted symbols.
type UnsortedContext struct {
	oldAllowUnsorted bool
}

// AllowUnsorted is a package-level flag controlling whether unsorted
// symbols are accepted. Corresponds to Python's allow_unsorted global.
var AllowUnsorted bool

// NewUnsortedContext creates a new unsorted context.
func NewUnsortedContext() *UnsortedContext {
	return &UnsortedContext{}
}

// Enter enables unsorted mode.
func (uc *UnsortedContext) Enter() {
	uc.oldAllowUnsorted = AllowUnsorted
	AllowUnsorted = true
}

// Exit restores the previous unsorted mode.
func (uc *UnsortedContext) Exit() {
	AllowUnsorted = uc.oldAllowUnsorted
}

// SortAsDefault provides enter/exit scoping for temporarily changing
// the default sort in a signature.
type SortAsDefault struct {
	sig     *Sig
	sort    lg.Sort
	oldSort lg.Sort
	hadOld  bool
}

// NewSortAsDefault creates a new SortAsDefault scope.
func NewSortAsDefault(sig *Sig, sort lg.Sort) *SortAsDefault {
	return &SortAsDefault{sig: sig, sort: sort}
}

// Enter sets the default sort.
func (sd *SortAsDefault) Enter() {
	sd.oldSort, sd.hadOld = sd.sig.Sorts["S"]
	sd.sig.Sorts["S"] = sd.sort
}

// Exit restores the previous default sort.
func (sd *SortAsDefault) Exit() {
	if sd.hadOld {
		sd.sig.Sorts["S"] = sd.oldSort
	} else {
		delete(sd.sig.Sorts, "S")
	}
}

// HasInfiniteInterpretation returns true if the sort has an infinite
// interpreted domain (e.g. int or nat). Corresponds to Python's
// has_infinite_interpretation.
func HasInfiniteInterpretation(sig *Sig, s lg.Sort) bool {
	name := SortName(s)
	interp, ok := sig.Interp[name]
	if !ok {
		return false
	}
	// Check if the interpretation is one of the infinite sorts
	switch v := interp.(type) {
	case string:
		return !quantifiersDecidable(v)
	}
	return false
}

// quantifiersDecidable returns true if quantifiers are decidable for the
// given theory name. Corresponds to Python's ivy_smtlib.quantifiers_decidable.
func quantifiersDecidable(theoryName string) bool {
	return theoryName != "int" && theoryName != "nat"
}

// AppsAst yields all function application subterms of an AST (excluding equality).
// Corresponds to Python's apps_ast in ivy_logic_utils.py.
func AppsAst(ast lg.Node) []lg.Node {
	var result []lg.Node
	appsAstRec(ast, &result)
	return result
}

func appsAstRec(ast lg.Node, result *[]lg.Node) {
	if IsApp(ast) {
		*result = append(*result, ast)
	}
	for _, arg := range NodeArgs(ast) {
		appsAstRec(arg, result)
	}
}

// SymbolsAst yields all function/relation symbols used in an AST.
// Corresponds to Python's symbols_ast in ivy_logic_utils.py.
func SymbolsAst(ast lg.Node) []*lg.Const {
	seen := make(map[string]bool)
	var result []*lg.Const
	symbolsAstRec(ast, &result, seen)
	return result
}

func symbolsAstRec(ast lg.Node, result *[]*lg.Const, seen map[string]bool) {
	if IsApp(ast) {
		var sym *lg.Const
		switch t := ast.(type) {
		case *lg.Apply:
			if c, ok := t.Func.(*lg.Const); ok {
				sym = c
			}
		case *lg.Const:
			sym = t
		}
		if sym != nil && !seen[sym.Name] {
			seen[sym.Name] = true
			*result = append(*result, sym)
		}
	}
	for _, arg := range NodeArgs(ast) {
		symbolsAstRec(arg, result, seen)
	}
}

// QuantifierVars returns the bound variables of a quantifier (ForAll or Exists).
func QuantifierVars(n lg.Node) []*lg.Var {
	switch t := n.(type) {
	case *lg.ForAll:
		return t.Variables
	case *lg.Exists:
		return t.Variables
	}
	return nil
}

// GetAppRep returns the function symbol (rep) of an application node.
// Returns nil if the node is not an application or has no named function.
func GetAppRep(n lg.Node) *lg.Const {
	switch t := n.(type) {
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Const); ok {
			return c
		}
	case *lg.Const:
		return t
	}
	return nil
}
