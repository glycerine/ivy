package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDuplicateFromListsNative(t *testing.T) {
	m := NewGiaManager()
	m.Name = "lists"
	m.Spec = "lists.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(b, c)
	z, _ := m.AppendAnd(x, y.Not())
	u, _ := m.AppendCO(z)
	v, _ := m.AppendCO(y.Not())
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, mapping, err := m.DuplicateFromLists([]GiaID{c.ID(), a.ID(), b.ID()}, []GiaID{y.ID(), x.ID(), z.ID()}, []GiaID{v.ID(), u.ID()}, 1, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_dup_lists.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct{ Fanins [][]GiaLit }
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if len(native.Fanins) != g.ObjectCount() {
		t.Fatal("count")
	}
	for i, obj := range g.objects {
		var fanins []GiaLit
		if obj.IsAnd() || obj.IsCO() {
			l, _ := obj.Fanin0(GiaID(i))
			fanins = append(fanins, l)
		}
		if obj.IsAnd() {
			l, _ := obj.Fanin1(GiaID(i))
			fanins = append(fanins, l)
		}
		if !slices.Equal(fanins, native.Fanins[i]) {
			t.Fatal(i, fanins, native.Fanins[i])
		}
	}
	if g.Name != m.Name || g.Spec != m.Spec || g.RegisterCount() != 1 || g.ConstraintCount() != 0 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	if mapping[c.ID()] != GiaLit(2) || mapping[v.ID()].ID() != g.cos[0] {
		t.Fatal("mapping")
	}
	for assignment := 0; assignment < 8; assignment++ {
		av, bv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		got, err := g.Evaluate([]bool{cv, av, bv})
		if err != nil || !slices.Equal(got, []bool{!(bv && cv), (av && bv) && !(bv && cv)}) {
			t.Fatal(assignment, got, err)
		}
	}
}

func TestGiaDuplicateFromListsValidation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	buf, _ := m.AppendBuffer(x)
	co, _ := m.AppendCO(buf)
	cis, gates, cos := []GiaID{a.ID(), b.ID()}, []GiaID{x.ID(), buf.ID()}, []GiaID{co.ID()}
	for _, tc := range []struct {
		ci, g, co []GiaID
		regs      int
	}{
		{cis, gates, cos, -1}, {cis, gates, cos, 2}, {[]GiaID{a.ID(), a.ID()}, gates, cos, 0},
		{cis, []GiaID{buf.ID(), x.ID()}, cos, 0}, {cis, []GiaID{x.ID(), x.ID()}, cos, 0},
		{[]GiaID{a.ID()}, gates, cos, 0}, {cis, nil, cos, 0}, {[]GiaID{x.ID()}, nil, nil, 0},
		{cis, []GiaID{co.ID()}, nil, 0}, {cis, gates, []GiaID{a.ID()}, 0}, {[]GiaID{InvalidGiaID}, nil, nil, 0},
	} {
		g, mapping, err := m.DuplicateFromLists(tc.ci, tc.g, tc.co, tc.regs, GiaCofactorOptions{})
		if g != nil || mapping != nil || err == nil {
			t.Fatal(tc, err)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxObjects: m.ObjectCount() - 1}, {MaxMappingEntries: m.ObjectCount() - 1}} {
		g, mapping, err := m.DuplicateFromLists(cis, gates, cos, 0, options)
		if g != nil || mapping != nil || err == nil {
			t.Fatal(options, err)
		}
	}
	for _, threshold := range []int{1, 3, 6} {
		calls := 0
		g, mapping, err := m.DuplicateFromLists(cis, gates, cos, 0, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || mapping != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
	g, mapping, err := m.DuplicateFromLists(nil, nil, nil, 0, GiaCofactorOptions{MaxObjects: 1})
	if err != nil || g.ObjectCount() != 1 || mapping[a.ID()] != InvalidGiaLit {
		t.Fatal("empty", err)
	}
}
