package ivylogic

import (
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
	return "alpha" + string(rune('0'+idx))
}

// --- Type predicates ---

// IsVariable returns true if the node is a logic.Var.
func IsVariable(n lg.Node) bool {
	_, ok := n.(*lg.Var)
	return ok
}

// IsConstant returns true if the node is a logic.Const.
func IsConstant(n lg.Node) bool {
	_, ok := n.(*lg.Const)
	return ok
}

// IsApp returns true if the node is a function application,
// a constant, or a 0-arity named binder.
func IsApp(n lg.Node) bool {
	switch t := n.(type) {
	case *lg.Apply:
		return true
	case *lg.Const:
		return true
	case *lg.NamedBinder:
		return len(t.Variables) == 0
	}
	return false
}

// IsAtom returns true if the node is an atomic formula.
func IsAtom(n lg.Node) bool {
	if _, ok := n.(*lg.Eq); ok {
		return true
	}
	if IsApp(n) && lg.SortEqual(n.NodeSort(), lg.Boolean) {
		return true
	}
	return false
}

// IsRelApp returns true if the node is a relation application.
func IsRelApp(n lg.Node) bool {
	app, ok := n.(*lg.Apply)
	if !ok {
		return false
	}
	if c, ok := app.Func.(*lg.Const); ok {
		return IsRelationalSort(c.CSort)
	}
	return false
}

// IsForall returns true if the node is a ForAll.
func IsForall(n lg.Node) bool {
	_, ok := n.(*lg.ForAll)
	return ok
}

// IsExists returns true if the node is an Exists.
func IsExists(n lg.Node) bool {
	_, ok := n.(*lg.Exists)
	return ok
}

// IsLambda returns true if the node is a Lambda.
func IsLambda(n lg.Node) bool {
	_, ok := n.(*lg.Lambda)
	return ok
}

// IsQuantifier returns true if the node is ForAll or Exists.
func IsQuantifier(n lg.Node) bool {
	return IsForall(n) || IsExists(n)
}

// IsBinder returns true for ForAll, Exists, Lambda, NamedBinder, or Some.
func IsBinder(n lg.Node) bool {
	switch n.(type) {
	case *lg.ForAll, *lg.Exists, *lg.Lambda, *lg.NamedBinder, *Some:
		return true
	}
	return false
}

// IsNamedBinder returns true if the node is a NamedBinder.
func IsNamedBinder(n lg.Node) bool {
	_, ok := n.(*lg.NamedBinder)
	return ok
}

// IsTemporal returns true if the node is a temporal operator.
func IsTemporal(n lg.Node) bool {
	switch n.(type) {
	case *lg.Globally, *lg.Eventually, *lg.WhenOperator:
		return true
	}
	return false
}

// HasTemporal returns true if the formula contains a temporal operator.
func HasTemporal(n lg.Node) bool {
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
func IsEq(n lg.Node) bool {
	_, ok := n.(*lg.Eq)
	return ok
}

// IsIte returns true if the node is an Ite.
func IsIte(n lg.Node) bool {
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
func IsBoolean(n lg.Node) bool {
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
func IsIndividual(n lg.Node) bool {
	return !lg.SortEqual(n.NodeSort(), lg.Boolean)
}

// IsNumeral returns true if the node is a numeral constant.
func IsNumeral(n lg.Node) bool {
	c, ok := n.(*lg.Const)
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
func IsLiteralString(n lg.Node) bool {
	c, ok := n.(*lg.Const)
	if !ok {
		return false
	}
	return len(c.Name) > 0 && c.Name[0] == '"'
}

// IsConcretetlySorted returns true if the term has no TopSort or polymorphic
// elements.
func IsConcretetlySorted(n lg.Node) bool {
	return !lg.ContainsTopSort(n) && !lg.IsPolymorphic(n)
}

// IsTrue returns true if the node is logical true (empty And).
func IsTrue(n lg.Node) bool {
	a, ok := n.(*lg.And)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is logical false (empty Or).
func IsFalse(n lg.Node) bool {
	o, ok := n.(*lg.Or)
	return ok && len(o.Terms) == 0
}

// IsGprop returns true if the formula is Globally(phi) where phi has
// no temporal operators.
func IsGprop(n lg.Node) bool {
	g, ok := n.(*lg.Globally)
	if !ok {
		return false
	}
	return !HasTemporal(g.Body)
}

// --- Equals symbol ---

// Equals is the built-in equality symbol.
var Equals = lg.NewConst("=", RelationSort([]lg.Sort{lg.TopS, lg.TopS}))

// IsEquals returns true if the constant is the equality symbol.
func IsEquals(c *lg.Const) bool {
	return c.Name == "="
}

// NewEquals creates an Eq node from two terms.
func NewEquals(x, y lg.Node) *lg.Eq {
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
