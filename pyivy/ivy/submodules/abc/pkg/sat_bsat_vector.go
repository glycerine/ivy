package abc

// Port of src/sat/bsat/satVec.h using concrete checked vectors.

import (
	"math"
	"slices"
)

const (
	SatVariableUndefined = -1
	SatLiteralUndefined  = -2
)

type SatLBool int8

const (
	SatLUndefined SatLBool = 0
	SatLTrue      SatLBool = 1
	SatLFalse     SatLBool = -1
)

type SatLegacyStats struct {
	Starts, Clauses, Learnts                        uint64
	Decisions, Propagations, Inspections, Conflicts uint64
	ClauseLiterals, LearntLiterals, TotalLiterals   uint64
}

type SatIntVector struct {
	data    []int
	maximum int
	stopped bool
}

func NewSatIntVector(maximum int) (*SatIntVector, error) {
	if maximum < 0 {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = 16 << 20
	}
	capacity := min(4, maximum)
	return &SatIntVector{data: make([]int, 0, capacity), maximum: maximum}, nil
}

func (v *SatIntVector) Size() int {
	if v == nil || v.stopped {
		return 0
	}
	return len(v.data)
}

func (v *SatIntVector) Capacity() int {
	if v == nil || v.stopped {
		return 0
	}
	return cap(v.data)
}

func (v *SatIntVector) Values() []int {
	if v == nil || v.stopped {
		return nil
	}
	return slices.Clone(v.data)
}

func (v *SatIntVector) At(index int) (int, error) {
	if v == nil || v.stopped {
		return 0, ErrInvariant
	}
	if index < 0 || index >= len(v.data) {
		return 0, ErrInvalidLiteral
	}
	return v.data[index], nil
}

func (v *SatIntVector) Set(index, value int) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if index < 0 || index >= len(v.data) {
		return ErrInvalidLiteral
	}
	v.data[index] = value
	return nil
}

func satLegacyVectorNextCapacity(capacity, maximum int) int {
	if capacity == 0 {
		return min(4, maximum)
	}
	next := capacity * 2
	if capacity >= 4 {
		next = (capacity / 2) * 3
	}
	if next <= capacity || next > maximum {
		return maximum
	}
	return next
}

func (v *SatIntVector) Push(value int) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if len(v.data) == v.maximum {
		return ErrLimit
	}
	if len(v.data) == cap(v.data) {
		capacity := satLegacyVectorNextCapacity(cap(v.data), v.maximum)
		grown := make([]int, len(v.data), capacity)
		copy(grown, v.data)
		v.data = grown
	}
	v.data = append(v.data, value)
	return nil
}

func (v *SatIntVector) Shrink(size int) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if size < 0 || size > len(v.data) {
		return ErrDimension
	}
	v.data = v.data[:size]
	return nil
}

func (v *SatIntVector) Pop() (int, error) {
	if v == nil || v.stopped {
		return 0, ErrInvariant
	}
	if len(v.data) == 0 {
		return 0, ErrDimension
	}
	last := len(v.data) - 1
	value := v.data[last]
	v.data = v.data[:last]
	return value, nil
}

func (v *SatIntVector) Remove(value int) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	index := slices.Index(v.data, value)
	if index < 0 {
		return ErrInvalidLiteral
	}
	copy(v.data[index:], v.data[index+1:])
	v.data = v.data[:len(v.data)-1]
	return nil
}

func (v *SatIntVector) Stop() {
	if v == nil || v.stopped {
		return
	}
	v.data, v.stopped = nil, true
}

type SatPointerVector struct {
	data    []uintptr
	maximum int
	stopped bool
}

func NewSatPointerVector(maximum int) (*SatPointerVector, error) {
	if maximum < 0 {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = 16 << 20
	}
	return &SatPointerVector{data: make([]uintptr, 0, min(4, maximum)), maximum: maximum}, nil
}

func (v *SatPointerVector) Size() int {
	if v == nil || v.stopped {
		return 0
	}
	return len(v.data)
}

func (v *SatPointerVector) Capacity() int {
	if v == nil || v.stopped {
		return 0
	}
	return cap(v.data)
}

func (v *SatPointerVector) Values() []uintptr {
	if v == nil || v.stopped {
		return nil
	}
	return slices.Clone(v.data)
}

func (v *SatPointerVector) At(index int) (uintptr, error) {
	if v == nil || v.stopped {
		return 0, ErrInvariant
	}
	if index < 0 || index >= len(v.data) {
		return 0, ErrInvalidLiteral
	}
	return v.data[index], nil
}

func (v *SatPointerVector) Set(index int, value uintptr) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if index < 0 || index >= len(v.data) {
		return ErrInvalidLiteral
	}
	v.data[index] = value
	return nil
}

func (v *SatPointerVector) Push(value uintptr) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if len(v.data) == v.maximum {
		return ErrLimit
	}
	if len(v.data) == cap(v.data) {
		capacity := satLegacyVectorNextCapacity(cap(v.data), v.maximum)
		grown := make([]uintptr, len(v.data), capacity)
		copy(grown, v.data)
		v.data = grown
	}
	v.data = append(v.data, value)
	return nil
}

func (v *SatPointerVector) Shrink(size int) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	if size < 0 || size > len(v.data) {
		return ErrDimension
	}
	v.data = v.data[:size]
	return nil
}

func (v *SatPointerVector) Remove(value uintptr) error {
	if v == nil || v.stopped {
		return ErrInvariant
	}
	index := slices.Index(v.data, value)
	if index < 0 {
		return ErrInvalidLiteral
	}
	copy(v.data[index:], v.data[index+1:])
	v.data = v.data[:len(v.data)-1]
	return nil
}

func (v *SatPointerVector) Stop() {
	if v == nil || v.stopped {
		return
	}
	v.data, v.stopped = nil, true
}

func (l SatLit) DIMACS() int {
	value := int(l.Var()) + 1
	if l.Negative() {
		return -value
	}
	return value
}

func ParseSatDIMACSLiteral(value int) (SatLit, error) {
	if value == 0 || value == math.MinInt {
		return 0, ErrInvalidLiteral
	}
	negative := value < 0
	if negative {
		value = -value
	}
	variable := value - 1
	if uint64(variable) > uint64(MaxLiteralID) {
		return 0, ErrInvalidLiteral
	}
	return MakeSatLit(SatVar(variable), negative)
}

func (l SatLit) ValidForVariables(variableCount int) bool {
	return variableCount >= 0 && l.Valid() && uint64(l.Var()) < uint64(variableCount)
}
