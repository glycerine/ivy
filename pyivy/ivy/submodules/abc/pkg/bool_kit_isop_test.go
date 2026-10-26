package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func kitSmallTruth(t testing.TB, variables int, value uint64) *TruthTable {
	t.Helper()
	bits := 1 << uint(variables)
	if bits < 64 {
		value &= (uint64(1) << uint(bits)) - 1
		for shift := bits; shift < 64; shift *= 2 {
			value |= value << uint(shift)
		}
	}
	table, err := TruthTableFromWords(variables, []uint64{value}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func kitCoverValue(cubes []uint32, a int, n int) bool {
	for _, cube := range cubes {
		match := true
		for b := 0; b < n; b++ {
			pair := cube >> uint(2*b) & 3
			v := a>>uint(b)&1 != 0
			if pair == 3 || pair == 1 && v || pair == 2 && !v {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
func TestTruthISOPNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/kit_isop.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, TryBoth, Complemented, Literals int
		Truth                                            []uint64
		Cubes                                            []uint32
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 196 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		table, e := TruthTableFromWords(row.Variables, row.Truth, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := table.ISOP(TruthISOPOptions{TryBoth: row.TryBoth != 0})
		if e != nil || r.Complemented != (row.Complemented != 0) || r.Literals != row.Literals || !slices.Equal(r.Cubes, row.Cubes) {
			t.Fatal(i, row.Variables, row.Mode, r, e)
		}
		for a := 0; a < 1<<uint(row.Variables); a++ {
			word := row.Truth[a/64]
			want := word>>uint(a%64)&1 != 0
			if kitCoverValue(r.Cubes, a, row.Variables) != (want != r.Complemented) {
				t.Fatal(i, a, r)
			}
		}
		if !slices.Equal(table.words, row.Truth) {
			t.Fatal("input changed")
		}
	}
}
func TestTruthISOPLimits(t *testing.T) {
	table, _ := TruthTableFromWords(6, []uint64{0x6996966996696996}, 0)
	for _, o := range []TruthISOPOptions{{MaxElements: 1}, {MaxVisits: 1}, {MaxElements: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		r, e := table.ISOP(o)
		if e == nil || r.Cubes != nil {
			t.Fatal(r, e)
		}
	}
	big, _ := NewConstantTruthTable(17, false, 0)
	if _, e := big.ISOP(TruthISOPOptions{}); e == nil {
		t.Fatal("wide")
	}
	for _, value := range []bool{false, true} {
		c, _ := NewConstantTruthTable(0, value, 0)
		r, e := c.ISOP(TruthISOPOptions{TryBoth: true})
		if e != nil || r.Complemented || len(r.Cubes) != kitISOPBoolCount(value) {
			t.Fatal(r, e)
		}
	}
}

func TestTruthISOPIntervalAndFormat(t *testing.T) {
	on := kitSmallTruth(t, 3, 1<<3|1<<7)
	allowed := kitSmallTruth(t, 3, 1<<1|1<<3|1<<5|1<<7)
	beforeOn, beforeAllowed := on.Words(), allowed.Words()
	for _, both := range []bool{false, true} {
		r, err := on.ISOPInterval(allowed, TruthISOPOptions{TryBoth: both})
		if err != nil || r.Table == nil || r.Table.variables != 3 {
			t.Fatal(both, r, err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			cover := kitCoverValue(r.Cubes, assignment, 3)
			onValue, _ := on.Get(uint64(assignment))
			allowedValue, _ := allowed.Get(uint64(assignment))
			tableValue, tableErr := r.Table.Get(uint64(assignment))
			if tableValue != cover {
				t.Fatal("returned table", both, assignment, cover, tableValue, tableErr, r.Table.words, r)
			}
			if !r.Complemented && (onValue && !cover || !allowedValue && cover) {
				t.Fatal("direct interval", assignment, r)
			}
			if r.Complemented && (!allowedValue && !cover || onValue && cover) {
				t.Fatal("complement interval", assignment, r)
			}
		}
		if _, err = r.Format(3); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(on.Words(), beforeOn) || !slices.Equal(allowed.Words(), beforeAllowed) {
		t.Fatal("interval input changed")
	}
	r, _ := on.ISOPInterval(allowed, TruthISOPOptions{})
	r.Table.words[0] = 0
	again, err := on.ISOPInterval(allowed, TruthISOPOptions{})
	if err != nil || again.Table.words[0] == 0 {
		t.Fatal("returned table aliases", again, err)
	}
	for _, test := range []struct {
		result TruthISOP
		vars   int
		want   string
		bad    bool
	}{
		{TruthISOP{Cubes: []uint32{1 | 2<<2}}, 2, "01 1\n", false},
		{TruthISOP{Cubes: []uint32{1 | 2<<2}, Complemented: true}, 2, "01 0\n", false},
		{TruthISOP{}, 2, "Constant 0\n", false},
		{TruthISOP{Cubes: []uint32{0}}, 2, "Constant 1\n", false},
		{TruthISOP{Cubes: []uint32{3}}, 1, "", true},
	} {
		got, err := test.result.Format(test.vars)
		if got != test.want || (err != nil) != test.bad {
			t.Fatal(test, got, err)
		}
	}
	if _, err := (TruthISOP{}).Format(17); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	other := kitSmallTruth(t, 2, 1)
	if r, err := on.ISOPInterval(other, TruthISOPOptions{}); r.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(r, err)
	}
	badOn := kitSmallTruth(t, 3, 1)
	badAllowed := kitSmallTruth(t, 3, 0)
	if r, err := badOn.ISOPInterval(badAllowed, TruthISOPOptions{}); r.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(r, err)
	}
	if r, err := (*TruthTable)(nil).ISOPInterval(allowed, TruthISOPOptions{}); r.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(r, err)
	}
}
func kitISOPBoolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}
func FuzzTruthISOP(f *testing.F) {
	f.Add(uint8(6), uint64(0x6996966996696996), true)
	f.Fuzz(func(t *testing.T, size uint8, word uint64, both bool) {
		n := int(size % 7)
		if n < 6 {
			bits := 1 << uint(n)
			mask := ^uint64(0)
			if bits < 64 {
				mask = (uint64(1) << uint(bits)) - 1
			}
			word &= mask
			for shift := bits; shift < 64; shift *= 2 {
				word |= word << uint(shift)
			}
		}
		table, e := TruthTableFromWords(n, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := table.ISOP(TruthISOPOptions{TryBoth: both})
		if e != nil {
			t.Fatal(e)
		}
		for a := 0; a < 1<<uint(n); a++ {
			want := word>>uint(a)&1 != 0
			if kitCoverValue(r.Cubes, a, n) != (want != r.Complemented) {
				t.Fatal(n, word, a, r)
			}
		}
		count, e := CNFISOPLiteralCount(r.Cubes, n)
		if e != nil || count != r.Literals {
			t.Fatal(r, count, e)
		}
	})
}

func FuzzTruthISOPInterval(f *testing.F) {
	f.Add(uint8(4), uint64(0x00a0), uint64(0x0ff0), true)
	f.Fuzz(func(t *testing.T, size uint8, rawOn, rawAllowed uint64, both bool) {
		n := int(size % 7)
		bits := 1 << uint(n)
		mask := ^uint64(0)
		if bits < 64 {
			mask = (uint64(1) << uint(bits)) - 1
		}
		allowedValue := rawAllowed & mask
		onValue := rawOn & allowedValue
		on := kitSmallTruth(t, n, onValue)
		allowed := kitSmallTruth(t, n, allowedValue)
		r, err := on.ISOPInterval(allowed, TruthISOPOptions{TryBoth: both})
		if err != nil || r.Table == nil {
			t.Fatal(n, onValue, allowedValue, r, err)
		}
		for assignment := 0; assignment < 1<<uint(n); assignment++ {
			cover := kitCoverValue(r.Cubes, assignment, n)
			required := onValue>>uint(assignment)&1 != 0
			permitted := allowedValue>>uint(assignment)&1 != 0
			if !r.Complemented && (required && !cover || !permitted && cover) {
				t.Fatal("direct", n, assignment, r)
			}
			if r.Complemented && (!permitted && !cover || required && cover) {
				t.Fatal("complement", n, assignment, r)
			}
			tableValue, tableErr := r.Table.Get(uint64(assignment))
			if tableValue != cover {
				t.Fatal("table", n, assignment, cover, tableValue, tableErr, r.Table.words, r)
			}
		}
		literals, err := CNFISOPLiteralCount(r.Cubes, n)
		if err != nil || literals != r.Literals {
			t.Fatal(literals, r, err)
		}
		if _, err = r.Format(n); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTruthISOPComplementFallback(t *testing.T) {
	table, _ := TruthTableFromWords(6, []uint64{^uint64(1)}, 0)
	normal, e := table.ISOP(TruthISOPOptions{TryBoth: true})
	if e != nil || !normal.Complemented {
		t.Fatal(normal, e)
	}
	found := false
	for limit := 1; limit < 1000; limit++ {
		direct, e := table.ISOP(TruthISOPOptions{MaxElements: limit})
		if e != nil {
			continue
		}
		limited, e := table.ISOP(TruthISOPOptions{MaxElements: limit, TryBoth: true})
		if e != nil || limited.Complemented || !slices.Equal(direct.Cubes, limited.Cubes) {
			t.Fatal(limit, direct, limited, e)
		}
		found = true
		break
	}
	if !found {
		t.Fatal("missing bounded direct cover")
	}
	calls := 0
	table.ISOP(TruthISOPOptions{Stop: func() bool { calls++; return false }})
	threshold := calls
	calls = 0
	r, e := table.ISOP(TruthISOPOptions{TryBoth: true, Stop: func() bool { calls++; return calls > threshold }})
	if e != ErrCancelled || r.Cubes != nil {
		t.Fatal("complement cancellation", r, e)
	}
}
