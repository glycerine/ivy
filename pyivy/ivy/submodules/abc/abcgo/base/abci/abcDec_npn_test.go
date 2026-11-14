package abci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTruthStoreTextAndBinaryRoundTrip(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "truths.txt")
	if err := os.WriteFile(textPath, []byte("53\nA5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var vars, count int32
	Abc_TruthGetParams(textPath, &vars, &count)
	if vars != 3 || count < 2 {
		t.Fatalf("params vars=%d count=%d", vars, count)
	}
	p := Abc_TtStoreLoad(textPath, -1)
	if p == nil || p.nVars != 3 || p.nFuncs != 2 || Abc_TruthGetHex(&p.pFuncs[0][0], 0) != 3 || Abc_TruthGetHex(&p.pFuncs[0][0], 1) != 5 {
		t.Fatalf("loaded text tables = %#v", p)
	}
	textOut := filepath.Join(dir, "roundtrip.txt")
	Abc_TtStoreWrite(textOut, p, 0)
	out, err := os.ReadFile(textOut)
	if err != nil || len(out) < 2 || string(out[:2]) != "35" {
		// The C writer emits low hex digits first, while the C reader consumes
		// the file's textual digits in reverse order. Keep that source behavior.
		t.Fatalf("source text encoding prefix=%q, err=%v", out, err)
	}

	binaryPath := filepath.Join(dir, "truths.bin")
	if err := os.WriteFile(binaryPath, []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}, 0600); err != nil {
		t.Fatal(err)
	}
	r := Abc_TtStoreLoad(binaryPath, 6)
	if r == nil || r.nFuncs != 2 || r.pFuncs[0][0] != 0x0807060504030201 || r.pFuncs[1][0] != 0x0102030405060708 {
		t.Fatalf("binary truth-table values: %#v", r)
	}
}

func TestTruthStoreAllocationAndHexAccessors(t *testing.T) {
	p := Abc_TruthStoreAlloc(7, 2)
	if p == nil || p.nWords != 2 || len(p.pFuncs) != 2 || len(p.backing) != 4 || p.pFuncs[0][0] != p.backing[0] || p.pFuncs[1][0] != p.backing[2] {
		t.Fatal("store allocation is not a contiguous C-style table block")
	}
	Abc_TruthSetBit(&p.pFuncs[0][0], 70)
	Abc_TruthSetHex(&p.pFuncs[0][0], 31, 0xA)
	if Abc_TruthGetBit(&p.pFuncs[0][0], 70) != 1 || Abc_TruthGetHex(&p.pFuncs[0][0], 31) != 0xA {
		t.Fatal("truth bit/hex accessor mismatch")
	}
	Abc_TruthXorBit(&p.pFuncs[0][0], 70)
	Abc_TruthXorHex(&p.pFuncs[0][0], 31, 0xA)
	if p.pFuncs[0][1] != 0 || p.pFuncs[0][0] != 0 {
		t.Fatal("truth xor did not clear bits")
	}
}

func TestNpnUniqueHashChainAndSort(t *testing.T) {
	p := Abc_TruthStoreAlloc(6, 4)
	p.pFuncs[0][0] = 0x10
	p.pFuncs[1][0] = 0x20
	p.pFuncs[2][0] = 0x10
	p.pFuncs[3][0] = 0x20
	table, next := []int32{-1, -1, -1, -1, -1}, []int32{-1, -1, -1, -1}
	key := Abc_TruthHashKey(p.pFuncs[0], p.nWords, int32(len(table)))
	if key < 0 || Abc_TruthHashLookup(p.pFuncs, 0, p.nWords, table, next, key) != 0 {
		t.Fatal("empty chain lookup mismatch")
	}
	next[0], table[key] = -1, 0
	if Abc_TruthHashLookup(p.pFuncs, 2, p.nWords, table, next, key) != 1 {
		t.Fatal("equal table was not found in collision chain")
	}
	if got := Abc_TruthNpnCountUnique(p); got != 2 || p.pFuncs[0][0] != 0x10 || p.pFuncs[1][0] != 0x20 {
		t.Fatalf("unique count=%d funcs=%v", got, p.pFuncs)
	}
	p.pFuncs = append(p.pFuncs, []word{0x10}, []word{0x20})
	p.nFuncs = 4
	if got := Abc_TruthNpnCountUniqueSort(p); got != 2 || p.pFuncs[0][0] != 0x10 || p.pFuncs[1][0] != 0x20 {
		t.Fatalf("sorted unique count=%d funcs=%v", got, p.pFuncs)
	}
}

func TestNpnSemiCanonicalizationKeepsEquivalentFunctionsTogether(t *testing.T) {
	p := Abc_TruthStoreAlloc(3, 2)
	// x0 is 0xAAAAAAAA in the low 32 bits; x0 xor 1 is equivalent under
	// the input phase operation and should reach the same canonical table.
	p.pFuncs[0][0] = 0xAAAAAAAAAAAAAAAA
	p.pFuncs[1][0] = 0x5555555555555555
	Abc_TruthNpnPerform(p, 2, 0)
	if p.nFuncs != 1 {
		t.Fatalf("NPN phase canonicalization left %d classes", p.nFuncs)
	}
}
