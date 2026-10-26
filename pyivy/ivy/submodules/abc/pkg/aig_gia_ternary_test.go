package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaTernaryExtendedTruthTables(t *testing.T) {
	// Each gate is checked against independent Boolean completions of its own
	// distinct input CIs, giving an exact local abstract truth-table oracle.
	for kind := 0; kind < 4; kind++ {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		var root GiaLit
		var e error
		switch kind {
		case 0:
			root, e = m.AppendAnd(a.Not(), b)
		case 1:
			root, e = m.AppendXorReal(a, b.Not())
		case 2:
			root, e = m.AppendMuxReal(c.Not(), a, b)
		case 3:
			root, e = m.AppendBuffer(a.Not())
		}
		if e != nil {
			t.Fatal(e)
		}
		m.AppendCO(root.Not())
		before := slices.Clone(m.objects)
		for av := GiaTernaryZero; av <= GiaTernaryUnknown; av++ {
			for bv := GiaTernaryZero; bv <= GiaTernaryUnknown; bv++ {
				for cv := GiaTernaryZero; cv <= GiaTernaryUnknown; cv++ {
					inputs := []GiaTernary{av, bv, cv}
					possible := uint8(0)
					for bits := 0; bits < 8; bits++ {
						booleans := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
						allowed := true
						for i, v := range inputs {
							if v != GiaTernaryUnknown && (v == GiaTernaryOne) != booleans[i] {
								allowed = false
							}
						}
						if !allowed {
							continue
						}
						a, b, c := booleans[0], booleans[1], booleans[2]
						value := false
						switch kind {
						case 0:
							value = !a && b
						case 1:
							value = a != !b
						case 2:
							value = b
							if !c {
								value = a
							}
						case 3:
							value = !a
						}
						value = !value
						if value {
							possible |= 2
						} else {
							possible |= 1
						}
					}
					got, e := m.EvaluateTernary(inputs, GiaCofactorOptions{})
					if e != nil || len(got) != 1 || got[0] != GiaTernary(possible) {
						t.Fatal(kind, inputs, got, possible, e)
					}
				}
			}
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source mutation")
		}
	}
}
func TestGiaTernaryLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	inputs := []GiaTernary{GiaTernaryUnknown, GiaTernaryZero}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		v, e := m.EvaluateTernary(inputs, options)
		if v != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 2, 3} {
		calls := 0
		v, e := m.EvaluateTernary(inputs, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if v != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, input := range [][]GiaTernary{nil, {0, GiaTernaryOne}, {4, GiaTernaryUnknown}} {
		if _, e := m.EvaluateTernary(input, GiaCofactorOptions{}); e == nil {
			t.Fatal(input)
		}
	}
}
func FuzzGiaTernarySoundness(f *testing.F) {
	f.Add(uint8(5), uint8(17))
	f.Fuzz(func(t *testing.T, phase, encoding uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a.NotIf(phase&1 != 0), b)
		y, _ := m.AppendMuxReal(c, x, a)
		z, _ := m.AppendAnd(y, b.NotIf(phase&2 != 0))
		m.AppendCO(z.Not())
		input := []GiaTernary{1 + GiaTernary(encoding%3), 1 + GiaTernary((encoding/3)%3), 1 + GiaTernary((encoding/9)%3)}
		got, e := m.EvaluateTernary(input, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for bits := 0; bits < 8; bits++ {
			values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			allowed := true
			for i, v := range input {
				if v != GiaTernaryUnknown && (v == GiaTernaryOne) != values[i] {
					allowed = false
				}
			}
			if !allowed {
				continue
			}
			a, b, c := values[0], values[1], values[2]
			v := a
			if c {
				v = (a != (phase&1 != 0)) != b
			}
			want := !(v && (b != (phase&2 != 0)))
			if got[0] != GiaTernaryUnknown && (got[0] == GiaTernaryOne) != want {
				t.Fatal(input, got, values, want)
			}
		}
	})
}

func TestGiaTernaryNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_ternary.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		A, B     GiaTernary
		Ca, Cb   int
		And, Not GiaTernary
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 36 {
		t.Fatal(e)
	}
	for _, row := range rows {
		a, b := row.A, row.B
		if row.Ca != 0 {
			a = giaTernaryNot(a)
		}
		if row.Cb != 0 {
			b = giaTernaryNot(b)
		}
		if giaTernaryAnd(a, b) != row.And || a != row.Not {
			t.Fatal(row, a, b)
		}
	}
}
