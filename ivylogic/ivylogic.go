package ivylogic

import (
	"fmt"

	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
)

// SortName extracts the name from a sort.
func SortName(s lg.Sort) string {
	switch t := s.(type) {
	case *lg.UninterpretedSort:
		return t.Name
	case *lg.EnumeratedSort:
		return t.Name
	case *lg.RangeSort:
		return t.Name
	case *lg.BooleanSort:
		return "bool"
	case *lg.TopSort:
		return t.Name
	case *lg.FunctionSort:
		return t.String()
	default:
		return s.String()
	}
}

// SortDomain returns the domain sorts of a sort, or nil for first-order sorts.
func SortDomain(s lg.Sort) []lg.Sort {
	if fs, ok := s.(*lg.FunctionSort); ok {
		return fs.Domain()
	}
	return nil
}

// SortRange returns the range sort of a sort, or the sort itself for first-order sorts.
func SortRange(s lg.Sort) lg.Sort {
	if fs, ok := s.(*lg.FunctionSort); ok {
		return fs.Range()
	}
	return s
}

// IsRelationalSort returns true if the sort is a relation sort
// (a function sort with Boolean range, or Boolean itself).
func IsRelationalSort(s lg.Sort) bool {
	if _, ok := s.(*lg.BooleanSort); ok {
		return true
	}
	if fs, ok := s.(*lg.FunctionSort); ok {
		return lg.SortEqual(fs.Range(), lg.Boolean)
	}
	return false
}

// RelationSort creates a relation sort (function sort with Boolean range).
func RelationSort(dom []lg.Sort) lg.Sort {
	if len(dom) == 0 {
		return lg.Boolean
	}
	sorts := make([]lg.Sort, len(dom)+1)
	copy(sorts, dom)
	sorts[len(dom)] = lg.Boolean
	fs, err := lg.NewFunctionSort(sorts...)
	if err != nil {
		return lg.Boolean
	}
	return fs
}

// FuncConstSort creates a function sort, or returns the single sort if arity 0.
func FuncConstSort(sorts ...lg.Sort) lg.Sort {
	if len(sorts) == 1 {
		return sorts[0]
	}
	fs, err := lg.NewFunctionSort(sorts...)
	if err != nil {
		return sorts[len(sorts)-1]
	}
	return fs
}

// TopFunctionSort creates a polymorphic function sort of given arity.
func TopFunctionSort(arity int) lg.Sort {
	if arity == 0 {
		return &lg.TopSort{Name: "alpha"}
	}
	sorts := make([]lg.Sort, arity+1)
	for i := range sorts {
		sorts[i] = &lg.TopSort{Name: alphaName(i)}
	}
	fs, _ := lg.NewFunctionSort(sorts...)
	return fs
}

func alphaName(idx int) string {
	if idx == 0 {
		return "alpha"
	}
	return fmt.Sprintf("alpha%d", idx)
}

// --- Type predicates ---

// IsVariable returns true if the node is a logic.Variable.
func IsVariable(n lg.Expr) bool {
	_, ok := n.(*lg.Variable)
	return ok
}

// IsConstant returns true if the node is a logic.Symbol.
func IsConstant(n lg.Expr) bool {
	_, ok := n.(*lg.Symbol)
	return ok
}

// IsApp returns true if the node is a function application,
// a constant, or a 0-arity named binder.
func IsApp(n lg.Expr) bool {
	switch t := n.(type) {
	case *lg.Apply:
		return true
	case *lg.Symbol:
		return true
	case *lg.NamedBinder:
		return len(t.Variables) == 0
	}
	return false
}

// IsAtom returns true if the node is an atomic formula.
func IsAtom(n lg.Expr) bool {
	if _, ok := n.(*lg.Eq); ok {
		return true
	}
	if IsApp(n) && lg.SortEqual(n.NodeSort(), lg.Boolean) {
		return true
	}
	return false
}

// IsRelApp returns true if the node is a relation application.
func IsRelApp(n lg.Expr) bool {
	app, ok := n.(*lg.Apply)
	if !ok {
		return false
	}
	if c, ok := app.Func.(*lg.Symbol); ok {
		return IsRelationalSort(c.CSort)
	}
	return false
}

