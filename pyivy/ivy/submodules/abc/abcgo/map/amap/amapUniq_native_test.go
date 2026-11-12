package amap

import "testing"

func TestAmapUniqueGateRules(t *testing.T) {
	lib := &Amap_Lib_t_{vRules3: Vec_IntAlloc(8)}
	if id := Amap_LibCreateVar(lib); id != 0 {
		t.Fatalf("first variable ID = %d", id)
	}
	second := Amap_LibCreateObj(lib)
	second.Type, second.nSuppSize = AMAP_OBJ_PI, 1
	third := Amap_LibCreateObj(lib)
	third.Type, third.nSuppSize = AMAP_OBJ_PI, 1
	and := Amap_LibCreateNode(lib, Abc_Var2Lit(0, 0), Abc_Var2Lit(1, 1), 0)
	if and != 3 {
		t.Fatalf("AND node ID = %d, want 3", and)
	}
	if got := Amap_LibFindNode(lib, Abc_Var2Lit(1, 1), Abc_Var2Lit(0, 0), 0); got != and {
		t.Fatalf("AND rule lookup = %d, want %d", got, and)
	}
	xor := Amap_LibCreateNode(lib, Abc_Var2Lit(0, 0), Abc_Var2Lit(2, 0), 1)
	if got := Amap_LibFindNode(lib, Abc_Var2Lit(2, 0), Abc_Var2Lit(0, 0), 1); got != xor {
		t.Fatalf("XOR rule lookup = %d, want %d", got, xor)
	}
	mux := Amap_LibCreateMux(lib, Abc_Var2Lit(0, 0), Abc_Var2Lit(1, 0), Abc_Var2Lit(2, 0))
	if got := Amap_LibFindMux(lib, 0, 2, 4); got != mux {
		t.Fatalf("mux lookup = %d, want %d", got, mux)
	}
	if got := Amap_LibFindMux(lib, 4, 2, 0); got != -1 {
		t.Fatalf("wrongly permuted mux lookup = %d", got)
	}
}

func TestAmapPackedRulesSortByFanKeyAndKeepSentinel(t *testing.T) {
	rules := Vec_IntAlloc(4)
	Vec_IntPushOrderWithMask(rules, 9<<16|7)
	Vec_IntPushOrderWithMask(rules, 4<<16|2)
	Vec_IntPushOrderWithMask(rules, 6<<16|5)
	items := Vec_IntArray(rules)
	if len(items) != 3 || items[0]&0xffff != 2 || items[1]&0xffff != 5 || items[2]&0xffff != 7 {
		t.Fatalf("packed rule order = %#v", items)
	}
	if got := Vec_IntCheckWithMask(rules, 5); got != 6 {
		t.Fatalf("masked lookup = %d, want 6", got)
	}
	lookup := Amap_LibLookupTableAlloc(ptrVector(rules), 0)
	if len(lookup) != 1 || len(lookup[0]) != 4 || lookup[0][3] != 0 {
		t.Fatalf("lookup table = %#v, want a copied zero-terminated row", lookup)
	}
	lookup[0][0] = 0
	if Vec_IntEntry(rules, 0) == 0 {
		t.Fatal("lookup table unexpectedly aliases the source vector")
	}
}

func ptrVector(values ...any) *Vec_Ptr_t {
	v := Vec_PtrAlloc(int32(len(values)))
	for _, value := range values {
		Vec_PtrPush(v, value)
	}
	return v
}
