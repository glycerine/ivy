package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestMsatIntVectorConstructionOwnershipAndTransfer(t *testing.T) {
	vector, err := NewMsatIntVector(1, 32)
	if err != nil || vector.Len() != 0 || vector.Capacity() != 16 {
		t.Fatal(vector, err)
	}
	input := []int{3, 1, 4}
	taken, err := TakeMsatIntVector(&input, 32)
	if err != nil || input != nil || !slices.Equal(taken.Values(), []int{3, 1, 4}) || taken.Capacity() != 3 {
		t.Fatal(input, taken, err)
	}
	copyVector, err := CopyMsatIntVector(taken.Values(), 32)
	if err != nil {
		t.Fatal(err)
	}
	if err = copyVector.Set(0, 9); err != nil || !slices.Equal(taken.Values(), []int{3, 1, 4}) {
		t.Fatal("copy aliases", err, taken.Values())
	}
	clone, err := taken.Clone()
	if err != nil || clone.Capacity() != taken.Capacity() {
		t.Fatal(clone, err)
	}
	clone.Set(1, 8)
	if !slices.Equal(taken.Values(), []int{3, 1, 4}) {
		t.Fatal("clone aliases")
	}
	moved, err := taken.Transfer()
	if err != nil || taken.Len() != 0 || taken.Capacity() != 0 || !slices.Equal(moved.Values(), []int{3, 1, 4}) {
		t.Fatal(moved, err, taken.Len(), taken.Capacity())
	}
	if err = taken.Push(7); err != nil || !slices.Equal(taken.Values(), []int{7}) {
		t.Fatal("transferred source not reusable", err, taken.Values())
	}
	released, err := moved.Release()
	if err != nil || !slices.Equal(released, []int{3, 1, 4}) || moved.Len() != 0 || moved.Capacity() != 0 {
		t.Fatal(released, moved, err)
	}
	released[0] = 10
	if err = moved.Push(6); err != nil || !slices.Equal(moved.Values(), []int{6}) {
		t.Fatal(err, moved.Values())
	}
}

func TestMsatIntVectorOperationsAndGrowth(t *testing.T) {
	vector, err := NewMsatIntVector(0, 32)
	if err != nil {
		t.Fatal(err)
	}
	for value := 0; value < 17; value++ {
		if err = vector.Push(value); err != nil {
			t.Fatal(value, err)
		}
	}
	if vector.Capacity() != 32 || vector.Len() != 17 {
		t.Fatal(vector.Capacity(), vector.Len())
	}
	if err = vector.Push(17); err != nil {
		t.Fatal(err)
	}
	if err = vector.Grow(31); err != nil || vector.Capacity() != 32 {
		t.Fatal(err, vector.Capacity())
	}
	if err = vector.Grow(33); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = vector.Fill(5, -4); err != nil || !slices.Equal(vector.Values(), []int{-4, -4, -4, -4, -4}) {
		t.Fatal(err, vector.Values())
	}
	if got, err := vector.At(2); err != nil || got != -4 {
		t.Fatal(got, err)
	}
	if err = vector.Set(2, 11); err != nil {
		t.Fatal(err)
	}
	if got, err := vector.Last(); err != nil || got != -4 {
		t.Fatal(got, err)
	}
	if err = vector.Shrink(3); err != nil || !slices.Equal(vector.Values(), []int{-4, -4, 11}) {
		t.Fatal(err, vector.Values())
	}
	if got, err := vector.Pop(); err != nil || got != 11 || vector.Len() != 2 {
		t.Fatal(got, err, vector.Len())
	}
	if err = vector.Clear(); err != nil || vector.Len() != 0 || vector.Capacity() != 32 {
		t.Fatal(err, vector.Len(), vector.Capacity())
	}
}