// IsForall returns true if the node is a ForAll.
func IsForall(n lg.Expr) bool {
	_, ok := n.(*lg.ForAll)
	return ok
}

// IsExists returns true if the node is an Exists.
func IsExists(n lg.Expr) bool {
	_, ok := n.(*lg.Exists)
	return ok
}

// IsLambda returns true if the node is a Lambda.
func IsLambda(n lg.Expr) bool {
	_, ok := n.(*lg.Lambda)
	return ok
}

// IsQuantifier returns true if the node is ForAll or Exists.
func IsQuantifier(n lg.Expr) bool {
	return IsForall(n) || IsExists(n)
}

// IsBinder returns true for ForAll, Exists, Lambda, NamedBinder, or Some.
func IsBinder(n lg.Expr) bool {
	switch n.(type) {
	case *lg.ForAll, *lg.Exists, *lg.Lambda, *lg.NamedBinder, *Some:
		return true
	}
	return false
}

// IsNamedBinder returns true if the node is a NamedBinder.
func IsNamedBinder(n lg.Expr) bool {
	_, ok := n.(*lg.NamedBinder)
	return ok
}

// IsTemporal returns true if the node is a temporal operator.
func IsTemporal(n lg.Expr) bool {
	switch n.(type) {
	case *lg.Globally, *lg.Eventually, *lg.WhenOperator:
		return true
	}
	return false
}

// HasTemporal returns true if the formula contains a temporal operator.
func HasTemporal(n lg.Expr) bool {
	if IsTemporal(n) {
		return true
	}
	for _, c := range n.Children() {
		if HasTemporal(c) {
			return true
		}
	}
	return false
}

// IsEq returns true if the node is an Eq.
func IsEq(n lg.Expr) bool {
	_, ok := n.(*lg.Eq)
	return ok
}

// IsIte returns true if the node is an Ite.
func IsIte(n lg.Expr) bool {
	_, ok := n.(*lg.Ite)
	return ok
}

// IsEnumeratedSort returns true if the sort is an EnumeratedSort.
func IsEnumeratedSort(s lg.Sort) bool {
	_, ok := s.(*lg.EnumeratedSort)
	return ok
}

// IsRangeSort returns true if the sort is a RangeSort.
func IsRangeSort(s lg.Sort) bool {
	_, ok := s.(*lg.RangeSort)
	return ok
}

// IsBooleanSort returns true if the sort is the Boolean sort.
func IsBooleanSort(s lg.Sort) bool {
	return lg.SortEqual(s, lg.Boolean)
}

// IsBoolean returns true if the term has Boolean sort.
func IsBoolean(n lg.Expr) bool {
	return lg.SortEqual(n.NodeSort(), lg.Boolean)
}

// IsFirstOrderSort returns true if the sort is an UninterpretedSort.
func IsFirstOrderSort(s lg.Sort) bool {
	_, ok := s.(*lg.UninterpretedSort)
	return ok
}

// IsFunctionSort returns true if the sort is a FunctionSort.
func IsFunctionSort(s lg.Sort) bool {
	_, ok := s.(*lg.FunctionSort)
	return ok
}

// IsTopSort returns true if the sort is a TopSort.
func IsTopSort(s lg.Sort) bool {
	_, ok := s.(*lg.TopSort)
	return ok
}

// IsIndividual returns true if the term has a non-Boolean sort.
func IsIndividual(n lg.Expr) bool {
	return !lg.SortEqual(n.NodeSort(), lg.Boolean)
}

