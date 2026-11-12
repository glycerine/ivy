package scl

import (
	"bytes"
	abcmain "github.com/glycerine/abcgo/base/main"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryCommandsAndLifecycle(t *testing.T) {
	f := abcmain.NewFrame()
	defer f.Close()
	f.Out = &bytes.Buffer{}
	f.Err = &bytes.Buffer{}
	Scl_Init(f)
	if f.Execute("write_lib missing") != 1 {
		t.Fatal("missing library accepted")
	}
	p, e := ParseLiberty("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	f.SetResource("pLibScl", p)
	path := filepath.Join(t.TempDir(), "test.scl")
	if Scl_CommandWriteScl(f, 2, []string{"write_scl", path}) != 0 {
		t.Fatal(f.Err)
	}
	if _, e = os.Stat(path); e != nil {
		t.Fatal(e)
	}
	if f.Execute("print_lib -s") != 0 {
		t.Fatal(f.Err)
	}
	if Scl_CommandReadScl(f, 2, []string{"read_scl", path}) != 0 {
		t.Fatal(f.Err)
	}
	if len(p.Cells) != 0 {
		t.Fatal("old library not released")
	}
	if f.Execute("read_lib -S nope") != 1 {
		t.Fatal("invalid option accepted")
	}
	Scl_End(f)
	if f.Resource("pLibScl") != nil {
		t.Fatal("end retained library")
	}
}
func TestTimingConeStoreRestore(t *testing.T) {
	p := Scl_ManAlloc(&SC_Lib{}, []int32{0}, []int32{2}, 3)
	p.Times[1] = SC_Pair{3, 4}
	p.Slews[1] = SC_Pair{5, 6}
	Scl_ConeStore(p, []int32{1})
	Scl_ConeClear(p, []int32{1})
	if p.Times[1] != (SC_Pair{}) {
		t.Fatal(p)
	}
	Scl_ConeRestore(p, []int32{1})
	if p.Times[1] != (SC_Pair{3, 4}) || p.Slews[1] != (SC_Pair{5, 6}) {
		t.Fatal(p)
	}
	p.Departures[1] = SC_Pair{2, 3}
	if Scl_ObjGetSlack(p, 1, 10) != 3 {
		t.Fatal("slack")
	}
	Scl_ManCleanTime(p)
	if p.Times[1] != (SC_Pair{}) || p.Departures[1] != (SC_Pair{}) {
		t.Fatal("clear")
	}
}

func TestReadSCLInstallsMIOAndResetClearsLegacyConstraints(t *testing.T) {
	f := abcmain.NewFrame()
	defer f.Close()
	f.Out = io.Discard
	f.Err = io.Discard
	Scl_Init(f)
	p, e := ParseLiberty("test", testLiberty)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "test.scl")
	if e = Abc_SclWriteScl(path, p); e != nil {
		t.Fatal(e)
	}
	if status := Scl_CommandReadScl(f, 2, []string{"read_scl", path}); status != 0 {
		t.Fatal(status)
	}
	if f.Resource("pLibGen") == nil {
		t.Fatal("SCL read did not install MIO")
	}
	f.SetStringValue("pDrivingCell", "inv")
	f.MaxLoad = 2
	if status := f.Execute("reset_constr"); status != 0 || f.StringValue("pDrivingCell") != "" || f.MaxLoad != 0 {
		t.Fatal(status)
	}
}

func TestNumericOptionRulesDoNotCrossTalk(t *testing.T) {
	o, e := parseSclOptions([]string{"upsize", "-G", "0", "-S", "0"}, "", "GS")
	if e != nil || o.numbers['G'] != 0 {
		t.Fatal(o, e)
	}
	if o.positive("SG") {
		t.Fatal("explicit library slew/gain zero accepted")
	}
	f := abcmain.NewFrame()
	defer f.Close()
	f.Out = io.Discard
	f.Err = io.Discard
	Scl_Init(f)
	for _, name := range []string{"topo", "unbuffer", "buffer", "upsize", "dnsize"} {
		command, ok := f.LookupCommand(name)
		if !ok || !command.ChangesNetwork {
			t.Fatal(name, command, ok)
		}
	}
	count := 0
	for _, command := range f.CommandList() {
		if command.Group == "SCL mapping" {
			count++
		}
	}
	if count != 21 {
		t.Fatal("registered SCL command count", count)
	}
}
