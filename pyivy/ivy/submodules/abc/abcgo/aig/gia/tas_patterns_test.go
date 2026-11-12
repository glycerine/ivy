package gia

import "testing"

func TestTasStorePatternMergesAndRejectsAtomically(t *testing.T) {
	info := [][]uint32{{0, 0}, {0, 0}, {0, 0}}
	present := [][]uint32{{0, 0}, {0, 0}, {0, 0}}
	if Tas_StorePatternTry(info, present, 33, []int32{2, 5}, 2) != 1 {
		t.Fatal("initial assignment rejected")
	}
	if info[1][1]&2 == 0 || info[2][1]&2 != 0 || present[1][1]&2 == 0 || present[2][1]&2 == 0 {
		t.Fatal("literal values not recorded")
	}
	if Tas_StorePatternTry(info, present, 33, []int32{0, 3}, 2) != 0 {
		t.Fatal("conflicting assignment accepted")
	}
	if present[0][1]&2 != 0 {
		t.Fatal("conflict partially modified the pattern")
	}
	if Tas_StorePatternTry(info, present, 33, []int32{2}, 1) != 1 {
		t.Fatal("compatible assignment rejected")
	}
}

func TestTasStorePatternUsesAlternativeBits(t *testing.T) {
	info, present := [][]uint32{{0}}, [][]uint32{{0}}
	if Tas_StorePattern(info, present, []int32{0}) != 1 || Tas_StorePattern(info, present, []int32{1}) != 1 {
		t.Fatal("failed to pack opposing values")
	}
	if present[0][0]&6 != 6 || info[0][0]&6 != 2 {
		t.Fatalf("wrong packed values: info=%b present=%b", info[0][0], present[0][0])
	}
}
