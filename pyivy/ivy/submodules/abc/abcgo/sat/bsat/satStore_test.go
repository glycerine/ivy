package bsat

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestStoredClauseRoundTrip(t *testing.T) {
	p := Sto_ManAlloc()
	input := []lit{4, 0}
	if Sto_ManAddClause(p, input) != 1 || !reflect.DeepEqual(input, []lit{4, 0}) {
		t.Fatal("input mutation")
	}
	Sto_ManMarkClausesA(p)
	Sto_ManAddClause(p, []lit{3})
	Sto_ManMarkRoots(p)
	Sto_ManAddClause(p, nil)
	var output bytes.Buffer
	if err := WriteClauseStore(&output, p); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadClauseStore(strings.NewReader(output.String()))
	if err != nil || loaded.nClauses != 3 || loaded.nVars != 3 || loaded.nRoots != 2 || loaded.nClausesA != 1 {
		t.Fatal(loaded, err, output.String())
	}
	if loaded.pHead.Id != 0 || loaded.pHead.fA != 1 || loaded.pHead.pNext.fRoot != 1 || loaded.pTail.nLits != 0 || loaded.pEmpty != loaded.pTail {
		t.Fatal("loaded metadata")
	}
	var again bytes.Buffer
	WriteClauseStore(&again, loaded)
	if again.String() != output.String() {
		t.Fatal("round-trip changed")
	}
}
func TestStoredClauseValidation(t *testing.T) {
	p := Sto_ManAlloc()
	for _, c := range [][]lit{{0, 0}, {0, 1}, {-2}} {
		if Sto_ManAddClause(p, c) != 0 || p.nClauses != 0 || p.pHead != nil {
			t.Fatal("invalid clause mutated store")
		}
	}
	if Sto_ManAddClause(p, nil) != 1 || Sto_ManAddClause(p, nil) != 0 || p.nClauses != 1 {
		t.Fatal("duplicate empty clause")
	}
	for _, input := range []string{"", "p 1 2 0 0\n1 0", "p 1 1 0 0\n2 0", "p 1 1 0 0\n1", "p 1 1 2 0\n1 0", "p 1 1 0 0\n1 1 0", "p 1 1 0 0\n1 -1 0", "p 1 1 0 0\n2147483647 0"} {
		if _, err := ReadClauseStore(strings.NewReader(input)); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
}
func TestStoreNumbersAtEOFAndFileOffsets(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "numbers"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.WriteString("12 -34\n c comment\n0 57")
	file.Seek(0, 0)
	for _, want := range []int32{12, -34, 0, 57} {
		var got int32
		if Sto_ManLoadNumber(file, &got) != 1 || got != want {
			t.Fatal(got, want)
		}
	}
	var got int32
	if Sto_ManLoadNumber(file, &got) != 0 {
		t.Fatal("EOF")
	}
}
func TestStoredClauseMemoryAndLastA(t *testing.T) {
	p := Sto_ManAlloc()
	a := Sto_ManMemoryFetch(p, 32)
	a[0] = 7
	Sto_ManMemoryFetch(p, p.nChunkSize)
	if a[0] != 7 || len(p.chunks) != 2 {
		t.Fatal("chunk ownership")
	}
	Sto_ManAddClause(p, []lit{6})
	Sto_ManMarkClausesA(p)
	if Sto_ManChangeLastClause(p) != 3 || p.nClausesA != 0 || p.pTail.fA != 0 {
		t.Fatal("last A clause")
	}
	if Sto_ManMemoryReport(p) < p.nChunkSize {
		t.Fatal("memory report")
	}
	Sto_ManMemoryStop(p)
	if p.pHead != nil || p.pChunkLast != nil || p.nClauses != 0 {
		t.Fatal("release")
	}
}
