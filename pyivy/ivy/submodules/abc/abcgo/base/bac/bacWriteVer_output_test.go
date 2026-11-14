package bac

import (
	"os"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// This expected byte string follows bacWriteVer.c's Psr_ManWriteVerilogArray2
// formatting for named formal/actual pairs.
func TestPsrWriteVerilogArray2MatchesCFormatting(t *testing.T) {
	names := util.Abc_NamStart(16, 16)
	p := &Psr_Ntk_t{pStrs: names}
	formalA := util.Abc_NamStrFindOrAdd(names, "A", nil)
	actualA := util.Abc_NamStrFindOrAdd(names, "alpha", nil)
	formalB := util.Abc_NamStrFindOrAdd(names, "B", nil)
	actualB := util.Abc_NamStrFindOrAdd(names, "beta", nil)
	vSigs := vec.Vec_IntAllocArrayCopy([]int32{formalA, actualA << 2, formalB, actualB << 2}, 4)

	f, err := os.CreateTemp(t.TempDir(), "verilog-array-*.v")
	if err != nil {
		t.Fatal(err)
	}
	Psr_ManWriteVerilogArray2(f, p, vSigs)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	const want = ".A(alpha), .B(beta)"
	if string(got) != want {
		t.Fatalf("writer output %q, want C-compatible output %q", got, want)
	}
}
