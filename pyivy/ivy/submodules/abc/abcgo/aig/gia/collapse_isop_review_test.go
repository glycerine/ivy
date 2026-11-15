package gia

import (
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCollapseRebuildProducesFactoredIsopCover(t *testing.T) {
	dd := cudd.Cudd_Init(3, 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	a, b, c := cudd.Cudd_bddIthVar(dd, 0), cudd.Cudd_bddIthVar(dd, 1), cudd.Cudd_bddIthVar(dd, 2)
	majority := cudd.Cudd_bddOr(dd, cudd.Cudd_bddAnd(dd, a, b), cudd.Cudd_bddOr(dd, cudd.Cudd_bddAnd(dd, a, c), cudd.Cudd_bddAnd(dd, b, c)))
	cudd.Cudd_Ref(majority)
	defer cudd.Cudd_RecursiveDeref(dd, majority)
	g := Gia_ManStart(32)
	leaves := Vec_IntAlloc(3)
	for i := 0; i < 3; i++ {
		Vec_IntPush(leaves, Gia_ManAppendCi(g))
	}
	Gia_ManHashStart(g)
	sop, cube := Vec_StrAlloc(0), Vec_StrAlloc(0)
	root := Gia_ManRebuildIsop(dd, majority, g, leaves, sop, cube)
	if root < 0 {
		t.Fatal("majority ISOP reconstruction failed")
	}
	cover := string(Vec_StrArray(sop))
	if strings.Count(cover, "\n") != 3 || !strings.Contains(cover, " 1\n") {
		t.Fatalf("positive minimum ISOP cover was not produced: %q", cover)
	}
	Gia_ManAppendCo(g, root)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		ones := 0
		for _, x := range inputs {
			if x {
				ones++
			}
		}
		checkQuantificationOutputs(t, g, inputs, []bool{ones >= 2})
	}
}

func TestCollapseRebuildRetainsInterfaceMetadata(t *testing.T) {
	p := Gia_ManStart(16)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, r))
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, r)
	Gia_ManSetRegNum(p, 1)
	if err := p.SetAigerProperties(1, 1); err != nil {
		t.Fatal(err)
	}
	p.aigerLatchInits = []int32{1}
	p.aigerComment = []byte("collapse-state")
	out := Gia_ManCollapseTest(p, 0)
	if out.nRegs != 1 || out.nConstrs != 1 || out.aigerBadOutputs != 1 || len(out.aigerLatchInits) != 1 || out.aigerLatchInits[0] != 1 || string(out.aigerComment) != "collapse-state" {
		t.Fatalf("rebuilt interface lost: regs%d constraints%d bad%d resets%v comment%q", out.nRegs, out.nConstrs, out.aigerBadOutputs, out.aigerLatchInits, out.aigerComment)
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		want, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		checkQuantificationOutputs(t, out, inputs, want)
	}
	out.aigerLatchInits[0] = 0
	out.aigerComment[0] = 'X'
	if p.aigerLatchInits[0] != 1 || string(p.aigerComment) != "collapse-state" {
		t.Fatal("rebuilt metadata aliases input")
	}
}

func TestCollapseIsopCubeLimit(t *testing.T) {
	for _, n := range []int{10, 11, 12} {
		dd := cudd.Cudd_Init(uint32(n), 0, 0, 0, 0)
		f := dd.LogicZero
		g := Gia_ManStart(32)
		leaves := Vec_IntAlloc(n)
		for i := 0; i < n; i++ {
			f = cudd.Cudd_bddXor(dd, f, cudd.Cudd_bddIthVar(dd, i))
			Vec_IntPush(leaves, Gia_ManAppendCi(g))
		}
		cudd.Cudd_Ref(f)
		Gia_ManHashStart(g)
		sop, cube := Vec_StrAlloc(0), Vec_StrAlloc(0)
		root := Gia_ManRebuildIsop(dd, f, g, leaves, sop, cube)
		if n > 10 {
			if root != -1 || Gia_ManObjNum(g) != int32(n+1) || len(Vec_StrArray(sop)) != 0 || len(Vec_StrArray(cube)) != 0 {
				t.Fatalf("%d-input parity must reject its oversized cover before mutation", n)
			}
		} else {
			if root < 0 || strings.Count(string(Vec_StrArray(sop)), "\n") != 512 {
				t.Fatalf("512-cube cover rejected: %d", root)
			}
			Gia_ManAppendCo(g, root)
			for mask := 0; mask < 1<<n; mask++ {
				inputs, parity := make([]bool, n), false
				for i := range inputs {
					inputs[i] = mask>>i&1 != 0
					parity = parity != inputs[i]
				}
				checkQuantificationOutputs(t, g, inputs, []bool{parity})
			}
		}
		cudd.Cudd_RecursiveDeref(dd, f)
		cudd.Cudd_Quit(dd)
	}
}
