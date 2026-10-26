package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDuplicateLevelizedNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, c)
	z, _ := m.AppendAnd(b, c)
	w, _ := m.AppendAnd(y, z)
	m.AppendAnd(a, c)
	m.AppendCO(w)
	m.AppendCO(y.Not())
	raw, err := os.ReadFile("test_vectors/oracle/gia_dup_levelized.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct{ Words [][]GiaLit }
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	g, mapping, err := m.DuplicateLevelized(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(native.Words) != g.ObjectCount() {
		t.Fatal("object count")
	}
	for i, obj := range g.objects {
		var got []GiaLit
		if obj.IsAnd() || obj.IsCO() {
			l, _ := obj.Fanin0(GiaID(i))
			got = append(got, l)
		}
		if obj.IsAnd() {
			l, _ := obj.Fanin1(GiaID(i))
			got = append(got, l)
		}
		if !slices.Equal(got, native.Words[i]) {
			t.Fatal(i, got, native.Words[i])
		}
	}
	if mapping[y.ID()].ID() <= mapping[z.ID()].ID() {
		t.Fatal("levels were not reordered")
	}
	checkGiaLevelizedTruth(t, m, g, mapping)
}

func checkGiaLevelizedTruth(t *testing.T, m, g *GiaManager, mapping []GiaLit) {
	t.Helper()
	if len(mapping) != m.ObjectCount() || g.ObjectCount() != m.ObjectCount() {
		t.Fatal("mapping dimensions")
	}
	for assignment := 0; assignment < 1<<m.InputCount(); assignment++ {
		input := make([]bool, m.InputCount())
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		want, e := m.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got, e := g.Evaluate(input)
		if e != nil || !slices.Equal(want, got) {
			t.Fatal(assignment, want, got, e)
		}
	}
}

func TestGiaDuplicateLevelizedExtendedAndLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "levels"
	m.Spec = "levels.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(x.Not())
	buf2, _ := m.AppendBuffer(buf)
	mux, _ := m.AppendMuxReal(c, buf2, b)
	m.AppendCO(mux)
	m.AppendCO(buf)
	m.AppendCO(a)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, mapping, err := m.DuplicateLevelized(GiaCofactorOptions{MaxObjects: m.ObjectCount(), MaxMappingEntries: m.ObjectCount()})
	if err != nil {
		t.Fatal(err)
	}
	checkGiaLevelizedTruth(t, m, g, mapping)
	if g.Name != m.Name || g.Spec != m.Spec || g.registers != 1 || g.constraints != 1 || g.XorCount() != 1 || g.MuxCount() != 1 || g.BufferCount() != 2 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/kinds/source")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxObjects: m.ObjectCount() - 1}, {MaxMappingEntries: m.ObjectCount() - 1}} {
		g, mapping, err := m.DuplicateLevelized(options)
		if g != nil || mapping != nil || err == nil {
			t.Fatal(options, err)
		}
	}
	for _, threshold := range []int{1, 3, 7, 10} {
		calls := 0
		g, mapping, err := m.DuplicateLevelized(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || mapping != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
	empty := NewGiaManager()
	g, mapping, err = empty.DuplicateLevelized(GiaCofactorOptions{})
	if err != nil || g.ObjectCount() != 1 || !slices.Equal(mapping, []GiaLit{GiaConst0}) {
		t.Fatal("empty", err)
	}
}
