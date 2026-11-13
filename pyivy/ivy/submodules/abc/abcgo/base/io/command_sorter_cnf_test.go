package io

import (
	"bytes"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestSorterCNFCommand(t *testing.T) {
	sorterCNFWriter.RLock()
	old := sorterCNFWriter.fn
	sorterCNFWriter.RUnlock()
	defer RegisterSorterCNFWriter(old)
	called := false
	RegisterSorterCNFWriter(func(path string, vars, queens int32) { called = path == "out.cnf" && vars == 8 && queens == 3 })
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	if IoCommandWriteSortCnf(frame, []string{"write_sorter_cnf", "-N", "8", "-Q", "3", "out.cnf"}) != 0 || !called {
		t.Fatalf("dispatch failed: %s", errors.String())
	}
	if IoCommandWriteSortCnf(frame, []string{"write_sorter_cnf", "-N", "3", "-Q", "3", "out.cnf"}) != 1 {
		t.Fatal("accepted invalid bounds")
	}
}
