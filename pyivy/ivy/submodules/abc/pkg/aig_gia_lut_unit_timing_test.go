package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTUnitTimingNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_unit_timing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Start, Maximum     int
		Arrivals, Required []int
		Fanins             [][]GiaID
		Objects            [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = m.AppendCI()
			case 2:
				_, err = m.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		records := make([]GiaLUTRecord, 0)
		for id, fanins := range row.Fanins {
			if len(fanins) > 0 {
				records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: fanins})
			}
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(m.objects)
		r, err := n.UnitTiming(m, row.Start, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.Maximum || !slices.Equal(r.Arrivals, row.Arrivals) || !slices.Equal(r.Required, row.Required) {
			t.Fatal(index, r, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source graph changed")
		}
	}
}
func TestGiaLUTUnitTimingDanglingDepthAndLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	first, _ := m.AppendAnd(a, b)
	root := first
	records := []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}
	for i := 0; i < 10000; i++ {
		next, _ := m.AppendAnd(root, a)
		records = append(records, GiaLUTRecord{Root: next.ID(), Fanins: []GiaID{root.ID(), a.ID()}})
		root = next
	}
	m.AppendCO(first)
	mapping, _ := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	n, _ := mapping.Network(GiaLUTAnalysisOptions{})
	r, err := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{MaxVisits: 40004})
	if err != nil || r.Maximum != 10001 || r.Required[root.ID()] != 1000000000 || r.Required[a.ID()] != 10000 {
		t.Fatal(r.Maximum, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 40003}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := n.UnitTiming(m, 0, options); r.Arrivals != nil || r.Required != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	calls := 0
	if r, err := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 10 }}); r.Arrivals != nil || err == nil {
		t.Fatal(r, err)
	}
	r.Arrivals[0] = 99
	fresh, err := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
	if err != nil || fresh.Arrivals[0] != 0 {
		t.Fatal("result alias", err)
	}
}
func FuzzGiaLUTUnitTiming(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, int8(0))
	f.Add([]byte{255, 0, 17, 42}, int8(12))
	f.Fuzz(func(t *testing.T, data []byte, start int8) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		records := make([]GiaLUTRecord, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() || id%5 == 0 {
				continue
			}
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: []GiaID{a.ID(), b.ID()}})
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		r, err := n.UnitTiming(m, int(start), GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		arrivals := make([]int, m.ObjectCount())
		seen := make([]bool, m.ObjectCount())
		var arrival func(GiaID) int
		arrival = func(id GiaID) int {
			if seen[id] {
				return arrivals[id]
			}
			seen[id] = true
			record, ok := mapping.Record(id)
			if !ok {
				return 0
			}
			for _, fanin := range record.Fanins {
				arrivals[id] = max(arrivals[id], arrival(fanin)+1)
			}
			return arrivals[id]
		}
		maximum := int(start)
		for _, root := range n.Roots() {
			maximum = max(maximum, arrival(root))
		}
		base := make([]int, m.ObjectCount())
		for i := range base {
			base[i] = 1000000000
		}
		for _, id := range m.cos {
			lit, _ := m.objects[id].Fanin0(id)
			base[lit.ID()] = min(base[lit.ID()], maximum)
		}
		required := make([]int, m.ObjectCount())
		seen = make([]bool, m.ObjectCount())
		var requirement func(GiaID) int
		requirement = func(id GiaID) int {
			if seen[id] {
				return required[id]
			}
			seen[id] = true
			required[id] = base[id]
			for _, record := range records {
				for _, fanin := range record.Fanins {
					if fanin == id {
						required[id] = min(required[id], requirement(record.Root)-1)
					}
				}
			}
			return required[id]
		}
		for id := range required {
			requirement(GiaID(id))
		}
		if maximum != r.Maximum || !slices.Equal(arrivals, r.Arrivals) || !slices.Equal(required, r.Required) {
			t.Fatal("independent path models", r, arrivals, required)
		}
	})
}
