package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"path/filepath"
	"testing"
)

func TestBCoreSequentialReachability(t *testing.T) {
	p := gia.Gia_ManStart(16)
	pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, ro)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 1)
	for _, tc := range []struct {
		frame int32
		want  bsat.Status
	}{{0, bsat.Unsatisfiable}, {1, bsat.Satisfiable}} {
		s := bsat.NewSolver()
		mapping := Bmc_ManBCoreCollect(p, tc.frame, 0, s)
		if got := s.Solve(); got != tc.want {
			t.Errorf("frame %d got %v want %v; map=%v", tc.frame, got, tc.want, vec.Vec_IntArray(mapping))
		}
	}
}
func TestBCorePivotMapping(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	m := vec.Vec_IntAllocArrayCopy([]int32{-1, -1, 1, 0, 2, 0, 1, 1}, 8)
	path := filepath.Join(t.TempDir(), "pivots")
	if err := os.WriteFile(path, []byte("1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got := vec.Vec_IntArray(Bmc_ManBCoreCollectPivots(p, path, m))
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("pivots %v", got)
	}
}
func TestBCoreProofFile(t *testing.T) {
	p := gia.Gia_ManStart(8)
	pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, ro)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 1)
	path := filepath.Join(t.TempDir(), "proof")
	Bmc_ManBCorePerform(p, &Bmc_BCorePar_t{pFileProof: path})
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) == 0 {
		t.Fatal("empty proof")
	}
}
