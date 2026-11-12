package gia

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCollapseAndRebuild(t *testing.T) {
	p := Gia_ManStart(32)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	f := Gia_ManHashMux(p, a, Gia_ManHashXor(p, b, c), Gia_ManHashAnd(p, b^1, d))
	Gia_ManAppendCo(p, f)
	Gia_ManAppendCo(p, f^1)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, 1)
	dd := cudd.Cudd_Init(4, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	roots := Gia_ManCollapse(p, dd, 10000, 0)
	if roots == nil || vec.Vec_PtrSize(roots) != 4 {
		t.Fatal("collapse failed")
	}
	defer Gia_ManCollapseDeref(dd, roots)
	rebuilt := Gia_ManCollapseTest(p, 0)
	for mask := 0; mask < 16; mask++ {
		inputs := make([]bool, 4)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		expected, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := rebuilt.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, want := range expected {
			if actual[i] != want {
				t.Fatalf("inputs %v output %d: %v != %v", inputs, i, actual[i], want)
			}
			got, err := dd.Evaluate(vec.Vec_PtrEntry(roots, i).(*cudd.DdNode), inputs)
			if err != nil || got != want {
				t.Fatalf("BDD %v output %d: %v %v", inputs, i, got, err)
			}
		}
	}
}

func TestCollapseConstantAndSequential(t *testing.T) {
	p := Gia_ManStart(8)
	x, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, s)
	Gia_ManAppendCo(p, x)
	Gia_ManSetRegNum(p, 1)
	result := Gia_ManCollapseTest(p, 0)
	if Gia_ManRegNum(result) != 1 {
		t.Fatal("lost register count")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		want, _ := p.Evaluate(in)
		got, err := result.Evaluate(in)
		if err != nil || len(want) != len(got) {
			t.Fatalf("invalid rebuild: %v", err)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatal("wrong output")
			}
		}
	}
}

func TestCollapseAllThreeInputFunctions(t *testing.T) {
	for truth := 0; truth < 256; truth++ {
		p := Gia_ManStart(32)
		vars := []int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
		Gia_ManHashAlloc(p)
		var root int32
		for assignment := 0; assignment < 8; assignment++ {
			if truth&(1<<assignment) == 0 {
				continue
			}
			cube := int32(1)
			for i, v := range vars {
				if assignment&(1<<i) == 0 {
					v ^= 1
				}
				cube = Gia_ManHashAnd(p, cube, v)
			}
			root = Gia_ManHashOr(p, root, cube)
		}
		Gia_ManAppendCo(p, root)
		rebuilt := Gia_ManCollapseTest(p, 0)
		for assignment := 0; assignment < 8; assignment++ {
			in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			got, err := rebuilt.Evaluate(in)
			if err != nil || got[0] != (truth&(1<<assignment) != 0) {
				t.Fatalf("truth %02x input %d: %v %v", truth, assignment, got, err)
			}
		}
	}
}
