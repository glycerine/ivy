package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func giaLUTVerificationModel(mapping *GiaLUTMapping, m *GiaManager) bool {
	available := make([]bool, m.ObjectCount())
	for id, obj := range m.objects {
		if !obj.IsAnd() || obj.IsBuffer() {
			available[id] = true
			continue
		}
		record, ok := mapping.Record(GiaID(id))
		if !ok {
			continue
		}
		available[id] = true
		for _, fanin := range record.Fanins {
			available[id] = available[id] && available[fanin]
		}
	}
	for id, obj := range m.objects {
		if obj.IsBuffer() || obj.IsCO() {
			driver, _ := obj.Fanin0(GiaID(id))
			if !available[driver.ID()] {
				return false
			}
		}
	}
	return true
}
func TestGiaLUTVerifyNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_verify.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Verified int
		Packed   []int
		Objects  [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 128 {
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
			case 5:
				_, err = m.AppendBuffer(GiaLit(obj[1]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		mapping, err := GiaLUTMappingFromPacked(m.ObjectCount(), row.Packed, 0)
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(m.objects)
		valid, err := mapping.Verify(m, GiaLUTAnalysisOptions{})
		if err != nil || valid != (row.Verified != 0) || valid != giaLUTVerificationModel(mapping, m) {
			t.Fatal(index, valid, row.Verified, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}
func TestGiaLUTVerifyLimitsAndDepth(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	records := []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}
	for i := 0; i < 10000; i++ {
		next, _ := m.AppendAnd(root, a)
		records = append(records, GiaLUTRecord{Root: next.ID(), Fanins: []GiaID{root.ID(), a.ID()}})
		root = next
	}
	m.AppendCO(root)
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := mapping.Verify(m, GiaLUTAnalysisOptions{MaxVisits: 10001, MaxObjects: m.ObjectCount()}); !valid || err != nil {
		t.Fatal(valid, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 10000}, {MaxObjects: m.ObjectCount() - 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if valid, err := mapping.Verify(m, options); valid || err == nil {
			t.Fatal(options, valid, err)
		}
	}
	for _, threshold := range []int{2, 5, 100} {
		calls := 0
		if valid, err := mapping.Verify(m, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= threshold }}); valid || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, valid, err)
		}
	}
	var missing *GiaLUTMapping
	if _, err := missing.Verify(m, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := mapping.Verify(nil, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	m.AppendCI()
	if _, err := mapping.Verify(m, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	empty := NewGiaManager()
	emptyMap, err := NewGiaLUTMapping(1, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := emptyMap.Verify(empty, GiaLUTAnalysisOptions{}); !valid || err != nil {
		t.Fatal(valid, err)
	}
}
func FuzzGiaLUTVerify(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(0))
	f.Add([]byte{255, 0, 17, 42}, uint8(255))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		if flags&128 != 0 {
			driver, _ := m.objects[m.cos[0]].Fanin0(m.cos[0])
			m.AppendBuffer(driver)
		}
		records := make([]GiaLUTRecord, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() || obj.IsBuffer() || uint8(id)%4 == flags%4 {
				continue
			}
			leaves := slices.Clone(m.cis)
			if flags&2 != 0 {
				a, _ := obj.Fanin0(GiaID(id))
				b, _ := obj.Fanin1(GiaID(id))
				leaves = []GiaID{a.ID(), b.ID(), a.ID()}
			}
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: leaves})
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(m.objects)
		actual, err := mapping.Verify(m, GiaLUTAnalysisOptions{})
		if err != nil || actual != giaLUTVerificationModel(mapping, m) {
			t.Fatal(actual, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	})
}

func TestGiaLUTVerifyBufferChains(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	first, _ := m.AppendAnd(a, b)
	buffer1, _ := m.AppendBuffer(first)
	buffer2, _ := m.AppendBuffer(buffer1)
	root, _ := m.AppendAnd(buffer2, b)
	m.AppendCO(root)
	records := []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{buffer2.ID(), b.ID()}}}
	partial, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := partial.Verify(m, GiaLUTAnalysisOptions{}); valid || err != nil {
		t.Fatal("buffer driver omitted", valid, err)
	}
	records = append(records, GiaLUTRecord{Root: first.ID(), Fanins: []GiaID{a.ID(), b.ID()}})
	complete, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := complete.Verify(m, GiaLUTAnalysisOptions{}); !valid || err != nil {
		t.Fatal("complete buffer chain", valid, err)
	}
}
