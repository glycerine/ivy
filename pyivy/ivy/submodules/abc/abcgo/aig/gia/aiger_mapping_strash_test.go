package gia

import "testing"

// Gia_AigerRead with structural hashing reads a mapped AIGER file, warns
// that hashing invalidated the mapping, and drops it (giaAiger.c). The file
// was written by native ABC after Lf mapping of sort16.
func TestAigerReadStrashDropsMapping(t *testing.T) {
	p, err := Gia_AigerRead("test_vectors/native_mapped_sort16.aig", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManHasMapping(p) != 0 {
		t.Fatal("mapping survived structural hashing")
	}
	if Gia_ManAndNum(p) != 958 { // native: "&r file; &ps" reports and = 958
		t.Fatalf("got %d ANDs, native 958", Gia_ManAndNum(p))
	}
	q, err := Gia_AigerRead("test_vectors/native_mapped_sort16.aig", 0, 1, 0)
	if err != nil || Gia_ManHasMapping(q) == 0 {
		t.Fatalf("reading with SkipStrash lost the mapping: %v", err)
	}
}

// Gia_AigerRead names the result after the file, overriding any name stored
// in the file (this one stores "st16").
func TestAigerReadNamesAfterFile(t *testing.T) {
	p, err := Gia_AigerRead("test_vectors/native_mapped_sort16.aig", 0, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.pName != "test_vectors/native_mapped_sort16" { // Gia_FileNameGeneric keeps the directory
		t.Fatalf("name %q, want the file name", p.pName)
	}
}
