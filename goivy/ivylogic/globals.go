package ivylogic

import (
	lg "github.com/glycerine/ivy/goivy/logic"
)

// IsDefaultSort returns true if s is the default sort of the given signature.
func IsDefaultSort(sig *Sig, s lg.Sort) bool {
	if sig == nil || sig.DefaultSort == nil {
		return false
	}
	return lg.SortEqual(s, sig.DefaultSort)
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
// in the given signature. Matches Python ivy_logic.py:1484-1486:
// canonizes the sort first, then checks type and sig.interp membership.
func IsInterpretedSort(sig *Sig, s lg.Sort) bool {
	s = CanonizeSort(sig, s)
	switch cs := s.(type) {
	case *lg.UninterpretedSort:
		_, ok := sig.Interp[cs.Name]
		return ok
	case *lg.EnumeratedSort:
		_, ok := sig.Interp[cs.Name]
		return ok
	default:
		return false
	}
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
func IsInterpretedSymbol(sig *Sig, s lg.Expr) bool {
	name := lg.ExprName(s)
	sort := s.NodeSort()
	// Check if it's a numeral with interpreted sort
	if IsNumeralName(name) && IsInterpretedSort(sig, SortRange(sort)) {
		return true
	}
	// Check if it's a polymorphic symbol over an interpreted sort
	if SymbolIsPolymorphic(name) {
		dom := SortDomain(sort)
		if len(dom) > 0 && IsInterpretedSort(sig, dom[0]) {
			if _, uninterp := UninterpretedPolymorphicSymbols[name]; !uninterp {
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
// map[NodeKey]Node environment. Uses structural identity keys to match
// Python's set of Symbol/Variable objects with structural equality.
type BindSymbols struct {
	env     map[lg.NodeKey]lg.Expr
	symbols []lg.Expr
	saved   []lg.Expr // symbols that were already in env
}

// NewBindSymbols creates a new BindSymbols scope.
func NewBindSymbols(env map[lg.NodeKey]lg.Expr, symbols []lg.Expr) *BindSymbols {
	return &BindSymbols{env: env, symbols: symbols}
}

// Enter adds the symbols to the environment, saving any that were already present.
func (bs *BindSymbols) Enter() {
	bs.saved = nil
	for _, sym := range bs.symbols {
		k := lg.Key(sym)
		if _, exists := bs.env[k]; exists {
			bs.saved = append(bs.saved, sym)
			delete(bs.env, k)
		}
		bs.env[k] = sym
	}
}

// Exit removes the symbols from the environment and restores saved ones.
func (bs *BindSymbols) Exit() {
	for _, sym := range bs.symbols {
		delete(bs.env, lg.Key(sym))
	}
	for _, sym := range bs.saved {
		bs.env[lg.Key(sym)] = sym
	}
}

// BindSymbolValues provides enter/exit scoping for key-value bindings
// in a map[NodeKey]Node environment. Uses structural identity keys to match
// Python's dict with Symbol/Variable keys using structural equality.
type BindSymbolValues struct {
	env      map[lg.NodeKey]lg.Expr
	bindings []SymbolBinding
	saved    []SymbolBinding
}

// SymbolBinding is a (symbol, value) pair for BindSymbolValues.
// Sym is the Symbol/Variable used as the dict key (structural equality).
type SymbolBinding struct {
	Sym   lg.Expr
	Value lg.Expr
}

// NewBindSymbolValues creates a new BindSymbolValues scope.
func NewBindSymbolValues(env map[lg.NodeKey]lg.Expr, bindings []SymbolBinding) *BindSymbolValues {
	return &BindSymbolValues{env: env, bindings: bindings}
}

// Enter adds the bindings to the environment.
func (bsv *BindSymbolValues) Enter() {
	bsv.saved = nil
	for _, b := range bsv.bindings {
		k := lg.Key(b.Sym)
		if old, ok := bsv.env[k]; ok {
			bsv.saved = append(bsv.saved, SymbolBinding{b.Sym, old})
			delete(bsv.env, k)
		}
		bsv.env[k] = b.Value
	}
}

// Exit removes the bindings and restores saved ones.
func (bsv *BindSymbolValues) Exit() {
	for _, b := range bsv.bindings {
		delete(bsv.env, lg.Key(b.Sym))
	}
	for _, b := range bsv.saved {
		bsv.env[lg.Key(b.Sym)] = b.Value
	}
}

// UnsortedContext provides enter/exit scoping for allowing unsorted symbols
// on a given Sig.
type UnsortedContext struct {
	sig              *Sig
	oldAllowUnsorted bool
}

// NewUnsortedContext creates a new unsorted context for the given Sig.
func NewUnsortedContext(sig *Sig) *UnsortedContext {
	return &UnsortedContext{sig: sig}
}

// Enter enables unsorted mode on the Sig.
func (uc *UnsortedContext) Enter() {
	uc.oldAllowUnsorted = uc.sig.AllowUnsorted
	uc.sig.AllowUnsorted = true
}

// Exit restores the previous unsorted mode on the Sig.
func (uc *UnsortedContext) Exit() {
	uc.sig.AllowUnsorted = uc.oldAllowUnsorted
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
func AppsAst(ast lg.Expr) []lg.Expr {
	var result []lg.Expr
	appsAstRec(ast, &result)
	return result
}

func appsAstRec(ast lg.Expr, result *[]lg.Expr) {
	if IsApp(ast) {
		*result = append(*result, ast)
	}
	for _, arg := range NodeArgs(ast) {
		appsAstRec(arg, result)
	}
}

// SymbolsAst yields all function/relation symbols used in an AST.
// Corresponds to Python's symbols_ast in ivy_logic_utils.py.
func SymbolsAst(ast lg.Expr) []*lg.Const {
	seen := make(map[lg.NodeKey]bool)
	var result []*lg.Const
	symbolsAstRec(ast, &result, seen)
	return result
}

func symbolsAstRec(ast lg.Expr, result *[]*lg.Const, seen map[lg.NodeKey]bool) {
	// Matches Python symbols_ast (ivy_logic_utils.py:534-545):
	// For Apply with binder rep: recurse into rep.body.
	// For Apply with const rep: yield rep.
	// Then iterate ast.args (Terms only).
	if IsApp(ast) {
		switch t := ast.(type) {
		case *lg.Apply:
			if nb, ok := t.Func.(*lg.NamedBinder); ok {
				// Binder as function head: recurse into body
				symbolsAstRec(nb.Body, result, seen)
			} else if c, ok := t.Func.(*lg.Const); ok {
				if !seen[lg.Key(c)] {
					seen[lg.Key(c)] = true
					*result = append(*result, c)
				}
			}
		case *lg.Const:
			if !seen[lg.Key(t)] {
				seen[lg.Key(t)] = true
				*result = append(*result, t)
			}
		}
	}
	for _, arg := range NodeArgs(ast) {
		symbolsAstRec(arg, result, seen)
	}
}

// QuantifierVars returns the bound variables of a quantifier (ForAll or Exists).
func QuantifierVars(n lg.Expr) []*lg.Variable {
	switch t := n.(type) {
	case *lg.ForAll:
		return t.Variables
	case *lg.Exists:
		return t.Variables
	}
	return nil
}

// GetAppRep returns the "representative" symbol of a node.
// Matches Python's .rep property (ivy_logic.py:128-129,282-283,295):
//   Symbol.rep = self
//   Apply.rep  = self.func
//   Eq.rep     = Symbol('=', RelationSort([t1.sort, t2.sort]))
// Returns nil if the node has no representative.
func GetAppRep(n lg.Expr) *lg.Const {
	switch t := n.(type) {
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Const); ok {
			return c
		}
	case *lg.Const:
		return t
	case *lg.Eq:
		// Python: Eq.rep = Symbol('=', RelationSort([t1.sort, t2.sort]))
		relSort := RelationSort([]lg.Sort{t.T1.NodeSort(), t.T2.NodeSort()})
		return lg.NewConst("=", relSort)
	}
	return nil
}
