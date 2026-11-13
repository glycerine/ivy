package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFaultDump(t *testing.T) {
	p := gia.Gia_ManStart(8)
	pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, pi, ro))
	gia.Gia_ManSetRegNum(p, 1)
	tests := vec.Vec_IntAllocArrayCopy([]int32{1, 1, 0, 0, 1, 1}, 6)
	name := filepath.Join(t.TempDir(), "tests")
	if err := Gia_ManDumpTests(tests, 2, name); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(name)
	if err != nil || string(data) != "110\n011\n" {
		t.Fatalf("dump %q %v", data, err)
	}
	if err := Gia_ManDumpTestsDelay(tests, 2, name, p); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(name)
	if err != nil || string(data) != "11\n01\n10\n10\n" {
		t.Fatalf("delay %q %v", data, err)
	}
}

func TestFaultPatternsAndDuplicate(t *testing.T) {
	name := filepath.Join(t.TempDir(), "patterns")
	if err := os.WriteFile(name, []byte("01 1\n0\t"), 0600); err != nil {
		t.Fatal(err)
	}
	got := Gia_ManGetTestPatterns(name)
	if got == nil || !reflect.DeepEqual(vec.Vec_IntArray(got), []int32{0, 1, 1, 0}) {
		t.Fatalf("patterns %v", got)
	}
	if err := os.WriteFile(name, []byte("0x1"), 0600); err != nil {
		t.Fatal(err)
	}
	if Gia_ManGetTestPatterns(name) != nil {
		t.Fatal("invalid pattern accepted")
	}
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	q := Gia_ManDeriveDup(p, 2)
	out, err := q.Evaluate([]bool{true, false, true})
	if err != nil || !out[0] || len(q.InputLiterals()) != 3 {
		t.Fatalf("duplicate %v %v", out, err)
	}
}