// IsNumeral returns true if the node is a numeral constant.
func IsNumeral(n lg.Expr) bool {
	c, ok := n.(*lg.Symbol)
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
func IsLiteralString(n lg.Expr) bool {
	c, ok := n.(*lg.Symbol)
	if !ok {
		return false
	}
	return len(c.Name) > 0 && c.Name[0] == '"'
}

// IsConcretetlySorted returns true if the term has no TopSort or polymorphic
// elements.
func IsConcretetlySorted(n lg.Expr) bool {
	return !lg.ContainsTopSort(n) && !lg.IsPolymorphic(n)
}

// IsTrue returns true if the node is logical true (empty And).
func IsTrue(n lg.Expr) bool {
	return lg.IsTrue(n)
}

// IsFalse returns true if the node is logical false (empty Or).
func IsFalse(n lg.Expr) bool {
	return lg.IsFalse(n)
}

// IsGprop returns true if the formula is Globally(phi) where phi has
// no temporal operators.
func IsGprop(n lg.Expr) bool {
	g, ok := n.(*lg.Globally)
	if !ok {
		return false
	}
	return !HasTemporal(g.Body)
}

// --- Equals symbol ---

// Equals is the built-in equality symbol.
var Equals = lg.NewSymbol("=", RelationSort([]lg.Sort{lg.TopS, lg.TopS}))

// IsEquals returns true if the constant is the equality symbol.
func IsEquals(c *lg.Symbol) bool {
	return c.Name == "="
}

// NewEquals creates an Eq node from two terms.
func NewEquals(x, y lg.Expr) *lg.Eq {
	return &lg.Eq{T1: x, T2: y}
}

// --- Sort predicates ---

// IsUISort returns true if the sort is specifically an UninterpretedSort
// (not a subclass).
func IsUISort(s lg.Sort) bool {
	_, ok := s.(*lg.UninterpretedSort)
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
func NormalizeSymbol(sym *lg.Symbol) *lg.Symbol {
	if iu.IvyUsePolymorphicMacros {
		if canonical, ok := PolymorphicMacrosMap[sym.Name]; ok {
			return lg.NewSymbol(canonical, sym.CSort)
		}
	}
	return sym
}

// GetSortTerm returns the sort of a term.
// Corresponds to Python's get_sort_term (ivy_logic.py:384-387).
// If the term has a .sort attribute, return it; otherwise return rep.sort.rng.
func GetSortTerm(term lg.Expr) lg.Sort {
	// In Go, all nodes have NodeSort(). For Apply nodes, this is
	// the range of the function sort, matching Python's term.rep.sort.rng.
	return term.NodeSort()
}

// GetDefaultSort returns the default sort for the given signature,
// creating it if necessary (for version <= 1.2 compatibility).
// Corresponds to Python's default_sort (ivy_logic.py:1118-1126).
func GetDefaultSort(sig *Sig) lg.Sort {
	if sig.DefaultSort != nil {
		return sig.DefaultSort
	}
	// Create default sort 'S' and add it to the signature
	ds := &lg.UninterpretedSort{Name: "S"}
	sig.Sorts["S"] = ds
	sig.DefaultSort = ds
	return ds
}

// Sorts returns all sorts in the given signature as a slice.
// Corresponds to Python's sorts() (ivy_logic.py:1188-1189).
func Sorts(sig *Sig) []lg.Sort {
	result := make([]lg.Sort, 0, len(sig.Sorts))
	for _, s := range sig.Sorts {
		result = append(result, s)
	}
	return result
}

// IsEnumerated returns true if the term is a function application with
// an EnumeratedSort. Corresponds to Python's is_enumerated (ivy_logic.py:1150-1151).
func IsEnumerated(term lg.Expr) bool {
	return IsApp(term) && IsEnumeratedSort(term.NodeSort())
}

// IsCanonicalSort returns true if the sort is canonical — i.e., it is
// not an uninterpreted sort that maps to another uninterpreted sort via
// the interpretation. Corresponds to Python's is_canonical_sort (ivy_logic.py:1451-1455).
func IsCanonicalSort(sig *Sig, sort lg.Sort) bool {
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		interp, exists := sig.Interp[SortName(sort)]
		if !exists {
			return true
		}
		_, isUI := interp.(*lg.UninterpretedSort)
		return !isUI
	}
	return true
}

// CanonizeSort follows the interpretation chain for uninterpreted sorts
// until a canonical sort is reached. Corresponds to Python's canonize_sort
// (ivy_logic.py:1457-1462).
func CanonizeSort(sig *Sig, sort lg.Sort) lg.Sort {
	if _, ok := sort.(*lg.UninterpretedSort); ok {
		interp, exists := sig.Interp[SortName(sort)]
		if exists {
			if uiSort, ok := interp.(*lg.UninterpretedSort); ok {
				return CanonizeSort(sig, uiSort)
			}
		}
	}
	return sort
}
