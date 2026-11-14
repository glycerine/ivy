package gia

import "testing"

func TestChoiceTfiControlDependencies(t *testing.T) {
	for name, check := range map[string]func(*Gia_Man_t, *Gia_Obj_t, *Gia_Obj_t) int32{
		"gia": Gia_ObjCheckTfi,
		"hcd": Hcd_ObjCheckTfi,
	} {
		t.Run(name, func(t *testing.T) {
			p := Gia_ManStart(1)
			a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			old := Gia_ManAppendAnd(p, a, b)
			p.EnableMuxes()
			root := Gia_ManAppendMuxReal(p, old^1, c, d)
			p.StartRepresentatives()
			p.PNexts = make([]int32, p.nObjs)
			if got := check(p, Gia_ManObj(p, old>>1), Gia_ManObj(p, root>>1)); got != 1 {
				t.Fatalf("control-only dependency = %d, want 1", got)
			}
			for _, obj := range p.pObjs {
				if obj.fMark0 != 0 {
					t.Fatal("TFI query retained a traversal mark")
				}
			}
		})
	}
}

func TestChoiceTfiConstantLeaves(t *testing.T) {
	for name, check := range map[string]func(*Gia_Man_t, *Gia_Obj_t, *Gia_Obj_t) int32{
		"gia": Gia_ObjCheckTfi,
		"hcd": Hcd_ObjCheckTfi,
	} {
		t.Run(name, func(t *testing.T) {
			p := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			old := Gia_ManAppendAnd(p, a, b)
			root := Gia_ManAppendAnd(p, 0, a)
			p.PNexts = make([]int32, p.nObjs)
			if got := check(p, Gia_ManObj(p, old>>1), Gia_ManObj(p, root>>1)); got != 0 {
				t.Fatalf("unrelated cone = %d, want 0", got)
			}
		})
	}
}

func TestChoiceTfiMixedGateReachability(t *testing.T) {
	p := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.EnableMuxes()
	// Keep an independent adjacency list from the construction recipe, rather
	// than deriving the expected cone from the same GIA accessors under test.
	deps := make(map[int32][]int32)
	record := func(lit int32, fanins ...int32) int32 {
		for _, fanin := range fanins {
			deps[lit>>1] = append(deps[lit>>1], fanin>>1)
		}
		return lit
	}
	x := record(Gia_ManAppendAnd(p, a, b), a, b)
	y := record(Gia_ManAppendAnd(p, c, d), c, d)
	z := record(Gia_ManAppendXorReal(p, x^1, c), x, c)
	control := record(Gia_ManAppendBuf(p, z^1), z)
	mux := record(Gia_ManAppendMuxReal(p, control, d^1, a), control, d, a)
	root := record(Gia_ManAppendAnd(p, mux, y^1), mux, y)
	record(Gia_ManAppendCo(p, root^1), root)
	record(Gia_ManAppendAnd(p, 0, a), 0, a)
	for _, obj := range p.pObjs {
		obj.fMark1 = uint32(obj.id & 1)
	}
	for old := range deps {
		for node := int32(0); node < p.nObjs; node++ {
			var want int32
			seen := make(map[int32]bool)
			queue := []int32{node}
			for len(queue) > 0 {
				id := queue[0]
				queue = queue[1:]
				if id == old {
					want = 1
					break
				}
				if !seen[id] {
					seen[id] = true
					queue = append(queue, deps[id]...)
				}
			}
			for _, check := range []func(*Gia_Man_t, *Gia_Obj_t, *Gia_Obj_t) int32{Gia_ObjCheckTfi, Hcd_ObjCheckTfi} {
				if got := check(p, Gia_ManObj(p, old), Gia_ManObj(p, node)); got != want {
					t.Fatalf("old %d node %d: reachability %d, want %d", old, node, got, want)
				}
				for _, obj := range p.pObjs {
					if obj.fMark0 != 0 || obj.fMark1 != uint32(obj.id&1) {
						t.Fatal("TFI query retained marks or changed another mark bit")
					}
				}
			}
		}
	}
}
