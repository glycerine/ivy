package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestFaultPrepareFlip(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	options := &Bmc_ParFf_t{Algo: 3}
	mapping := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	tests := vec.Vec_IntAlloc(4)
	scratch := vec.Vec_IntAlloc(4)
	var m *gia.Gia_Man_t
	var data *cnf.Cnf_Dat_t
	var s *bsat.Solver
	if Gia_ManFaultPrepare(p, p, options, 2, mapping, tests, scratch, &m, &data, &s, 0) != 1 {
		t.Fatal("prepare failed")
	}
	if s.Solve() != bsat.Satisfiable {
		t.Fatal("expected detectable fault")
	}
	fault := data.VarNums[m.InputLiterals()[2]>>1]
	if value, ok := s.Value(fault); !ok || !value {
		t.Fatalf("fault enabled: %v %v", value, ok)
	}
}
