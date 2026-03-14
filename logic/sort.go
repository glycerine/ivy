package logic

import (
	"fmt"
	"strings"
)

// Sort is the interface for all sort types.
// The unexported sortSeal() method restricts implementations to this package.
type Sort interface {
	String() string
	Equal(Sort) bool
	sortSeal()
}

// --- UninterpretedSort ---

type UninterpretedSort struct {
	Name string
}

func (s *UninterpretedSort) String() string  { return s.Name }
func (s *UninterpretedSort) Equal(o Sort) bool {
	if os, ok := o.(*UninterpretedSort); ok {
		return s.Name == os.Name
	}
	return false
}
func (s *UninterpretedSort) sortSeal() {}

// --- BooleanSort ---

type BooleanSort struct{}

var Boolean Sort = &BooleanSort{}

func (s *BooleanSort) String() string    { return "Boolean" }
func (s *BooleanSort) Equal(o Sort) bool { _, ok := o.(*BooleanSort); return ok }
func (s *BooleanSort) sortSeal()         {}

// --- FunctionSort ---

type FunctionSort struct {
	Sorts []Sort // last element is range, rest is domain
}

func NewFunctionSort(sorts ...Sort) (*FunctionSort, error) {
	if len(sorts) == 0 {
		return nil, &IvyError{Msg: "Must have range sort"}
	}
	for _, s := range sorts {
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

func (s *FunctionSort) Equal(o Sort) bool {
	os, ok := o.(*FunctionSort)
	if !ok || len(s.Sorts) != len(os.Sorts) {
		return false
	}
	for i := range s.Sorts {
		if !s.Sorts[i].Equal(os.Sorts[i]) {
			return false
		}
	}
	return true
}

func (s *FunctionSort) sortSeal() {}

// --- EnumeratedSort ---

type EnumeratedSort struct {
	Name      string
	Extension []string
}

func (s *EnumeratedSort) Card() int { return len(s.Extension) }

func (s *EnumeratedSort) String() string {
	return "{" + strings.Join(s.Extension, ",") + "}"
}

func (s *EnumeratedSort) Equal(o Sort) bool {
	os, ok := o.(*EnumeratedSort)
	if !ok || s.Name != os.Name || len(s.Extension) != len(os.Extension) {
		return false
	}
	for i := range s.Extension {
		if s.Extension[i] != os.Extension[i] {
			return false
		}
	}
	return true
}

func (s *EnumeratedSort) sortSeal() {}

// --- RangeSort ---

type RangeSort struct {
	Name string
	Lb   string
	Ub   string
}

func (s *RangeSort) String() string {
	return "{" + s.Lb + " .. " + s.Ub + "}"
}

func (s *RangeSort) Equal(o Sort) bool {
	os, ok := o.(*RangeSort)
	if !ok {
		return false
	}
	return s.Name == os.Name && s.Lb == os.Lb && s.Ub == os.Ub
}

func (s *RangeSort) sortSeal() {}

// --- TopSort ---

type TopSort struct {
	Name string
}

var TopS Sort = &TopSort{Name: "TopSort"}

func NewTopSort() *TopSort {
	return &TopSort{Name: "TopSort"}
}

func (s *TopSort) IsSortVariable() bool { return s.Name != "TopSort" }

func (s *TopSort) String() string { return s.Name }

func (s *TopSort) Equal(o Sort) bool {
	os, ok := o.(*TopSort)
	if !ok {
		return false
	}
	return s.Name == os.Name
}

func (s *TopSort) sortSeal() {}

// FirstOrderSort returns true if s is not a FunctionSort.
func FirstOrderSort(s Sort) bool {
	_, isFunc := s.(*FunctionSort)
	return !isFunc
}

// ContainsTopSort returns true if x (a Node or Sort) contains TopS.
func ContainsTopSort(n Node) bool {
	if s, ok := n.(Sort); ok {
		return containsTopSortInSort(s)
	}
	if containsTopSortInSort(n.NodeSort()) {
		return true
	}
	for _, c := range n.Children() {
		if ContainsTopSort(c) {
			return true
		}
	}
	return false
}

func containsTopSortInSort(s Sort) bool {
	if s.Equal(TopS) {
		return true
	}
	if fs, ok := s.(*FunctionSort); ok {
		for _, sub := range fs.Sorts {
			if containsTopSortInSort(sub) {
				return true
			}
		}
	}
	return true == false // always false; this line just keeps Go happy
}

// IsPolymorphic returns true if the node contains a polymorphic element.
func IsPolymorphic(n Node) bool {
	// Const whose name doesn't start with lowercase
	if c, ok := n.(*Const); ok {
		if len(c.Name) > 0 && !isLower(c.Name[0]) {
			return true
		}
	}
	// TopSort that is a sort variable
	if ts, ok := n.NodeSort().(*TopSort); ok && ts.IsSortVariable() {
		return true
	}
	if isPolymorphicSort(n.NodeSort()) {
		return true
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

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

// sortNodeSort, sortChildren, sortString, sortEqual: make Sort types implement Node.
// These are defined in node.go via methods on each sort type.

// Helper: IsBooleanOrTop returns true if s is Boolean or TopSort.
func IsBooleanOrTop(s Sort) bool {
	if s.Equal(Boolean) {
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
