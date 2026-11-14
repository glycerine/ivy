package acb

import (
	"fmt"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"reflect"
	"testing"
)

func TestABCSOPRoundTrip(t *testing.T) {
	for bits := uint64(0); bits < 16; bits++ {
		for _, preserveIDs := range []bool{false, true} {
			t.Run(fmt.Sprintf("truth_%x_ids_%v", bits, preserveIDs), func(t *testing.T) {
				n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_LOGIC, baseabc.ABC_FUNC_SOP, 1)
				n.Name = "fixture"
				n.Spec = "fixture.blif"
				defer baseabc.Abc_NtkDelete(n)
				a, b := baseabc.Abc_NtkCreatePi(n), baseabc.Abc_NtkCreatePi(n)
				baseabc.Abc_ObjAssignName(a, "a", "")
				baseabc.Abc_ObjAssignName(b, "b", "")
				gate := baseabc.Abc_NtkCreateNode(n)
				baseabc.Abc_ObjAddFanin(gate, a)
				baseabc.Abc_ObjAddFanin(gate, b)
				var sop string
				for i := 0; i < 4; i++ {
					if bits>>uint(i)&1 != 0 {
						sop += fmt.Sprintf("%d%d 1\n", i&1, i>>1)
					}
				}
				if sop == "" {
					sop = "-- 0\n"
				}
				gate.Data = sop
				output := baseabc.Abc_NtkCreatePo(n)
				baseabc.Abc_ObjAssignName(output, "y", "")
				baseabc.Abc_ObjAddFanin(output, gate)
				var p *Acb_Ntk_t
				if preserveIDs {
					p = Acb_NtkFromAbc(n)
				} else {
					p = Acb_NtkFromAbc2(n)
				}
				defer Acb_ManFree(p.pDesign)
				for i := 0; i < 4; i++ {
					if acbEval(p, i)[0] != (bits>>uint(i)&1 != 0) {
						t.Fatal("ABC import changed truth")
					}
				}
				restored := Acb_NtkToAbc(n, p)
				if restored == nil {
					t.Fatal("restore failed")
				}
				defer baseabc.Abc_NtkDelete(restored)
				for i := 0; i < 4; i++ {
					inputs := map[*baseabc.Object]bool{baseabc.Abc_NtkCi(restored, 0): i&1 != 0, baseabc.Abc_NtkCi(restored, 1): i&2 != 0}
					got, err := restored.Evaluate(baseabc.Abc_NtkCo(restored, 0).Signal(), inputs)
					if err != nil || got != (bits>>uint(i)&1 != 0) {
						t.Fatalf("ABC roundtrip assignment=%d got=%v err=%v", i, got, err)
					}
				}
			})
		}
	}
}

func TestGateDualRailAndMiter(t *testing.T) {
	cases := []struct {
		kind  Acb_ObjType_t
		truth uint64
	}{
		{ABC_OPER_CONST_F, 0}, {ABC_OPER_CONST_T, ^uint64(0)},
		{ABC_OPER_BIT_AND, s_Truths6[0] & s_Truths6[1]},
		{ABC_OPER_BIT_NAND, ^(s_Truths6[0] & s_Truths6[1])},
		{ABC_OPER_BIT_OR, s_Truths6[0] | s_Truths6[1]},
		{ABC_OPER_BIT_NOR, ^(s_Truths6[0] | s_Truths6[1])},
		{ABC_OPER_BIT_XOR, s_Truths6[0] ^ s_Truths6[1]},
		{ABC_OPER_BIT_NXOR, ^(s_Truths6[0] ^ s_Truths6[1])},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.kind), func(t *testing.T) {
			p := acbNetwork(t, 2)
			n := Acb_ObjAlloc(p, tc.kind, 2, 0)
			if tc.kind != ABC_OPER_CONST_F && tc.kind != ABC_OPER_CONST_T {
				Acb_ObjAddFanin(p, n, 1)
				Acb_ObjAddFanin(p, n, 2)
			}
			acbOutput(p, n)
			dual := Acb_NtkGiaDeriveDual(p)
			defer Gia_ManStop(dual)
			for a := 0; a < 4; a++ {
				out, err := dual.Evaluate([]bool{a&1 != 0, a&2 != 0})
				if err != nil || !reflect.DeepEqual(out, []bool{tc.truth>>uint(a)&1 != 0, false}) {
					t.Fatalf("dual a=%d got=%v err=%v", a, out, err)
				}
			}
			for kind := int32(0); kind < 3; kind++ {
				m := Acb_NtkGiaDeriveMiter(dual, dual, kind)
				for a := 0; a < 4; a++ {
					out, err := m.Evaluate([]bool{a&1 != 0, a&2 != 0})
					want := []bool{false}
					if kind == 0 {
						v := tc.truth>>uint(a)&1 != 0
						want = []bool{v, v}
					} else if kind == 1 {
						want = []bool{false, false}
					}
					if err != nil || !reflect.DeepEqual(out, want) {
						t.Fatalf("miter kind=%d a=%d got=%v want=%v err=%v", kind, a, out, want, err)
					}
				}
				Gia_ManStop(m)
			}
		})
	}
}

