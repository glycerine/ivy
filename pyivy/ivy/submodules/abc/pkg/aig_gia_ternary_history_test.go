package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaTernaryHistoryConvergence(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	m.AppendCI()
	m.AppendCO(a)
	m.AppendCO(GiaConst0)
	m.AppendCO(p)
	m.AppendCO(b.Not())
	m.SetRegisterCount(3)
	before := slices.Clone(m.objects)
	r, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{})
	if e != nil {
		t.Fatal(e)
	}
	expected := [][]GiaTernary{{1, 1, 1}, {1, 3, 2}, {1, 3, 3}, {1, 3, 3}}
	if len(r.States) != len(expected) || r.LoopStart != 2 || !slices.Equal(r.ZeroCounts, []int{4, 1, 1}) || !slices.Equal(r.UnknownCounts, []int{0, 3, 2}) || !slices.Equal(before, m.objects) {
		t.Fatal(r)
	}
	for i := range expected {
		if !slices.Equal(r.States[i], expected[i]) {
			t.Fatal(r.States, expected)
		}
	}
	r.States[2][0] = 3
	if r.States[3][0] != 1 {
		t.Fatal("state alias")
	}
	for _, options := range []GiaTernaryHistoryOptions{{MaxStates: 1}, {MaxStateEntries: 3}, {MaxStates: -1}, {MaxStateEntries: -1}, {GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}, {GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}} {
		r, e := m.SimulateTernaryHistory(options)
		if r.States != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 8, 15} {
		calls := 0
		r, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}})
		if r.States != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func giaTernaryCounter(t *testing.T, width int) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	state := make([]GiaLit, width)
	for i := range state {
		state[i], _ = m.AppendCI()
	}
	m.HashStart()
	carry := GiaConst1
	next := make([]GiaLit, width)
	for i, bit := range state {
		var e error
		next[i], e = m.HashXor(bit, carry)
		if e != nil {
			t.Fatal(e)
		}
		carry, e = m.HashAnd(carry, bit)
		if e != nil {
			t.Fatal(e)
		}
	}
	m.HashStop()
	for _, l := range next {
		m.AppendCO(l)
	}
	m.SetRegisterCount(width)
	return m
}
func TestGiaTernaryHistoryRetirement(t *testing.T) {
	m := giaTernaryCounter(t, 9)
	r, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if len(r.States) < 303 || r.LoopStart < 301 {
		t.Fatal("counter did not exercise retirement", len(r.States), r.LoopStart)
	}
	for i, state := range r.States[:302] {
		for bit, v := range state {
			want := GiaTernaryZero
			if i&(1<<bit) != 0 {
				want = GiaTernaryOne
			}
			if v != want {
				t.Fatal(i, bit, v, want)
			}
		}
	}
	if _, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{MaxStates: 301}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
}
func FuzzGiaTernaryHistorySoundness(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(x.NotIf(phase&1 != 0))
		m.AppendCO(b.NotIf(phase&2 != 0))
		m.SetRegisterCount(2)
		r, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{MaxStates: 16})
		if e != nil {
			t.Fatal(e)
		}
		state := []bool{false, false}
		for frame := 0; frame < 8; frame++ {
			index := frame
			if index >= len(r.States)-1 {
				index = r.LoopStart + (index-r.LoopStart)%(len(r.States)-1-r.LoopStart)
			}
			abstract := r.States[index]
			for bit, v := range abstract {
				if v != GiaTernaryUnknown && (v == GiaTernaryOne) != state[bit] {
					t.Fatal(frame, abstract, state)
				}
			}
			pv := sequence&(1<<frame) != 0
			state = []bool{(pv != state[0]) != (phase&1 != 0), state[1] != (phase&2 != 0)}
		}
	})
}

func TestGiaTernaryHistoryNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_ternary_history.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []GiaTernaryHistory
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		var m *GiaManager
		if mode == 0 {
			m = NewGiaManager()
			p, _ := m.AppendCI()
			a, _ := m.AppendCI()
			b, _ := m.AppendCI()
			m.AppendCI()
			m.AppendCO(a)
			m.AppendCO(GiaConst0)
			m.AppendCO(p)
			m.AppendCO(b.Not())
			m.SetRegisterCount(3)
		} else {
			m = giaTernaryCounter(t, 9)
		}
		r, e := m.SimulateTernaryHistory(GiaTernaryHistoryOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.LoopStart != row.LoopStart || len(r.States) != len(row.States) || !slices.Equal(r.ZeroCounts, row.ZeroCounts) || !slices.Equal(r.UnknownCounts, row.UnknownCounts) {
			t.Fatal(mode, r.LoopStart, len(r.States), row.LoopStart, len(row.States))
		}
		for i, state := range r.States {
			if !slices.Equal(state, row.States[i]) {
				t.Fatal("native state history", mode, i, state, row.States[i])
			}
		}
	}
}
