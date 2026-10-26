package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthMemoryNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/vec_truth_memory.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Complement, Words, PageLog int
		Entries                               [][]uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 20 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		m, e := NewSeededTruthTableMemory(r.Variables, r.Complement != 0, WordMemoryOptions{})
		if e != nil {
			t.Fatal(e)
		}
		id, e := m.AddMuxTruthTable(r.Variables)
		if e != nil || id != 2 {
			t.Fatal(id, e)
		}
		if m.EntryWords() != r.Words || m.PageLog() != r.PageLog || m.Len() != len(r.Entries) || m.PageCount() != 1 {
			t.Fatal("metadata")
		}
		for i, want := range r.Entries {
			got, e := m.Read(i)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(r.Variables, r.Complement, i, got, want, e)
			}
		}
		falseTable, _ := m.Read(0)
		variable, _ := m.Read(1)
		mux, _ := m.Read(2)
		for assignment := 0; assignment < 1<<uint(r.Variables); assignment++ {
			word, bit := assignment/64, uint(assignment%64)
			if falseTable[word]&(1<<bit) != 0 {
				t.Fatal("constant")
			}
			wantVar := (assignment&1 != 0) != (r.Complement != 0)
			gotVar := variable[word]&(1<<bit) != 0
			wantMux := assignment&1 != 0
			if assignment&4 != 0 {
				wantMux = assignment&2 != 0
			}
			gotMux := mux[word]&(1<<bit) != 0
			if gotVar != wantVar || gotMux != wantMux {
				t.Fatal(r.Variables, assignment)
			}
		}
		id, e = m.AddMuxTruthTable(r.Variables)
		if e != nil || id != 2 || m.Len() != 3 {
			t.Fatal("idempotent", id, e)
		}
	}
}
func TestTruthMemoryLayoutsAndLimits(t *testing.T) {
	empty, e := NewTruthTableMemory(6, WordMemoryOptions{})
	if e != nil || empty.Len() != 0 || empty.PageCount() != 0 || !empty.HashEnabled() {
		t.Fatal(e)
	}
	if _, e = empty.AddMuxTruthTable(6); !errors.Is(e, ErrInvariant) || empty.Len() != 0 {
		t.Fatal(e)
	}
	for _, variables := range []int{-1, 37} {
		_, e := TruthTableWordCount(variables)
		want := ErrDimension
		if variables > 0 {
			want = ErrLimit
		}
		if !errors.Is(e, want) {
			t.Fatal(e)
		}
	}
	if _, e := NewTruthTableMemory(15, WordMemoryOptions{}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e := NewSeededTruthTableMemory(3, false, WordMemoryOptions{MaxEntries: 1}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	m, _ := NewSeededTruthTableMemory(3, false, WordMemoryOptions{MaxEntries: 2})
	before, _ := m.HexDump(1024)
	if _, e = m.AddMuxTruthTable(3); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	after, _ := m.HexDump(1024)
	if after != before || m.Len() != 2 {
		t.Fatal("limit mutation")
	}
	conflicting, _ := NewSeededTruthTableMemory(3, false, WordMemoryOptions{})
	conflicting.Intern([]uint64{123})
	before, _ = conflicting.HexDump(1024)
	if _, e = conflicting.AddMuxTruthTable(3); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	after, _ = conflicting.HexDump(1024)
	if before != after {
		t.Fatal("conflict mutation")
	}
	if _, e = conflicting.AddMuxTruthTable(7); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	conflicting.DisableHash()
	if _, e = conflicting.AddMuxTruthTable(3); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	var nilMemory *WordMemory
	if _, e = nilMemory.AddMuxTruthTable(3); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	width, _ := TruthTableWordCount(36)
	if width != 1<<30 {
		t.Fatal(width)
	}
}
func FuzzTruthMemory(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(128), []byte{0, 255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		variables := int(flags % 10)
		m, e := NewSeededTruthTableMemory(variables, flags&128 != 0, WordMemoryOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.AddMuxTruthTable(variables); e != nil {
			t.Fatal(e)
		}
		model := [][]uint64{}
		for i := 0; i < 3; i++ {
			entry, _ := m.Read(i)
			model = append(model, entry)
		}
		for i, v := range data {
			entry := make([]uint64, m.EntryWords())
			for k := range entry {
				entry[k] = uint64(v) + uint64(k)*0x0101010101010101
				if i&1 != 0 {
					entry[k] = ^entry[k]
				}
			}
			want := -1
			for k, prior := range model {
				if slices.Equal(prior, entry) {
					want = k
					break
				}
			}
			id, exists, e := m.Intern(entry)
			if e != nil || exists != (want >= 0) {
				t.Fatal(e)
			}
			if want < 0 {
				want = len(model)
				model = append(model, slices.Clone(entry))
			}
			if id != want {
				t.Fatal("stable ID")
			}
			repeat, found, e := m.Intern(entry)
			if e != nil || !found || repeat != id {
				t.Fatal("repeat")
			}
			if mux, e := m.AddMuxTruthTable(variables); e != nil || mux != 2 {
				t.Fatal("mux reuse", mux, e)
			}
		}
		if m.Len() != len(model) {
			t.Fatal("size")
		}
		for i, want := range model {
			got, e := m.Read(i)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(i, e)
			}
		}
		if e = m.Check(); e != nil {
			t.Fatal(e)
		}
	})
}
