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
