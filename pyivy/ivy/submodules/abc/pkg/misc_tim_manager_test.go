package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func timTerminalWords(p TimTerminal) [6]int64 {
	return [6]int64{int64(p.ID), int64(p.TravID), int64(p.Box), int64(p.Position), int64(math.Float32bits(p.Arrival)), int64(math.Float32bits(p.Required))}
}
func timSnapshot(m *TimManager) []int64 {
	r := []int64{int64(m.epoch)}
	if m.useTraversal {
		r = append(r, 1)
	} else {
		r = append(r, 0)
	}
	for _, objects := range [][]TimTerminal{m.cis, m.cos} {
		for _, p := range objects {
			words := timTerminalWords(p)
			r = append(r, words[:]...)
		}
	}
	return r
}
func TestTimTerminalsNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_terminals.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Cis, Cos, Epoch int
		Value           uint32
		CI, CO          [][6]int64
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 128 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		m, err := NewTimManager(row.Cis, row.Cos, 0)
		if err != nil {
			t.Fatal(err)
		}
		v := math.Float32frombits(row.Value)
		m.InitPIArrivalAll(v)
		m.InitPORequiredAll(v)
		m.IncrementTraversal()
		for i := 0; i < row.Cos; i++ {
			if err = m.SetCOArrival(i, v); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < row.Cis; i++ {
			if err = m.SetCIRequired(i, v); err != nil {
				t.Fatal(err)
			}
		}
		m.IncrementTraversal()
		for i := 0; i < row.Cis; i++ {
			m.CIArrival(i)
		}
		for i := 0; i < row.Cos; i++ {
			m.CORequired(i)
		}
		if m.TraversalID() != row.Epoch {
			t.Fatal(index, "epoch")
		}
		for i, want := range row.CI {
			actual, err := m.CI(i)
			if err != nil || timTerminalWords(actual) != want {
				t.Fatal(index, i, actual, want, err)
			}
		}
		for i, want := range row.CO {
			actual, err := m.CO(i)
			if err != nil || timTerminalWords(actual) != want {
				t.Fatal(index, i, actual, want, err)
			}
		}
	}
}
func TestTimTraversalAndAtomicity(t *testing.T) {
	m, err := NewTimManager(2, 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetCOArrival(0, 1); err == nil {
		t.Fatal("epoch-zero arrival assertion")
	}
	if err = m.SetCIRequired(0, 1); err == nil {
		t.Fatal("epoch-zero required assertion")
	}
	if err = m.SetCORequired(0, 1); err != nil {
		t.Fatal(err)
	}
	if err = m.SetCORequired(0, 2); err != nil {
		t.Fatal(err)
	}
	m.IncrementTraversal()
	m.SetCORequired(1, 3)
	before := timSnapshot(m)
	if err = m.InitPORequiredAll(8); err == nil || !slices.Equal(before, timSnapshot(m)) {
		t.Fatal("partial bulk write", err)
	}
	clone := m.Clone()
	clone.InitPIArrival(0, 99)
	if m.cis[0].Arrival != 0 {
		t.Fatal("clone alias")
	}
	m.SetTraversalEnabled(false)
	if err = m.SetCOArrival(1, 5); err != nil {
		t.Fatal(err)
	}
	if err = m.SetCOArrival(1, 6); err != nil {
		t.Fatal(err)
	}
	if _, err = m.IsCurrent(false, 1); err == nil {
		t.Fatal("disabled traversal assertion")
	}
	m.SetTraversalEnabled(true)
	m.epoch = timTraversalLimit
	m.cis[0].TravID = 100
	m.cos[0].TravID = 200
	m.IncrementTraversal()
	if m.epoch != 1 {
		t.Fatal(m.epoch)
	}
	for _, p := range append(slices.Clone(m.cis), m.cos...) {
		if p.TravID != 0 {
			t.Fatal("rollover", p)
		}
	}
	for _, dims := range [][3]int{{-1, 2, 0}, {1, -1, 0}, {2, 2, 3}, {1, 1, -1}} {
		if _, err = NewTimManager(dims[0], dims[1], dims[2]); err == nil {
			t.Fatal(dims)
		}
	}
	for _, index := range []int{-1, 2} {
		if err = m.InitPIArrival(index, 1); err == nil {
			t.Fatal(index)
		}
		if _, err = m.CORequired(index); err == nil {
			t.Fatal(index)
		}
	}
	var missing *TimManager
	if err = missing.IncrementTraversal(); err == nil {
		t.Fatal("nil manager")
	}
	if missing.Clone() != nil {
		t.Fatal("nil clone")
	}
}
func FuzzTimTerminalSequence(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint32(0x7fc01234))
	f.Add([]byte{0, 8, 1, 9}, uint32(0x80000000))
	f.Fuzz(func(t *testing.T, data []byte, bits uint32) {
		if len(data) > 100 {
			data = data[:100]
		}
		m, err := NewTimManager(2, 2, 4)
		if err != nil {
			t.Fatal(err)
		}
		value := math.Float32frombits(bits)
		for _, op := range data {
			index := int(op>>4) % 2
			before := timSnapshot(m)
			expectedError := false
			switch op % 10 {
			case 0:
				expectedError = m.useTraversal && m.cos[index].TravID == m.epoch
				err = m.SetCOArrival(index, value)
			case 1:
				expectedError = m.useTraversal && m.cis[index].TravID == m.epoch
				err = m.SetCIRequired(index, value)
			case 2:
				expectedError = m.useTraversal && m.epoch != 0 && m.cos[index].TravID == m.epoch
				err = m.SetCORequired(index, value)
			case 3:
				err = m.InitPIArrival(index, value)
			case 4:
				err = m.InitPORequired(index, value)
			case 5:
				err = m.IncrementTraversal()
			case 6:
				_, err = m.CIArrival(index)
			case 7:
				_, err = m.CORequired(index)
			case 8:
				err = m.SetTraversalEnabled(!m.useTraversal)
			case 9:
				for _, p := range m.cos {
					expectedError = expectedError || m.useTraversal && m.epoch != 0 && p.TravID == m.epoch
				}
				err = m.InitPORequiredAll(value)
			}
			if (err != nil) != expectedError {
				t.Fatal(op, err, expectedError)
			}
			if err != nil && !slices.Equal(before, timSnapshot(m)) {
				t.Fatal("failed operation changed state")
			}
			copy := m.Clone()
			if !slices.Equal(timSnapshot(m), timSnapshot(copy)) {
				t.Fatal("clone changed float bits")
			}
		}
	})
}
