package abc

import (
	"errors"
	"slices"
	"testing"
)

func msatVectorTestClauses(count int) []*MsatClause {
	result := make([]*MsatClause, count)
	for index := range result {
		result[index] = &MsatClause{number: index}
	}
	return result
}

func TestMsatClauseVectorOperationsIdentityAndGrowth(t *testing.T) {
	vector, err := NewMsatClauseVector(1, 32)
	if err != nil || vector.Len() != 0 || vector.Capacity() != 16 || vector.Values() == nil {
		t.Fatal(vector, err)
	}
	clauses := msatVectorTestClauses(18)
	clauses[5] = nil
	for _, clause := range clauses {
		if err = vector.Push(clause); err != nil {
			t.Fatal(err)
		}
	}
	if vector.Len() != 18 || vector.Capacity() != 32 || !slices.Equal(vector.Values(), clauses) {
		t.Fatal(vector.Len(), vector.Capacity(), vector.Values())
	}
	values := vector.Values()
	values[0] = nil
	if got, err := vector.At(0); err != nil || got != clauses[0] {
		t.Fatal("pointer array aliases", got, err)
	}
	replacement := &MsatClause{number: 99}
	if err = vector.Set(1, replacement); err != nil {
		t.Fatal(err)
	}
	clauses[1] = replacement
	if got, err := vector.At(1); err != nil || got != replacement {
		t.Fatal(got, err)
	}
	if got, err := vector.Pop(); err != nil || got != clauses[17] || vector.Len() != 17 {
		t.Fatal(got, err, vector.Len())
	}
	if err = vector.Shrink(3); err != nil || !slices.Equal(vector.Values(), clauses[:3]) {
		t.Fatal(err, vector.Values())
	}
	if err = vector.Clear(); err != nil || vector.Len() != 0 || vector.Capacity() != 32 {
		t.Fatal(err, vector.Len(), vector.Capacity())
	}
	if err = vector.Push(clauses[0]); err != nil {
		t.Fatal(err)
	}
}

func TestMsatClauseVectorLimitsLifecycleAndErrors(t *testing.T) {
	for _, test := range []struct {
		capacity, maximum int
		want              error
	}{
		{-1, 0, ErrDimension},
		{0, -1, ErrDimension},
		{1, 8, ErrLimit},
		{9, 8, ErrLimit},
	} {
		vector, err := NewMsatClauseVector(test.capacity, test.maximum)
		if vector != nil || !errors.Is(err, test.want) {
			t.Fatal(test, vector, err)
		}
	}
	vector, err := NewMsatClauseVector(0, 8)
	if err != nil {
		t.Fatal(err)
	}
	clauses := msatVectorTestClauses(9)
	for index := 0; index < 8; index++ {
		if err = vector.Push(clauses[index]); err != nil {
			t.Fatal(index, err)
		}
	}
	before := vector.Values()
	if err = vector.Push(clauses[8]); !errors.Is(err, ErrLimit) || !slices.Equal(before, vector.Values()) {
		t.Fatal(err, vector.Values())
	}
	for _, operation := range []func() error{
		func() error { return vector.Grow(-1) },
		func() error { return vector.Grow(9) },
		func() error { return vector.Shrink(-1) },
		func() error { return vector.Shrink(9) },
		func() error { return vector.Set(-1, nil) },
		func() error { _, err := vector.At(8); return err },
	} {
		before = vector.Values()
		if err = operation(); err == nil || !slices.Equal(before, vector.Values()) {
			t.Fatal(err, before, vector.Values())
		}
	}
	empty, _ := NewMsatClauseVector(0, 4)
	if clause, err := empty.Pop(); clause != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(clause, err)
	}
	vector.Close()
	vector.Close()
	if vector.Len() != 0 || vector.Capacity() != 0 || vector.Values() != nil {
		t.Fatal("closed vector exposes storage")
	}
	if err = vector.Push(nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = (*MsatClauseVector)(nil).Push(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzMsatClauseVectorOperations(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {0, 1, 0, 2, 1}, {0, 255, 3, 0, 4, 2}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		const maximum = 32
		vector, err := NewMsatClauseVector(0, maximum)
		if err != nil {
			t.Fatal(err)
		}
		clauses := msatVectorTestClauses(256)
		model := make([]*MsatClause, 0, maximum)
		for index := 0; index+1 < len(operations); index += 2 {
			op, raw := operations[index]%6, operations[index+1]
			clause := clauses[raw]
			switch op {
			case 0:
				err = vector.Push(clause)
				if len(model) == maximum {
					if !errors.Is(err, ErrLimit) {
						t.Fatal("push limit", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					model = append(model, clause)
				}
			case 1:
				got, popErr := vector.Pop()
				if len(model) == 0 {
					if got != nil || !errors.Is(popErr, ErrDimension) {
						t.Fatal(got, popErr)
					}
				} else {
					want := model[len(model)-1]
					model = model[:len(model)-1]
					if popErr != nil || got != want {
						t.Fatal(got, want, popErr)
					}
				}
			case 2:
				if err = vector.Clear(); err != nil {
					t.Fatal(err)
				}
				model = model[:0]
			case 3:
				if len(model) != 0 {
					position := int(raw) % len(model)
					if err = vector.Set(position, clause); err != nil {
						t.Fatal(err)
					}
					model[position] = clause
				}
			case 4:
				if len(model) != 0 {
					size := int(raw) % (len(model) + 1)
					if err = vector.Shrink(size); err != nil {
						t.Fatal(err)
					}
					model = model[:size]
				}
			case 5:
				if len(model) != 0 {
					position := int(raw) % len(model)
					got, atErr := vector.At(position)
					if atErr != nil || got != model[position] {
						t.Fatal(got, model[position], atErr)
					}
				}
			}
			if vector.Len() != len(model) || !slices.Equal(vector.Values(), model) {
				t.Fatal(op, raw, vector.Values(), model)
			}
		}
	})
}
