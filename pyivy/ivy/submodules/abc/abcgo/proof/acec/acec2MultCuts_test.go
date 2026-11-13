package acec

import "testing"

func testCut(mask int) *Sdb_Cut_t {
	c := &Sdb_Cut_t{}
	for i := 0; i < 6; i++ {
		if mask&(1<<i) != 0 {
			c.pLeaves[c.nLeaves] = int32(i)
			c.nLeaves++
		}
	}
	c.Sign = Sdb_CutGetSign(c)
	return c
}
func TestCutMergeAndContainmentExhaustive(t *testing.T) {
	for a := 0; a < 64; a++ {
		for b := 0; b < 64; b++ {
			ca, cb := testCut(a), testCut(b)
			want := a | b
			for limit := int32(0); limit <= 6; limit++ {
				var count int32
				for i := 0; i < 6; i++ {
					if want&(1<<i) != 0 {
						count++
					}
				}
				var out Sdb_Cut_t
				got := Sdb_CutMergeOrder(ca, cb, &out, limit)
				if (got != 0) != (count <= limit) {
					t.Fatalf("%x | %x, limit %d: %d", a, b, limit, got)
				}
				if got != 0 {
					if out.nLeaves != uint32(count) || out.Sign != uint64(want) || out.iFunc != -1 {
						t.Fatalf("invalid result %x | %x: %+v", a, b, out)
					}
					for i := uint32(1); i < out.nLeaves; i++ {
						if out.pLeaves[i-1] >= out.pLeaves[i] {
							t.Fatal("unsorted merge")
						}
					}
				}
			}
			got := Sdb_CutCheck(ca, cb) != 0
			if got != (a&b == b) {
				t.Fatalf("containment %x %x: %v", a, b, got)
			}
		}
	}
}
func TestCutSetPrunesSupersetsAndSorts(t *testing.T) {
	a, b, c := testCut(7), testCut(24), testCut(3)
	a.nTreeLeaves = 3
	b.nTreeLeaves = 1
	c.nTreeLeaves = 0
	cuts := []*Sdb_Cut_t{b, a, c}
	n := Sdb_CutSetAddCut(cuts, 2, 5)
	if n != 2 || cuts[0] != c || cuts[1] != b {
		t.Fatalf("n=%d cuts=%v", n, cuts)
	}
	if Sdb_CutSetCheckArray(cuts, n) != 1 {
		t.Fatal("invalid cut set")
	}
	if Sdb_CutSetLastCutIsContained([]*Sdb_Cut_t{testCut(3), testCut(7)}, 1) != 1 {
		t.Fatal("failed to reject superset")
	}
}

// acec2Mult.c compares full-size cuts before copying each matching leaf.
// A mismatch at the first leaf returns without modifying the destination.
func TestCutMergeFullSizeMismatchPreservesDestination(t *testing.T) {
	a, b := testCut(5), testCut(6)
	out := Sdb_Cut_t{pLeaves: [SDB_MAX_CUTSIZE]int32{91, 92}, iFunc: 42, Sign: 17}
	if got := Sdb_CutMergeOrder(a, b, &out, 2); got != 0 {
		t.Fatalf("merge returned %d", got)
	}
	if out.pLeaves[0] != 91 || out.pLeaves[1] != 92 || out.iFunc != 42 || out.Sign != 17 {
		t.Fatalf("failed merge changed destination: %+v", out)
	}
}

func TestCutMergeFailedTailAndAlternateBehavior(t *testing.T) {
	a, b := testCut(1), testCut(14)
	out := Sdb_Cut_t{pLeaves: [SDB_MAX_CUTSIZE]int32{91, 92, 93}}
	if Sdb_CutMergeOrder(a, b, &out, 2) != 0 || out.pLeaves[0] != 0 || out.pLeaves[1] != 92 {
		t.Fatalf("optimized tail should remain untouched: %+v", out)
	}
	out = Sdb_Cut_t{pLeaves: [SDB_MAX_CUTSIZE]int32{91, 92, 93}}
	if Sdb_CutMergeOrder2(a, b, &out, 2) != 0 || out.pLeaves[0] != 0 || out.pLeaves[1] != 1 || out.pLeaves[2] != 93 {
		t.Fatalf("alternate merge should copy up to limit: %+v", out)
	}
}
