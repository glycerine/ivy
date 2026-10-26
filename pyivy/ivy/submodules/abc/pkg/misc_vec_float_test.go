package abc

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

func floatVectorBits(values []float32) []uint32 {
	r := make([]uint32, len(values))
	for i, v := range values {
		r[i] = math.Float32bits(v)
	}
	return r
}
func TestFloatVectorNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/vec_float.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Stage, Size, Capacity int
		Bits                  []uint32
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 6 {
		t.Fatal(len(rows))
	}
	v, e := NewFullFloatVector(3, 128)
	if e != nil {
		t.Fatal(e)
	}
	for _, r := range rows {
		switch r.Stage {
		case 1:
			v.Fill(3, float32(math.Copysign(0, -1)))
			v.Write(1, 1.25)
			v.Add(1, .5)
			v.Update(2, -1)
			v.Update(2, 2)
			v.Pop()
		case 2:
			v.FillExtra(5, 3.5)
			v.Grow(17)
		case 3:
			v.PushOrdered(-1)
			found, e := v.PushUnique(3.5)
			if e != nil || !found {
				t.Fatal(e)
			}
			v.Remove(3.5)
			v.Sort(true)
		case 4:
			v.Clear()
			for i := 0; i < 20; i++ {
				v.Push(float32(i%7 - 3))
			}
		case 5:
			v.Sort(false)
		}
		if v.Len() != r.Size || v.Cap() != r.Capacity || !slices.Equal(floatVectorBits(v.Values()), r.Bits) {
			t.Fatal(r, v.Len(), v.Cap(), floatVectorBits(v.Values()))
		}
	}
}
func TestFloatVectorIEEEAndOwnership(t *testing.T) {
	nan := math.Float32frombits(0xffffffff)
	v, _ := FloatVectorFrom([]float32{nan, 0, float32(math.Copysign(0, -1)), float32(math.Inf(1)), float32(math.Inf(-1))}, 64)
	if v.Find(nan) != -1 || v.Equal(v) || v.Find(0) != 1 {
		t.Fatal("IEEE equality")
	}
	lo, ok := v.Min()
	if !ok || math.Float32bits(lo) != 0xffffffff {
		t.Fatal(lo, ok)
	}
	hi, _ := v.Max()
	if !math.IsNaN(float64(hi)) {
		t.Fatal(hi)
	}
	before := floatVectorBits(v.Values())
	if e := v.Sort(false); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if e := v.PushOrdered(1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if !slices.Equal(before, floatVectorBits(v.Values())) {
		t.Fatal("NaN failure mutation")
	}
	exists, e := v.PushUnique(nan)
	if e != nil || exists || v.Len() != 6 {
		t.Fatal(e)
	}
	v.Update(0, 100)
	if math.Float32bits(v.Values()[0]) != 0xffffffff {
		t.Fatal("NaN update")
	}
	source := []float32{2, 1}
	owned, _ := FloatVectorFrom(source, 64)
	source[0] = 99
	clone := owned.Clone()
	clone.Write(0, 100)
	snapshot := owned.Values()
	snapshot[0] = 100
	if owned.Values()[0] != 2 {
		t.Fatal("alias")
	}
	moved := owned.Move()
	if owned.Len() != 0 || owned.Cap() != 0 || moved.Values()[0] != 2 {
		t.Fatal("move")
	}
	released := moved.ReleaseValues()
	if moved.Len() != 0 || moved.Cap() != 0 || released[0] != 2 {
		t.Fatal("release")
	}
	sorted, _ := FloatVectorFrom([]float32{0, float32(math.Copysign(0, -1)), float32(math.Inf(1)), float32(math.Inf(-1))}, 64)
	sorted.Sort(false)
	if math.Float32bits(sorted.Values()[1]) != 0 || math.Float32bits(sorted.Values()[2]) != 0x80000000 {
		t.Fatal("stable zeros")
	}
	empty, _ := NewFloatVector(0, 0)
	if value, ok := empty.Min(); value != 0 || ok {
		t.Fatal(value, ok)
	}
	if _, ok := empty.Max(); ok {
		t.Fatal("empty max")
	}
	if _, e := empty.Pop(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	formatted, _ := FloatVectorFrom([]float32{1.25, -2, float32(math.Copysign(0, -1))}, 8)
	if got := formatted.Format(); got != "Vector has 3 entries: { 1.250000 -2.000000 -0.000000 }\n" {
		t.Fatal(got)
	}
	if got := (*FloatVector)(nil).Format(); got != "Vector has 0 entries: { }\n" {
		t.Fatal(got)
	}
}
func TestFloatVectorLimits(t *testing.T) {
	v, _ := NewFloatVectorExact(2, 2)
	v.Fill(2, 3)
	before := floatVectorBits(v.Values())
	for _, op := range []func() error{func() error { return v.Push(1) }, func() error { return v.Fill(3, 1) }, func() error { return v.FillExtra(3, 1) }, func() error { return v.Grow(3) }, func() error { return v.PushOrdered(1) }} {
		if e := op(); !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
		if !slices.Equal(before, floatVectorBits(v.Values())) || v.Len() != 2 || v.Cap() != 2 {
			t.Fatal("failure mutation")
		}
	}
	for _, op := range []func() error{func() error { return v.Fill(-1, 1) }, func() error { return v.Grow(-1) }, func() error { return v.Shrink(3) }, func() error { return v.Write(-1, 1) }, func() error { return v.Add(2, 1) }, func() error { return v.Update(2, 1) }} {
		if e := op(); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if _, e := NewFloatVector(-1, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFloatVectorExact(0, -1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFloatVector(1, 8); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	var zero FloatVector
	zero.Push(1)
	if zero.Len() != 1 || zero.Cap() != 16 {
		t.Fatal("zero value")
	}
	zero.Erase()
	if zero.Len() != 0 || zero.Cap() != 0 {
		t.Fatal("erase")
	}
	var nilVector *FloatVector
	if nilVector.Len() != 0 || nilVector.Clone() != nil || nilVector.Move() != nil || nilVector.Values() != nil {
		t.Fatal("nil")
	}
	if e := nilVector.Push(1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzFloatVector(f *testing.F) {
	f.Add([]byte{0, 1, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{8, 3, 9, 2, 10, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		v, _ := NewFloatVectorExact(0, 256)
		model := []float32{}
		for i := 0; i+1 < len(data); i += 2 {
			op, arg := data[i]%13, int(data[i+1])
			value := float32(int8(data[i+1])) / 4
			switch op {
			case 0:
				if len(model) < 256 {
					v.Push(value)
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
				v.Fill(arg, value)
				model = make([]float32, arg)
				for j := range model {
					model[j] = value
				}
			case 3:
				v.FillExtra(arg, value)
				for len(model) < arg {
					model = append(model, value)
				}
			case 4:
				if len(model) > 0 {
					j := arg % len(model)
					v.Write(j, value)
					model[j] = value
				}
			case 5:
				if len(model) > 0 {
					j := arg % len(model)
					v.Add(j, value)
					model[j] = float32(model[j] + value)
				}
			case 6:
				if len(model) > 0 {
					j := arg % len(model)
					v.Update(j, value)
					if model[j] < value {
						model[j] = value
					}
				}
			case 7:
				if len(model) > 0 {
					n := arg % (len(model) + 1)
					v.Shrink(n)
					model = model[:n]
				}
			case 8:
				v.Clear()
				model = nil
			case 9:
				reverse := arg&1 != 0
				v.Sort(reverse)
				for a := 0; a < len(model); a++ {
					for b := a + 1; b < len(model); b++ {
						if (!reverse && model[b] < model[a]) || (reverse && model[b] > model[a]) {
							model[a], model[b] = model[b], model[a]
						}
					}
				}
			case 10:
				index := slices.Index(model, value)
				removed := v.Remove(value)
				if removed != (index >= 0) {
					t.Fatal("remove")
				}
				if index >= 0 {
					model = append(model[:index], model[index+1:]...)
				}
			case 11:
				if len(model) < 256 {
					index := slices.Index(model, value)
					exists, e := v.PushUnique(value)
					if e != nil || exists != (index >= 0) {
						t.Fatal(e)
					}
					if index < 0 {
						model = append(model, value)
					}
				}
			case 12:
				clone := v.Clone()
				if !slices.Equal(floatVectorBits(clone.Values()), floatVectorBits(model)) {
					t.Fatal("clone")
				}
				clone.Erase()
			}
			if v.Len() != len(model) || v.Cap() < v.Len() || v.StorageBytes() != v.Cap()*4 || !slices.Equal(floatVectorBits(v.Values()), floatVectorBits(model)) {
				t.Fatal(op, arg, v.Values(), model)
			}
		}
	})
}

func FuzzFloatVectorIEEE(f *testing.F) {
	f.Add(uint32(0xffffffff), uint32(0x80000000))
	f.Add(uint32(0x7f7fffff), uint32(0x00000001))
	f.Add(uint32(0x3f800000), uint32(0x33800000))
	f.Fuzz(func(t *testing.T, first, second uint32) {
		a, b := math.Float32frombits(first), math.Float32frombits(second)
		v, _ := FloatVectorFrom([]float32{a, b}, 32)
		if got := floatVectorBits(v.Clone().Values()); !slices.Equal(got, []uint32{first, second}) {
			t.Fatal("payload copy")
		}
		v.Add(0, b)
		got, _ := v.Entry(0)
		want := float32(float64(a) + float64(b))
		if math.IsNaN(float64(want)) {
			if !math.IsNaN(float64(got)) {
				t.Fatal("NaN arithmetic")
			}
		} else if math.Float32bits(got) != math.Float32bits(want) {
			t.Fatal("rounding", first, second, got, want)
		}
		v.Write(0, a)
		v.Update(0, b)
		updated, _ := v.Entry(0)
		expected := a
		if a < b {
			expected = b
		}
		if math.Float32bits(updated) != math.Float32bits(expected) {
			t.Fatal("update")
		}
		v.Write(0, a)
		low, _ := v.Min()
		high, _ := v.Max()
		wantLow, wantHigh := a, a
		if a > b {
			wantLow = b
		}
		if a < b {
			wantHigh = b
		}
		if math.Float32bits(low) != math.Float32bits(wantLow) || math.Float32bits(high) != math.Float32bits(wantHigh) {
			t.Fatal("extrema")
		}
		before := floatVectorBits(v.Values())
		e := v.Sort(false)
		if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
			if !errors.Is(e, ErrDimension) || !slices.Equal(before, floatVectorBits(v.Values())) {
				t.Fatal("NaN sorting")
			}
		} else {
			if e != nil {
				t.Fatal(e)
			}
			wantBits := []uint32{first, second}
			if b < a {
				wantBits = []uint32{second, first}
			}
			if !slices.Equal(wantBits, floatVectorBits(v.Values())) {
				t.Fatal("order")
			}
		}
	})
}
