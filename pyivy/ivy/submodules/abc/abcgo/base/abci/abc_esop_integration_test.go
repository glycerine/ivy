package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/exor"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestESOPPreservesCubeLiteralsAndOutputIndices(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(n)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	// Deliberately reorder fanins: row literals use global input positions.
	for _, input := range []*abc.Object{c, a, b} {
		abc.Abc_ObjAddFanin(node, input)
	}
	node.Data = "10- 1\n-11 1\n"
	for _, driver := range []*abc.Object{node, b, abc.Abc_NtkCreateNodeConst1(n), abc.Abc_NtkCreateNodeConst0(n)} {
		abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), driver)
	}
	rows := abcNtkESOP(n)
	if rows == nil {
		t.Fatal("ESOP conversion failed")
	}
	defer vec.Vec_WecFree(rows)
	want := [][]int32{{4, 1, -1}, {0, 2, -1}, {2, -2}, {-3}}
	if got := abcESOPRows(rows); !reflect.DeepEqual(got, want) {
		t.Fatalf("rows=%v want=%v", got, want)
	}
}

func TestExorcismCommandRoundTripAndWriteError(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	dir := t.TempDir()
	in, out := filepath.Join(dir, "input.pla"), filepath.Join(dir, "output.pla")
	if err := os.WriteFile(in, []byte(".i 2\n.o 2\n.p 3\n1- 10\n-1 10\n11 01\n.e\n"), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"&exorcism", "-Q", "1", in, out}
	if status := Abc_CommandAbc9Exorcism(frame, int32(len(args)), args); status != 0 {
		t.Fatalf("ESOP command returned %d", status)
	}
	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cover, err := exor.ReadPLA(file)
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0}
		got, err := cover.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || got[0] != (values[0] != values[1]) || got[1] != (values[0] && values[1]) {
			t.Fatalf("bits=%d outputs=%v", bits, got)
		}
	}
	args[len(args)-1] = filepath.Join(dir, "missing", "output.pla")
	status, message := captureIfifPrint(t, func() int32 {
		return Abc_CommandAbc9Exorcism(frame, int32(len(args)), args)
	})
	if status != 1 || !strings.Contains(message, "ESOP minimization failed:") {
		t.Fatalf("write failure returned status %d, message %q", status, message)
	}
}
