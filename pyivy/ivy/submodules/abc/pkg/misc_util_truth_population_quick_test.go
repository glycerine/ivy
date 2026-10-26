package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestTruthQuickByteCountTableOwnedAndFormatted(t *testing.T) {
	table := TruthQuickByteCountTable()
	original := table[1][0]
	table[1][0] = ^uint64(0)
	fresh := TruthQuickByteCountTable()
	if fresh[1][0] != original || truthQuickByteCounts[1][0] != original {
		t.Fatal("table snapshot aliases generated storage")
	}
	formatted := FormatTruthQuickByteCountTable()
	lines := strings.Split(strings.TrimSuffix(formatted, "\n"), "\n")
	if len(lines) != 256 || !strings.HasPrefix(lines[0], "00: 0000000000000000") ||
		!strings.HasPrefix(lines[255], "FF: 0808080808040404") || len(strings.Fields(lines[1])) != 33 {
		t.Fatal("formatted table shape")
	}
	if formatted != FormatTruthQuickByteCountTable() {
		t.Fatal("formatted table is nondeterministic")
	}
}

func TestTruthPopulationQuickNativeAndCorrections(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_population_quick.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Table [][]uint64
		Cases []struct {
			Variables, Mode          int
			NativeOnes, Ones         uint64
			Source, NativeZero, Zero []uint64
		}
	}
	if e = json.Unmarshal(data, &fixture); e != nil {
		t.Fatal(e)
	}
	if len(fixture.Table) != 256 || len(fixture.Cases) != 104 {
		t.Fatal("fixture dimensions")
	}
	for value, row := range fixture.Table {
		if len(row) != 32 {
			t.Fatal("table row width")
		}
		for index, native := range row {
			if truthQuickByteCounts[value][index] != native {
				t.Fatal("lookup", value, index)
			}
		}
	}
	totals, cofactors := 0, 0
	for _, row := range fixture.Cases {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		actual, e := source.CofactorPopulationsQuick()
		if e != nil || actual.Ones != row.Ones || !slices.Equal(actual.Zero, row.Zero) {
			t.Fatal(row, actual, e)
		}
		for i := range actual.One {
			if actual.One[i] != row.Ones-row.Zero[i] {
				t.Fatal("one cofactor")
			}
		}
		if row.NativeOnes != row.Ones {
			totals++
		}
		if !slices.Equal(row.NativeZero, row.Zero) {
			cofactors++
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
	if totals != 9 || cofactors != 21 {
		t.Fatal("native bug evidence changed", totals, cofactors)
	}
}
func TestTruthPopulationQuickOverflowAndRecursion(t *testing.T) {
	if r, e := (*TruthTable)(nil).CofactorPopulationsQuick(); r.Zero != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	full, _ := NewConstantTruthTable(8, true, 0)
	r, e := full.CofactorPopulationsQuick()
	if e != nil || r.Ones != 256 {
		t.Fatal("packed byte overflow", r, e)
	}
	for _, value := range r.Zero {
		if value != 128 {
			t.Fatal(r)
		}
	}
	source, _ := NewConstantTruthTable(20, true, 0)
	source.Set(0, false)
	before := source.Words()
	r, e = source.CofactorPopulationsQuick()
	if e != nil || r.Ones != (1<<20)-1 {
		t.Fatal("wide total", r, e)
	}
	for v := range r.Zero {
		if r.Zero[v] != (1<<19)-1 || r.One[v] != 1<<19 {
			t.Fatal("recursive accumulation", v, r)
		}
	}
	r.Zero[0] = 0
	r.One[0] = 0
	fresh, e := source.CofactorPopulationsQuick()
	if e != nil || fresh.Zero[0] != (1<<19)-1 || fresh.One[0] != 1<<19 || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, value := range []bool{false, true} {
		constant, _ := NewConstantTruthTable(0, value, 0)
		r, e := constant.CofactorPopulationsQuick()
		expected := uint64(0)
		if value {
			expected = 1
		}
		if e != nil || r.Ones != expected || len(r.Zero) != 0 || len(r.One) != 0 {
			t.Fatal("zero convention", r, e)
		}
	}
	small, _ := NewVariableTruthTable(3, 0, 0)
	quick, _ := small.CofactorPopulationsQuick()
	fast, _ := small.CofactorPopulationsFast()
	if quick.Ones != 32 || fast.Ones != 4 {
		t.Fatal("count conventions", quick, fast)
	}
}
func FuzzTruthPopulationQuick(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(8), []byte{255})
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
		actual, e := source.CofactorPopulationsQuick()
		if e != nil {
			t.Fatal(e)
		}
		ones := uint64(0)
		zero, one := make([]uint64, variables), make([]uint64, variables)
		limit := count * 64
		if variables == 0 {
			limit = 1
		}
		for bit := 0; bit < limit; bit++ {
			x := bit % n
			if data[(x/8)%len(data)]&(1<<uint(x%8)) == 0 {
				continue
			}
			ones++
			for v := range zero {
				if bit&(1<<uint(v)) == 0 {
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
		if variables > 0 {
			if actual.Ones != uint64(packed.Ones) {
				t.Fatal("packed total")
			}
			for v := range zero {
				if zero[v] != uint64(packed.Zero[v]) || one[v] != uint64(packed.One[v]) {
					t.Fatal("packed cofactor")
				}
			}
		}
		fast, e := source.CofactorPopulationsFast()
		if e != nil {
			t.Fatal(e)
		}
		factor := uint64(max(1, 64/n))
		if variables == 0 {
			factor = 1
		}
		if actual.Ones != fast.Ones*factor {
			t.Fatal("logical scaling")
		}
		if !slices.Equal(source.Words(), words) {
			t.Fatal("source changed")
		}
	})
}
