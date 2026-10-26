package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTAnalysisNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_analysis.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Overlap, Overlap2         int
		References, Sizes, Packed []int
		Objects                   [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 30 {
		t.Fatal(len(rows))
	}
	positive, negative := false, false
	for index, row := range rows {
		graph := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = graph.AppendCI()
			case 2:
				_, err = graph.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = graph.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		mapping, err := GiaLUTMappingFromPacked(graph.ObjectCount(), row.Packed, 0)
		if err != nil {
			t.Fatal(index, err)
		}
		before := slices.Clone(graph.objects)
		r, err := mapping.Analyze(graph, GiaLUTAnalysisOptions{})
		if err != nil || r.Overlap != row.Overlap || r.Overlap != row.Overlap2 || !slices.Equal(r.References, row.References) || !slices.Equal(r.ConeSizes, row.Sizes) {
			t.Fatal(index, r, row, err)
		}
		positive = positive || r.Overlap > 0
		negative = negative || r.Overlap < 0
		if !slices.Equal(before, graph.objects) {
			t.Fatal("source changed")
		}
		checkGiaLUTAnalysisModel(t, mapping, graph, r)
	}
	if !positive || !negative {
		t.Fatal("missing signed overlap cases")
	}
}

// Reverse topological propagation is independent of the production DFS stack.
func checkGiaLUTAnalysisModel(t *testing.T, mapping *GiaLUTMapping, m *GiaManager, r GiaLUTAnalysis) {
	t.Helper()
	refs := make([]int, m.ObjectCount())
	sizes := make([]int, m.ObjectCount())
	overlap := 0
	for _, obj := range m.objects {
		if obj.IsAnd() {
			overlap--
		}
	}
	for _, id := range m.cos {
		lit, _ := m.objects[id].Fanin0(id)
		refs[lit.ID()]++
	}
	for _, root := range mapping.Roots() {
		record, _ := mapping.Record(root)
		boundary := make([]bool, m.ObjectCount())
		active := make([]bool, m.ObjectCount())
		boundary[0] = true
		for _, id := range record.Fanins {
			refs[id]++
			boundary[id] = true
		}
		active[root] = true
		for id := int(root); id > 0; id-- {
			if !active[id] || boundary[id] {
				continue
			}
			obj := m.objects[id]
			if obj.IsCI() {
				t.Fatal("uncovered model CI", id)
			}
			sizes[root]++
			overlap++
			a, err := obj.Fanin0(GiaID(id))
			if err != nil {
				t.Fatal(err)
			}
			active[a.ID()] = true
			if obj.IsAnd() && !obj.IsBuffer() {
				b, _ := obj.Fanin1(GiaID(id))
				active[b.ID()] = true
			}
			if c, mux := m.MuxControl(GiaID(id)); mux {
				active[c.ID()] = true
			}
		}
	}
	if overlap != r.Overlap || !slices.Equal(refs, r.References) || !slices.Equal(sizes, r.ConeSizes) {
		t.Fatal("independent analysis model", r, refs, sizes, overlap)
	}
}
func TestGiaLUTAnalysisExtensionsLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	y, _ := m.AppendMuxReal(c, x, a)
	z, _ := m.AppendBuffer(y)
	m.AppendCO(z)
	m.AppendCO(GiaConst1)
	records := []GiaLUTRecord{{Root: z.ID(), Fanins: []GiaID{a.ID(), b.ID(), c.ID()}, MuxID: int(z.ID())}}
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	for id := 1; id < m.ObjectCount(); id++ {
		m.objects[id].word0 |= 1 << 30
		m.objects[id].word1 |= 1 << 30
		m.objects[id].value = uint32(id * 19)
	}
	before := slices.Clone(m.objects)
	r, err := mapping.Analyze(m, GiaLUTAnalysisOptions{MaxVisits: 3, MaxObjects: m.ObjectCount()})
	if err != nil || r.ConeSizes[z.ID()] != 3 || r.References[0] != 1 {
		t.Fatal(r, err)
	}
	checkGiaLUTAnalysisModel(t, mapping, m, r)
	if !slices.Equal(before, m.objects) {
		t.Fatal("scratch changed")
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxObjects: m.ObjectCount() - 1}, {MaxVisits: 2}, {Stop: func() bool { return true }}} {
		result, err := mapping.Analyze(m, options)
		if err == nil || result.References != nil || result.ConeSizes != nil {
			t.Fatal(options, result, err)
		}
	}
	for _, threshold := range []int{2, 5, 7, 8} {
		calls := 0
		result, err := mapping.Analyze(m, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.References != nil {
			t.Fatal(threshold, result, err)
		}
	}
	broken, err := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: z.ID(), Fanins: []GiaID{a.ID()}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := broken.Analyze(m, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) || result.References != nil {
		t.Fatal(result, err)
	}
	var missing *GiaLUTMapping
	if _, err := missing.Analyze(m, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := mapping.Analyze(nil, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	deep := NewGiaManager()
	first, _ := deep.AppendCI()
	second, _ := deep.AppendCI()
	root, _ := deep.AppendAnd(first, second)
	for i := 0; i < 10000; i++ {
		root, _ = deep.AppendAnd(root, first)
	}
	deep.AppendCO(root)
	deepMap, err := NewGiaLUTMapping(deep.ObjectCount(), []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{first.ID(), second.ID()}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, err := deepMap.Analyze(deep, GiaLUTAnalysisOptions{MaxVisits: 10001})
	if err != nil || result.ConeSizes[root.ID()] != 10001 || result.Overlap != 0 {
		t.Fatal(result.Overlap, err)
	}
	r.References[0] = 99
	fresh, err := mapping.Analyze(m, GiaLUTAnalysisOptions{})
	if err != nil || fresh.References[0] != 1 {
		t.Fatal("owned result", err)
	}
}
func FuzzGiaLUTAnalysis(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(0))
	f.Add([]byte{255, 0, 17, 42}, uint8(255))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		records := make([]GiaLUTRecord, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() || flags&1 != 0 && id%2 != 0 {
				continue
			}
			leaves := slices.Clone(m.cis)
			if flags&2 != 0 {
				a, _ := obj.Fanin0(GiaID(id))
				b, _ := obj.Fanin1(GiaID(id))
				leaves = []GiaID{a.ID(), b.ID(), a.ID()}
			}
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: leaves, MuxID: id})
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		r, err := mapping.Analyze(m, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaLUTAnalysisModel(t, mapping, m, r)
	})
}
