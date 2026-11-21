package ivy

import (
	"strconv"
	"testing"
)

func TestIvyRecycleRejectsCopiedObject(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	live := Ivy_ObjCreatePi(p)
	copyOfLive := *live
	func() {
		defer func() {
			if recover() == nil {
				t.Error("accepted a value copy outside the manager's pages")
			}
		}()
		Ivy_ManRecycleMemory(p, &copyOfLive)
	}()
	if next := Ivy_ManFetchMemory(p); next == &copyOfLive || next == live {
		t.Fatal("returned a copied or still-live object from the free list")
	}
}

func TestIvyRecycleAcrossPageBoundary(t *testing.T) {
	p := &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	var saved []*Ivy_Obj_t
	for i := range (1 << IVY_PAGE_SIZE) + 5 {
		n := Ivy_ManFetchMemory(p)
		if i == 0 || i == 1023 || i >= (1<<IVY_PAGE_SIZE)-1 {
			Ivy_Not(n)
			saved = append(saved, n)
		}
	}
	for _, n := range saved {
		Ivy_ManRecycleMemory(p, n)
	}
	for i := len(saved) - 1; i >= 0; i-- {
		if n := Ivy_ManFetchMemory(p); n != saved[i] || n.recycled || n.neg.regular != n {
			t.Fatal("interleaved pages lost free-list order or complemented identity")
		}
	}
	extra := Ivy_ManFetchMemory(p)
	if len(p.pages) != 2 || extra == saved[len(saved)-1] {
		t.Fatal("failed to resume the partially issued page")
	}
	Ivy_ManStopMemory(p)
	for _, n := range append(saved, extra) {
		if n.owner != nil || !n.recycled || n.Id != -1 {
			t.Fatal("retained handle still live after stop")
		}
	}
}

func TestIvyAddMemoryRequiresExhaustedPage(t *testing.T) {
	p := &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	defer Ivy_ManStopMemory(p)
	Ivy_ManAddMemory(p)
	rejectAdd := func() {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("allocated another page with a free slot available")
			}
		}()
		Ivy_ManAddMemory(p)
	}
	rejectAdd()
	var last *Ivy_Obj_t
	for range 1 << IVY_PAGE_SIZE {
		last = Ivy_ManFetchMemory(p)
	}
	Ivy_ManRecycleMemory(p, last)
	rejectAdd()
	if Ivy_ManFetchMemory(p) != last {
		t.Fatal("lost the recycled slot on an exhausted page")
	}
	Ivy_ManAddMemory(p)
	if len(p.pages) != 2 || Ivy_ManFetchMemory(p) == last {
		t.Fatal("explicit page allocation failed")
	}
}

func BenchmarkIvyFetchSlots(b *testing.B) {
	for _, count := range []int{16, 5000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				p := &Ivy_Man_t{}
				Ivy_ManStartMemory(p)
				var last *Ivy_Obj_t
				for i := range count {
					last = Ivy_ManFetchMemory(p)
					last.Id = int32(i)
				}
				Ivy_ManRecycleMemory(p, last)
				if got := Ivy_ManFetchMemory(p); got != last || got.Id != 0 {
					b.Fatal("wrong recycled slot")
				}
				Ivy_ManStopMemory(p)
			}
		})
	}
}
