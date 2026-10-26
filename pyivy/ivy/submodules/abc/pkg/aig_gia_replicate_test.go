package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaReplicationIndependentBanks(t *testing.T) {
	m := NewGiaManager()
	m.Name = "copies"
	m.Spec = "copies.aig"
	p, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, r)
	buf, _ := m.AppendBuffer(x.Not())
	m.AppendCO(buf)
	m.AppendCO(p)
	m.AppendCO(x)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for copies := 1; copies <= 3; copies++ {
		g, e := m.Replicate(copies, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != copies || g.constraints != copies || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != copies || g.BufferCount() != copies || g.ObjectCount() != 1+copies*(m.ObjectCount()-1) || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for assignment := 0; assignment < 1<<(2*copies); assignment++ {
			input := make([]bool, 2*copies)
			for i := range input {
				input[i] = assignment&(1<<i) != 0
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for copy := 0; copy < copies; copy++ {
				p, r := input[copy], input[copies+copy]
				if got[copy] != (p == r) || got[copies+copy] != p || got[2*copies+copy] != (p != r) {
					t.Fatal(copies, assignment, copy, got)
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.Replicate(2, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 16} {
		calls := 0
		g, e := m.Replicate(2, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, copies := range []int{-1, 0, int(^uint(0) >> 1)} {
		if g, e := m.Replicate(copies, GiaCofactorOptions{}); g != nil || e == nil {
			t.Fatal(copies, e)
		}
	}
	empty := NewGiaManager()
	g, e := empty.Replicate(3, GiaCofactorOptions{})
	if e != nil || g.ObjectCount() != 1 {
		t.Fatal("empty", e)
	}
}
func FuzzGiaReplication(f *testing.F) {
	f.Add(uint8(2), uint8(42))
	f.Fuzz(func(t *testing.T, n, assignment uint8) {
		copies := 1 + int(n%3)
		m := NewGiaManager()
		p, _ := m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendAnd(p, r)
		m.AppendCO(x.Not())
		m.AppendCO(p.Not())
		m.SetRegisterCount(1)
		g, e := m.Replicate(copies, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		input := make([]bool, 2*copies)
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		got, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for copy := 0; copy < copies; copy++ {
			if got[copy] != !(input[copy] && input[copies+copy]) || got[copies+copy] != !input[copy] {
				t.Fatal(got, input)
			}
		}
	})
}

func TestGiaReplicationNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, r)
	m.AppendCO(x.Not())
	m.AppendCO(p.Not())
	m.SetRegisterCount(1)
	raw, e := os.ReadFile("test_vectors/oracle/gia_replicate.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Cis     []GiaID
		Drivers []GiaLit
		Ands    [][3]int
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for i, row := range rows {
		g, e := m.Replicate(i+1, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || len(g.cis) != len(row.Cis) {
			t.Fatal(i, g.cis, drivers, row)
		}
		for assignment := 0; assignment < 1<<len(row.Cis); assignment++ {
			input := make([]bool, len(row.Cis))
			nativeValues := make([]bool, row.Objects)
			for index, id := range row.Cis {
				input[index] = assignment&(1<<index) != 0
				nativeValues[id] = input[index]
			}
			read := func(l int) bool { return nativeValues[l>>1] != (l&1 != 0) }
			for _, gate := range row.Ands {
				nativeValues[gate[0]] = read(gate[1]) && read(gate[2])
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for output, driver := range row.Drivers {
				if got[output] != read(int(driver)) {
					t.Fatal("native replication truth", i, assignment, output)
				}
			}
		}
	}
}
