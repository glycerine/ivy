package abc

import (
	"math/bits"
	"slices"
	"testing"
)

func dauDivisorModel(t *TruthTable) []DauDivisorPair {
	rows := make([]DauDivisorPair, 0, t.variables*(t.variables-1)/2)
	for first := 0; first < t.variables; first++ {
		for second := first + 1; second < t.variables; second++ {
			row := DauDivisorPair{First: first, Second: second}
			others := make([]int, 0, t.variables-2)
			for v := 0; v < t.variables; v++ {
				if v != first && v != second {
					others = append(others, v)
				}
			}
			for rest := 0; rest < 1<<uint(len(others)); rest++ {
				digit := 0
				for b := 0; b < 4; b++ {
					assignment := uint64(0)
					for k, v := range others {
						assignment |= uint64(rest>>uint(k)&1) << uint(v)
					}
					assignment |= uint64(b&1) << uint(first)
					assignment |= uint64(b>>1) << uint(second)
					value, _ := t.Get(assignment)
					if value {
						digit |= 1 << uint(b)
					}
				}
				for category, masks := range [5][2]int{{1, 14}, {2, 13}, {4, 11}, {8, 7}, {6, 9}} {
					if digit == masks[0] || digit == masks[1] {
						row.Counts[category]++
					}
				}
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func TestDauDSDDivisors(t *testing.T) {
	DauDSDTest000()
	for digit, want := range map[uint64][5]int{0: {}, 1: {1, 0, 0, 0, 0}, 2: {0, 1, 0, 0, 0}, 4: {0, 0, 1, 0, 0}, 8: {0, 0, 0, 1, 0}, 6: {0, 0, 0, 0, 1}, 9: {0, 0, 0, 0, 1}, 14: {1, 0, 0, 0, 0}} {
		table, e := TruthTableFromWords(2, []uint64{digit * 0x1111111111111111}, 1)
		if e != nil {
			t.Fatal(e)
		}
		rows, e := DauDSDDivisors(table, DauDivisorOptions{})
		if e != nil || len(rows) != 1 || rows[0].Counts != want {
			t.Fatal(digit, rows, e)
		}
	}
	table, _ := NewConstantTruthTable(1, false, 1)
	rows, e := DauDSDDivisors(table, DauDivisorOptions{})
	if e != nil || len(rows) != 0 {
		t.Fatal(rows, e)
	}
	if text := FormatDauDSDDivisors([]DauDivisorPair{{First: 0, Second: 2, Counts: [5]int{1, 2, 3, 4, 5}}}); text != "      !a *!b !a * b  a *!b  a * b  a + b\na c        1      2      3      4      5\n" {
		t.Fatal(text)
	}
}

func TestDauDSDDivisorLimits(t *testing.T) {
	table, _ := NewVariableTruthTable(4, 0, 1)
	for _, o := range []DauDivisorOptions{{MaxPairs: -1}, {MaxDigits: -1}, {MaxPairs: 5}, {MaxDigits: 23}, {Stop: func() bool { return true }}} {
		if rows, e := DauDSDDivisors(table, o); rows != nil || e == nil {
			t.Fatal(rows, e)
		}
	}
	if rows, e := DauDSDDivisors(nil, DauDivisorOptions{}); rows != nil || e == nil {
		t.Fatal(rows, e)
	}
}

func FuzzDauDSDDivisors(f *testing.F) {
	f.Add(uint8(5), uint64(0x9ef7a8d9c7193a0f))
	f.Fuzz(func(t *testing.T, rawVariables uint8, seed uint64) {
		variables := int(rawVariables % 9)
		table, e := NewConstantTruthTable(variables, false, 4)
		if e != nil {
			t.Fatal(e)
		}
		assignments := 1 << uint(variables)
		for a := 0; a < assignments; a++ {
			value := bits.RotateLeft64(seed, a&63)>>uint(a%17)&1 != 0
			if e = table.Set(uint64(a), value); e != nil {
				t.Fatal(e)
			}
		}
		got, e := DauDSDDivisors(table, DauDivisorOptions{})
		want := dauDivisorModel(table)
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}
