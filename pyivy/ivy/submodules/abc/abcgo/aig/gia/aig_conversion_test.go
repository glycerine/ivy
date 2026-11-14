package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestActualAigGiaRoundTrip(t *testing.T) {
	source := aig.NewGraph()
	source.Name = "round-trip"
	source.Spec = "source.aig"
	a, b := source.NewInput(), source.NewInput()
	source.AddOutput(source.XOR(a, b))
	source.AddOutput(source.Const1())
	source.AddOutput(source.Const0())
	for _, simple := range []bool{false, true} {
		var converted *Gia_Man_t
		if simple {
			converted = Gia_ManFromAigSimple(source)
		} else {
			converted = Gia_ManFromAig(source)
		}
		if converted.pName != source.Name || converted.PSpec != source.Spec {
			t.Fatal("metadata")
		}
		restored := Gia_ManToAig(converted, 0)
		for mask := 0; mask < 4; mask++ {
			input := []bool{mask&1 != 0, mask&2 != 0}
			expected, err := source.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			got, err := converted.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			back, err := restored.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			for i := range expected {
				if got[i] != expected[i] || back[i] != expected[i] {
					t.Fatal(simple, mask, expected, got, back)
				}
			}
		}
		again := Gia_ManFromAig(restored)
		if Cec_ManVerifyTwo(converted, again, 0) != 1 {
			t.Fatal("SAT round-trip comparison")
		}
	}
}

func TestAigNonTopologicalBufferImport(t *testing.T) {
	p := aig.NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	old := p.AND(a, b)
	p.AddOutput(old)
	replacement := p.AND(b, c)
	aig.Aig_ObjReplace(p, old, replacement, 1)
	converted := Gia_ManFromAig(p)
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		expected, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := converted.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		if expected[0] != got[0] {
			t.Fatal(mask, expected, got)
		}
	}
}

func TestAigSimpleImportRetainsDanglingLogic(t *testing.T) {
	p := aig.NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	p.AddOutput(root)
	p.AND(a, c)
	reachable, all := Gia_ManFromAig(p), Gia_ManFromAigSimple(p)
	if Gia_ManAndNum(all) <= Gia_ManAndNum(reachable) {
		t.Fatal("simple mode dropped dangling logic")
	}
}

func TestAigRegisterAndConstraintMetadata(t *testing.T) {
	p := Gia_ManStart(6)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, pi)
	Gia_ManAppendCo(p, ro)
	p.nRegs = 1
	p.nConstrs = 1
	converted := Gia_ManToAig(p, 0)
	if aig.Aig_ManRegNum(converted) != 1 || aig.Aig_ManConstrNum(converted) != 1 {
		t.Fatal("AIG metadata lost")
	}
	restored := Gia_ManFromAig(converted)
	if restored.nRegs != 1 || restored.nConstrs != 1 {
		t.Fatal("GIA metadata lost")
	}
}
