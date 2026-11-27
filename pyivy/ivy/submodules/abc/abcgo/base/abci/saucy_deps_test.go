package abci

import (
	"testing"

	ioPkg "github.com/glycerine/abcgo/base/io"
)

// getDependencies (abcSaucy.c) takes each output's input dependencies from
// Sim_ComputeFunSupp, the functional support that print_supp -s prints
// (equal to ABC's, see TestPrintSymmLikeNative). The port used a separate
// graph-based support computation that misses dependencies (i10: 4807
// support entries where ABC finds 5438).
func TestSaucyDependenciesAreFunctionalSupport(t *testing.T) {
	ntk := ioPkg.Io_Read("test_vectors/i10_native_demo.aig", ioPkg.IO_FILE_AIGER, 1, 0)
	if ntk == nil {
		t.Fatal("cannot read i10")
	}
	ntk = Abc_NtkStrash(ntk, 0, 1, 0)
	nPis, nPos := Abc_NtkPiNum(ntk), Abc_NtkPoNum(ntk)
	iDep, oDep := make([]*Vec_Int_t, nPis), make([]*Vec_Int_t, nPos)
	for i := range iDep {
		iDep[i] = Vec_IntAlloc(1)
	}
	for i := range oDep {
		oDep[i] = Vec_IntAlloc(1)
	}
	saucyGetDependencies(ntk, iDep, oDep)
	supports := Sim_ComputeFunSupp(ntk, 0)
	total := 0
	for o := int32(0); o < nPos; o++ {
		bits := Vec_PtrEntry(supports, o).([]uint32)
		var want []int32
		for i := int32(0); i < nPis; i++ {
			if symHasBit(bits, i) {
				want = append(want, i)
			}
		}
		total += len(want)
		got := Vec_IntArray(oDep[o])
		if len(got) != len(want) {
			t.Fatalf("output %d: %d dependencies, functional support has %d", o, len(got), len(want))
		}
		for k := range want {
			if got[k] != want[k] {
				t.Fatalf("output %d: dependencies %v, support %v", o, got, want)
			}
		}
	}
	if total != 5438 {
		t.Fatalf("total dependencies %d, ABC's print_supp -s finds 5438", total)
	}
}
