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

func TestBookAndListCommands(t *testing.T) {
	formats.Lock()
	old := formats.handlers[IO_FILE_BOOK]
	formats.Unlock()
	defer func() { formats.Lock(); formats.handlers[IO_FILE_BOOK] = old; formats.Unlock() }()
	called := false
	RegisterFormat(IO_FILE_BOOK, nil, func(n *abc.Network, name string) error { called = n != nil && name == "out.aux"; return nil })
	frame := abcmain.NewFrame()
	var out bytes.Buffer
	frame.Out = &out
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	pi := n.NewObject(abc.ObjectPI)
	pi.SetName("input")
	frame.ReplaceCurrentNetwork(n)
	if IoCommandWriteBook(frame, []string{"write_book", "out.aux"}) != 0 || !called {
		t.Fatal("book writer not called")
	}
	file := filepath.Join(t.TempDir(), "graph.txt")
	if IoCommandWriteList(frame, []string{"write_list", file}) != 0 || !strings.Contains(out.String(), "does not work") {
		t.Fatal("ABC disabled behavior lost")
	}
	if ioCommandWriteList(frame, []string{"write_list", file}) != 0 {
		t.Fatal("list export failed")
	}
	data, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(data), "HOST") {
		t.Fatalf("list output %q: %v", data, err)
	}
}
