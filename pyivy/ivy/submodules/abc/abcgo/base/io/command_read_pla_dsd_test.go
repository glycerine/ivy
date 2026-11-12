package io

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestReadDsdAndPlaCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var errors bytes.Buffer
	frame.Err = &errors
	if IoCommandReadDsd(frame, []string{"read_dsd", "(a*b)"}) != 0 || frame.Network() == nil {
		t.Fatalf("DSD read failed: %s", errors.String())
	}
	file := filepath.Join(t.TempDir(), "input.pla")
	if err := os.WriteFile(file, []byte(".i 2\n.o 1\n.p 1\n11 1\n.e\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, options := range [][]string{nil, {"-b"}, {"-x"}} {
		args := append([]string{"read_pla"}, options...)
		args = append(args, file)
		if IoCommandReadPla(frame, args) != 0 {
			t.Fatalf("PLA options %v: %s", options, errors.String())
		}
		n, ok := frame.Network().(*abc.Network)
		if !ok || abc.Abc_NtkPiNum(n) != 2 || abc.Abc_NtkPoNum(n) != 1 {
			t.Fatalf("bad PLA network with %v", options)
		}
	}
}
