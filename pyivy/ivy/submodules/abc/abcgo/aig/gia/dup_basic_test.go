package gia

import "testing"

func TestNativeDuplicateRealGatesAndMetadataOwnership(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.EnableMuxes()
	xor := Gia_ManAppendXorReal(p, a^1, b)
	mux := Gia_ManAppendMuxReal(p, c, xor, a)
	buffer := Gia_ManAppendBuf(p, mux^1)
	Gia_ManAppendCo(p, buffer)
	Gia_ManAppendCo(p, c)
	p.nConstrs = 1
	p.SetIdentity("clone", "clone.aig")
	p.aigerExtensions = map[byte][]byte{'h': {1, 2, 3}}
	p.aigerSymbols = map[byte][]string{'i': {"a", "b", "c"}}
	p.aigerComment = []byte("comment")
	clone := Gia_ManDup(p)
	if clone.nConstrs != 1 || clone.pName != "clone" || clone.PSpec != "clone.aig" || clone.nXors != 1 || clone.nMuxes != 1 || clone.nBufs != 1 {
		t.Fatal("duplicate tags/metadata")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		before, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		after, err := clone.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		for i := range before {
			if before[i] != after[i] {
				t.Fatal("duplicate function", mask, before, after)
			}
		}
	}
	clone.aigerExtensions['h'][0] = 9
	clone.aigerSymbols['i'][0] = "changed"
	clone.aigerComment[0] = 'C'
	if p.aigerExtensions['h'][0] != 1 || p.aigerSymbols['i'][0] != "a" || string(p.aigerComment) != "comment" {
		t.Fatal("duplicate aliases serialized metadata")
	}
	for _, obj := range p.pObjs {
		if obj.Value != uint32(obj.id<<1) {
			t.Fatal("source copy map")
		}
	}
}
