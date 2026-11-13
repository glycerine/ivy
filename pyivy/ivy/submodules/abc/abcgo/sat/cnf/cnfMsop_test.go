package cnf

import "testing"

func TestABCCompressedMsopsAllTruthTables(t *testing.T) {
	sizes, sops := cnfReadMsops()
	if len(sizes) != 1<<16 || len(sops) != 1<<16 {
		t.Fatal("wrong table size")
	}
	masks := [4][2]uint16{{0x5555, 0xaaaa}, {0x3333, 0xcccc}, {0x0f0f, 0xf0f0}, {0x00ff, 0xff00}}
	for expected, sop := range sops {
		if int(sizes[expected]) != len(sop) {
			t.Fatalf("truth %04x bad cube count", expected)
		}
		var truth uint16
		for i := 0; i < len(sop); i++ {
			cube := uint16(0xffff)
			lit := int(sop[i])
			for b := 3; b >= 0; b-- {
				if lit%3 < 2 {
					cube &= masks[b][lit%3]
				}
				lit /= 3
			}
			truth |= cube
		}
		if int(truth) != expected {
			t.Fatalf("truth %04x decoded %04x", expected, truth)
		}
	}
}

func TestReadMsopsReturnsIndependentTables(t *testing.T) {
	first := Cnf_ReadMsops()
	second := Cnf_ReadMsops()
	if len(first[1]) == 0 || len(second[1]) == 0 {
		t.Fatal("missing MSOP")
	}
	before := second[1][0]
	first[1][0] ^= 1
	if second[1][0] != before || cnfMsopCovers()[1][0] != before {
		t.Fatal("caller changed shared MSOP table")
	}
}
