package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// SortName extracts the name from a sort.
func IvySortName(s Sort) string {
	switch t := s.(type) {
	case *UninterpretedSort:
		return t.Name
	case *LogicEnumeratedSort:
		return t.Name
	case *RangeSort:
		return t.Name
	case *BooleanSort:
		return "bool"
	case *TopSort:
		return t.Name
	case *LogicFunctionSort:
		return t.String()
	default:
		return s.String()
	}
}

// SortDomain returns the domain sorts of a sort, or nil for first-order sorts.
func SortDomain(s Sort) []Sort {
	if fs, ok := s.(*LogicFunctionSort); ok {
		return fs.Domain()
	}
	return nil
}

// SortRange returns the range sort of a sort, or the sort itself for first-order sorts.
func SortRange(s Sort) Sort {
	if fs, ok := s.(*LogicFunctionSort); ok {
		return fs.Range()
	}
	return s
}

// IsRelationalSort returns true if the sort is a relation sort
// (a function sort with Boolean range, or Boolean itself).
func IsRelationalSort(s Sort) bool {
	if _, ok := s.(*BooleanSort); ok {
		return true
	}
	if fs, ok := s.(*LogicFunctionSort); ok {
		return SortEqual(fs.Range(), Boolean)
	}
	return false
}

// RelationSort creates a relation sort (function sort with Boolean range).
func LogicRelationSort(dom []Sort) Sort {
	if len(dom) == 0 {
		return Boolean
	}
	sorts := make([]Sort, len(dom)+1)
	copy(sorts, dom)
	sorts[len(dom)] = Boolean
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		return Boolean
	}
	return fs
}

// FuncConstSort creates a function sort, or returns the single sort if arity 0.
func FuncConstSort(sorts ...Sort) Sort {
	if len(sorts) == 1 {
		return sorts[0]
	}
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		return sorts[len(sorts)-1]
	}
	return fs
}

// TopFunctionSort creates a polymorphic function sort of given arity.
func TopFunctionSort(arity int) Sort {
	if arity == 0 {
		return &TopSort{Name: "alpha"}
	}
	sorts := make([]Sort, arity+1)
	for i := range sorts {
		sorts[i] = &TopSort{Name: alphaName(i)}
	}
	fs, _ := NewFunctionSort(sorts...)
	return fs
}

func alphaName(idx int) string {
	return fmt.Sprintf("alpha%d", idx)
}

// --- Type predicates ---

// IsVariable returns true if the node is a logic.Variable.
func IsVariable(n Expr) bool {
	_, ok := n.(*LogicVariable)
	return ok
}

// IsConstant returns true if the node is a logic.Const.
func IsConstant(n Expr) bool {
	_, ok := n.(*Const)
	return ok
}

// IsApp returns true if the node is a function application,
// a constant, or a 0-arity named binder.
func IsApp(n Expr) bool {
	switch t := n.(type) {
	case *Apply:
		return true
	case *Const:
		return true
	case *LogicNamedBinder:
		return len(t.Variables) == 0
	}
	return false
}

// IsAtom returns true if the node is an atomic formula.
func IsAtom(n Expr) bool {
	if _, ok := n.(*Eq); ok {
		return true
	}
	if IsApp(n) && SortEqual(n.NodeSort(), Boolean) {
		return true
	}
	return false
}

// IsRelApp returns true if the node is a relation application.
func IsRelApp(n Expr) bool {
	app, ok := n.(*Apply)
	if !ok {
		return false
	}
	if c, ok := app.Func.(*Const); ok {
		return IsRelationalSort(c.CSort)
	}
	return false
}

// IsForall returns true if the node is a ForAll.
func IsForall(n Expr) bool {
	_, ok := n.(*ForAll)
	return ok
}

// IsExists returns true if the node is an Exists.
func IsExists(n Expr) bool {
	_, ok := n.(*LogicExists)
	return ok
}

// IsLambda returns true if the node is a Lambda.
func IsLambda(n Expr) bool {
	_, ok := n.(*Lambda)
	return ok
}

// IsQuantifier returns true if the node is ForAll or Exists.
func IsQuantifier(n Expr) bool {
	return IsForall(n) || IsExists(n)
}

// IsBinder returns true for ForAll, Exists, Lambda, NamedBinder, or Some.
func IsBinder(n Expr) bool {
	switch n.(type) {
	case *ForAll, *LogicExists, *Lambda, *LogicNamedBinder, *LogicSome:
		return true
	}
	return false
}

