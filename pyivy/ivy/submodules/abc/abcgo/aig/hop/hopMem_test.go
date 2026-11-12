package hop

import "testing"

func TestNodePageFreeList(t *testing.T) {
	p := new(Hop_Man_t)
	Hop_ManStartMemory(p)
	Hop_ManAddMemory(p)
	seen := make(map[*Hop_Obj_t]bool)
	for node := p.pListFree; node != nil; node = node.pNext {
		if seen[node] {
			t.Fatal("free-list cycle")
		}
		seen[node] = true
	}
	if len(seen) != 1<<IVY_PAGE_SIZE {
		t.Fatalf("nodes: %d", len(seen))
	}
	if len(p.nodePages) != 1 {
		t.Fatal("page ownership not retained")
	}
	p.pListFree = nil
	Hop_ManAddMemory(p)
	if len(p.nodePages) != 2 {
		t.Fatal("second page missing")
	}
	Hop_ManStopMemory(p)
	if p.nodePages != nil || p.pListFree != nil {
		t.Fatal("pool retained after stop")
	}
}
