package kit

import "testing"

func TestReviewTrivialDsdRootPreserved(t *testing.T) {
	for _, truth := range []uint32{0, ^uint32(0), 0xaaaaaaaa, 0x55555555} {
		original := Kit_DsdDecompose([]uint32{truth}, 3)
		root := Kit_DsdNtkRoot(original)
		for _, transform := range []func(*Kit_DsdNtk_t) *Kit_DsdNtk_t{Kit_DsdExpand, func(n *Kit_DsdNtk_t) *Kit_DsdNtk_t { return Kit_DsdShrink(n, nil) }} {
			result := transform(original)
			node := Kit_DsdNtkRoot(result)
			if node == nil || root == nil || node.Type != root.Type || node.nFans != root.nFans || result.Root&1 != original.Root&1 {
				t.Fatalf("trivial root representation changed truth=%x", truth)
			}
			got := make([]uint32, 1)
			Kit_DsdTruth(result, got)
			if got[0] != truth {
				t.Fatal("truth changed")
			}
		}
	}
}

func TestReviewGraphAndDsdDerivedCounts(t *testing.T) {
	g := Kit_GraphCreate(2)
	root := Kit_EdgeCreate(0, 0)
	for i := 0; i < 100; i++ {
		root = Kit_GraphAddNodeAnd(g, root, Kit_EdgeCreate(1, 0))
		if Kit_GraphNodeNum(g) != int32(i+1) || Kit_GraphNodeLast(g) != g.pNodes[len(g.pNodes)-1] {
			t.Fatal("derived graph count")
		}
	}
	Kit_GraphSetRoot(g, root)
	if Kit_GraphToTruth(g) != 0x88888888 {
		t.Fatal("graph truth")
	}
	ntk := Kit_DsdNtkAlloc(3)
	for i := 0; i < 20; i++ {
		Kit_DsdObjAlloc(ntk, KIT_DSD_CONST1, 0)
		if Kit_DsdNtkObjNum(ntk) != int32(3+len(ntk.pNodes)) {
			t.Fatal("derived DSD count")
		}
	}
}
func TestReviewSopDerivedCountAndFailureReuse(t *testing.T) {
	var s Kit_Sop_t
	Kit_SopCreate(&s, []uint32{1, 2, 3}, 3, nil)
	Kit_SopShrink(&s, 1)
	Kit_SopPushCube(&s, 4)
	if Kit_SopCubeNum(&s) != 2 || len(s.pCubes) != 2 || Kit_SopCube(&s, 1) != 4 {
		t.Fatal("SOP count")
	}
	memory := make([]uint32, KIT_ISOP_MEM_LIMIT)
	if Kit_TruthIsop_rec([]uint32{0xaaaaaaaa}, []uint32{0xaaaaaaaa}, 1, &s, &memory) != nil || Kit_SopCubeNum(&s) != -1 {
		t.Fatal("expected arena exhaustion")
	}
	memory = nil
	if out := Kit_TruthIsop_rec([]uint32{0xaaaaaaaa}, []uint32{0xaaaaaaaa}, 1, &s, &memory); out == nil || out[0] != 0xaaaaaaaa || Kit_SopCubeNum(&s) != 1 {
		t.Fatal("failed state survived reuse")
	}
}
func TestReviewCloudCacheBudget(t *testing.T) {
	dd := Cloud_Init(3, 12)
	defer Cloud_Quit(dd)
	Cloud_CacheAllocate(dd, CLOUD_OPER_AND, 12)
	if dd.cacheLimit != 1 {
		t.Fatal("cache ratio not applied")
	}
	for f := uint32(0); f < 256; f++ {
		left := cloudFromTruth(dd, f*0x01010101, 3)
		other := (f*73 + 19) & 255
		right := cloudFromTruth(dd, other*0x01010101, 3)
		node := Cloud_bddAnd(dd, left, right)
		if node == nil {
			t.Fatal("BDD limit")
		}
		if len(dd.andCache) > dd.cacheLimit {
			t.Fatal("cache exceeds limit")
		}
		for a := 0; a < 8; a++ {
			values := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
			if Cloud_Evaluate(dd, node, values) != ((f&other)>>uint(a)&1 != 0) {
				t.Fatal("cache eviction changed function")
			}
		}
	}
	if dd.nCacheMisses == 0 {
		t.Fatal("test did not exercise the cache")
	}
	if len(dd.tUnique) == 0 {
		t.Fatal("missing nodes")
	}
}
