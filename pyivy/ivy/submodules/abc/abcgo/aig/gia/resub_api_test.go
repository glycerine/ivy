package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"path/filepath"
	"testing"
)

func TestResubPublicComputeAndDump(t *testing.T) {
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	on := a & b
	divs := [][]word{{^on}, {on}, {a}, {b}}
	Abc_ResubPrepareManager(1)
	defer Abc_ResubPrepareManager(0)
	result := Abc_ResubComputeFunction(divs, 1, 1, 20, 0, 1, 1, 0)
	if len(result) != 3 {
		t.Fatalf("gates %v", result)
	}
	path := filepath.Join(t.TempDir(), "divs.hex")
	Abc_ResubDumpProblem(path, divs, 1)
	var n int
	words := vec.Vec_WrdReadHex(path, &n, 0)
	if words == nil || n != 1 || vec.Vec_PtrSize(Gia_ManDeriveDivs(words, 1)) != 4 {
		t.Fatal("dump round trip failed")
	}
}
