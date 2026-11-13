package abci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestVerificationLogSourceRoundTrip(t *testing.T) {
	cex := util.Abc_CexAlloc(2, 2, 2)
	cex.IPo = 3
	cex.IFrame = 1
	for _, bit := range []int32{0, 2, 5} {
		util.Abc_InfoSetBit(cex.PData, bit)
	}
	name := filepath.Join(t.TempDir(), "witness.log")
	Abc_NtkWriteLogFile(name, cex, 0, 4, "pdr")
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "snl_SAT 4 pdr 3 1\n10\n1001\n"; got != want {
		t.Fatalf("file=%q want=%q", got, want)
	}
	var got *util.Abc_Cex_t
	var frames int32
	if status := Abc_NtkReadLogFile(name, &got, &frames); status != 0 || frames != 4 || got == nil || got.IPo != 3 || got.IFrame != 1 || got.NRegs != 2 || got.NPis != 2 || got.NBits != 6 {
		t.Fatalf("status=%d frames=%d cex=%+v", status, frames, got)
	}
	for i := int32(0); i < cex.NBits; i++ {
		if util.Abc_InfoHasBit(cex.PData, i) != util.Abc_InfoHasBit(got.PData, i) {
			t.Fatalf("bit %d differs", i)
		}
	}
}

func TestVerificationLogSourceEmptyAndBadTrace(t *testing.T) {
	name := filepath.Join(t.TempDir(), "status.log")
	Abc_NtkWriteLogFile(name, nil, 1, 5, "engine")
	var got *util.Abc_Cex_t
	var frames int32
	if status := Abc_NtkReadLogFile(name, &got, &frames); status != 1 || frames != 5 || got == nil || got.NBits != 0 || got.IFrame != 5 || got.IPo != -1 {
		t.Fatalf("status=%d frames=%d cex=%+v", status, frames, got)
	}
	if err := os.WriteFile(name, []byte("snl_SAT 2 engine 0\n01\n1010\n"), 0600); err != nil {
		t.Fatal(err)
	}
	sentinel := got
	if status := Abc_NtkReadLogFile(name, &got, &frames); status != -1 || got != sentinel {
		t.Fatalf("invalid trace: status=%d cex=%v", status, got)
	}
}
