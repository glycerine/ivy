package abc

import (
	"errors"
	"slices"
	"testing"
)

func giaFanoutEdgesModel(t testing.TB, m *GiaManager) [][]GiaFanoutEdge {
	t.Helper()
	result := make([][]GiaFanoutEdge, len(m.objects))
	for id, object := range m.objects {
		target := GiaID(id)
		if object.IsAnd() {
			first, err := object.Fanin0(target)
			if err != nil {
				t.Fatal(err)
			}
			result[first.ID()] = append(result[first.ID()], GiaFanoutEdge{target, 0})
			if !object.IsBuffer() {
				second, err := object.Fanin1(target)
				if err != nil {
					t.Fatal(err)
				}
				result[second.ID()] = append(result[second.ID()], GiaFanoutEdge{target, 1})
			}
		} else if object.IsCO() {
			first, err := object.Fanin0(target)
			if err != nil {
				t.Fatal(err)
			}
			result[first.ID()] = append(result[first.ID()], GiaFanoutEdge{target, 0})
		}
	}
	return result
}

func TestGiaDynamicAndMappingFanouts(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, a)
	co, _ := m.AppendCO(y)
	dynamic, err := m.DynamicFanouts(GiaFanoutOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model := giaFanoutEdgesModel(t, m)
	total := 0
	for id, want := range model {
		got, err := dynamic.Fanouts(GiaID(id))
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(id, got, want, err)
		}
		total += len(want)
	}
	if dynamic.EdgeCount() != total {
		t.Fatal(dynamic.EdgeCount(), total)
	}
	if err = dynamic.Remove(a.ID(), x.ID(), 0); err != nil {
		t.Fatal(err)
	}
	if err = dynamic.Add(m, a.ID(), x.ID(), 0); err != nil {
		t.Fatal(err)
	}
	got, _ := dynamic.Fanouts(a.ID())
	want := append(slices.Clone(model[a.ID()][1:]), model[a.ID()][0])
	if !slices.Equal(got, want) {
		t.Fatal(got, want)
	}
	if err = dynamic.Add(m, a.ID(), x.ID(), 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = dynamic.Remove(a.ID(), co.ID(), 0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: x.ID(), Fanins: []GiaID{a.ID(), b.ID(), a.ID()}}, {Root: y.ID(), Fanins: []GiaID{x.ID(), a.ID()}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	static, err := m.StaticMappingFanouts(mapping, GiaFanoutOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range map[GiaID][]GiaID{
		a.ID(): {x.ID(), x.ID(), y.ID()},
		b.ID(): {x.ID()},
		x.ID(): {y.ID()},
		y.ID(): {co.ID()},
	} {
		got, err := static.Fanouts(id)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(id, got, want, err)
		}
	}
	if _, err = m.DynamicFanouts(GiaFanoutOptions{MaxEdges: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = m.StaticMappingFanouts(mapping, GiaFanoutOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzGiaDynamicFanouts(f *testing.F) {
	f.Add(uint16(0x1234))
	f.Add(uint16(0xabcd))
	f.Fuzz(func(t *testing.T, seed uint16) {
		m := NewGiaManager()
		nodes := []GiaLit{}
		for range 3 {
			literal, err := m.AppendCI()
			if err != nil {
				t.Fatal(err)
			}
			nodes = append(nodes, literal)
		}
		for i := 0; i < 8; i++ {
			first := nodes[int(seed>>uint(i%8))%len(nodes)]
			second := nodes[int(seed>>uint((i+5)%12))%len(nodes)]
			if seed>>uint(i%16)&1 != 0 {
				first = first.Not()
			}
			literal, err := m.AppendAndSimplified(first, second)
			if err != nil {
				t.Fatal(err)
			}
			nodes = append(nodes, literal)
		}
		for i := 0; i < 3; i++ {
			if _, err := m.AppendCO(nodes[(int(seed)+i)%len(nodes)]); err != nil {
				t.Fatal(err)
			}
		}
		index, err := m.DynamicFanouts(GiaFanoutOptions{})
		if err != nil {
			t.Fatal(err)
		}
		model := giaFanoutEdgesModel(t, m)
		for id, want := range model {
			got, err := index.Fanouts(GiaID(id))
			if err != nil || !slices.Equal(got, want) {
				t.Fatal(id, got, want, err)
			}
			for _, edge := range slices.Clone(got) {
				if err = index.Remove(GiaID(id), edge.Fanout, edge.Slot); err != nil {
					t.Fatal(err)
				}
				if err = index.Add(m, GiaID(id), edge.Fanout, edge.Slot); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}
