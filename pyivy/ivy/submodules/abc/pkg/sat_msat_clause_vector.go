package abc

// Owned clause-pointer vector from src/sat/msat/msatClauseVec.c.

import "slices"

// MsatClause is the concrete clause identity shared by the MSAT port. Its
// behavior is added with the clause engine; vectors deliberately store stable
// pointers because watched lists and reason arrays compare clause identity.
type MsatClause struct {
	number                 int
	coreID                 int
	learned, marked, typeA bool
	literals               []SatLit
	activity               float32
	closed                 bool
}

// MsatClauseVector owns a bounded ordered array of clause references.
type MsatClauseVector struct {
	values  []*MsatClause
	size    int
	maximum int
	closed  bool
}

func NewMsatClauseVector(capacity, maximum int) (*MsatClauseVector, error) {
	maximum, err := msatVectorMaximum(maximum)
	if err != nil || capacity < 0 {
		return nil, ErrDimension
	}
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity > maximum {
		return nil, ErrLimit
	}
	return &MsatClauseVector{values: make([]*MsatClause, capacity), maximum: maximum}, nil
}

func (v *MsatClauseVector) checkLive() error {
	if v == nil {
		return ErrDimension
	}
	if v.closed || v.maximum <= 0 || v.size < 0 || v.size > len(v.values) || len(v.values) > v.maximum {
		return ErrInvariant
	}
	return nil
}

func (v *MsatClauseVector) Len() int {
	if v == nil || v.closed {
		return 0
	}
	return v.size
}

func (v *MsatClauseVector) Capacity() int {
	if v == nil || v.closed {
		return 0
	}
	return len(v.values)
}

// Values copies the pointer array while retaining clause identity.
func (v *MsatClauseVector) Values() []*MsatClause {
	if v == nil || v.closed || v.size < 0 || v.size > len(v.values) {
		return nil
	}
	return slices.Clone(v.values[:v.size])
}

func (v *MsatClauseVector) Grow(capacity int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(v.values) {
		return nil
	}
	if capacity > v.maximum {
		return ErrLimit
	}
	grown := make([]*MsatClause, capacity)
	copy(grown, v.values[:v.size])
	v.values = grown
	return nil
}

func (v *MsatClauseVector) Shrink(size int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if size < 0 || size > v.size {
		return ErrDimension
	}
	clear(v.values[size:v.size])
	v.size = size
	return nil
}

func (v *MsatClauseVector) Clear() error { return v.Shrink(0) }

func (v *MsatClauseVector) Push(clause *MsatClause) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if v.size == len(v.values) {
		capacity := min(16, v.maximum)
		if len(v.values) >= 16 {
			if len(v.values) > v.maximum/2 {
				capacity = v.maximum
			} else {
				capacity = 2 * len(v.values)
			}
		}
		if capacity <= len(v.values) {
			return ErrLimit
		}
		if err := v.Grow(capacity); err != nil {
			return err
		}
	}
	v.values[v.size] = clause
	v.size++
	return nil
}

func (v *MsatClauseVector) Pop() (*MsatClause, error) {
	if err := v.checkLive(); err != nil {
		return nil, err
	}
	if v.size == 0 {
		return nil, ErrDimension
	}
	v.size--
	clause := v.values[v.size]
	v.values[v.size] = nil
	return clause, nil
}

func (v *MsatClauseVector) Set(index int, clause *MsatClause) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if index < 0 || index >= v.size {
		return ErrDimension
	}
	v.values[index] = clause
	return nil
}

func (v *MsatClauseVector) At(index int) (*MsatClause, error) {
	if err := v.checkLive(); err != nil {
		return nil, err
	}
	if index < 0 || index >= v.size {
		return nil, ErrDimension
	}
	return v.values[index], nil
}

func (v *MsatClauseVector) Close() {
	if v == nil || v.closed {
		return
	}
	clear(v.values)
	v.values, v.size = nil, 0
	v.closed = true
}
