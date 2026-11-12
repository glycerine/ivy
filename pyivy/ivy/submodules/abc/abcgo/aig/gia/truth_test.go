package gia

import "testing"

func TestOriginalTruthWordOperations(t *testing.T) {
	for _, word := range []uint32{0, 1, 2, 3, 0x80000000, 0xffffffff, 0x55555555} {
		count := int32(0)
		first := int32(-1)
		for i := int32(0); i < 32; i++ {
			if word&(1<<uint(i)) != 0 {
				count++
				if first < 0 {
					first = i
				}
			}
		}
		if Gia_WordCountOnes(word) != count || Gia_WordFindFirstBit(word) != first {
			t.Fatal(word)
		}
	}
	if Gia_WordHasOneBit(0) != 1 || Gia_WordHasOneBit(3) != 0 || Gia_ObjCutSign(33) != 2 {
		t.Fatal("source predicates")
	}
	for vars := int32(0); vars < 10; vars++ {
		n := giaTruthWords(vars)
		a, b := make([]uint32, n+1), make([]uint32, n+1)
		a[n] = 123
		b[n] = 456
		Gia_ManTruthFill(a, vars)
		if Gia_ManTruthIsConst1(a, vars) != 1 || a[n] != 123 {
			t.Fatal(vars)
		}
		Gia_ManTruthNot(b, a, vars)
		if Gia_ManTruthIsConst0(b, vars) != 1 || b[n] != 456 {
			t.Fatal(vars)
		}
		Gia_ManTruthCopy(b, a, vars)
		if Gia_ManTruthIsConst1(b, vars) != 1 {
			t.Fatal(vars)
		}
		Gia_ManTruthNot(b, b, vars)
		if Gia_ManTruthIsConst0(b, vars) != 1 {
			t.Fatal("in-place not")
		}
		Gia_ManTruthClear(a, vars)
		if Gia_ManTruthIsConst0(a, vars) != 1 {
			t.Fatal(vars)
		}
	}
}
