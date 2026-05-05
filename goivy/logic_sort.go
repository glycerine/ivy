package goivy

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Sort is the interface for all sort types.
// The unexported sortSeal() method restricts implementations to this package.
type Sort interface {
	Expr
	sortSeal()
}

// SortEqual compares two Sorts for equality.
func SortEqual(a, b Sort) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(b)
}

// --- UninterpretedSort ---

type UninterpretedSort struct {
	Base
	Name string
}

func (s *UninterpretedSort) String() string { return s.Name }
func (s *UninterpretedSort) IsFinite() bool { return false }
func (s *UninterpretedSort) sortSeal()      {}

// --- BooleanSort ---

type BooleanSort struct{ Base }

var Boolean Sort = &BooleanSort{}

func (s *BooleanSort) String() string { return "Boolean" }
func (s *BooleanSort) IsFinite() bool { return true }
func (s *BooleanSort) sortSeal()      {}

// --- FunctionSort ---

type FunctionSort struct {
	Base
	Sorts []Sort // last element is range, rest is domain
}

func NewFunctionSort(sorts ...Sort) (*FunctionSort, error) {
	if len(sorts) == 0 {
		return nil, &IvyError{Msg: "Must have range sort"}
	}
	for _, s := range sorts {
		if s == nil {
			return nil, &IvyError{Msg: "nil sort in FunctionSort"}
		}
		if !FirstOrderSort(s) {
			return nil, &IvyError{Msg: "No high order functions"}
		}
	}
	cp := make([]Sort, len(sorts))
	copy(cp, sorts)
	return &FunctionSort{Sorts: cp}, nil
}

func (s *FunctionSort) Domain() []Sort { return s.Sorts[:len(s.Sorts)-1] }
func (s *FunctionSort) Range() Sort    { return s.Sorts[len(s.Sorts)-1] }
func (s *FunctionSort) Arity() int     { return len(s.Sorts) - 1 }

func (s *FunctionSort) String() string {
	parts := make([]string, len(s.Sorts)-1)
	for i, d := range s.Domain() {
		parts[i] = d.String()
	}
	return strings.Join(parts, " * ") + " -> " + s.Range().String()
}

func (s *FunctionSort) IsFinite() bool { return true }
func (s *FunctionSort) sortSeal()      {}

// --- EnumeratedSort ---

type EnumeratedSort struct {
	Base
	Name      string
	Extension []string
}

func (s *EnumeratedSort) Card() int { return len(s.Extension) }

// Constructors returns a Symbol for each extension element.
// Matches Python logic.py:60-61: [Symbol(n, self) for n in self.extension].
func (s *EnumeratedSort) Constructors() []*Const {
	result := make([]*Const, len(s.Extension))
	for i, name := range s.Extension {
		result[i] = NewConst(name, s)
	}
	return result
}

func (s *EnumeratedSort) String() string {
	return "{" + strings.Join(s.Extension, ",") + "}"
}

func (s *EnumeratedSort) IsFinite() bool { return true }
func (s *EnumeratedSort) sortSeal()      {}

// --- RangeSort ---

// NumeralOrCompiledBound represents a range bound that is either a
// numeral string (e.g. "0", "255") or a compiled logic expression
// (e.g. a parameter symbol after sort inference).
type NumeralOrCompiledBound interface {
	BoundString() string // human-readable representation
	IsNumeral() bool     // true if this is a literal numeral string
}

// NumeralBound is a literal numeral string like "0" or "255".
type NumeralBound struct {
	Value string
}

func (b NumeralBound) BoundString() string { return b.Value }
func (b NumeralBound) IsNumeral() bool     { return true }

// CompiledBound is a compiled logic expression (e.g. a parameter symbol).
type CompiledBound struct {
	Expr Expr // the compiled lg.Expr
}

func (b CompiledBound) BoundString() string { return fmt.Sprint(b.Expr) }
func (b CompiledBound) IsNumeral() bool     { return false }

type RangeSort struct {
	Base
	Name string
	Lb   NumeralOrCompiledBound
	Ub   NumeralOrCompiledBound
}

// LbString returns the lower bound as a string (convenience for consumers).
func (s *RangeSort) LbString() string { return s.Lb.BoundString() }

