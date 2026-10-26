package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthPopulationFastNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_population_fast.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode int
		Ones            uint64
		Source, Zero    []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 80 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.CofactorPopulationsFast()
		if e != nil || r.Ones != row.Ones || !slices.Equal(r.Zero, row.Zero) {
			t.Fatal(row, r, e)
		}
		for v := range r.One {
			if r.One[v] != row.Ones-row.Zero[v] {
				t.Fatal("one cofactor", v, r)
			}
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthPopulationFastSmallAndConstantBlocks(t *testing.T) {
	if r, e := (*TruthTable)(nil).CofactorPopulationsFast(); r.Zero != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for variables := 0; variables <= 6; variables++ {
		for _, value := range []bool{false, true} {
			source, _ := NewConstantTruthTable(variables, value, 0)
			r, e := source.CofactorPopulationsFast()
			expected := uint64(0)
			if value {
				expected = uint64(1) << uint(variables)
			}
			if e != nil || r.Ones != expected || len(r.Zero) != variables || len(r.One) != variables {
				t.Fatal(variables, value, r, e)
			}
			for v := range r.Zero {
				if r.Zero[v] != expected/2 || r.One[v] != expected/2 {
					t.Fatal("constant halves", r)
				}
			}
		}
	}
	source, _ := NewConstantTruthTable(20, true, 0)
	source.Set(0, false)
	before := source.Words()
	r, e := source.CofactorPopulationsFast()
	if e != nil || r.Ones != (1<<20)-1 {
		t.Fatal("wide mixed block", r, e)
	}
	for v := range r.Zero {
		if r.Zero[v] != (1<<19)-1 || r.One[v] != 1<<19 {
			t.Fatal("wide cofactor", v, r)
		}
	}
	r.Zero[0] = 0
	r.One[0] = 0
	fresh, e := source.CofactorPopulationsFast()
	if e != nil || fresh.Zero[0] != (1<<19)-1 || fresh.One[0] != 1<<19 || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	sparse, _ := NewConstantTruthTable(20, false, 0)
	sparse.Set(1<<19, true)
	sparse.Set(0, true)
	r, e = sparse.CofactorPopulationsFast()
	if e != nil || r.Ones != 2 {
		t.Fatal("sparse wide", r, e)
	}
	for v := range r.Zero {
		expected := uint64(2)
		if v == 19 {
			expected = 1
		}
		if r.Zero[v] != expected {
			t.Fatal("sparse cofactor", v, r)
		}
	}
	// Fast follows logical counts, unlike the existing packed API below six.
	one, _ := NewVariableTruthTable(3, 0, 0)
	logical, _ := one.CofactorPopulationsFast()
	packed, _ := one.CofactorPopulations()
	if logical.Ones != 4 || packed.Ones != 32 || logical.Zero[1] != 2 || packed.Zero[1] != 16 {
		t.Fatal("small-count convention", logical, packed)
	}
}
func FuzzTruthPopulationFast(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Add(uint8(12), []byte{0, 255, 0, 255})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 256 {
			data = data[:256]
		}
		variables := int(width % 13)
		n := 1 << uint(variables)
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for bit := 0; bit < count*64; bit++ {
			x := bit % n
			if data[(x/8)%len(data)]&(1<<uint(x%8)) != 0 {
				words[bit/64] |= uint64(1) << uint(bit%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		actual, e := source.CofactorPopulationsFast()
		if e != nil {
			t.Fatal(e)
		}
		ones := uint64(0)
		zero, one := make([]uint64, variables), make([]uint64, variables)
		for x := 0; x < n; x++ {
			if data[(x/8)%len(data)]&(1<<uint(x%8)) == 0 {
				continue
			}
			ones++
			for v := 0; v < variables; v++ {
				if x&(1<<uint(v)) == 0 {
					zero[v]++
				} else {
					one[v]++
				}
			}
		}
		if actual.Ones != ones || !slices.Equal(actual.Zero, zero) || !slices.Equal(actual.One, one) {
			t.Fatal(variables, actual, ones, zero, one)
		}
		packed, e := source.CofactorPopulations()
		if e != nil {
			t.Fatal(e)
		}
		factor := uint64(max(1, 64/n))
		if uint64(packed.Ones) != actual.Ones*factor {
			t.Fatal("packed total")
		}
		for v := range zero {
			if uint64(packed.Zero[v]) != zero[v]*factor || uint64(packed.One[v]) != one[v]*factor {
				t.Fatal("packed cofactor")
			}
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
