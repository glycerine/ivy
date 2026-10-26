package abc

// Stand-alone tests for the owned clause/proof store from satStore.c.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"
)

func satStoreTestLit(t testing.TB, variable int, negative bool) SatLit {
	t.Helper()
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func TestSatProofStoreAddMarkChangeAndOwnership(t *testing.T) {
	store, err := NewSatProofStore(SatProofStoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	input := []SatLit{satStoreTestLit(t, 2, false), satStoreTestLit(t, 0, true), satStoreTestLit(t, 1, false)}
	original := slices.Clone(input)
	first, err := store.AddClause(input)
	if err != nil || first != 0 || !slices.Equal(input, original) {
		t.Fatal(first, input, original, err)
	}
	second, _ := store.AddClause([]SatLit{satStoreTestLit(t, 4, true)})
	if err = store.MarkRoots(); err != nil {
		t.Fatal(err)
	}
	third, _ := store.AddClause([]SatLit{satStoreTestLit(t, 3, false), satStoreTestLit(t, 5, true)})
	if err = store.MarkClausesA(); err != nil {
		t.Fatal(err)
	}
	fourth, _ := store.AddClause([]SatLit{satStoreTestLit(t, 6, true)})
	if err = store.MarkClausesA(); err != nil {
		t.Fatal(err)
	}
	variable, err := store.ChangeLastClause()
	if err != nil || variable != 6 {
		t.Fatal(variable, err)
	}
	empty, err := store.AddClause(nil)
	if err != nil || empty != 4 {
		t.Fatal(empty, err)
	}
	if err = store.SetVisited(third, true); err != nil {
		t.Fatal(err)
	}
	clauses := store.Clauses()
	if len(clauses) != 5 || clauses[first].ID != first || clauses[second].ID != second || clauses[fourth].PartA || !clauses[third].PartA || !clauses[third].Visited || !clauses[first].Root || !clauses[second].Root || clauses[third].Root {
		t.Fatal(clauses)
	}
	wantFirst := []SatLit{satStoreTestLit(t, 0, true), satStoreTestLit(t, 1, false), satStoreTestLit(t, 2, false)}
	if !slices.Equal(clauses[first].Literals, wantFirst) || len(store.RootClauses()) != 2 {
		t.Fatal(clauses[first], store.RootClauses())
	}
	clauses[first].Literals[0] = 99
	got, _ := store.Clause(first)
	if got.Literals[0] != wantFirst[0] {
		t.Fatal("clause snapshot aliases store")
	}
	stats := store.Stats()
	if stats.Variables != 7 || stats.Roots != 2 || stats.Clauses != 5 || stats.ClausesA != 3 || stats.Literals != 7 || stats.EmptyClause != empty || stats.Chunks != 1 || stats.MemoryBytes == 0 {
		t.Fatal(stats)
	}
	before := store.Stats()
	if _, err = store.AddClause(nil); !errors.Is(err, ErrInvariant) || store.Stats() != before {
		t.Fatal(err, store.Stats(), before)
	}
	duplicateVariable := []SatLit{satStoreTestLit(t, 1, false), satStoreTestLit(t, 1, true)}
	if _, err = store.AddClause(duplicateVariable); !errors.Is(err, ErrInvariant) || store.Stats() != before {
		t.Fatal(err, store.Stats(), before)
	}
}

func TestSatProofStoreChunkAccounting(t *testing.T) {
	store, _ := NewSatProofStore(SatProofStoreLimits{MaxVariables: 2000, MaxClauses: 100, MaxLiterals: 100000, MaxBytes: 1 << 20})
	clause := make([]SatLit, 1000)
	for i := range clause {
		clause[i] = satStoreTestLit(t, i, false)
	}
	for i := 0; i < 17; i++ {
		if _, err := store.AddClause(clause); err != nil {
			t.Fatal(i, err)
		}
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	recordBytes := (3*pointerBytes + 8 + 4*len(clause) + pointerBytes - 1) / pointerBytes * pointerBytes
	wantLast := pointerBytes + recordBytes
	wantMemory := satProofStoreChunkBytes + wantLast
	stats := store.Stats()
	if stats.Chunks != 2 || stats.LastChunkBytes != wantLast || stats.MemoryBytes != wantMemory {
		t.Fatal(stats, wantLast, wantMemory)
	}
}

func TestSatProofStoreMarshalParseAndFile(t *testing.T) {
	store, _ := NewSatProofStore(SatProofStoreLimits{})
	_, _ = store.AddClause([]SatLit{satStoreTestLit(t, 2, false), satStoreTestLit(t, 0, true)})
	store.MarkClausesA()
	_, _ = store.AddClause([]SatLit{satStoreTestLit(t, 1, false)})
	store.MarkRoots()
	_, _ = store.AddClause([]SatLit{satStoreTestLit(t, 3, true), satStoreTestLit(t, 2, true)})
	data, err := store.MarshalText()
	want := "p 4 3 2 1\n -1 3 0\n 2 0\n -3 -4 0\n"
	if err != nil || string(data) != want {
		t.Fatalf("got %q want %q err=%v", data, want, err)
	}
	parsed, err := ParseSatProofStore(append([]byte("c generated in Go\n\n"), data...), SatProofStoreLimits{})
	if err != nil || !reflect.DeepEqual(parsed.Clauses(), store.Clauses()) || parsed.Stats().Variables != 4 || parsed.Stats().Roots != 2 || parsed.Stats().ClausesA != 1 {
		t.Fatal(parsed, err)
	}
	name := filepath.Join(t.TempDir(), "store.cnf")
	if err = store.WriteFile(name); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadSatProofStore(name, SatProofStoreLimits{})
	if err != nil || !reflect.DeepEqual(loaded.Clauses(), store.Clauses()) {
		t.Fatal(loaded, err)
	}
}

func TestSatProofStoreErrorsLimitsAndLifecycle(t *testing.T) {
	if store, err := NewSatProofStore(SatProofStoreLimits{MaxBytes: -1}); store != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(store, err)
	}
	store, _ := NewSatProofStore(SatProofStoreLimits{MaxVariables: 2, MaxClauses: 1, MaxLiterals: 2, MaxBytes: 100})
	if _, err := store.AddClause([]SatLit{satStoreTestLit(t, 2, false)}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := store.AddClause([]SatLit{satStoreTestLit(t, 0, false)}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddClause(nil); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := store.ChangeLastClause(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := store.SetVisited(9, true); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err := store.WriteFile(""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	badInputs := []string{
		"",
		"p 2 1 2 0\n 1 0\n",
		"p 2 2 0 0\n 1 0\n",
		"p 2 1 0 0\n 1\n",
		"p 1 1 0 0\n 2 0\n",
		"p 2 1 0 0\n 1 -1 0\n",
		"p 2 2 0 0\n 0\n 0\n",
		"p x 0 0 0\n",
		"p 1 1 0 0\n c 0\n",
	}
	for _, input := range badInputs {
		if parsed, err := ParseSatProofStore([]byte(input), SatProofStoreLimits{}); err == nil {
			t.Fatal("accepted", input, parsed)
		}
	}
	store.Stop()
	store.Stop()
	if _, err := store.AddClause(nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := LoadSatProofStore(filepath.Join(t.TempDir(), "missing"), SatProofStoreLimits{}); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func FuzzSatProofStoreRoundTrip(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
		}
		clauseCount := len(data)
		aCount, rootCount := clauseCount/3, 2*clauseCount/3
		store, _ := NewSatProofStore(SatProofStoreLimits{MaxVariables: 512, MaxClauses: 128, MaxLiterals: 1024, MaxBytes: 1 << 20})
		for i, value := range data {
			size := int(value%7) + 1
			clause := make([]SatLit, size)
			for j := range clause {
				variable := (i*11 + j) % 511
				clause[j] = satStoreTestLit(t, variable, value>>uint(j)&1 != 0)
			}
			if _, err := store.AddClause(clause); err != nil {
				t.Fatal(i, err)
			}
			if i+1 == aCount {
				store.MarkClausesA()
			}
			if i+1 == rootCount {
				store.MarkRoots()
			}
		}
		if aCount == 0 {
			store.clausesA = 0
		}
		if rootCount == 0 {
			store.roots = 0
		}
		encoded, err := store.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseSatProofStore(encoded, SatProofStoreLimits{MaxVariables: 512, MaxClauses: 128, MaxLiterals: 1024, MaxBytes: 1 << 20})
		if err != nil || !reflect.DeepEqual(parsed.Clauses(), store.Clauses()) {
			t.Fatal(string(encoded), parsed, err)
		}
		left, right := parsed.Stats(), store.Stats()
		if left.Variables != right.Variables || left.Roots != right.Roots || left.Clauses != right.Clauses || left.ClausesA != right.ClausesA || left.Literals != right.Literals {
			t.Fatal(left, right)
		}
	})
}

func TestSatProofStoreHeaderWhitespace(t *testing.T) {
	parsed, err := ParseSatProofStore([]byte(strings.Join([]string{" c", "", "p 3 1 1 1", " 1 -3 0", ""}, "\n")), SatProofStoreLimits{})
	if err != nil || parsed.Stats().Variables != 3 || len(parsed.RootClauses()) != 1 {
		t.Fatal(parsed, err)
	}
}
