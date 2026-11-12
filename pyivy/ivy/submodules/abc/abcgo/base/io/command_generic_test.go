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

type ioCommandRecorder struct{ command string }

func (r *ioCommandRecorder) Execute(_ *abcmain.Frame, command string) int32 {
	r.command = command
	return 0
}

func TestGenericReadWriteCommands(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.blif")
	if err := os.WriteFile(input, []byte(".model top\n.inputs a b\n.outputs y\n.names a b y\n11 1\n.end\n"), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.NewFrame()
	var out, errors bytes.Buffer
	frame.Out, frame.Err = &out, &errors
	if got := IoCommandRead(frame, []string{"read", input}); got != 0 {
		t.Fatalf("read returned %d", got)
	}
	if n, ok := frame.Network().(*abc.Network); !ok || n == nil || abc.Abc_NtkPiNum(n) != 2 {
		t.Fatalf("read network = %v, errors %s", n, errors.String())
	}
	output := filepath.Join(dir, "out.blif")
	if got := IoCommandWrite(frame, []string{"write", output}); got != 0 {
		t.Fatalf("write returned %d", got)
	}
	data, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(data), ".model") {
		t.Fatalf("output %q: %v", data, err)
	}
	if IoCommandRead(frame, []string{"read", "-h"}) != 1 || !strings.Contains(errors.String(), "usage: read") {
		t.Fatal("missing read usage")
	}
}

func TestGenericReadWriteLibraryDispatch(t *testing.T) {
	frame := abcmain.NewFrame()
	recorder := &ioCommandRecorder{}
	frame.SetDispatcher(recorder)
	if IoCommandRead(frame, []string{"read", "library.genlib"}) != 0 || recorder.command != "read_genlib library.genlib" {
		t.Fatal("read library dispatch")
	}
	if IoCommandWrite(frame, []string{"write", "library.lib"}) != 0 || recorder.command != "write_liberty library.lib" {
		t.Fatal("write library dispatch")
	}
}
