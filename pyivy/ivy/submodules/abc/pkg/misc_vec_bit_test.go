package abc

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

func checkBitVector(t *testing.T, v *BitVector, model []bool) {
	t.Helper()
	if v.Len() != len(model) || v.Cap()%32 != 0 || v.Cap() < v.Len() || v.StorageBytes() != v.Cap()/8 {
		t.Fatal("dimensions", v.Len(), len(model), v.Cap())
	}
	count := 0
	want := make([]uint32, (len(model)+31)/32)
	for i, b := range model {
		got, e := v.Entry(i)
		if e != nil || got != b {
			t.Fatal(i, got, b, e)
		}
		if b {
			count++
			want[i/32] |= uint32(1) << uint(i%32)
		}
	}
	if count != v.Count() || !slices.Equal(want, v.Words()) {
		t.Fatal("packed/count", want, v.Words(), count, v.Count())
	}
}
func TestBitVectorNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/vec_bit.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Stage, Size, Capacity, Count int
		Words                        []uint32
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 9 {
		t.Fatal(len(rows))
	}
	v, e := NewFilledBitVector(5, true, 1024)
	if e != nil {
		t.Fatal(e)
	}
	for _, r := range rows {
		switch r.Stage {
		case 1:
			v.Shrink(5)
		case 2:
			v.FillExtra(9, false)
		case 3:
			v.FillExtra(40, true)
		case 4:
			v.Grow(65)
		case 5:
			v.Fill(33, false)
			v.Write(63, true)
		case 6:
			v.Pop()
		case 7:
			v.Clear()
			for i := 0; i < 70; i++ {
				v.Push(i%3 == 0)
			}
		case 8:
			v.Reset()
		}
		if v.Len() != r.Size || v.Cap() != r.Capacity || v.Count() != r.Count || !slices.Equal(v.Words(), r.Words) {
			t.Fatal(r, v.Len(), v.Cap(), v.Count(), v.Words())
		}
	}
}
func TestBitVectorBoundariesAndOwnership(t *testing.T) {
	for _, start := range []int{0, 1, 31, 32, 33, 63, 64, 65, 95, 96} {
		for _, end := range []int{0, 1, 31, 32, 33, 63, 64, 65, 95, 96} {
			if end < start {
				continue
			}
			for _, value := range []bool{false, true} {
				v, _ := NewFilledBitVector(96, true, 256)
				v.Shrink(start)
				if e := v.FillExtra(end, value); e != nil {
					t.Fatal(e)
				}
				model := make([]bool, end)
				for i := range model {
					model[i] = i < start || value
				}
				checkBitVector(t, v, model)
				clone := v.Clone()
				if clone == v {
					t.Fatal("clone")
				}
				words := v.Words()
				if len(words) > 0 {
					words[0] ^= math.MaxUint32
					checkBitVector(t, v, model)
				}
				released := clone.ReleaseWords()
				if clone.Len() != 0 || clone.Cap() != 0 || !slices.Equal(released, v.Words()) {
					t.Fatal("release")
				}
				if len(released) > 0 {
					released[0] ^= math.MaxUint32
					checkBitVector(t, v, model)
				}
			}
		}
	}
	v, _ := NewBitVector(0, 64)
	v.Set(63, true)
	words := v.Words()
	size, capacity := v.Len(), v.Cap()
	for _, op := range []func() error{func() error { return v.Push(true) }, func() error { return v.Grow(65) }, func() error { return v.Fill(65, true) }, func() error { return v.FillExtra(65, false) }, func() error { return v.Set(math.MaxInt, true) }} {
		if e := op(); !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
		if !slices.Equal(words, v.Words()) || v.Len() != size || v.Cap() != capacity {
			t.Fatal("non-atomic limit")
		}
	}
	for _, op := range []func() error{func() error { return v.Write(-1, true) }, func() error { return v.Shrink(65) }, func() error { return v.Fill(-1, true) }, func() error { return v.Grow(-1) }, func() error { return v.FillExtra(-1, true) }} {
		if e := op(); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if _, e := NewBitVector(-1, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewBitVector(1, -1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFilledBitVector(1, true, 31); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	var zero BitVector
	if e := zero.Push(true); e != nil {
		t.Fatal(e)
	}
	if old, e := zero.Add(0); e != nil || !old {
		t.Fatal(old, e)
	}
	zero.Reset()
	if old, e := zero.Add(0); e != nil || old {
		t.Fatal(old, e)
	}
	zero.Clear()
	if _, e := zero.Pop(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	var nilVector *BitVector
	if nilVector.Len() != 0 || nilVector.Cap() != 0 || nilVector.Count() != 0 || nilVector.StorageBytes() != 0 || nilVector.Clone() != nil || nilVector.Words() != nil {
		t.Fatal("nil accessors")
	}
	if e := nilVector.Push(true); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzBitVector(f *testing.F) {
	f.Add([]byte{0, 31, 1, 32, 2, 63, 3, 64, 4, 65})
	f.Add([]byte{5, 0, 6, 127, 7, 128})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		v, _ := NewBitVector(0, 512)
		model := []bool{}
		for i := 0; i+1 < len(data); i += 2 {
			op, arg := data[i]%13, int(data[i+1])
			value := data[i]&128 != 0
			switch op {
			case 0:
				if len(model) < 512 {
					if e := v.Push(value); e != nil {
						t.Fatal(e)
					}
					model = append(model, value)
				}
			case 1:
				if len(model) > 0 {
					got, e := v.Pop()
					want := model[len(model)-1]
					model = model[:len(model)-1]
					if e != nil || got != want {
						t.Fatal(e)
					}
				}
			case 2:
				if e := v.Set(arg, value); e != nil {
					t.Fatal(e)
				}
				for len(model) <= arg {
					model = append(model, false)
				}
				model[arg] = value
			case 3:
				got, e := v.Get(arg)
				for len(model) <= arg {
					model = append(model, false)
				}
				if e != nil || got != model[arg] {
					t.Fatal(e)
				}
			case 4:
				if e := v.FillExtra(arg, value); e != nil {
					t.Fatal(e)
				}
				for len(model) < arg {
					model = append(model, value)
				}
			case 5:
				if e := v.Fill(arg, value); e != nil {
					t.Fatal(e)
				}
				model = make([]bool, (arg+31)/32*32)
				for j := range model {
					model[j] = value
				}
			case 6:
				if len(model) > 0 {
					n := arg % (len(model) + 1)
					v.Shrink(n)
					model = model[:n]
				}
			case 7:
				v.Clear()
				model = nil
			case 8:
				v.Reset()
				clear(model)
			case 9:
				if len(model) > 0 {
					j := arg % len(model)
					old, e := v.Add(j)
					if e != nil || old != model[j] {
						t.Fatal(e)
					}
					model[j] = true
				}
			case 10:
				clone := v.Clone()
				checkBitVector(t, clone, model)
				if len(model) > 0 {
					clone.Write(0, !model[0])
				}
			case 11:
				v.Grow(arg)
			case 12:
				words := v.ReleaseWords()
				if len(words) != (len(model)+31)/32 {
					t.Fatal("release size")
				}
				model = nil
			}
			checkBitVector(t, v, model)
		}
	})
}
