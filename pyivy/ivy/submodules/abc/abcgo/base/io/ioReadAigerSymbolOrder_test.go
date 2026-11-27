package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// Duplicate symbol names make the name table's namesake ring depend on the
// order the names are stored. ABC stores them in symbol-table order (AIGER section order),
// so a lookup of the shared name finds the first input.
func TestReadAigerStoresSymbolsInFileOrder(t *testing.T) {
	src := []byte("aag 5 4 0 1 1\n2\n4\n6\n8\n10\n10 4 2\ni0 x\ni1 x\ni2 x\ni3 x\no0 y\n")
	for run := 0; run < 20; run++ {
		n, err := ReadAigerFromMemory(src, true)
		if err != nil {
			t.Fatal(err)
		}
		got := abc.Abc_NtkFindCi(n, "x")
		if got == nil || got != abc.Abc_NtkPi(n, 0) {
			t.Fatalf("run %d: name x resolved to object %d, want the first input %d", run, got.ID, abc.Abc_NtkPi(n, 0).ID)
		}
	}
}
