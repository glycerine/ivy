package typeinfer

import (
	"fmt"

	"github.com/glycerine/goivy/logic"
)

// SortOrVar is implemented by both logic.Sort and *SortVar.
type SortOrVar interface {
	fmt.Stringer
	isSortOrVar()
}

// SortWrapper wraps a logic.Sort to implement SortOrVar.
type SortWrapper struct {
	Sort logic.Sort
}

func (sw *SortWrapper) String() string { return sw.Sort.String() }
func (sw *SortWrapper) isSortOrVar()   {}

// Wrap converts a logic.Sort to SortOrVar.
func Wrap(s logic.Sort) SortOrVar {
	return &SortWrapper{Sort: s}
}

// Unwrap extracts the logic.Sort from a SortOrVar, returning nil if it's a SortVar.
func Unwrap(s SortOrVar) logic.Sort {
	if sw, ok := s.(*SortWrapper); ok {
		return sw.Sort
	}
	return nil
}

// SortVar is a mutable sort variable used for union-find type inference.
type SortVar struct {
	Instance SortOrVar // nil, another *SortVar, or a *SortWrapper
}

func NewSortVar() *SortVar {
	return &SortVar{}
}

func (sv *SortVar) String() string {
	if sv.Instance != nil {
		return sv.Instance.String()
	}
	return fmt.Sprintf("SortVar(%p)", sv)
}

func (sv *SortVar) isSortOrVar() {}

// FunctionSortVar is a FunctionSort whose elements are SortOrVar (not logic.Sort).
// This mirrors Python's ability to store SortVar objects inside FunctionSort.
// It is used during type inference when we need to unify a sort variable
// with a function sort whose domain/range may themselves be sort variables.
type FunctionSortVar struct {
	Sorts []SortOrVar // last element is range, rest is domain
}

func NewFunctionSortVar(sorts ...SortOrVar) *FunctionSortVar {
	cp := make([]SortOrVar, len(sorts))
	copy(cp, sorts)
	return &FunctionSortVar{Sorts: cp}
}

func (fsv *FunctionSortVar) Arity() int { return len(fsv.Sorts) - 1 }

func (fsv *FunctionSortVar) String() string {
	parts := make([]string, len(fsv.Sorts)-1)
	for i := 0; i < len(fsv.Sorts)-1; i++ {
		parts[i] = fsv.Sorts[i].String()
	}
	rng := fsv.Sorts[len(fsv.Sorts)-1].String()
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " * "
		}
		result += p
	}
	return result + " -> " + rng
}

func (fsv *FunctionSortVar) isSortOrVar() {}
