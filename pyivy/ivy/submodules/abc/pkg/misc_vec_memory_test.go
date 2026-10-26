package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestWordMemoryNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/vec_memory.json")
	if e != nil {
		t.Fatal(e)
	}
	var row struct {
		IDs, ExtraIDs                    []int
		Pages, BoundaryPages, EmptyPages int
		Entries                          [][]uint64
	}
	if e = json.Unmarshal(data, &row); e != nil {
		t.Fatal(e)
	}
	m, e := NewWordMemory(2, 2, WordMemoryOptions{MaxEntries: 64, MaxWords: 128})
	if e != nil {
		t.Fatal(e)
	}
	m.EnableHash()
	for i, n := range []int{0, 1, 2, 1, 3, 0, 4, 5, 2} {
		word := uint64(n) * 0x0102030405060708
		id, _, e := m.Intern([]uint64{word, ^word})
		if e != nil || id != row.IDs[i] {
			t.Fatal(i, id, e)
		}
	}
	for i := 6; i < 32; i++ {
		value := i
		if i == 31 {
			value = 2
		}
		word := uint64(value) * 0x0102030405060708
		id, found, e := m.Intern([]uint64{word, ^word})
		if e != nil || id != row.ExtraIDs[i-6] || found != (i == 31) {
			t.Fatal("resize IDs", i, id, found, e)
		}
	}
	if m.Len() != len(row.Entries) || m.PageCount() != row.Pages {
		t.Fatal(m.Len(), m.PageCount())
	}
	for i, want := range row.Entries {
		got, e := m.Read(i)
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(i, got, want, e)
		}
	}
	m.Shrink(4)
	if m.PageCount() != 1 || row.BoundaryPages != 2 {
		t.Fatal("C boundary correction")
	}
	m.Shrink(0)
	if m.PageCount() != 0 || row.EmptyPages != 1 {
		t.Fatal("C empty correction")
	}
	if e = m.Check(); e != nil {
		t.Fatal(e)
	}
}
func TestWordMemoryOwnershipBoundsAndDump(t *testing.T) {
	m, _ := NewWordMemory(2, 1, WordMemoryOptions{MaxEntries: 5, MaxWords: 12})
	source := []uint64{0x1234, 0xfedc}
	m.Push(source)
	source[0] = 0
	read, _ := m.Read(0)
	if read[0] != 0x1234 {
		t.Fatal("borrowed input")
	}
	read[0] = 0
	read, _ = m.Read(0)
	if read[0] != 0x1234 {
		t.Fatal("borrowed read")
	}
	m.Set(3, []uint64{7, 8})
	for _, id := range []int{1, 2} {
		words, _ := m.Read(id)
		if !slices.Equal(words, []uint64{0, 0}) {
			t.Fatal("gap", id, words)
		}
	}
	dump, e := m.HexDump(200)
	if e != nil || !strings.HasPrefix(dump, "000000000000FEDC0000000000001234\n") || len(dump) != 132 {
		t.Fatal(dump, e)
	}
	if _, e = m.HexDump(131); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	clone := m.Clone()
	clone.Write(0, []uint64{99, 99})
	got, _ := m.Read(0)
	if got[0] != 0x1234 {
		t.Fatal("clone alias")
	}
	if e = m.EnableHash(); !errors.Is(e, ErrInvariant) {
		t.Fatal("duplicates must reject", e)
	}
	m.Write(2, []uint64{4, 5})
	if e = m.EnableHash(); e != nil {
		t.Fatal(e)
	}
	clone = m.Clone()
	clone.Shrink(1)
	if m.Len() != 4 || !m.HashEnabled() {
		t.Fatal("hash clone")
	}
	for _, op := range []func() error{func() error { return m.Write(0, []uint64{1, 2}) }, func() error { return m.Set(4, []uint64{1, 2}) }, func() error { _, e := m.Get(4); return e }} {
		if e := op(); !errors.Is(e, ErrInvariant) {
			t.Fatal(e)
		}
	}
	id, found, e := m.Intern([]uint64{0x1234, 0xfedc})
	if e != nil || !found || id != 0 {
		t.Fatal(id, found, e)
	}
	m.Intern([]uint64{10, 11})
	before := m.Clone()
	if _, _, e = m.Intern([]uint64{12, 13}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if m.Len() != before.Len() || m.PageCount() != before.PageCount() {
		t.Fatal("limit mutation")
	}
	m.Shrink(4)
	id, found, e = m.Intern([]uint64{10, 11})
	if e != nil || found || id != 4 {
		t.Fatal("truncated hash key", id, found, e)
	}
	m.DisableHash()
	if e = m.Set(0, []uint64{2}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e = m.Read(-1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if e = m.Shrink(6); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	small, _ := NewWordMemory(2, 1, WordMemoryOptions{MaxEntries: 64, MaxWords: 4})
	small.Push([]uint64{1, 2})
	if e = small.Set(2, []uint64{3, 4}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if small.Len() != 1 || small.PageCount() != 1 {
		t.Fatal("page limit mutation")
	}
	for _, args := range [][2]int{{0, 1}, {1, -1}, {1, 31}} {
		if _, e := NewWordMemory(args[0], args[1], WordMemoryOptions{}); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if _, e := NewWordMemory(2, 2, WordMemoryOptions{MaxWords: 7}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e := NewWordMemory(1, 0, WordMemoryOptions{MaxEntries: -1}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	var zero WordMemory
	if _, e = zero.Read(0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if e = zero.EnableHash(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}

func TestWordMemoryNativeHashAndTruthDump(t *testing.T) {
	words := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	got, err := WordMemoryNativeHashKey(words, 1009)
	if err != nil {
		t.Fatal(err)
	}
	primes := []uint32{1699, 4177, 5147, 5647}
	parts := []uint32{0x89abcdef, 0x01234567, 0x76543210, 0xfedcba98}
	var sum uint32
	for i := range parts {
		sum += parts[i] * primes[i]
	}
	if got != sum%1009 {
		t.Fatal(got, sum%1009)
	}
	for _, invalid := range []struct {
		words []uint64
		size  int
	}{{nil, 1}, {words, 0}, {words, -1}} {
		if _, err = WordMemoryNativeHashKey(invalid.words, invalid.size); !errors.Is(err, ErrDimension) {
			t.Fatal(invalid, err)
		}
	}
	m, err := NewSeededTruthTableMemory(3, false, WordMemoryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	dump, err := m.TruthTableDump("demo", 3, 1024)
	if err != nil || dump.FileName != "tt_demo_03.txt" || dump.Text != "0000000000000000\nAAAAAAAAAAAAAAAA\n" || dump.Summary != "Dumped 2 3-var truth tables into file \"tt_demo_03.txt\" (0.00 MB).\n" {
		t.Fatal(dump, err)
	}
	stdout, err := m.TruthTableDump("", 3, 1024)
	if err != nil || stdout.FileName != "stdout" || stdout.Text != dump.Text || stdout.Summary != "Dumped 2 3-var truth tables into file \"stdout\" (0.00 MB).\n" {
		t.Fatal(stdout, err)
	}
	if dump, err = m.TruthTableDump("x", -1, 1024); !errors.Is(err, ErrDimension) || dump.Text != "" {
		t.Fatal(dump, err)
	}
	if dump, err = m.TruthTableDump("x", 3, 1); !errors.Is(err, ErrLimit) || dump.Text != "" {
		t.Fatal(dump, err)
	}
}

func FuzzWordMemoryNativeHashKey(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8}, uint16(17))
	f.Fuzz(func(t *testing.T, data []byte, modulus uint16) {
		if len(data) == 0 {
			data = []byte{0}
		}
		if len(data) > 128 {
			data = data[:128]
		}
		words := make([]uint64, (len(data)+7)/8)
		for i, value := range data {
			words[i/8] |= uint64(value) << uint(8*(i%8))
		}
		tableSize := int(modulus) + 1
		got, err := WordMemoryNativeHashKey(words, tableSize)
		if err != nil {
			t.Fatal(err)
		}
		primes := [...]uint32{1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147}
		var want uint32
		for i, word := range words {
			want += uint32(word) * primes[(2*i)&7]
			want += uint32(word>>32) * primes[(2*i+1)&7]
		}
		if got != want%uint32(tableSize) {
			t.Fatal(got, want, tableSize)
		}
	})
}
func FuzzWordMemory(f *testing.F) {
	f.Add([]byte{0, 1, 1, 7, 2, 3, 3, 2, 4, 1})
	f.Add([]byte{5, 1, 6, 3, 7, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		log := 0
		if len(data) > 0 {
			log = int(data[0] % 4)
		}
		m, _ := NewWordMemory(2, log, WordMemoryOptions{MaxEntries: 64, MaxWords: 128})
		model := [][]uint64{}
		hashed := false
		for i := 0; i+1 < len(data); i += 2 {
			op, arg := data[i]%9, int(data[i+1]%32)
			entry := []uint64{uint64(data[i+1]), uint64(data[i])}
			switch op {
			case 0:
				if !hashed {
					if e := m.Set(arg, entry); e != nil {
						t.Fatal(e)
					}
					for len(model) <= arg {
						model = append(model, []uint64{0, 0})
					}
					model[arg] = slices.Clone(entry)
				}
			case 1:
				if !hashed {
					got, e := m.Get(arg)
					if e != nil {
						t.Fatal(e)
					}
					for len(model) <= arg {
						model = append(model, []uint64{0, 0})
					}
					if !slices.Equal(got, model[arg]) {
						t.Fatal("get")
					}
				}
			case 2:
				if len(model) > 0 {
					n := arg % (len(model) + 1)
					m.Shrink(n)
					model = model[:n]
				}
			case 3:
				if !hashed {
					unique := true
					for a := range model {
						for b := 0; b < a; b++ {
							if slices.Equal(model[a], model[b]) {
								unique = false
							}
						}
					}
					e := m.EnableHash()
					if unique {
						if e != nil {
							t.Fatal(e)
						}
						hashed = true
					} else if !errors.Is(e, ErrInvariant) {
						t.Fatal(e)
					}
				}
			case 4:
				m.DisableHash()
				hashed = false
			case 5:
				if hashed {
					want := -1
					for j, w := range model {
						if slices.Equal(w, entry) {
							want = j
							break
						}
					}
					id, found, e := m.Intern(entry)
					if want < 0 && len(model) == 64 {
						if !errors.Is(e, ErrLimit) {
							t.Fatal(e)
						}
					} else {
						if e != nil || found != (want >= 0) {
							t.Fatal(e)
						}
						if want < 0 {
							want = len(model)
							model = append(model, slices.Clone(entry))
						}
						if id != want {
							t.Fatal("ID")
						}
					}
				}
			case 6:
				clone := m.Clone()
				if e := clone.Check(); e != nil {
					t.Fatal(e)
				}
				clone.Shrink(0)
			case 7:
				if !hashed && len(model) < 64 {
					if e := m.Push(entry); e != nil {
						t.Fatal(e)
					}
					model = append(model, slices.Clone(entry))
				}
			case 8:
				if !hashed && len(model) > 0 {
					j := arg % len(model)
					m.Write(j, entry)
					model[j] = slices.Clone(entry)
				}
			}
			if e := m.Check(); e != nil {
				t.Fatal(e)
			}
			if m.Len() != len(model) || m.HashEnabled() != hashed {
				t.Fatal("dimensions")
			}
			for j, want := range model {
				got, e := m.Read(j)
				if e != nil || !slices.Equal(got, want) {
					t.Fatal(j, got, want, e)
				}
			}
		}
	})
}
