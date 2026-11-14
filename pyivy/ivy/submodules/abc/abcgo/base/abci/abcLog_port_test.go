package abci

import (
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"path/filepath"
	"testing"
)

func TestAbcLogOutputAndReadRoundTrip(t *testing.T) {
	name := filepath.Join(t.TempDir(), "verify.log")
	cex := util.Abc_CexAlloc(1, 2, 2)
	cex.IPo, cex.IFrame = 3, 1
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 2)
	Abc_NtkWriteLogFile(name, cex, 0, 1, "bmc")
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if want := "snl_SAT 1 bmc 3\n1\n0100\n"; string(data) != want {
		t.Fatalf("got %q, want %q", data, want)
	}
	var got *util.Abc_Cex_t
	var frames int32
	if status := Abc_NtkReadLogFile(name, &got, &frames); status != 0 {
		t.Fatalf("status %d", status)
	}
	if frames != 1 || got == nil || got.NRegs != 1 || got.NPis != 2 || got.NBits != 5 || got.IPo != 3 || got.IFrame != 1 {
		t.Fatalf("unexpected parsed cex: frames=%d cex=%+v", frames, got)
	}
	for i, want := range []bool{true, false, true, false, false} {
		if have := util.Abc_InfoHasBit(got.PData, int32(i)); have != want {
			t.Errorf("bit %d=%t, want %t", i, have, want)
		}
	}
}

func TestAbcLogNilCommandUsesSourceFallback(t *testing.T) {
	name := filepath.Join(t.TempDir(), "unknown.log")
	Abc_NtkWriteLogFile(name, nil, 1, 0, nil)
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if want := "snl_UNSAT 0 unknown\nNULL\nNULL\n"; string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