// IsNamedBinder returns true if the node is a NamedBinder.
func IsNamedBinder(n Expr) bool {
	_, ok := n.(*LogicNamedBinder)
	return ok
}

// IsTemporal returns true if the node is a temporal operator.
func IsTemporal(n Expr) bool {
	switch n.(type) {
	case *LogicGlobally, *LogicEventually, *LogicWhenOperator:
		return true
	}
	return false
}

// HasTemporal returns true if the formula contains a temporal operator.
func IvyHasTemporal(n Expr) bool {
	if IsTemporal(n) {
		return true
	}
	for _, c := range n.Children() {
		if IvyHasTemporal(c) {
			return true
		}
	}
	return false
}

// IsEq returns true if the node is an Eq.
func IsEq(n Expr) bool {
	_, ok := n.(*Eq)
	return ok
}

// IsIte returns true if the node is an Ite.
func IsIte(n Expr) bool {
	_, ok := n.(*LogicIte)
	return ok
}

// IsEnumeratedSort returns true if the sort is an EnumeratedSort.
func IsEnumeratedSort(s Sort) bool {
	_, ok := s.(*LogicEnumeratedSort)
	return ok
}

// IsRangeSort returns true if the sort is a RangeSort.
func IsRangeSort(s Sort) bool {
	_, ok := s.(*RangeSort)
	return ok
}

// IsBooleanSort returns true if the sort is the Boolean sort.
func IsBooleanSort(s Sort) bool {
	return SortEqual(s, Boolean)
}

// IsBoolean returns true if the term has Boolean sort.
func IsBoolean(n Expr) bool {
	return SortEqual(n.NodeSort(), Boolean)
}

// IsFirstOrderSort returns true if the sort is an UninterpretedSort.
func IsFirstOrderSort(s Sort) bool {
	_, ok := s.(*UninterpretedSort)
	return ok
}

// IsFunctionSort returns true if the sort is a FunctionSort.
func IsFunctionSort(s Sort) bool {
	_, ok := s.(*LogicFunctionSort)
	return ok
}

// IsTopSort returns true if the sort is a TopSort.
func IvyIsTopSort(s Sort) bool {
	_, ok := s.(*TopSort)
	return ok
}

// IsIndividual returns true if the term has a non-Boolean sort.
func IsIndividual(n Expr) bool {
	return !SortEqual(n.NodeSort(), Boolean)
}

// IsNumeral returns true if the node is a numeral constant.
func IsNumeral(n Expr) bool {
	c, ok := n.(*Const)
	if !ok {
		return false
	}
	return IsNumeralName(c.Name)
}

// IsNumeralName returns true if the name starts with a digit or quote.
func IsNumeralName(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	if s[0] == '"' {
		return true
	}
	if s[0] == '-' && len(s) > 1 && s[1] >= '0' && s[1] <= '9' {
		return true
	}
	return false
}

// IsLiteralString returns true if the name starts with a double quote.
func IsLiteralString(n Expr) bool {
	c, ok := n.(*Const)
	if !ok {
		return false
	}
	return len(c.Name) > 0 && c.Name[0] == '"'
}

// IsConcretetlySorted returns true if the term has no TopSort or polymorphic
// elements.
func IsConcretetlySorted(n Expr) bool {
	return !ContainsTopSort(n) && !IsPolymorphic(n)
}

// IsTrue returns true if the node is logical true (empty And).
func IvyIsTrue(n Expr) bool {
	return IsTrue(n)
}

// IsFalse returns true if the node is logical false (empty Or).
func IvyIsFalse(n Expr) bool {
	return IsFalse(n)
}

// IsGprop returns true if the formula is Globally(phi) where phi has
// no temporal operators.
func IsGprop(n Expr) bool {
	g, ok := n.(*LogicGlobally)
	if !ok {
		return false
	}
	return !IvyHasTemporal(g.Body)
}

// --- IvyEquals symbol ---

// IvyEquals is the built-in equality symbol.
var IvyEquals = NewConst("=", LogicRelationSort([]Sort{TopS, TopS}))

// IsEquals returns true if the constant is the equality symbol.
func IvyIsEquals(c *Const) bool {
	return c.Name == "="
}

// NewEquals creates an Eq node from two terms.
func NewEquals(x, y Expr) *Eq {
	return &Eq{T1: x, T2: y}
}

// --- Sort predicates ---

// IsUISort returns true if the sort is specifically an UninterpretedSort
// (not a subclass).
func IsUISort(s Sort) bool {
	_, ok := s.(*UninterpretedSort)
	return ok
}