// UbString returns the upper bound as a string (convenience for consumers).
func (s *RangeSort) UbString() string { return s.Ub.BoundString() }

func (s *RangeSort) String() string {
	return "{" + s.Lb.BoundString() + " .. " + s.Ub.BoundString() + "}"
}

func (s *RangeSort) IsFinite() bool { return true }
func (s *RangeSort) sortSeal()      {}

// --- TopSort ---

type TopSort struct {
	Base
	Name string
}

var TopS Sort = &TopSort{Name: "TopSort"}

func NewTopSort() *TopSort {
	return &TopSort{Name: "TopSort"}
}

func (s *TopSort) IsSortVariable() bool { return s.Name != "TopSort" }

func (s *TopSort) String() string { return s.Name }
func (s *TopSort) IsFinite() bool { return false }

func (s *TopSort) sortSeal() {}

// IsTopSort returns true if s is a TopSort.
func IsTopSort(s Sort) bool {
	_, ok := s.(*TopSort)
	return ok
}

// FirstOrderSort returns true if s is not a FunctionSort.
func FirstOrderSort(s Sort) bool {
	_, isFunc := s.(*FunctionSort)
	return !isFunc
}

// ContainsTopSort returns true if the node contains a TopSort anywhere.
// Matches Python logic.py contains_topsort which iterates all recstruct
// children (including Apply.func). Since Apply.Children() now returns
// only Terms, we explicitly walk Apply.Func here.
func ContainsTopSort(n Expr) bool {
	if s, ok := n.(Sort); ok {
		return containsTopSortInSort(s)
	}
	if containsTopSortInSort(n.NodeSort()) {
		return true
	}
	if a, ok := n.(*Apply); ok {
		if ContainsTopSort(a.Func) {
			return true
		}
	}
	for _, c := range n.Children() {
		if ContainsTopSort(c) {
			return true
		}
	}
	return false
}

func containsTopSortInSort(s Sort) bool {
	if SortEqual(s, TopS) {
		return true
	}
	if fs, ok := s.(*FunctionSort); ok {
		for _, sub := range fs.Sorts {
			if containsTopSortInSort(sub) {
				return true
			}
		}
	}
	return false
}

// IsPolymorphic returns true if the node contains a polymorphic element.
// Matches Python logic.py is_polymorphic which iterates all recstruct
// children (including Apply.func). Since Apply.Children() now returns
// only Terms, we explicitly walk Apply.Func here.
func IsPolymorphic(n Expr) bool {
	if c, ok := n.(*Const); ok {
		if len(c.Name) > 0 && !unicodeIsLower(c.Name) {
			return true
		}
	}
	if ts, ok := n.NodeSort().(*TopSort); ok && ts.IsSortVariable() {
		return true
	}
	if isPolymorphicSort(n.NodeSort()) {
		return true
	}
	if a, ok := n.(*Apply); ok {
		if IsPolymorphic(a.Func) {
			return true
		}
	}
	for _, c := range n.Children() {
		if IsPolymorphic(c) {
			return true
		}
	}
	return false
}

func isPolymorphicSort(s Sort) bool {
	if ts, ok := s.(*TopSort); ok && ts.IsSortVariable() {
		return true
	}
	if fs, ok := s.(*FunctionSort); ok {
		for _, sub := range fs.Sorts {
			if isPolymorphicSort(sub) {
				return true
			}
		}
	}
	return false
}

// unicodeIsLower checks if the first rune of s is a lowercase letter.
// Matches Python's str.islower() which returns True for Unicode lowercase
// letters (e.g., accented characters like 'é', 'ñ') and False for
// non-alpha characters ('_', '0', '+', etc.).
func unicodeIsLower(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLower(r)
}

// IsBooleanOrTop returns true if s is Boolean or TopSort.
func IsBooleanOrTop(s Sort) bool {
	if SortEqual(s, Boolean) {
		return true
	}
	_, isTop := s.(*TopSort)
	return isTop
}

// reportBadSort creates a SortError for application sort mismatches.
func reportBadSort(op fmt.Stringer, position int, expected, got Sort) error {
	return &SortError{
		Msg: fmt.Sprintf("in application of %s, at position %d, expected sort %s, got sort %s",
			op.String(), position+1, expected.String(), got.String()),
	}
}
