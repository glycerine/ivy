package scl

import (
	"github.com/glycerine/abcgo/map/mio"
	"strings"
	"testing"
)

func TestGenlibProfileSkipsUnselectedTiming(t *testing.T) {
	text := `GATE INV 1 Y=!A;
PIN A INV 1 999 1 0 1 0
GATE NAND 2 Y=!(A*B);
PIN A INV 1 999 1 0 1 0
PIN B INV 1 999 1 0 1 0
.end
`
	profile, e := mio.ReadLibrary("test", text, nil)
	if e != nil {
		t.Fatal(e)
	}
	lib, e := Abc_SclReadFromGenlib(profile)
	if e != nil {
		t.Fatal(e)
	}
	nand := lib.Cells[Abc_SclCellFind(lib, "NAND")]
	nand.Pins[nand.Inputs].Timings = nil
	result, e := Abc_SclDeriveGenlib(lib, profile, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if mio.Mio_LibraryReadGateByName(result, "NAND", "") != nil {
		t.Fatal("unprofiled NAND retained")
	}
	count := 0
	output, e := Abc_SclProduceGenlibStrProfile(lib, profile, 0, 1, 0, &count)
	if e != nil || count != 3 || strings.Contains(output, "GATE NAND") {
		t.Fatal(count, output, e)
	}
}
