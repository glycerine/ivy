package abci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNpnTruthPaddingAndCanonicalization(t *testing.T) {
	// A two-input truth is repeated across the six-input truth word exactly as
	// ABC's Npn_TruthPadWord does; canonicalization only chooses complement in
	// this source (polarity and permutation flags are compile-time zero).
	got := Npn_TruthPadWord(0x6, 2)
	if got != 0x6666666666666666 {
		t.Fatalf("pad = %016x", got)
	}
	phase := int32(-1)
	if canon := Npn_TruthCanon(^word(0), 6, &phase); canon != 0 || phase != 1<<6 {
		t.Fatalf("canon=%016x phase=%x", canon, phase)
	}
	if Npn_TruthSupport(0xaaaaaaaaaaaaaaaa) != 1 || Npn_TruthSuppSize(0xaaaaaaaaaaaaaaaa, 6) != 1 {
		t.Fatal("single-variable support mismatch")
	}
}

func TestNpnManagerSerializationAndCounts(t *testing.T) {
	p := Npn_ManStart("")
	Npn_ManAdd(p, 0x1234)
	Npn_ManAdd(p, 0x1234)
	Npn_ManAdd(p, 0x5678)
	if p.nEntries != 3 || Npn_ManObj(p, 1).Count != 2 {
		t.Fatalf("manager entries/count = %d/%d", p.nEntries, Npn_ManObj(p, 1).Count)
	}
	name := filepath.Join(t.TempDir(), "npn.txt")
	Npn_ManWrite(p, name)
	loaded := Npn_ManStart(name)
	defer Npn_ManStop(loaded)
	if loaded == nil || loaded.nEntries != 3 {
		t.Fatalf("loaded manager = %#v", loaded)
	}
	if got := Npn_ManObj(loaded, 1); got == nil || got.Count != 2 {
		t.Fatalf("loaded frequency = %#v", got)
	}
	Npn_ManStop(p)
	if _, err := os.Stat(name); err != nil {
		t.Fatal(err)
	}
}
