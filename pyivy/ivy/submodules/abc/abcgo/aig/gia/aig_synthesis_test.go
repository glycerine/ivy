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
	p.vNamesIn = Vec_PtrAlloc(3)
	for _, name := range []string{"a", "b", "c"} {
		Vec_PtrPush(p.vNamesIn, name)
	}
	out, err := Gia_ManCompress2Result(p, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.NetworkName() != "compress-test" || out.NetworkSpec() != "compress.aig" || p.vNamesIn != nil || len(out.InputNames()) != 3 {
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
