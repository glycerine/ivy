package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLiteralAndBitHelpers(t *testing.T) {
	for i := int32(0); i < 100; i++ {
		l := Abc_Var2Lit(i, 1)
		if Abc_Lit2Var(l) != i || Abc_LitIsCompl(l) != 1 || Abc_LitNot(l) != i*2 {
			t.Fatal(l)
		}
	}
	p := make([]uint32, 4)
	Abc_InfoSetBit(p, 63)
	if !Abc_InfoHasBit(p, 63) {
		t.Fatal(p)
	}
	Abc_InfoXorBit(p, 63)
	if Abc_InfoHasBit(p, 63) {
		t.Fatal(p)
	}
}
func TestGlobalFileAndLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	var n int32
	b := Abc_FileReadContents(path, &n)
	if n != 3 || len(b) != 35 || string(b[:3]) != "abc" {
		t.Fatal(n, b)
	}
	for _, x := range b[3:] {
		if x != 0 {
			t.Fatal(b)
		}
	}
	for _, tc := range [][2]uint32{{0, 0}, {1, 1}, {2, 1}, {3, 2}, {4, 2}, {5, 3}} {
		if Abc_Base2Log(tc[0]) != int32(tc[1]) {
			t.Fatal(tc)
		}
	}
}

func TestNumericPointerPayloadsAndLargeBitCounts(t *testing.T) {
	for _, value := range []uint64{0, 1, 3, 1 << 40, ^uint64(0)} {
		if got := Abc_Ptr2Wrd(Abc_Wrd2Ptr(value)); got != value {
			t.Fatal(value, got)
		}
	}
	if got := Abc_Ptr2Int(Abc_Int2Ptr(-17)); got != -17 {
		t.Fatal(got)
	}
	if Abc_BitByteNum(2147483647) != 268435456 || Abc_BitWordNum(2147483647) != 67108864 || Abc_Bit6WordNum(2147483647) != 33554432 {
		t.Fatal("rounding overflow")
	}
}
