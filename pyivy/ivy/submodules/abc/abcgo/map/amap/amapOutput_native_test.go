package amap

import "testing"

func TestAmapOutputSerializesInputAndOutput(t *testing.T) {
	manager := Amap_ManStart(1)
	manager.pLib = &Amap_Lib_t_{}
	Amap_ManCreateConst1(manager)
	input := Amap_ManCreatePi(manager)
	input.nFouts[0] = 1
	Amap_ManCreatePo(manager, input)

	result := Amap_ManProduceMapped(manager)
	if result == nil || Vec_PtrSize(result) != 3 {
		t.Fatalf("mapped records = %d, want two records plus memory marker", Vec_PtrSize(result))
	}
	pi := Vec_PtrEntry(result, 0).(*Amap_Out_t)
	po := Vec_PtrEntry(result, 1).(*Amap_Out_t)
	if pi.Type != -1 || pi.nFans != 0 {
		t.Fatalf("PI record = %#v", pi)
	}
	if po.Type != 1 || po.nFans != 1 || po.pFans[0] != 0 {
		t.Fatalf("PO record = %#v, want driver index 0", po)
	}
	mem := Vec_PtrEntry(result, 2).(*Aig_MmFlex_t)
	Aig_MmFlexStop(mem, 0)
	Vec_PtrFree(result)
	Amap_ManStop(manager)
}

func TestAmapOutputRecordOwnsFaninSliceAndName(t *testing.T) {
	gate := &Amap_Gat_t_{pName: "INV", nPins: 1}
	record := Amap_OutputStructAlloc(nil, gate)
	if record.pName != "INV" || len(record.pFans) != 1 || record.pFans[0] != -1 {
		t.Fatalf("allocated record = %#v", record)
	}
	record.pFans[0] = 7
	other := Amap_OutputStructAlloc(nil, gate)
	if other.pFans[0] != -1 {
		t.Fatal("output records share fanin storage")
	}
}