func TestMsatIntVectorUniqueOrderedSortAndErrors(t *testing.T) {
	vector, _ := NewMsatIntVector(0, 64)
	for _, value := range []int{5, 1, 3, 1, -2} {
		found, err := vector.PushUnique(value)
		if err != nil || found != (value == 1 && vector.Len() > 2) {
			t.Fatal(value, found, err, vector.Values())
		}
	}
	if !slices.Equal(vector.Values(), []int{5, 1, 3, -2}) {
		t.Fatal(vector.Values())
	}
	if err := vector.Sort(false); err != nil {
		t.Fatal(err)
	}
	for _, value := range []int{4, 0, 5, -3, 4} {
		if _, err := vector.PushUniqueOrdered(value, true); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(vector.Values(), []int{-3, -2, 0, 1, 3, 4, 5}) {
		t.Fatal(vector.Values())
	}
	if err := vector.Sort(true); err != nil || !slices.Equal(vector.Values(), []int{5, 4, 3, 1, 0, -2, -3}) {
		t.Fatal(err, vector.Values())
	}
	for _, value := range []int{2, 8, -4, 2} {
		if _, err := vector.PushUniqueOrdered(value, false); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(vector.Values(), []int{8, 5, 4, 3, 2, 1, 0, -2, -3, -4}) {
		t.Fatal(vector.Values())
	}

	for _, operation := range []func() error{
		func() error { _, err := vector.At(-1); return err },
		func() error { _, err := vector.At(vector.Len()); return err },
		func() error { return vector.Set(vector.Len(), 0) },
		func() error { return vector.Shrink(vector.Len() + 1) },
		func() error { return vector.Fill(-1, 0) },
		func() error { return vector.Grow(-1) },
	} {
		before := vector.Values()
		if err := operation(); !errors.Is(err, ErrDimension) || !slices.Equal(before, vector.Values()) {
			t.Fatal(err, before, vector.Values())
		}
	}
	empty, _ := NewMsatIntVector(0, 8)
	if _, err := empty.Pop(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := empty.Last(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for index := 0; index < 8; index++ {
		if err := empty.Push(index); err != nil {
			t.Fatal(err)
		}
	}
	if err := empty.Push(8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	vector.Close()
	vector.Close()
	if vector.Values() != nil || vector.Len() != 0 || vector.Capacity() != 0 {
		t.Fatal("closed vector exposes data")
	}
	if err := vector.Push(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if vector, err := NewMsatIntVector(-1, 0); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewMsatIntVector(1, 8); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
	if vector, err := CopyMsatIntVector(make([]int, 9), 8); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
}

func FuzzMsatIntVectorOperations(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1}, {0, 1, 0, 2, 4, 0}, {3, 7, 5, 7, 6, 2, 7, 1}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		const maximum = 32
		vector, err := NewMsatIntVector(0, maximum)
		if err != nil {
			t.Fatal(err)
		}
		model := make([]int, 0, maximum)
		for index := 0; index+1 < len(operations); index += 2 {
			op, raw := operations[index]%8, operations[index+1]
			value := int(int8(raw))
			err = nil
			switch op {
			case 0:
				err = vector.Push(value)
				if len(model) == maximum {
					if !errors.Is(err, ErrLimit) {
						t.Fatal("push limit", err)
					}
					err = nil
				} else {
					if err != nil {
						t.Fatal(err)
					}
					model = append(model, value)
				}
			case 1:
				got, popErr := vector.Pop()
				if len(model) == 0 {
					if !errors.Is(popErr, ErrDimension) {
						t.Fatal(popErr)
					}
				} else {
					want := model[len(model)-1]
					model = model[:len(model)-1]
					if popErr != nil || got != want {
						t.Fatal(got, want, popErr)
					}
				}
			case 2:
				err = vector.Clear()
				model = model[:0]
			case 3:
				size := int(raw) % (maximum + 1)
				err = vector.Fill(size, value)
				model = slices.Grow(model[:0], size)[:size]
				for i := range model {
					model[i] = value
				}
			case 4:
				err = vector.Sort(raw&1 != 0)
				slices.Sort(model)
				if raw&1 != 0 {
					slices.Reverse(model)
				}
			case 5:
				beforeLen := len(model)
				found, uniqueErr := vector.PushUnique(value)
				wantFound := slices.Contains(model, value)
				if !wantFound && beforeLen < maximum {
					model = append(model, value)
				}
				if beforeLen == maximum && !wantFound && !errors.Is(uniqueErr, ErrLimit) {
					t.Fatal("unique limit", uniqueErr)
				} else if beforeLen < maximum && uniqueErr != nil || found != wantFound {
					t.Fatal(found, wantFound, uniqueErr)
				}
			case 6:
				slices.Sort(model)
				_ = vector.Sort(false)
				beforeLen := len(model)
				found, orderedErr := vector.PushUniqueOrdered(value, true)
				wantFound := slices.Contains(model, value)
				if !wantFound && beforeLen < maximum {
					position, _ := slices.BinarySearch(model, value)
					model = append(model, 0)
					copy(model[position+1:], model[position:])
					model[position] = value
				}
				if beforeLen == maximum && !wantFound && !errors.Is(orderedErr, ErrLimit) {
					t.Fatal("ordered limit", orderedErr)
				} else if beforeLen < maximum && orderedErr != nil || found != wantFound {
					t.Fatal(found, wantFound, orderedErr)
				}
			case 7:
				if len(model) != 0 {
					size := int(raw) % (len(model) + 1)
					err = vector.Shrink(size)
					model = model[:size]
				}
			}
			if err != nil || vector.Len() != len(model) || !slices.Equal(vector.Values(), model) {
				t.Fatal(op, raw, err, vector.Values(), model)
			}
		}
	})
}
