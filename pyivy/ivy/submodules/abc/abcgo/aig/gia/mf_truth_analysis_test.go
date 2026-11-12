package gia

import "testing"

func TestMfTruthCanonicize(t *testing.T) {
	value := uint64(0xaaaaaaaaaaaaaaaa)
	config := Mf_ManTruthCanonicize(&value, 1)
	if value != 0x5555555555555555 || config != 2 {
		t.Fatalf("truth %x config %d", value, config)
	}
	Mf_ManTruthQuit()
	if len(mfTruthClasses) != 595 {
		t.Fatal("missing DSD classes")
	}
}
