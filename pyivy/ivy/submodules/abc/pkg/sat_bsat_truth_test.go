package abc

// Go-only model, lifecycle and operation tests for satTruth.c/satTruth.h.

import (
	"errors"
	"slices"
	"testing"
)

func TestSatTruthWordOperations(t *testing.T) {
	left := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	right := []uint64{0x1111111111111111, 0x2222222222222222}
	if SatTruthEqual(left, right) || !SatTruthEqual(left, slices.Clone(left)) || SatTruthEqual(left, left[:1]) {
		t.Fatal("equality")
	}
	zero := make([]uint64, 2)
	if !SatTruthEqualZero(zero) || SatTruthEqualOne(zero) {
		t.Fatal("zero")
	}
	if SatTruthFill(zero)[0] != ^uint64(0) || !SatTruthEqualOne(zero) {
		t.Fatal("fill")
	}
	if SatTruthNot(zero)[0] != 0 || !SatTruthEqualZero(zero) {
		t.Fatal("not")
	}
	output := make([]uint64, 2)
	if got, err := SatTruthCopy(output, left, false); err != nil || !slices.Equal(got, left) {
		t.Fatal(got, err)
	}
	if got, err := SatTruthCopy(output, left, true); err != nil || got[0] != ^left[0] || got[1] != ^left[1] {
		t.Fatal(got, err)
	}
	output = slices.Clone(left)
	if got, err := SatTruthAnd(output, right, false); err != nil || got[0] != left[0]&right[0] || got[1] != left[1]&right[1] {
		t.Fatal(got, err)
	}
	output = slices.Clone(left)
	if got, err := SatTruthAnd(output, right, true); err != nil || got[0] != left[0]&^right[0] || got[1] != left[1]&^right[1] {
		t.Fatal(got, err)
	}
	output = slices.Clone(left)
	if got, err := SatTruthOr(output, right, false); err != nil || got[0] != left[0]|right[0] || got[1] != left[1]|right[1] {
		t.Fatal(got, err)
	}
	output = slices.Clone(left)
	if got, err := SatTruthOr(output, right, true); err != nil || got[0] != left[0]|^right[0] || got[1] != left[1]|^right[1] {
		t.Fatal(got, err)
	}
	if result, err := SatTruthCopy(make([]uint64, 1), right, false); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if SatTruthClear(output)[0] != 0 || SatTruthClear(output)[1] != 0 {
		t.Fatal(output)
	}
}

func TestSatTruthStoreVariablesAndInterning(t *testing.T) {
	for variables := 1; variables <= 16; variables++ {
		store, err := NewSatTruthStore(variables, SatTruthStoreOptions{})
		if err != nil {
			t.Fatal(variables, err)
		}
		stats, err := store.Stats()
		wantWords := 1
		if variables >= 6 {
			wantWords = 1 << uint(variables-6)
		}
		if err != nil || stats.Variables != variables || stats.Words != wantWords || stats.Entries != variables || stats.Lookups != variables {
			t.Fatal(variables, stats, err)
		}
		for variable := 0; variable < variables; variable++ {
			words, variableErr := store.Variable(variable)
			if variableErr != nil || len(words) != wantWords {
				t.Fatal(variables, variable, variableErr)
			}
			assignments := 1 << uint(variables)
			for _, assignment := range []int{0, 1, assignments / 3, assignments / 2, assignments - 1} {
				got := words[assignment/64]>>uint(assignment%64)&1 != 0
				want := assignment&(1<<uint(variable)) != 0
				if got != want {
					t.Fatal(variables, variable, assignment, got, want)
				}
			}
			words[0] = 0 // returned storage is owned
			again, _ := store.Variable(variable)
			if variable == 0 && again[0] == 0 {
				t.Fatal("variable storage aliased")
			}
		}
		if err = store.Close(); err != nil {
			t.Fatal(err)
		}
	}

	store, err := NewSatTruthStore(7, SatTruthStoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	zero := make([]uint64, 2)
	ones := []uint64{^uint64(0), ^uint64(0)}
	if handle, err := store.Insert(zero); err != nil || handle != 0 {
		t.Fatal(handle, err)
	}
	if handle, err := store.Insert(ones); err != nil || handle != 1 {
		t.Fatal(handle, err)
	}
	truth := []uint64{0x123456789abcdef0, 0x0fedcba987654321}
	before := slices.Clone(truth)
	handle, err := store.Insert(truth)
	if err != nil || handle&1 != 0 || !slices.Equal(truth, before) {
		t.Fatal(handle, truth, before, err)
	}
	again, err := store.Insert(truth)
	if err != nil || again != handle {
		t.Fatal(again, handle, err)
	}
	complement := []uint64{^truth[0], ^truth[1]}
	signed, err := store.Insert(complement)
	if err != nil || signed != handle^1 || !slices.Equal(complement, []uint64{^truth[0], ^truth[1]}) {
		t.Fatal(signed, handle, err)
	}
	stored, err := store.Function(handle)
	if err != nil || !slices.Equal(stored, truth) {
		t.Fatal(stored, truth, err)
	}
	stored[0] = 0
	if next, _ := store.Function(handle); next[0] != truth[0] {
		t.Fatal("function storage aliased")
	}
}

func TestSatTruthStoreErrorsLimitsAndClose(t *testing.T) {
	for _, test := range []struct {
		variables int
		options   SatTruthStoreOptions
		want      error
	}{
		{0, SatTruthStoreOptions{}, ErrDimension},
		{17, SatTruthStoreOptions{}, ErrDimension},
		{2, SatTruthStoreOptions{MaxVariables: -1}, ErrDimension},
		{8, SatTruthStoreOptions{MaxWords: 1}, ErrLimit},
		{3, SatTruthStoreOptions{MaxEntries: 2}, ErrLimit},
	} {
		store, err := NewSatTruthStore(test.variables, test.options)
		if store != nil || !errors.Is(err, test.want) {
			t.Fatal(test, store, err)
		}
	}
	store, err := NewSatTruthStore(3, SatTruthStoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Insert(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = store.Variable(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = store.Function(1); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err = store.Function(1 << 20); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = store.Stats(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = (*SatTruthStore)(nil).Close(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzSatTruthStore(f *testing.F) {
	f.Add(uint8(7), []byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Add(uint8(1), []byte{255, 255, 255, 255, 255, 255, 255, 255})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		variables := 1 + int(width%10)
		store, err := NewSatTruthStore(variables, SatTruthStoreOptions{})
		if err != nil {
			t.Fatal(err)
		}
		stats, _ := store.Stats()
		words := make([]uint64, stats.Words)
		for index := range words {
			for offset := 0; offset < 8; offset++ {
				if len(data) != 0 {
					words[index] |= uint64(data[(8*index+offset)%len(data)]) << uint(8*offset)
				}
			}
		}
		before := slices.Clone(words)
		handle, insertErr := store.Insert(words)
		if insertErr != nil || !slices.Equal(words, before) {
			t.Fatal(handle, insertErr)
		}
		complement := slices.Clone(words)
		for index := range complement {
			complement[index] = ^complement[index]
		}
		complementHandle, complementErr := store.Insert(complement)
		if complementErr != nil || complementHandle != handle^1 {
			t.Fatal(handle, complementHandle, complementErr)
		}
		regular := handle &^ 1
		stored, functionErr := store.Function(regular)
		if functionErr != nil {
			t.Fatal(functionErr)
		}
		want := before
		if handle&1 != 0 {
			want = complement
		}
		if !slices.Equal(stored, want) {
			t.Fatal(handle, stored, want)
		}
	})
}
