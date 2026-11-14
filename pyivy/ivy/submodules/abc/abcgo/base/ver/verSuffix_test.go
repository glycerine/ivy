package ver

import "testing"

func TestSuffixTablePreservesFirstDeclaration(t *testing.T) {
	p := &Ver_Man_t{}
	var hi, lo int32
	Ver_ParseLookupSuffix(p, "bus", &hi, &lo)
	if hi != -1 || lo != -1 {
		t.Fatal(hi, lo)
	}
	Ver_ParseInsertsSuffix(p, "bus", 127, 3)
	Ver_ParseInsertsSuffix(p, "bus", 4, 2)
	Ver_ParseLookupSuffix(p, "bus", &hi, &lo)
	if hi != 127 || lo != 3 {
		t.Fatal(hi, lo)
	}
	Ver_ParseLookupSuffix(p, "missing", &hi, &lo)
	if hi != -1 || lo != -1 {
		t.Fatal(hi, lo)
	}
	Ver_ParseRemoveSuffixTable(p)
	if p.tName2Suffix != nil {
		t.Fatal("table not removed")
	}
}
