package io

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteHierarchyAndCellNetCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	pi := n.NewObject(abc.ObjectPI)
	pi.SetName("a")
	frame.ReplaceCurrentNetwork(n)
	dir := t.TempDir()
	cell := filepath.Join(dir, "out.cells")
	if IoCommandWriteCellNet(frame, []string{"write_cellnet", cell}) != 0 {
		t.Fatal("cellnet command failed")
	}
	if data, err := os.ReadFile(cell); err != nil || !strings.Contains(string(data), "CellNet") {
		t.Fatalf("cellnet %q: %v", data, err)
	}
	formats.Lock()
	old := formats.handlers[IO_FILE_BLIF]
	formats.Unlock()
	defer func() { formats.Lock(); formats.handlers[IO_FILE_BLIF] = old; formats.Unlock() }()
	called := false
	RegisterFormat(IO_FILE_BLIF, nil, func(got *abc.Network, filename string) error { called = got == n && filename == "out.blif"; return nil })
	if IoCommandWriteHie(frame, []string{"write_hie", "original.blif", "out.blif"}) != 0 || !called {
		t.Fatalf("hierarchy dispatch failed: %s", errors.String())
	}
}
