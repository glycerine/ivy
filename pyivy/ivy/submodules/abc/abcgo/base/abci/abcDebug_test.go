package abci

import (
	"io"
	"os"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestAbcNtkAutoDebugOriginalPredicateOutput(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	oldStdout := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = write
	Abc_NtkAutoDebug(n, func(*Abc_Ntk_t) int32 { return 0 })
	abc.Abc_NtkDelete(n)
	_ = write.Close()
	os.Stdout = oldStdout
	got, err := io.ReadAll(read)
	_ = read.Close()
	if err != nil {
		t.Fatal(err)
	}
	const want = "The original network does not cause the bug. Quitting.\n"
	if string(got) != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestDebugFaninEnumerationExcludesSingleOutputAndConstants(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := abc.Abc_NtkCreateNodeAnd(n, []*abc.Abc_Obj_t{a, b})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, and)
	abc.Abc_NtkOrderCisCos(n)
	if got := Abc_NtkCountFaninsTotal(n); got != 2 {
		t.Fatalf("single-output network fanin count = %d, want 2", got)
	}
	var obj, fanin *abc.Abc_Obj_t
	if Abc_NtkFindGivenFanin(n, 1, &obj, &fanin) == 0 || obj != and || fanin != b {
		t.Fatalf("step 1 = (%p, %p), want (%p, %p)", obj, fanin, and, b)
	}
	abc.Abc_NtkDelete(n)
}

func TestGlibcRandSequence(t *testing.T) {
	r := newGlibcRand(1)
	want := []uint32{1804289383, 846930886, 1681692777, 1714636915, 1957747793}
	for i, expected := range want {
		if got := r.next(); got != expected {
			t.Fatalf("rand[%d] = %d, want %d", i, got, expected)
		}
	}
}
