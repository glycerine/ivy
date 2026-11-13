package bmc

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func captureInductOutput(t *testing.T, run func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close() }()
	run()
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestInductionEntryChecksEmptyAndFullSets(t *testing.T) {
	// The two-register-copy explicit-state oracle finds a mismatch when the
	// register is omitted, and no mismatch when it is selected.
	g := gia.Gia_ManStart(8)
	gia.Gia_ManHashStart(g)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, truthLiteral(g, pi, ro, 1))
	gia.Gia_ManAppendCo(g, truthLiteral(g, pi, ro, 2))
	gia.Gia_ManSetRegNum(g, 1)
	gia.Gia_ManHashStop(g)
	if got := captureInductOutput(t, func() { Bmc_PerformICheck(g, 1, 0, 1, 0) }); !strings.Contains(got, "Inductiveness status:") {
		t.Fatalf("empty selection must be noninductive: %q", got)
	}
	got := captureInductOutput(t, func() { Bmc_PerformICheck(g, 1, 0, 0, 1) })
	if !strings.Contains(got, "Inductive assumptions:") || strings.Contains(got, "Inductiveness status:") {
		t.Fatalf("full selection must be inductive: %q", got)
	}
}

func TestInseEntryUsesSourceZeroInitialState(t *testing.T) {
	// The C entry point intentionally ignores the supplied initial vector and
	// starts at zero. A holding register remains zero even if the caller gives 1.
	g := gia.Gia_ManStart(8)
	ro := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManSetRegNum(g, 1)
	initial := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	result := Gia_ManInseTest(g, initial, 2, 1, 0, 0, 0)
	if got := vec.Vec_IntArray(result); len(got) != 1 || got[0] != 0 {
		t.Fatalf("holding register from zero: %v", got)
	}
	if vec.Vec_IntEntry(initial, 0) != 1 {
		t.Fatal("caller initial vector was modified")
	}
}
