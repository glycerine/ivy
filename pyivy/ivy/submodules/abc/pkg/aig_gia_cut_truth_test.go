package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaCutTruthNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_cut_truth.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Root          GiaLit
		Leaves, Nodes []GiaID
		Truth         []uint64
		Objects       [][3]int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 136 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source := NewGiaManager()
		for _, object := range row.Objects[1:] {
			if object[0] == 1 {
				source.AppendCI()
			} else {
				if _, e := source.AppendAnd(GiaLit(object[1]), GiaLit(object[2])); e != nil {
					t.Fatal(e)
				}
			}
		}
		for i := range source.objects {
			source.objects[i].value = uint32(i*3 + 7)
			if i != 0 {
				source.objects[i].word0 ^= 1 << 30
				source.objects[i].word1 ^= 1 << 30
			}
		}
		before := slices.Clone(source.objects)
		leaves := slices.Clone(row.Leaves)
		result, e := source.CutTruth(row.Root, row.Leaves, GiaCutTruthOptions{})
		if e != nil || !slices.Equal(result.Table.Words(), row.Truth) || !slices.Equal(result.Nodes, row.Nodes) || result.Table.Variables() != len(row.Leaves) {
			t.Fatal(index, row, result, e)
		}
		if e = result.Table.Check(); e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(source.objects, before) || !slices.Equal(row.Leaves, leaves) {
			t.Fatal("source scratch changed")
		}
	}
}
func TestGiaCutTruthBoundariesAndLimits(t *testing.T) {
	if r, e := (*GiaManager)(nil).CutTruth(GiaConst0, nil, GiaCutTruthOptions{}); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	n, _ := source.AppendAnd(a, b)
	root, _ := source.AppendAnd(n, c.Not())
	leaves := []GiaID{a.ID(), b.ID(), c.ID()}
	before := slices.Clone(source.objects)
	for _, options := range []GiaCutTruthOptions{{MaxObjects: 1}, {MaxMappingEntries: source.ObjectCount() - 1}, {MaxWords: 2}} {
		if r, e := source.CutTruth(root, leaves, options); r.Table != nil || r.Nodes != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(r, e)
		}
	}
	result, e := source.CutTruth(root, leaves, GiaCutTruthOptions{MaxObjects: 2, MaxWords: 3})
	if e != nil || len(result.Nodes) != 2 {
		t.Fatal("exact budget", result, e)
	}
	expected := result.Table.Words()
	result.Table.Set(0, true)
	result.Nodes[0] = InvalidGiaID
	fresh, e := source.CutTruth(root, leaves, GiaCutTruthOptions{})
	if e != nil || !slices.Equal(fresh.Table.Words(), expected) || fresh.Nodes[0] == InvalidGiaID || !slices.Equal(source.objects, before) {
		t.Fatal("ownership", e)
	}
	for _, options := range []GiaCutTruthOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxWords: -1}} {
		if _, e := source.CutTruth(root, leaves, options); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if _, e := source.CutTruth(root, leaves[:1], GiaCutTruthOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal("uncovered input", e)
	}
	if _, e := source.CutTruth(root, []GiaID{0}, GiaCutTruthOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := source.CutTruth(InvalidGiaLit, leaves, GiaCutTruthOptions{}); !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal(e)
	}
	if _, e := source.CutTruth(root, []GiaID{InvalidGiaID}, GiaCutTruthOptions{}); !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal(e)
	}
	for _, when := range []int{1, 4, 6, 8} {
		calls := 0
		r, e := source.CutTruth(root, leaves, GiaCutTruthOptions{Stop: func() bool { calls++; return calls == when }})
		if r.Table != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	trivial, e := source.CutTruth(a.Not(), []GiaID{b.ID(), a.ID(), a.ID()}, GiaCutTruthOptions{MaxWords: 1})
	if e != nil || len(trivial.Nodes) != 0 || trivial.Table.words[0] != ^truthVariablePattern(1) {
		t.Fatal("root leaf and duplicate slot", trivial, e)
	}
	for _, lit := range []GiaLit{GiaConst0, GiaConst1} {
		r, e := source.CutTruth(lit, nil, GiaCutTruthOptions{MaxWords: 1})
		if e != nil || len(r.Nodes) != 0 || !r.Table.IsConstant(lit == GiaConst1) {
			t.Fatal("constant root", r, e)
		}
	}
	co, _ := source.AppendCO(root)
	r, e := source.CutTruth(co, leaves, GiaCutTruthOptions{})
	if e != nil || !slices.Equal(r.Table.Words(), expected) || r.Nodes[len(r.Nodes)-1] != co.ID() {
		t.Fatal("CO root", r, e)
	}
	if _, e := source.CutTruth(root, []GiaID{co.ID()}, GiaCutTruthOptions{}); !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal("CO leaf", e)
	}
}
func TestGiaCutTruthDeepCone(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		root, _ = m.AppendAnd(root, a)
	}
	result, e := m.CutTruth(root, []GiaID{a.ID(), b.ID()}, GiaCutTruthOptions{})
	if e != nil || len(result.Nodes) != 10001 || result.Table.words[0] != truthVariablePattern(0)&truthVariablePattern(1) {
		t.Fatal(result, e)
	}
}
func FuzzGiaCutTruth(f *testing.F) {
	f.Add(uint8(0), uint64(71))
	f.Add(uint8(255), uint64(0x96254718abcdef31))
	f.Fuzz(func(t *testing.T, flags uint8, seed uint64) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		n0, _ := m.AppendAnd(a, b.Not())
		n1, _ := m.AppendXorReal(a.Not(), c)
		n2, _ := m.AppendMuxReal(b, n1, n0)
		n3, _ := m.AppendBuffer(n2.Not())
		root, _ := m.AppendAnd(n3, a)
		pool := []GiaLit{n0, n1, n2, n3, root}
		leaves := []GiaID{a.ID(), b.ID(), c.ID(), pool[int(seed%5)].ID(), pool[int((seed>>8)%5)].ID()}
		if flags&64 != 0 {
			leaves = append(leaves, root.ID())
		}
		if flags&32 != 0 {
			leaves = append(leaves, leaves[0])
		}
		if flags&16 != 0 {
			slices.Reverse(leaves)
		}
		root = root.NotIf(flags&4 != 0)
		co, _ := m.AppendCO(root)
		if flags&128 != 0 {
			root = co
		}
		if flags&8 != 0 {
			if e := m.HashStart(); e != nil {
				t.Fatal(e)
			}
		}
		before := slices.Clone(m.objects)
		result, e := m.CutTruth(root, leaves, GiaCutTruthOptions{})
		if e != nil {
			t.Fatal(e)
		}
		slots := map[GiaID]int{}
		for slot, id := range leaves {
			if _, exists := slots[id]; !exists {
				slots[id] = slot
			}
		}
		for assignment := 0; assignment < 1<<uint(len(leaves)); assignment++ {
			values := make([]bool, m.ObjectCount())
			read := func(lit GiaLit) bool { return values[lit.ID()] != lit.Complemented() }
			for id := 1; id < m.ObjectCount(); id++ {
				if slot, exists := slots[GiaID(id)]; exists {
					values[id] = assignment&(1<<uint(slot)) != 0
					continue
				}
				object := m.objects[id]
				if object.IsCI() {
					continue
				}
				first, _ := object.Fanin0(GiaID(id))
				value := read(first)
				if object.IsAnd() && !object.IsBuffer() {
					second, _ := object.Fanin1(GiaID(id))
					if control, mux := m.MuxControl(GiaID(id)); mux {
						if read(control) {
							value = read(second)
						}
					} else if object.IsXor() {
						value = value != read(second)
					} else {
						value = value && read(second)
					}
				}
				values[id] = value
			}
			got, e := result.Table.Get(uint64(assignment))
			if e != nil || got != read(root) {
				t.Fatal("free boundary assignment", assignment, leaves, result, e)
			}
		}
		visited := make([]bool, m.ObjectCount())
		visited[0] = true
		for _, id := range leaves {
			visited[id] = true
		}
		expected := []GiaID{}
		var collect func(GiaID)
		collect = func(id GiaID) {
			if visited[id] {
				return
			}
			visited[id] = true
			children, n, e := m.crossCutFanins(id)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < n; i++ {
				collect(children[i])
			}
			expected = append(expected, id)
		}
		collect(root.ID())
		if !slices.Equal(result.Nodes, expected) {
			t.Fatal("postorder", result.Nodes, expected)
		}
		if !slices.Equal(m.objects, before) {
			t.Fatal("source changed")
		}
		if e := result.Table.Check(); e != nil {
			t.Fatal(e)
		}
	})
}
