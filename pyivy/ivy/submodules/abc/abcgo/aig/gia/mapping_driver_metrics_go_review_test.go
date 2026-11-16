package gia_test

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Check real mapper results using topological scalar levels and an explicit
// set/stack cone enumeration, independently of the recursive production walks.
func mappingDriverReviewCheckMetrics(t *testing.T, p *gia.Gia_Man_t) {
	t.Helper()
	n := int(gia.Gia_ManObjNum(p))
	refs, levels, physical := make([]int32, n), make([]int32, n), make([]int32, n)
	var luts, edges, nodes, overlap, maximum, physicalMax int32
	sizeMax := int32(-1)
	label, visited := vec.Vec_StrStart(n), vec.Vec_IntAlloc(16)
	defer vec.Vec_StrFree(label)
	defer vec.Vec_IntFree(visited)
	for id := int32(1); id < int32(n); id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2, 3:
			nodes++
			physical[id] = max(physical[a>>1], physical[b>>1]) + 1
		case 4:
			nodes++
			physical[id] = max(physical[a>>1], physical[b>>1], physical[c>>1]) + 1
		case 5:
			nodes++
			physical[id] = physical[a>>1] + 1
		case 6:
			physical[id] = physical[a>>1]
		}
		if kind >= 2 && kind <= 5 {
			if sibling := gia.Gia_ObjSibl(p, id); sibling != 0 {
				physical[id] = max(physical[id], physical[sibling])
			}
		}
		if gia.Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		leaves := gia.Gia_ObjLutFanins(p, id)
		luts++
		edges += int32(len(leaves))
		sizeMax = max(sizeMax, int32(len(leaves)))
		boundary := make(map[int32]bool, len(leaves))
		for _, leaf := range leaves {
			refs[leaf]++
			levels[id] = max(levels[id], levels[leaf])
			boundary[leaf] = true
		}
		levels[id]++
		seen := map[int32]bool{}
		pending := []int32{id}
		for len(pending) != 0 {
			last := len(pending) - 1
			current := pending[last]
			pending = pending[:last]
			if boundary[current] || seen[current] {
				continue
			}
			seen[current] = true
			x, y, z, gate := p.Gate(current)
			switch gate {
			case 2, 3:
				pending = append(pending, x>>1, y>>1)
			case 4:
				pending = append(pending, x>>1, y>>1, z>>1)
			case 5:
				pending = append(pending, x>>1)
			default:
				t.Fatalf("cut%d reaches terminal%d outside its boundary", id, current)
			}
		}
		wantCone := int32(len(seen))
		overlap += wantCone
		for pass := 0; pass < 2; pass++ {
			if gia.Gia_ManComputeOverlapOne(p, id) != wantCone || gia.Gia_ManComputeOverlap2One(p, id, label, visited) != wantCone {
				t.Fatalf("cut%d cone count differs from scalar enumeration", id)
			}
			for i := 0; i < n; i++ {
				if vec.Vec_StrEntry(label, i) != 0 {
					t.Fatal("overlap scratch retained a mark")
				}
			}
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		obj := gia.Gia_ManCo(p, i)
		id := gia.Gia_ObjId(p, obj)
		driver := gia.Gia_ObjFaninId0p(p, obj)
		refs[driver]++
		levels[id] = levels[driver]
		maximum = max(maximum, levels[id])
		physicalMax = max(physicalMax, physical[id])
	}
	var actual []int32
	if gia.Gia_ManLutNum(p) != luts || gia.Gia_ManLutFaninCount(p) != edges || gia.Gia_ManLutSizeMax(p) != sizeMax || gia.Gia_ManLutLevel(p, &actual) != maximum || !slices.Equal(actual, levels) {
		t.Fatal("mapped summary/levels differ from scalar result")
	}
	var gotLuts, gotEdges, gotDepth int32
	gia.Gia_ManLutParams(p, &gotLuts, &gotEdges, &gotDepth)
	if gotLuts != luts || gotEdges != edges || gotDepth != maximum {
		t.Fatal("mapped parameters differ from scalar result")
	}
	if gia.Gia_ManComputeOverlap(p) != overlap-nodes || gia.Gia_ManComputeOverlap2(p) != overlap-nodes {
		t.Fatal("total overlap differs from scalar cones minus physical nodes")
	}
	if gia.Gia_ManChoiceLevel(p) != physicalMax {
		t.Fatal("physical choice depth differs from scalar topological depth")
	}
	gia.Gia_ManSetRefsMapped(p)
	gia.Gia_ManSetLutRefs(p)
	if !slices.Equal(p.Refs, refs) || !slices.Equal(p.LUTReferenceCounts(), refs) {
		t.Fatal("mapped reference counts differ from scalar counts")
	}
	p.Refs[0]++
	if gia.Gia_ObjLutRefNumId(p, 0) != refs[0] {
		t.Fatal("mapped reference arrays alias")
	}
	gia.Gia_ManSetRefsMapped(p)
	if !slices.Equal(p.Refs, refs) {
		t.Fatal("reference replacement retained old counts")
	}
	actual[0] = 99
	var again []int32
	gia.Gia_ManLutLevel(p, &again)
	if !slices.Equal(again, levels) {
		t.Fatal("mapped level results alias")
	}
}
