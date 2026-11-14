package amap

import "testing"

func TestAmapOutputAccessors(t *testing.T) {
	output := &Amap_Out_t{pName: "nand2", Type: 0, nFans: 2, pFans: []int32{3, 5}}
	if Amap_OutName(output) != "nand2" || Amap_OutType(output) != 0 || Amap_OutFanCount(output) != 2 || Amap_OutFan(output, 1) != 5 {
		t.Fatal("output accessors did not preserve the mapped record")
	}
}
