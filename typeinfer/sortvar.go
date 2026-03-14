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
