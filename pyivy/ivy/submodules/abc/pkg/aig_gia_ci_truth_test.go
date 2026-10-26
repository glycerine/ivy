package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func checkGiaCITruthAssignments(t *testing.T, m *GiaManager, root GiaLit, r GiaCITruth6) {
	t.Helper()
	if !slices.IsSorted(r.Inputs) || r.Fits != (len(r.Inputs) <= 6) {
		t.Fatal("support shape", r)
	}
	// Independent forward support propagation retains structurally used inputs
	// even when Boolean cancellation makes them functionally irrelevant.
	masks := make([]uint64, m.ObjectCount())
	for i, id := range m.cis {
		masks[id] = uint64(1) << uint(i)
	}
	for id, object := range m.objects {
		if !object.IsAnd() && !object.IsCO() {
			continue
		}
		first, err := object.Fanin0(GiaID(id))
		if err != nil {
			t.Fatal(err)
		}
		masks[id] = masks[first.ID()]
		if object.IsAnd() && !object.IsBuffer() {
			second, err := object.Fanin1(GiaID(id))
			if err != nil {
				t.Fatal(err)
			}
			masks[id] |= masks[second.ID()]
		}
		if control, mux := m.MuxControl(GiaID(id)); mux {
			masks[id] |= masks[control.ID()]
		}
	}
	var actualMask uint64
	for i, id := range r.Inputs {
		if uint64(id) >= uint64(m.ObjectCount()) || !m.objects[id].IsCI() || i > 0 && id == r.Inputs[i-1] {
			t.Fatal("invalid support", r.Inputs)
		}
		actualMask |= uint64(1) << uint(m.objects[id].Diff1())
	}
	if actualMask != masks[root.ID()] {
		t.Fatal("structural support", actualMask, masks[root.ID()])
	}
	if !r.Fits {
		if r.Word != 0 {
			t.Fatal("oversized word", r)
		}
		return
	}
	for assignment := 0; assignment < 1<<uint(len(m.cis)); assignment++ {
		in := make([]bool, len(m.cis))
		for i := range in {
			in[i] = assignment&(1<<uint(i)) != 0
		}
		values, err := m.evaluateValues(in)
		if err != nil {
			t.Fatal(err)
		}
		local := 0
		for i, id := range r.Inputs {
			if !m.objects[id].IsCI() {
				t.Fatal("non-CI support", id)
			}
			if values[id] {
				local |= 1 << uint(i)
			}
		}
		expected := values[root.ID()] != root.Complemented()
		if (r.Word>>uint(local)&1 != 0) != expected {
			t.Fatal("assignment", assignment, local, r)
		}
	}
	// The unused high input dimensions must repeat the actual local function.
	for bit := 0; bit < 64; bit++ {
		if r.Word>>uint(bit)&1 != r.Word>>uint(bit&((1<<uint(len(r.Inputs)))-1))&1 {
			t.Fatal("unstretched word")
		}
	}
}

func TestGiaCITruth6Native(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_ci_truth.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Root    GiaLit
		Word    uint64
		Inputs  []GiaID
		Objects [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 144 {
		t.Fatal(len(rows))
	}
	oversized := 0
	for index, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			if obj[0] == 1 {
				_, err = m.AppendCI()
			} else {
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		before := slices.Clone(m.objects)
		r, err := m.CITruth6(row.Root, GiaCITruthOptions{})
		if err != nil || !slices.Equal(r.Inputs, row.Inputs) || r.Word != row.Word || r.Fits != (len(row.Inputs) <= 6) {
			t.Fatal(index, r, row, err)
		}
		if !r.Fits {
			oversized++
		}
		checkGiaCITruthAssignments(t, m, row.Root, r)
		if !slices.Equal(before, m.objects) {
			t.Fatal("source mutated")
		}
	}
	if oversized == 0 {
		t.Fatal("missing oversized native cases")
	}
}

func TestGiaCITruth6ExtensionsAndLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	mux, _ := m.AppendMuxReal(c, x, a.Not())
	buf, _ := m.AppendBuffer(mux)
	co, _ := m.AppendCO(buf.Not())
	for id := 1; id < m.ObjectCount(); id++ {
		m.objects[id].word0 |= uint32(1) << 30
		m.objects[id].word1 |= uint32(1) << 30
		m.objects[id].value = uint32(id * 31)
	}
	before := slices.Clone(m.objects)
	for _, root := range []GiaLit{GiaConst0, GiaConst1, a, a.Not(), x, mux, buf, co} {
		r, err := m.CITruth6(root, GiaCITruthOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaCITruthAssignments(t, m, root, r)
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("source scratch mutated")
	}
	for _, options := range []GiaCITruthOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if _, err := m.CITruth6(mux, options); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
	for _, options := range []GiaCITruthOptions{{MaxObjects: 1}, {MaxMappingEntries: m.ObjectCount() - 1}} {
		if r, err := m.CITruth6(mux, options); !errors.Is(err, ErrLimit) || r.Inputs != nil {
			t.Fatal(r, err)
		}
	}
	for _, threshold := range []int{1, 3, 8, 15} {
		calls := 0
		r, err := m.CITruth6(co, GiaCITruthOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || r.Inputs != nil {
			t.Fatal(threshold, r, err)
		}
	}
	for _, root := range []GiaLit{InvalidGiaLit, GiaLit(m.ObjectCount() << 1)} {
		if _, err := m.CITruth6(root, GiaCITruthOptions{}); !errors.Is(err, ErrInvalidLiteral) {
			t.Fatal(err)
		}
	}
	var missing *GiaManager
	if _, err := missing.CITruth6(a, GiaCITruthOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	deep := NewGiaManager()
	first, _ := deep.AppendCI()
	second, _ := deep.AppendCI()
	root, _ := deep.AppendAnd(first, second)
	for i := 0; i < 10000; i++ {
		root, _ = deep.AppendAnd(root, first)
	}
	result, err := deep.CITruth6(root, GiaCITruthOptions{MaxObjects: 10003, MaxMappingEntries: deep.ObjectCount()})
	if err != nil || result.Word != 0x8888888888888888 {
		t.Fatal(result, err)
	}
	if _, err = deep.CITruth6(root, GiaCITruthOptions{MaxObjects: 10002}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzGiaCITruth6(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, false)
	f.Add([]byte{255, 0, 17, 42}, true)
	f.Fuzz(func(t *testing.T, data []byte, phase bool) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		root, _ := m.objects[m.cos[0]].Fanin0(m.cos[0])
		root = root.NotIf(phase)
		r, err := m.CITruth6(root, GiaCITruthOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaCITruthAssignments(t, m, root, r)
		r.Inputs = append(r.Inputs, 0)
		fresh, err := m.CITruth6(root, GiaCITruthOptions{})
		if err != nil || len(fresh.Inputs) == len(r.Inputs) {
			t.Fatal("owned support", err)
		}
	})
}
