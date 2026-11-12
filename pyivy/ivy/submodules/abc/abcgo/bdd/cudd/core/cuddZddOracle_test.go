package core

import "testing"

// This checksum comes from the independent original CUDD 2.4.2 implementation,
// using testdata/zdd_setops_oracle.c. No C compiler or C objects are needed to
// execute this fixture. Operation order and the numeric family-bit encoding are
// identical in the C harness and here; pointer layouts never enter the checksum.
func TestZDDOriginalCUDDOracleChecksum(t *testing.T) {
	dd := Cudd_Init(0, 3, 256, 256, 0)
	families := make([]*DdNode, 256)
	for i := range families {
		families[i] = zddFromTable(dd, uint16(i))
	}
	hash := uint64(14695981039346656037)
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			f, g, h := families[a], families[b], families[(a+3*b)&255]
			inclusion := uint16(0)
			if Cudd_zddDiffConst(dd, f, g) == dd.Zero {
				inclusion = 1
			}
			results := []uint16{
				zddTable(dd, Cudd_zddUnion(dd, f, g)),
				zddTable(dd, Cudd_zddIntersect(dd, f, g)),
				zddTable(dd, Cudd_zddDiff(dd, f, g)),
				zddTable(dd, Cudd_zddIte(dd, f, g, h)),
				inclusion,
			}
			for _, value := range results {
				hash ^= uint64(value)
				hash *= 1099511628211
			}
		}
	}
	const originalCUDDChecksum uint64 = 0x7292090e7f59522c
	if hash != originalCUDDChecksum {
		t.Fatalf("ZDD results differ from original CUDD: got %016x, want %016x", hash, originalCUDDChecksum)
	}
}
