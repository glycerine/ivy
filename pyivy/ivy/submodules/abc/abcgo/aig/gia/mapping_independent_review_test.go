package gia

import "testing"

func TestMappingIndependentMetricsAndOwnership(t *testing.T) {
	for seed := uint32(1); seed <= 128; seed++ {
		for mode := 0; mode < 3; mode++ {
			p := mappingReviewOrdinary(seed, mode)
			Gia_ManSetRegNum(p, int32(seed%3))
			refs, levels := make([]int32, p.nObjs), make([]int32, p.nObjs)
			var luts, edges, maximum, overlap int32
			overlap = -Gia_ManAndNum(p)
			for id := int32(1); id < p.nObjs; id++ {
				if Gia_ObjIsLut(p, id) == 0 {
					continue
				}
				luts++
				leaves := Gia_ObjLutFanins(p, id)
				edges += int32(len(leaves))
				cut := map[int32]bool{}
				for _, leaf := range leaves {
					refs[leaf]++
					cut[leaf] = true
					if levels[id] < levels[leaf] {
						levels[id] = levels[leaf]
					}
				}
				levels[id]++
				// Enumerate the physical cone using a set and an explicit stack, rather
				// than either recursive production traversal/marking representation.
				seen := map[int32]bool{}
				pending := []int32{id}
				for len(pending) > 0 {
					objID := pending[len(pending)-1]
					pending = pending[:len(pending)-1]
					if cut[objID] || seen[objID] {
						continue
					}
					seen[objID] = true
					obj := p.pObjs[objID]
					pending = append(pending, Gia_ObjFaninId0(obj, objID), Gia_ObjFaninId1(obj, objID))
				}
				expected := int32(len(seen))
				overlap += expected
				if got := Gia_ManComputeOverlapOne(p, id); got != expected {
					t.Fatalf("seed%d mode%d cone%d count%d want%d", seed, mode, id, got, expected)
				}
			}
			for i := int32(0); i < Gia_ManCoNum(p); i++ {
				co := Gia_ManCo(p, i)
				driver := Gia_ObjFaninId0(co, co.id)
				refs[driver]++
				levels[co.id] = levels[driver]
				if maximum < levels[co.id] {
					maximum = levels[co.id]
				}
			}
			var gotLevels []int32
			if Gia_ManLutLevel(p, &gotLevels) != maximum || Gia_ManLutNum(p) != luts || Gia_ManLutFaninCount(p) != edges || Gia_ManComputeOverlap(p) != overlap || Gia_ManComputeOverlap2(p) != overlap {
				t.Fatalf("seed%d mode%d summary mismatch", seed, mode)
			}
			for i, want := range levels {
				if gotLevels[i] != want {
					t.Fatal("independent mapped level mismatch")
				}
			}
			Gia_ManSetRefsMapped(p)
			Gia_ManSetLutRefs(p)
			for i, want := range refs {
				if p.Refs[i] != want || p.pLutRefs[i] != want {
					t.Fatal("independent reference counts mismatch")
				}
			}
			p.Refs[0]++
			if p.pLutRefs[0] != refs[0] {
				t.Fatal("reference arrays alias")
			}
			Gia_ManSetRefsMapped(p)
			if p.Refs[0] != refs[0] {
				t.Fatal("reference replacement retains stale counts")
			}
			gotLevels[0] = 99
			var second []int32
			Gia_ManLutLevel(p, &second)
			if second[0] != 0 {
				t.Fatal("level results alias")
			}
			Gia_ManStop(p)
		}
	}
}
