package abc

// Stand-alone tests for the concrete vector and literal helpers from satVec.h.

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"testing"
)

func TestSatIntVectorExactGrowthAndOperations(t *testing.T) {
	vector, err := NewSatIntVector(32)
	if err != nil || vector.Size() != 0 || vector.Capacity() != 4 {
		t.Fatal(vector, err)
	}
	wantCapacity := map[int]int{0: 4, 4: 4, 5: 6, 6: 6, 7: 9, 9: 9, 10: 12, 12: 12, 13: 18}
	for i := 0; i < 13; i++ {
		if err = vector.Push(i % 5); err != nil {
			t.Fatal(err)
		}
		if want, ok := wantCapacity[i+1]; ok && vector.Capacity() != want {
			t.Fatal(i+1, vector.Capacity(), want)
		}
	}
	if err = vector.Remove(2); err != nil {
		t.Fatal(err)
	}
	want := []int{0, 1, 3, 4, 0, 1, 2, 3, 4, 0, 1, 2}
	if !slices.Equal(vector.Values(), want) {
		t.Fatal(vector.Values(), want)
	}
	copyValues := vector.Values()
	copyValues[0] = 99
	if vector.Values()[0] != 0 {
		t.Fatal("values alias vector")
	}
	if err = vector.Set(0, -7); err != nil {
		t.Fatal(err)
	}
	if got, atErr := vector.At(0); atErr != nil || got != -7 {
		t.Fatal(got, atErr)
	}
	if err = vector.Set(0, 0); err != nil {
		t.Fatal(err)
	}
	value, err := vector.Pop()
	if err != nil || value != 2 {
		t.Fatal(value, err)
	}
	if err = vector.Shrink(3); err != nil || !slices.Equal(vector.Values(), []int{0, 1, 3}) || vector.Capacity() != 18 {
		t.Fatal(vector.Values(), vector.Capacity(), err)
	}
	vector.Stop()
	vector.Stop()
	if err = vector.Push(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatPointerVectorConcreteOpaqueHandles(t *testing.T) {
	vector, err := NewSatPointerVector(10)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []uintptr{0, 11, 22, 11, ^uintptr(0)} {
		if err = vector.Push(value); err != nil {
			t.Fatal(err)
		}
	}
	if vector.Capacity() != 6 || vector.Size() != 5 {
		t.Fatal(vector.Capacity(), vector.Size())
	}
	if err = vector.Remove(11); err != nil || !reflect.DeepEqual(vector.Values(), []uintptr{0, 22, 11, ^uintptr(0)}) {
		t.Fatal(vector.Values(), err)
	}
	if err = vector.Shrink(2); err != nil || !reflect.DeepEqual(vector.Values(), []uintptr{0, 22}) {
		t.Fatal(vector.Values(), err)
	}
	copyValues := vector.Values()
	copyValues[0] = 9
	if vector.Values()[0] != 0 {
		t.Fatal("values alias vector")
	}
	if err = vector.Set(1, 33); err != nil {
		t.Fatal(err)
	}
	if got, atErr := vector.At(1); atErr != nil || got != 33 {
		t.Fatal(got, atErr)
	}
	vector.Stop()
	if err = vector.Remove(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatVectorLiteralAndBooleanHelpers(t *testing.T) {
	if SatVariableUndefined != -1 || SatLiteralUndefined != -2 || SatLUndefined != 0 || SatLTrue != 1 || SatLFalse != -1 {
		t.Fatal(SatVariableUndefined, SatLiteralUndefined, SatLUndefined, SatLTrue, SatLFalse)
	}
	for variable := 0; variable < 1000; variable++ {
		for _, negative := range []bool{false, true} {
			literal, err := MakeSatLit(SatVar(variable), negative)
			if err != nil {
				t.Fatal(err)
			}
			want := variable + 1
			if negative {
				want = -want
			}
			if literal.DIMACS() != want || !literal.ValidForVariables(variable+1) || literal.ValidForVariables(variable) {
				t.Fatal(variable, negative, literal.DIMACS())
			}
			parsed, parseErr := ParseSatDIMACSLiteral(want)
			if parseErr != nil || parsed != literal || parsed.Not().Not() != parsed {
				t.Fatal(variable, negative, parsed, parseErr)
			}
		}
	}
	for _, value := range []int{0, math.MinInt, int(MaxLiteralID) + 2, -int(MaxLiteralID) - 2} {
		if _, err := ParseSatDIMACSLiteral(value); !errors.Is(err, ErrInvalidLiteral) {
			t.Fatal(value, err)
		}
	}
	if (SatLit(math.MaxUint32)).ValidForVariables(math.MaxInt) || (SatLit(0)).ValidForVariables(-1) {
		t.Fatal("invalid literal bound")
	}
	stats := SatLegacyStats{Starts: 1, Clauses: 2, Learnts: 3, Decisions: 4, Propagations: 5, Inspections: 6, Conflicts: 7, ClauseLiterals: 8, LearntLiterals: 9, TotalLiterals: 10}
	if stats.TotalLiterals != 10 || stats.Conflicts != 7 {
		t.Fatal(stats)
	}
}

func TestSatVectorErrorsAndLimits(t *testing.T) {
	if vector, err := NewSatIntVector(-1); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewSatPointerVector(-1); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	vector, _ := NewSatIntVector(2)
	if vector.Capacity() != 2 || vector.Push(1) != nil || vector.Push(2) != nil {
		t.Fatal(vector.Capacity(), vector.Values())
	}
	if err := vector.Push(3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := vector.Remove(9); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err := vector.Shrink(3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := vector.At(2); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err := vector.Set(-1, 0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	vector.Shrink(0)
	if _, err := vector.Pop(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzSatIntVectorModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		vector, _ := NewSatIntVector(512)
		var model []int
		for _, operation := range data {
			value := int(operation >> 2)
			switch operation & 3 {
			case 0:
				if err := vector.Push(value); err != nil {
					t.Fatal(err)
				}
				model = append(model, value)
			case 1:
				if len(model) == 0 {
					if _, err := vector.Pop(); !errors.Is(err, ErrDimension) {
						t.Fatal(err)
					}
					break
				}
				got, err := vector.Pop()
				want := model[len(model)-1]
				model = model[:len(model)-1]
				if err != nil || got != want {
					t.Fatal(got, want, err)
				}
			case 2:
				index := slices.Index(model, value)
				err := vector.Remove(value)
				if index < 0 {
					if !errors.Is(err, ErrInvalidLiteral) {
						t.Fatal(err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					model = append(model[:index], model[index+1:]...)
				}
			case 3:
				size := 0
				if len(model) > 0 {
					size = value % (len(model) + 1)
				}
				if err := vector.Shrink(size); err != nil {
					t.Fatal(err)
				}
				model = model[:size]
			}
			if !slices.Equal(vector.Values(), model) || vector.Size() != len(model) || vector.Capacity() < vector.Size() || vector.Capacity() > 512 {
				t.Fatal(operation, vector.Values(), model, vector.Capacity())
			}
		}
	})
}