func TestMffcReferenceRestoration(t *testing.T) {
	p := acbNetwork(t, 2)
	and := s_Truths6[0] & s_Truths6[1]
	a := acbNode(p, and, 1, 2)
	b := acbNode(p, and, a, 1)
	c := acbNode(p, and, a, b)
	acbOutput(p, c)
	Acb_NtkCreateFanout(p)
	refs := Vec_IntStart(Acb_NtkObjNumMax(p))
	for i := int32(1); i < Acb_NtkObjNumMax(p); i++ {
		Vec_IntWriteEntry(refs, i, Acb_ObjFanoutNum(p, i))
	}
	before := append([]int32(nil), Vec_IntArray(refs)...)
	if got := Acb_NtkNodeDeref_rec(refs, p, c); got != 3 {
		t.Fatalf("deref count=%d", got)
	}
	if got := Acb_NtkNodeRef_rec(refs, p, c); got != 3 {
		t.Fatalf("ref count=%d", got)
	}
	if !reflect.DeepEqual(Vec_IntArray(refs), before) {
		t.Fatalf("ref counts not restored: %v want %v", Vec_IntArray(refs), before)
	}
	removed := Vec_IntAlloc(1)
	Vec_IntPush(removed, c)
	if got := Vec_IntArray(Acb_NtkCollectMffc(p, nil, removed)); !reflect.DeepEqual(got, []int32{a, b, c}) {
		t.Fatalf("MFFC=%v", got)
	}
	retained := Vec_IntAlloc(1)
	Vec_IntPush(retained, a)
	if got := Vec_IntArray(Acb_NtkCollectMffc(p, retained, removed)); !reflect.DeepEqual(got, []int32{b, c}) {
		t.Fatalf("retained MFFC=%v", got)
	}
	acbCheckLinks(t, p)
}

func TestNetworkDuplicationAndMetadata(t *testing.T) {
	p := acbNetwork(t, 2)
	id := acbNode(p, s_Truths6[0]&s_Truths6[1], 1, 2)
	acbOutput(p, id)
	Acb_NtkAdd(p.pDesign, p)
	Acb_NtkCleanObjNames(p)
	Acb_NtkCleanObjWeights(p)
	Acb_NtkCleanObjRanges(p)
	Acb_NtkCleanObjAttrs(p)
	for i := int32(1); i < Acb_NtkObjNumMax(p); i++ {
		Acb_ObjSetName(p, i, Abc_NamStrFindOrAdd(p.pDesign.pStrs, fmt.Sprintf("n%d", i), nil))
		Acb_ObjSetWeight(p, i, i*3)
	}
	Acb_ObjHashRange(p, id, 7, 0)
	Acb_ObjSetRangeSign(p, id, Acb_ObjRangeId(p, id)|1)
	attrs := []int32{7, 42, 9, 19}
	Acb_ObjSetAttrs(p, id, attrs, int32(len(attrs)))
	if Acb_ObjRangeSize(p, id) != 8 || Acb_ObjSigned(p, id) != 1 {
		t.Fatal("range or signedness")
	}
	qman := Acb_ManDup(p.pDesign, nil)
	defer Acb_ManFree(qman)
	q := Acb_ManRoot(qman)
	if Acb_NtkCiNum(q) != 2 || Acb_NtkCoNum(q) != 1 {
		t.Fatal("duplicate terminals")
	}
	qi := Acb_ObjCopy(p, id)
	if Acb_ObjWeight(q, qi) != 9 || Acb_ObjNameStr(q, qi) != "n3" || Acb_ObjLeft(q, qi) != 7 || Acb_ObjRight(q, qi) != 0 {
		t.Fatal("metadata not duplicated")
	}
	if Acb_ObjAttrValue(q, qi, 7) != 42 || Acb_ObjAttrValue(q, qi, 9) != 19 {
		t.Fatal("attributes not duplicated")
	}
	Vec_IntWriteEntry(&q.vObjWeight, qi, 99)
	if Acb_ObjWeight(p, id) != 9 {
		t.Fatal("duplicate aliases source weights")
	}
	Acb_ManSetMap(qman, 100, 12)
	Acb_ManSetMap2(qman, 100, 34)
	Acb_ManCleanMap(qman)
	if Acb_ManGetMap(qman, 100) != 0 || Acb_ManGetMap2(qman, 100) != 34 {
		t.Fatal("map cleanup corrupted independent map")
	}
	Acb_ManCleanMap2(qman)
	if Acb_ManGetMap2(qman, 100) != 0 {
		t.Fatal("second map not cleared")
	}
}

func TestBoxFinalOutputCountAndDeletion(t *testing.T) {
	p := acbNetwork(t, 0)
	box := Acb_ObjAlloc(p, ABC_OPER_BOX, 0, 2)
	if got := Acb_ObjFonNum(p, box); got != 2 {
		t.Fatalf("box output count=%d want 2", got)
	}
	Acb_ObjDelete(p, box)
	for i := box; i < Acb_NtkObjNumMax(p); i++ {
		if Acb_ObjType(p, i) != ABC_OPER_NONE {
			t.Fatalf("object %d survived box deletion", i)
		}
	}
}
