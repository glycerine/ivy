package gia

import (
	"math/rand"
	"testing"
)

func independentMFFC(p *Gia_Man_t, root int32) map[int32]bool {
	cone := map[int32]bool{}
	stack := []int32{root}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cone[id] {
			continue
		}
		cone[id] = true
		stack = append(stack, giaObjectFanins(p, Gia_ManObj(p, id))...)
	}
	fanouts := make([][]int32, len(p.pObjs))
	for _, obj := range p.pObjs {
		for _, id := range giaObjectFanins(p, obj) {
			fanouts[id] = append(fanouts[id], obj.id)
		}
	}
	result := map[int32]bool{root: true}
	for changed := true; changed; {
		changed = false
		for id := range cone {
			if result[id] || Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 {
				continue
			}
			closed := len(fanouts[id]) > 0
			for _, fanout := range fanouts[id] {
				if !result[fanout] {
					closed = false
					break
				}
			}
			if closed {
				result[id] = true
				changed = true
			}
		}
	}
	return result
}
func TestMFFCIndependentFanoutClosureAndRestore(t *testing.T) {
	rng := rand.New(rand.NewSource(61237))
	for run := 0; run < 25; run++ {
		p := Gia_ManStart(8)
		for i := 0; i < 5; i++ {
			Gia_ManAppendCi(p)
		}
		for i := 0; i < 40; i++ {
			a := 1 + rng.Intn(int(p.nObjs)-1)
			b := 1 + rng.Intn(int(p.nObjs)-1)
			if a == b {
				b = 1 + b%(int(p.nObjs)-1)
			}
			Gia_ManAppendAnd(p, int32(a*2+rng.Intn(2)), int32(b*2+rng.Intn(2)))
		}
		for i := 0; i < 3; i++ {
			Gia_ManAppendCo(p, int32(6+rng.Intn(40))*2)
		}
		Gia_ManCreateRefs(p)
		saved := append([]int32(nil), p.Refs...)
		for id := int32(6); id < 46; id++ {
			want := independentMFFC(p, id)
			got := Gia_NodeMffcSizeMark(p, Gia_ManObj(p, id))
			if int(got) != len(want) {
				t.Fatalf("run%d root%d size%d want%v", run, id, got, want)
			}
			for i, value := range saved {
				if p.Refs[i] != value {
					t.Fatal("reference count not restored")
				}
				if Gia_ObjIsAnd(p.pObjs[i]) != 0 && (Gia_ObjIsTravIdCurrentId(p, int32(i)) != 0) != want[int32(i)] {
					t.Fatal("marked cone mismatch", id, i)
				}
			}
		}
	}
}
func TestMFFCMappingRetainsSiblingBoundaryAndReferenceOwner(t *testing.T) {
	p := Gia_ManStart(8)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	x := Gia_ManAppendAnd(p, 2, 4)
	y := Gia_ManAppendAnd(p, 6, 8)
	root := Gia_ManAppendAnd(p, x, y)
	Gia_ManAppendAnd(p, x, 10)
	Gia_ManAppendAnd(p, y, 10)
	Gia_ManAppendCo(p, root)
	Gia_ManCreateRefs(p)
	old := p.Refs
	saved := append([]int32(nil), old...)
	if got := Gia_NodeMffcMapping(p); got != 3 {
		t.Fatal("both boundary MFFCs must be mapped", got)
	}
	if Vec_IntEntry(p.VMapping, int(x>>1)) == 0 || Vec_IntEntry(p.VMapping, int(y>>1)) == 0 || Vec_IntEntry(p.VMapping, int(root>>1)) == 0 {
		t.Fatal("recursive scratch vector lost sibling boundary")
	}
	if &p.Refs[0] != &old[0] {
		t.Fatal("old reference owner replaced")
	}
	for i, value := range saved {
		if p.Refs[i] != value {
			t.Fatal("old references changed")
		}
	}
}
func TestMFFCRealMuxAndBuffer(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	control := Gia_ManAppendAnd(p, a, b)
	p.EnableMuxes()
	mux := Gia_ManAppendMuxReal(p, control, a, c)
	buffer := Gia_ManAppendBuf(p, mux)
	Gia_ManAppendCo(p, buffer)
	Gia_ManCreateRefs(p)
	if got := Gia_NodeMffcSize(p, Gia_ManObj(p, buffer>>1)); got != 3 {
		t.Fatal("real MUX control or buffer omitted", got)
	}
	support := Vec_IntAlloc(3)
	if got := Gia_NodeMffcSizeSupp(p, Gia_ManObj(p, buffer>>1), support); got != 3 || Vec_IntSize(support) != 3 {
		t.Fatal("real-gate MFFC boundary", got, Vec_IntArray(support))
	}
}
