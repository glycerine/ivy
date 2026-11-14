package gia

import "testing"

func TestDarCompressionRealBridgeAndMetadataMove(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	ab := Gia_ManAppendAnd(p, a, b)
	nested := Gia_ManAppendAnd(p, a, ab)
	root := Gia_ManAppendOr(p, nested, c)
	Gia_ManAppendCo(p, root)
	p.SetIdentity("compress-test", "compress.aig")
	p.VNamesIn = Vec_PtrAlloc(3)
	for _, name := range []string{"a", "b", "c"} {
		Vec_PtrPush(p.VNamesIn, name)
	}
	out, err := Gia_ManCompress2Result(p, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.NetworkName() != "compress-test" || out.NetworkSpec() != "compress.aig" || p.VNamesIn != nil || len(out.InputNames()) != 3 {
		t.Fatal("compression metadata ownership")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		before, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		after, err := out.Evaluate(input)
		if err != nil || before[0] != after[0] {
			t.Fatal("compression changed truth", mask, before, after, err)
		}
	}
}

func TestDefaultDchChoiceBridge(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, root)
	out := Gia_ManPerformDch(p, nil)
	if out == nil {
		t.Fatal("default DCH engine failed")
	}
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		got, err := out.Evaluate(input)
		if err != nil || got[0] != (input[0] && input[1]) {
			t.Fatal("DCH choice truth", mask, got, err)
		}
	}
}

func TestDefaultDchUsesChoiceMiterSourceOrder(t *testing.T) {
	p := Gia_ManStart(32)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	ab := p.AndLiteral(a, b, true)
	ac := p.AndLiteral(a, c, true)
	or := p.AndLiteral(ab^1, ac^1, true) ^ 1
	Gia_ManAppendCo(p, p.AndLiteral(or, d, true))
	result := Gia_ManPerformDch(p, nil)
	if result == nil {
		t.Fatal("default choice path failed")
	}
	if nodes, choices := Gia_ManAndNum(result), Gia_ManChoiceNum(result); nodes != 7 || choices != 1 {
		t.Fatalf("compiled C source path has 7 ANDs and 1 choice; got %d and %d", nodes, choices)
	}
}

func TestSequentialCleanupClassBridge(t *testing.T) {
	p := Gia_ManStart(8)
	input, first, second := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, first, second))
	Gia_ManAppendCo(p, input)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 2)
	Gia_ManSeqCleanupClasses(p, 1, 1, 0)
	if p.PReprs == nil || Gia_ObjReprSelf(p, first>>1) != Gia_ObjReprSelf(p, second>>1) || p.nRegs != 2 {
		t.Fatal("sequential representatives not imported")
	}
}
