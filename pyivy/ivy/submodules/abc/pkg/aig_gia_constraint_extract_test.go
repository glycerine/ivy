package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaConstraintExtractionNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x.Not(), c)
	m.AppendCO(y)
	m.Name = "extract"
	m.Spec = "extract.aig"
	before := slices.Clone(m.objects)
	g, err := m.ExtractConjunctiveConstraint(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_constraint_extract.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Objects int
		Masks   []uint
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	masks := []uint{0, 0}
	for assignment := 0; assignment < 8; assignment++ {
		input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		values, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		original, e := m.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		if (values[0] && values[1]) != original[0] {
			t.Fatal("conjunction", assignment)
		}
		for i, v := range values {
			if v {
				masks[i] |= 1 << assignment
			}
		}
	}
	if g.ObjectCount() != native.Objects || masks[0] != native.Masks[0] || masks[1] != (native.Masks[1]^255) {
		t.Fatal(masks, native)
	}
	if g.ConstraintCount() != 1 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.ExtractConjunctiveConstraint(options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 8, 12} {
		calls := 0
		g, e := m.ExtractConjunctiveConstraint(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaConstraintExtraction(f *testing.F) {
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(255), uint8(13))
	f.Fuzz(func(t *testing.T, phases, assignment uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		buf, _ := m.AppendBuffer(c.Not())
		root, e := m.AppendAnd(x.NotIf(phases&1 != 0), buf.NotIf(phases&2 != 0))
		if e != nil {
			t.Fatal(e)
		}
		m.AppendCO(root)
		g, e := m.ExtractConjunctiveConstraint(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		v, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		xv := (input[0] != input[1]) != (phases&1 != 0)
		bv := (!input[2]) != (phases&2 != 0)
		if v[0] != xv || v[1] != bv {
			t.Fatal(v, xv, bv)
		}
	})
}

func TestGiaConstraintExtractionRejectsInvalidRoots(t *testing.T) {
	for mode := 0; mode < 6; mode++ {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		root := a
		switch mode {
		case 0:
			root = GiaConst0
		case 1:
			root = GiaConst1
		case 2:
			root = a
		case 3:
			x, _ := m.AppendAnd(a, b)
			root = x.Not()
		case 4:
			root, _ = m.AppendXorReal(a, b)
		case 5:
			root, _ = m.AppendAnd(a, b)
		}
		m.AppendCO(root)
		if mode == 5 {
			m.SetConstraintCount(1)
		}
		g, err := m.ExtractConjunctiveConstraint(GiaCofactorOptions{})
		if g != nil || err == nil {
			t.Fatal(mode, err)
		}
	}
	var m *GiaManager
	if _, err := m.ExtractConjunctiveConstraint(GiaCofactorOptions{}); err == nil {
		t.Fatal("nil")
	}
}
