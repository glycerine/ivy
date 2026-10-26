package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func demiterSource(empty, extended bool) *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	var x, y, z GiaLit
	if extended {
		x, _ = m.AppendXorReal(a, b)
		y, _ = m.AppendMuxReal(c, x, a.Not())
		z, _ = m.AppendBuffer(b.Not())
	} else {
		x, _ = m.AppendAnd(a, b)
		y, _ = m.AppendAnd(x.Not(), c)
		z, _ = m.AppendAnd(b, c)
	}
	if !empty {
		m.AppendCO(y)
		m.AppendCO(z.Not())
		m.AppendCO(x)
		m.AppendCO(GiaConst0.Not())
		m.SetConstraintCount(1)
	}
	m.Name = "pairs"
	m.Spec = "pairs.aig"
	return m
}
func TestGiaDemiterNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_demiter.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Empty, Words, Side, Objects int
		Drivers                     []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 8 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := demiterSource(row.Empty != 0, false)
		r, e := m.SplitPairedOutputs(row.Words != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graphs[row.Side]
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), drivers)
		}
	}
}
func checkDemiter(t *testing.T, empty, words, extended bool) {
	t.Helper()
	m := demiterSource(empty, extended)
	before := slices.Clone(m.objects)
	r, e := m.SplitPairedOutputs(words, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for side, g := range r.Graphs {
		var indices []int
		for i := range m.cos {
			if (!words && i%2 == side) || (words && (i < len(m.cos)/2) == (side == 0)) {
				indices = append(indices, i)
			}
		}
		if !slices.Equal(r.Outputs[side], indices) || g.InputCount() != 3 || g.registers != 0 || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("projection/metadata")
		}
		for bits := 0; bits < 8; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			source, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(input)
			if e != nil || len(got) != len(indices) {
				t.Fatal(e)
			}
			for i, j := range indices {
				if got[i] != source[j] {
					t.Fatal(side, bits, got, source)
				}
			}
		}
	}
	// Verify independent graph and projection ownership after construction.
	if !empty {
		old := r.Graphs[1].ObjectCount()
		r.Graphs[0].AppendCO(GiaConst0)
		if r.Graphs[1].ObjectCount() != old || !slices.Equal(before, m.objects) {
			t.Fatal("aliased graphs")
		}
		r.Outputs[0][0] = 99
		if r.Outputs[1][0] == 99 {
			t.Fatal("aliased projections")
		}
	}
}
func TestGiaDemiterSemantics(t *testing.T) {
	for _, empty := range []bool{false, true} {
		for _, words := range []bool{false, true} {
			for _, extended := range []bool{false, true} {
				checkDemiter(t, empty, words, extended)
			}
		}
	}
	m := demiterSource(false, true)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.SplitPairedOutputs(false, o)
		if r.Graphs[0] != nil || r.Graphs[1] != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.SplitPairedOutputs(false, o)
		if r.Graphs[0] != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 5, 10, 15} {
		calls := 0
		r, e := m.SplitPairedOutputs(false, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graphs[0] != nil || r.Graphs[1] != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, source := range []*GiaManager{nil, func() *GiaManager { g := demiterSource(false, false); g.AppendCO(GiaConst0); return g }(), func() *GiaManager { g := demiterSource(false, false); g.SetRegisterCount(1); return g }()} {
		r, e := source.SplitPairedOutputs(false, GiaCofactorOptions{})
		if r.Graphs[0] != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
}
func FuzzGiaDemiter(f *testing.F) {
	f.Add([]byte{0, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{1, 9, 12, 15, 18})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		m := NewGiaManager()
		pool := []GiaLit{GiaConst0}
		for i := 0; i < 3; i++ {
			v, _ := m.AppendCI()
			pool = append(pool, v)
		}
		m.HashStart()
		for i := 1; i+1 < len(data); i += 2 {
			a := pool[int(data[i])%len(pool)].NotIf(data[i]&128 != 0)
			b := pool[int(data[i+1])%len(pool)].NotIf(data[i+1]&128 != 0)
			var v GiaLit
			var e error
			if data[i]&1 == 0 {
				v, e = m.HashAnd(a, b)
			} else {
				v, e = m.HashXor(a, b)
			}
			if e != nil {
				t.Fatal(e)
			}
			pool = append(pool, v)
		}
		m.HashStop()
		n := int(data[0]%6) * 2
		for i := 0; i < n; i++ {
			v := data[i%len(data)]
			m.AppendCO(pool[int(v)%len(pool)].NotIf(v&64 != 0))
		}
		words := data[0]&128 != 0
		r, e := m.SplitPairedOutputs(words, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for bits := 0; bits < 8; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			source, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for side, g := range r.Graphs {
				got, e := g.Evaluate(input)
				if e != nil || len(got) != n/2 {
					t.Fatal(e)
				}
				for i, v := range got {
					j := 2*i + side
					if words {
						j = i + side*(n/2)
					}
					if r.Outputs[side][i] != j || v != source[j] {
						t.Fatal(side, bits, got, source)
					}
				}
			}
		}
	})
}
