package fpga

import (
	"bytes"
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryCommands(t *testing.T) {
	f := abcmain.NewFrame()
	var out, errout bytes.Buffer
	f.Out = &out
	f.Err = &errout
	Fpga_Init(f)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lut")
	if e := os.WriteFile(path, []byte("1 1 1\n2 2 2\n3 3 3\n"), 0600); e != nil {
		t.Fatal(e)
	}
	read, ok := f.LookupCommand("read_lut")
	if !ok {
		t.Fatal("registration")
	}
	if read.Handler(f, []string{"read_lut", path}) != 0 {
		t.Fatal(errout.String())
	}
	lib := f.Resource("pLibLut").(*Fpga_LutLib_t)
	if lib.LutMax != 3 {
		t.Fatal("loaded wrong library")
	}
	print, _ := f.LookupCommand("print_lut")
	if print.Handler(f, []string{"print_lut"}) != 0 || out.Len() == 0 {
		t.Fatal("print")
	}
	if read.Handler(f, []string{"read_lut", "-bad"}) == 0 || read.Handler(f, []string{"read_lut", filepath.Join(dir, "missing")}) == 0 {
		t.Fatal("invalid read accepted")
	}
	if f.Resource("pLibLut") != lib {
		t.Fatal("failed load destroyed current library")
	}
	if print.Handler(f, []string{"print_lut", "-h"}) == 0 {
		t.Fatal("usage")
	}
	Fpga_End(f)
	if _, ok := f.LookupCommand("read_lut"); ok || f.Resource("pLibLut") != nil {
		t.Fatal("cleanup")
	}
}
