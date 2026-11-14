package gia

import (
	"fmt"
	"slices"
	"testing"
)

func checkDynamicFanoutRing(t *testing.T, p *Gia_Man_t, source int32, want []int32) {
	t.Helper()
	first := Gia_ObjFanout0Int(p, source)
	var got []int32
	if first != 0 {
		edge := first
		for {
			got = append(got, edge)
			if len(got) > len(want)+1 {
				t.Fatalf("fanout ring failed to close: %v, want %v", got, want)
			}
			next := Gia_ObjFanoutNext(p, edge)
			if next == 0 || *Gia_FanoutPrev(p.pFanData, next) != edge {
				t.Fatalf("broken reciprocal link after %d", edge)
			}
			edge = next
			if edge == first {
				break
			}
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("fanout edges %v, want %v", got, want)
	}
}

func TestDynamicFanoutsRepeatedPins(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		for signs := int32(0); signs < 4; signs++ {
			t.Run(fmt.Sprintf("incremental%t_signs%d", incremental, signs), func(t *testing.T) {
				p := Gia_ManStart(1)
				p.fGiaSimple = 1
				a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
				if incremental {
					Gia_ManFanoutStart(p)
				}
				left := Gia_ManAppendAnd(p, a, b)
				repeated := Gia_ManAppendAnd(p, a^(signs&1), a^((signs>>1)&1))
				right := Gia_ManAppendAnd(p, a, b^1)
				if !incremental {
					Gia_ManFanoutStart(p)
				}
				want := []int32{left, repeated}
				// Same-phase pins are represented as a buffer with one arc;
				// opposite phases are a real two-arc contradiction AND.
				count := 1
				if signs == 1 || signs == 2 {
					want = append(want, repeated|1)
					count = 2
				}
				want = append(want, right)
				checkDynamicFanoutRing(t, p, a>>1, want)
				for i := 0; i < count; i++ {
					Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, repeated>>1))
					remaining := []int32{left}
					if i == 0 && count == 2 {
						remaining = append(remaining, repeated|1)
					}
					remaining = append(remaining, right)
					checkDynamicFanoutRing(t, p, a>>1, remaining)
				}
				checkDynamicFanoutRing(t, p, a>>1, []int32{left, right})
				for i := 0; i < count; i++ {
					Gia_ObjAddFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, repeated>>1))
				}
				want = []int32{left, right, repeated}
				if count == 2 {
					want = append(want, repeated|1)
				}
				checkDynamicFanoutRing(t, p, a>>1, want)
				for _, edge := range want {
					Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, edge>>1))
				}
				checkDynamicFanoutRing(t, p, a>>1, nil)
			})
		}
	}
}

func TestDynamicFanoutStorageGrowth(t *testing.T) {
	p := Gia_ManStart(1)
	a := Gia_ManAppendCi(p)
	Gia_ManFanoutStart(p)
	initial := p.nFansAlloc
	edges := make([]int32, int(initial)+3)
	for i := range edges {
		edges[i] = Gia_ManAppendCo(p, a^int32(i&1))
	}
	if p.nFansAlloc <= initial {
		t.Fatal("fixture did not cross the initial fanout allocation")
	}
	checkDynamicFanoutRing(t, p, a>>1, edges)
	// Remove an interior element, the first, then the last after reallocation.
	for _, index := range []int{len(edges) / 2, 0, len(edges) - 3} {
		Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, edges[index]>>1))
		edges = slices.Delete(edges, index, index+1)
		checkDynamicFanoutRing(t, p, a>>1, edges)
	}
}
