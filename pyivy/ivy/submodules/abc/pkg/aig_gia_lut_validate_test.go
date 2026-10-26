package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func checkGiaLUTComposition(t *testing.T, m *GiaManager, n *GiaLUTNetwork, functions []GiaLUTFunction) {
	t.Helper()
	for mask := 0; mask < 1<<uint(len(m.cis)); mask++ {
		inputs := make([]bool, len(m.cis))
		v := make([]bool, len(m.objects))
		for i, id := range m.cis {
			inputs[i] = mask>>uint(i)&1 != 0
			v[id] = inputs[i]
		}
		for _, f := range functions {
			index := 0
			for i, id := range n.fanins[f.Root] {
				if v[id] {
					index |= 1 << uint(i)
				}
			}
			v[f.Root] = f.Truth>>uint(index)&1 != 0
		}
		want, e := m.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		for i, id := range m.cos {
			l, _ := m.objects[id].Fanin0(id)
			if (v[l.ID()] != l.Complemented()) != want[i] {
				t.Fatal(mask, id, v, want)
			}
		}
	}
}
func TestGiaWindowApplyNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_window_apply.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Rotation int
		Truth          uint64
		Fanins         [][]GiaID
		Refs           []int
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := NewGiaManager()
		var ci []GiaLit
		for i := 0; i < 4; i++ {
			l, _ := m.AppendCI()
			ci = append(ci, l)
		}
		a, _ := m.AppendAnd(ci[0].NotIf(row.Mode&1 != 0), ci[1].NotIf(row.Mode&2 != 0))
		b, _ := m.AppendAnd(ci[2], ci[3].NotIf(row.Mode&1 != 0))
		c, _ := m.AppendAnd(a, b)
		d, _ := m.AppendAnd(c, ci[0].NotIf(row.Mode&4 != 0))
		root, _ := m.AppendAnd(d, b)
		m.AppendCO(root)
		nodes := []GiaID{a.ID(), b.ID(), c.ID(), d.ID(), root.ID()}
		var records []GiaLUTRecord
		for _, id := range nodes {
			x, _ := m.objects[id].Fanin0(id)
			y, _ := m.objects[id].Fanin1(id)
			records = append(records, GiaLUTRecord{Root: id, Fanins: []GiaID{x.ID(), y.ID()}})
		}
		mapping, e := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if e != nil {
			t.Fatal(e)
		}
		n, _ := mapping.Network(GiaLUTAnalysisOptions{})
		window := make([][]GiaID, len(nodes))
		for i := 0; i < 4; i++ {
			window[4] = append(window[4], ci[(i+row.Rotation)%4].ID())
		}
		before := slices.Clone(m.objects)
		r, e := n.WithValidatedUnitWindow(m, nodes, window, 0, GiaLUTAnalysisOptions{})
		if e != nil || r.Network == nil {
			t.Fatal(row, e)
		}
		if len(r.Functions) != 1 || r.Functions[0].Root != root.ID() || r.Functions[0].Truth != row.Truth {
			t.Fatal(row, r.Functions)
		}
		refs := make([]int, m.ObjectCount())
		for i, f := range r.Network.fanins {
			if !slices.Equal(f, row.Fanins[i]) {
				t.Fatal(row, i, f)
			}
			for _, id := range f {
				refs[id]++
			}
			if len(r.Network.fanouts[i]) != row.Refs[i] && GiaID(i) != root.ID() {
				t.Fatal("fanouts", i)
			}
		}
		refs[root.ID()]++
		if !slices.Equal(refs, row.Refs) {
			t.Fatal(refs, row.Refs)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		if !slices.Equal(before, m.objects) {
			t.Fatal("graph changed")
		}
		window[4][0] = 0
		if r.Network.fanins[root.ID()][0] == 0 {
			t.Fatal("aliased proposal")
		}
	}
}
func TestGiaLUTValidateExtended(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	mux, _ := m.AppendMuxReal(c, x, a)
	buf, _ := m.AppendBuffer(mux.Not())
	m.AppendCO(buf.Not())
	n := &GiaLUTNetwork{fanins: make([][]GiaID, m.ObjectCount())}
	n.fanins[x.ID()] = []GiaID{a.ID(), b.ID()}
	n.fanins[mux.ID()] = []GiaID{a.ID(), x.ID(), c.ID()}
	n.fanins[buf.ID()] = []GiaID{mux.ID(), mux.ID(), 0}
	f, e := n.ValidateFunctions(m, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	checkGiaLUTComposition(t, m, n, f)
	if f[len(f)-1].Truth != ^uint64(0xcccccccccccccccc) {
		t.Fatal(f)
	}
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		r, e := n.WithValidatedUnitWindow(m, nil, nil, 0, o)
		if e == nil || r.Network != nil || r.Functions != nil {
			t.Fatal(r, e)
		}
	}
	if _, e := n.WithValidatedUnitWindow(m, []GiaID{x.ID()}, [][]GiaID{nil}, 0, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("removed dependency")
	}
	n.fanins[x.ID()] = []GiaID{a.ID()}
	if _, e := n.ValidateFunctions(m, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("uncovered CI")
	}
}
func TestGiaWindowApplyTiming(t *testing.T) {
	m, n, nodes, window := buildGiaUnitWindowTest(t)
	before := slices.Clone(m.objects)
	r, e := n.WithValidatedUnitWindow(m, nodes, window, 0, GiaLUTAnalysisOptions{})
	if e == nil || r.Network != nil {
		t.Fatal("late mapping accepted", r, e)
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("graph changed")
	}
}
func FuzzGiaWindowApply(f *testing.F) {
	f.Add(uint8(4), uint8(0))
	f.Fuzz(func(t *testing.T, size, mode uint8) {
		m, n, l, ids, _ := mappingStateFixture(t, 1+int(size%8), int(mode%2), 4)
		m.AppendCO(GiaLit(ids[len(ids)-1]) << 1)
		n.fanins = append(n.fanins, nil)
		n.fanouts = append(n.fanouts, nil)
		s, e := n.RemapUnitWindow(m, l, ids, []GiaID{ids[len(ids)-1]}, 0, GiaUnitWindowSearchOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := n.WithValidatedUnitWindow(m, ids, s.Search.BestWindow, 0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		if s.Search.Improved {
			if s.Replacement.Network == nil {
				t.Fatal("missing replacement")
			}
			checkGiaLUTComposition(t, m, s.Replacement.Network, s.Replacement.Functions)
		}
	})
}

func FuzzGiaLUTFunctions(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 1, 2, 0, 4, 2, 3, 1, 6})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		m := NewGiaManager()
		var ci, signals []GiaLit
		for i := 0; i < 4; i++ {
			l, _ := m.AppendCI()
			ci = append(ci, l)
			signals = append(signals, l)
		}
		var gates []GiaLit
		for i := 0; i+3 < len(data); i += 4 {
			a := signals[int(data[i])%len(signals)].NotIf(data[i]&128 != 0)
			b := signals[int(data[i+1])%len(signals)].NotIf(data[i+1]&128 != 0)
			c := signals[int(data[i+2])%len(signals)]
			var l GiaLit
			var e error
			if a.ID() == b.ID() {
				l, e = m.AppendBuffer(a)
			} else {
				switch data[i+3] % 3 {
				case 0:
					l, e = m.AppendAnd(a, b)
				case 1:
					l, e = m.AppendXorReal(a, b)
				case 2:
					if c.ID() != a.ID() && c.ID() != b.ID() {
						l, e = m.AppendMuxReal(c, a, b)
					} else {
						l, e = m.AppendAnd(a, b)
					}
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			signals = append(signals, l)
			gates = append(gates, l)
		}
		if len(gates) == 0 {
			return
		}
		m.AppendCO(gates[len(gates)-1])
		m.AppendCO(gates[len(gates)/2].Not())
		n := &GiaLUTNetwork{fanins: make([][]GiaID, m.ObjectCount())}
		for i, l := range gates {
			for j := 0; j < 4; j++ {
				n.fanins[l.ID()] = append(n.fanins[l.ID()], ci[(j+i)%4].ID())
			}
		}
		functions, e := n.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkGiaLUTComposition(t, m, n, functions)
	})
}
