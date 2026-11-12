package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestSourceSatokoDriverUsesOutputOR(t *testing.T) {
	p := Gia_ManStart(4)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a^1)
	result, err := Gia_ManSatokoCallOneResult(p, nil)
	if err != nil || result.Status != satoko.SATOKO_SAT {
		t.Fatal("source MF output clause is OR", result, err)
	}
	zero := Gia_ManStart(2)
	Gia_ManAppendCo(zero, 0)
	result, err = Gia_ManSatokoCallOneResult(zero, nil)
	if err != nil || result.Status != satoko.SATOKO_UNSAT || result.Conflicts != 0 {
		t.Fatal("constant output", result, err)
	}
	solver := Gia_ManSatokoCreate(p, nil)
	if solver == nil || solver.Solve() != satoko.SATOKO_SAT {
		t.Fatal("source create path")
	}
}
func TestSourceMappedConeCollection(t *testing.T) {
	// Variables 0 and1 are terminals;2 depends on0,1;3 depends on2.
	mapping := Vec_IntAlloc(12)
	for _, value := range []int32{0, 0, 4, 7, 2, 0, 1, 1, 2} {
		Vec_IntPush(mapping, value)
	}
	cone := Gia_ManCollectVars(3, mapping, 4)
	want := []int32{0, 1, 2, 3}
	if Vec_IntSize(cone) != len(want) {
		t.Fatal("cone size")
	}
	for i, value := range want {
		if Vec_IntEntry(cone, i) != value {
			t.Fatal("postorder", Vec_IntArray(cone))
		}
	}
}
func TestDfsConePreservesAllInputsAndOnlyRequestedOutput(t *testing.T) {
	src := Gia_ManStart(6)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x := Gia_ManAppendAnd(src, a, b^1)
	Gia_ManAppendCo(src, x^1)
	Gia_ManAppendCo(src, c)
	dst := Gia_ManDupDfsCone(src, Gia_ManCo(src, 0))
	if Gia_ManCiNum(dst) != 3 || Gia_ManCoNum(dst) != 1 || dst.nRegs != 0 {
		t.Fatal("cone interface")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		left, err := src.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		right, err := dst.Evaluate(input)
		if err != nil || left[0] != right[0] {
			t.Fatal("cone logic", mask, left, right, err)
		}
	}
}

func TestSourceMFDriverPublishesTypedData(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	cnf, err := Mf_ManGenerateCnfResult(p, 8, false, true, false)
	if err != nil || p.Data() != cnf {
		t.Fatal("published CNF data", err)
	}
	if p.TakeData() != cnf || p.Data() != nil {
		t.Fatal("CNF data ownership")
	}
}
