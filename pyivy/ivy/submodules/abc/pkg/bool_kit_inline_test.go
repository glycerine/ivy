package abc

import "testing"

func TestKitInlineTruth(t *testing.T) {
	const nVars = 6
	n := kitTruthWordNum(nVars)
	if n != 2 || kitTruthWordNum(3) != 1 || kitBitWordNum(33) != 2 || kitBitWordNum(32) != 1 {
		t.Fatal("word counts")
	}
	a, b, c, out := make([]uint32, n), make([]uint32, n), make([]uint32, n), make([]uint32, n)
	kitTruthIthVar(a, nVars, 0) // 0xAAAAAAAA
	kitTruthIthVar(b, nVars, 5) // second word all ones
	if a[0] != 0xAAAAAAAA || a[1] != 0xAAAAAAAA || b[0] != 0 || b[1] != ^uint32(0) {
		t.Fatalf("ithvar %x %x", a, b)
	}
	kitTruthAnd(out, a, b, nVars)
	if kitTruthCountOnes(out, nVars) != 16 || kitTruthFindFirstBit(out, nVars) != 33 {
		t.Fatalf("and: ones=%d first=%d", kitTruthCountOnes(out, nVars), kitTruthFindFirstBit(out, nVars))
	}
	kitTruthAndPhase(c, a, b, nVars, 1, 1) // !a & !b
	kitTruthOr(out, c, out, nVars)
	kitTruthNot(c, a, nVars)
	if kitTruthIsImply(out, c, nVars) != 0 { // out contains a&b so it does not imply !a
		t.Fatal("imply")
	}
	kitTruthFill(out, nVars)
	if kitTruthIsConst1(out, nVars) != 1 || kitTruthIsConst0(out, nVars) != 0 || kitTruthFindFirstZero(out, nVars) != -1 {
		t.Fatal("const1")
	}
	kitTruthMux(out, a, b, a, nVars) // ctrl=a: picks b where a=1
	for w := 0; w < n; w++ {
		if out[w] != (a[w]&^a[w])|(b[w]&a[w]) {
			t.Fatal("mux")
		}
	}
	if kitWordFindFirstBit(0) != -1 || kitWordFindFirstBit(8) != 3 || kitWordHasOneBit(8) != 1 || kitWordCountOnes(0xF0F0) != 8 {
		t.Fatal("word ops")
	}
	if kitCubeMask(5) != 31 || kitCubeContains(7, 5) != 1 || kitCubeContains(5, 7) != 0 || kitCubeIsMarked(kitCubeMark(0)) != 1 {
		t.Fatal("cube ops")
	}
	e := kitEdgeCreate(5, 1)
	if kitEdgeToInt(e) != 11 || kitIntToEdge(11) != e || kitIntToEdge_(kitEdgeToInt_(e)) != e {
		t.Fatal("edge")
	}
	if kitSuppIsMinBase(7) != 1 || kitSuppIsMinBase(5) != 0 || kitBitMask(32) != ^uint32(0) {
		t.Fatal("supp/mask")
	}
}