// IsPolymorphicName returns true if the name is in the polymorphic symbols table.
func IsPolymorphicName(name string) bool {
	_, ok := polymorphicSymbols[name]
	return ok
}

// IsInequalitySymbol returns true if the symbol name is <, <=, >, or >=.
func IsInequalitySymbol(name string) bool {
	switch name {
	case "<", "<=", ">", ">=":
		return true
	}
	return false
}

// IsStrictInequalitySymbol checks if the inequality is strict under the
// given number of negations (pol=0: even, pol=1: odd).
func IsStrictInequalitySymbol(name string, pol int) bool {
	if pol == 0 {
		return name == "<" || name == ">"
	}
	if pol == 1 {
		return name == "<=" || name == ">="
	}
	return false
}

// NormalizeSymbol maps polymorphic macro symbols to their canonical form.
// Corresponds to Python's normalize_symbol (ivy_logic.py:363-366).
// E.g. Symbol("<=", sort) -> Symbol("<", sort) when macros are active.
func NormalizeSymbol(sym *Const, iuCfg *IvyUtilsConfig) *Const {
	if iuCfg != nil && iuCfg.UsePolymorphicMacros {
		if canonical, ok := PolymorphicMacrosMap[sym.Name]; ok {
			return NewConst(canonical, sym.CSort)
		}
	}
	return sym
}

// GetSortTerm returns the sort of a term.
// Corresponds to Python's get_sort_term (ivy_logic.py:384-387).
// If the term has a .sort attribute, return it; otherwise return rep.sort.rng.
func GetSortTerm(term Expr) Sort {
	// In Go, all nodes have NodeSort(). For Apply nodes, this is
	// the range of the function sort, matching Python's term.rep.sort.rng.
	return term.NodeSort()
}

// GetDefaultSort returns the default sort for the given signature,
// creating it if necessary (for version <= 1.2 compatibility).
// Corresponds to Python's default_sort (ivy_logic.py:1129-1137).
func GetDefaultSort(sig *Sig) (Sort, error) {
	if sig.DefaultSort != nil {
		xtracer.Trace("ivylogic.GetDefaultSort CACHED HASH canon=%s", sig.Canon())
		return sig.DefaultSort, nil
	}
	if sig.IuCfg != nil && !VersionLE(sig.IuCfg.LanguageVersion, "1.2") {
		xtracer.Trace("ivylogic.GetDefaultSort VERSION_BLOCK HASH canon=%s", sig.Canon())
		return nil, &IvyError{Msg: "unspecified type"}
	}
	// Create default sort 'S' and add it to the signature
	ds := &UninterpretedSort{Name: "S"}
	sig.Sorts.Set("S", ds)
	sig.DefaultSort = ds
	xtracer.Trace("ivylogic.GetDefaultSort CREATED_S HASH canon=%s", sig.Canon())
	return ds, nil
}

// Sorts returns all sorts in the given signature as a slice.
// Corresponds to Python's sorts() (ivy_logic.py:1188-1189).
func Sorts(sig *Sig) []Sort {
	result := make([]Sort, 0, sig.Sorts.Len())
	for _, s := range sig.Sorts.All() {
		result = append(result, s)
	}
	return result
}

// IsEnumerated returns true if the term is a function application with
// an EnumeratedSort. Corresponds to Python's is_enumerated (ivy_logic.py:1150-1151).
func IsEnumerated(term Expr) bool {
	return IsApp(term) && IsEnumeratedSort(term.NodeSort())
}

// IsCanonicalSort returns true if the sort is canonical — i.e., it is
// not an uninterpreted sort that maps to another uninterpreted sort via
// the interpretation. Corresponds to Python's is_canonical_sort (ivy_logic.py:1451-1455).
func IsCanonicalSort(sig *Sig, sort Sort) bool {
	if _, ok := sort.(*UninterpretedSort); ok {
		interp, exists := sig.Interp[IvySortName(sort)]
		if !exists {
			return true
		}
		_, isUI := interp.(*UninterpretedSort)
		return !isUI
	}
	return true
}

// CanonizeSort follows the interpretation chain for uninterpreted sorts
// until a canonical sort is reached. Corresponds to Python's canonize_sort
// (ivy_logic.py:1457-1462).
func CanonizeSort(sig *Sig, sort Sort) Sort {
	if _, ok := sort.(*UninterpretedSort); ok {
		interp, exists := sig.Interp[IvySortName(sort)]
		if exists {
			if uiSort, ok := interp.(*UninterpretedSort); ok {
				return CanonizeSort(sig, uiSort)
			}
		}
	}
	return sort
}
