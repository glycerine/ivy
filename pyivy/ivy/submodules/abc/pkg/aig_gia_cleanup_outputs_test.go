package abc

// Tests for src/aig/gia/giaScl.c:Gia_ManCleanupOutputs.

import (
	"errors"
	"slices"
	"testing"
)

func TestGiaCleanupOutputsRetainsSkippedCones(t *testing.T) {
	m := NewGiaManager()
	m.Name, m.Spec = "skip", "skip.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	firstOnly, _ := m.AppendAnd(a, b)
	kept, _ := m.AppendXorReal(b, c)
	constraint, _ := m.AppendMuxReal(a, b, c)
	dead, _ := m.AppendAnd(a.Not(), b)
	bufferCone, _ := m.AppendAnd(a, c)
	buffer, _ := m.AppendBuffer(bufferCone.Not())
	firstCO, _ := m.AppendCO(firstOnly)
	m.AppendCO(kept)
	m.AppendCO(constraint)
	if err := m.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(m.objects)
	g, mapping, err := m.CleanupOutputs(1)
	if err != nil {
		t.Fatal(err)
	}
	if g.OutputCount() != 2 || g.ConstraintCount() != 1 || g.RegisterCount() != 0 || g.Name != m.Name || g.Spec != m.Spec {
		t.Fatal("metadata", g.OutputCount(), g.ConstraintCount())
	}
	if mapping[firstCO.ID()] != InvalidGiaLit || mapping[firstOnly.ID()] == InvalidGiaLit || mapping[dead.ID()] != InvalidGiaLit || mapping[bufferCone.ID()] == InvalidGiaLit || mapping[buffer.ID()] == InvalidGiaLit {
		t.Fatal("marked duplication mapping", mapping[firstCO.ID()], mapping[firstOnly.ID()], mapping[dead.ID()], mapping[bufferCone.ID()], mapping[buffer.ID()])
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("source mutated")
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, err := m.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := g.Evaluate(inputs)
		if err != nil || !slices.Equal(got, want[1:]) {
			t.Fatal(assignment, got, want, err)
		}
	}
	onlyConstraint, onlyMapping, err := m.CleanupOutputs(2)
	if err != nil || onlyConstraint.OutputCount() != 1 || onlyConstraint.ConstraintCount() != 1 || onlyMapping[kept.ID()] == InvalidGiaLit {
		t.Fatal("retain constraint suffix", onlyConstraint, err)
	}
	unchanged, _, err := m.CleanupOutputs(0)
	if err != nil || unchanged.OutputCount() != m.OutputCount() {
		t.Fatal("zero prefix", unchanged, err)
	}
}

func TestGiaCleanupOutputsRejectsInvalidRequests(t *testing.T) {
	if g, mapping, err := (*GiaManager)(nil).CleanupOutputs(0); g != nil || mapping != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(g, mapping, err)
	}
	empty := NewGiaManager()
	if g, mapping, err := empty.CleanupOutputs(0); g != nil || mapping != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(g, mapping, err)
	}
	m := NewGiaManager()
	a, _ := m.AppendCI()
	m.AppendCO(a)
	for _, skip := range []int{-1, 1, 2} {
		if g, mapping, err := m.CleanupOutputs(skip); g != nil || mapping != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(skip, g, mapping, err)
		}
	}
	sequential := NewGiaManager()
	pi, _ := sequential.AppendCI()
	sequential.AppendCI()
	sequential.AppendCO(pi)
	sequential.AppendCO(pi)
	sequential.SetRegisterCount(1)
	if g, mapping, err := sequential.CleanupOutputs(0); g != nil || mapping != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(g, mapping, err)
	}
	constrained := NewGiaManager()
	a, _ = constrained.AppendCI()
	constrained.AppendCO(a)
	constrained.AppendCO(a.Not())
	constrained.SetConstraintCount(2)
	if g, mapping, err := constrained.CleanupOutputs(1); g != nil || mapping != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(g, mapping, err)
	}
}

func FuzzGiaCleanupOutputs(f *testing.F) {
	f.Add(uint8(1), uint8(5))
	f.Fuzz(func(t *testing.T, rawSkip, phases uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendAnd(a.NotIf(phases&1 != 0), b)
		y, _ := m.AppendXorReal(b.NotIf(phases&2 != 0), c)
		z, _ := m.AppendMuxReal(a, x, y.NotIf(phases&4 != 0))
		m.AppendAnd(a.Not(), c.Not())
		m.AppendCO(x)
		m.AppendCO(y)
		m.AppendCO(z)
		m.AppendCO(a.NotIf(phases&8 != 0))
		m.SetConstraintCount(1)
		skip := int(rawSkip % 4)
		g, mapping, err := m.CleanupOutputs(skip)
		if err != nil || g.Check() != nil || g.OutputCount() != 4-skip || g.ConstraintCount() != 1 {
			t.Fatal(skip, g, err)
		}
		for _, root := range []GiaLit{x, y, z} {
			if mapping[root.ID()] == InvalidGiaLit {
				t.Fatal("removed marked cone", skip, root)
			}
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			want, _ := m.Evaluate(inputs)
			got, err := g.Evaluate(inputs)
			if err != nil || !slices.Equal(got, want[skip:]) {
				t.Fatal(skip, assignment, got, want, err)
			}
		}
	})
}
