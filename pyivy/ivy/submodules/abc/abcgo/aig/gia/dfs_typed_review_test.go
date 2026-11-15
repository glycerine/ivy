package gia

import (
	"fmt"
	"testing"
)

// Construct the reference dependency relation independently while building the
// graph. Reachability below uses fixed-point set expansion, not the port's DFS.
func TestDfsTypedDependencyClosure(t *testing.T) {
	p := Gia_ManStart(32)
	defer Gia_ManStop(p)
	p.EnableMuxes()
	deps := map[int32][]int32{}
	internal, cis, cos := map[int32]bool{}, map[int32]bool{}, map[int32]bool{}
	add := func(lit int32, inputs ...int32) int32 {
		id := lit >> 1
		internal[id] = true
		for _, in := range inputs {
			deps[id] = append(deps[id], in>>1)
		}
		return lit
	}
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, lit := range []int32{a, b, r0, r1} {
		cis[lit>>1] = true
	}
	x := add(Gia_ManAppendAnd(p, a, b), a, b)
	z := add(Gia_ManAppendXorReal(p, x, r0^1), x, r0)
	buf := add(Gia_ManAppendBuf(p, z), z)
	m := add(Gia_ManAppendMuxReal(p, r1, buf, a^1), r1, buf, a)
	y := add(Gia_ManAppendAnd(p, m, r0), m, r0)
	add(Gia_ManAppendAnd(p, b, 1), b, 1) // dangling constant-driven logic
	var outputs []int32
	for _, lit := range []int32{y, 0, 1, r1, z, m} {
		Gia_ManAppendCo(p, lit)
		id := p.nObjs - 1
		cos[id], deps[id] = true, []int32{lit >> 1}
		outputs = append(outputs, id)
	}
	Gia_ManSetRegNum(p, 2)
	registers := map[int32]int32{r0 >> 1: outputs[4], r1 >> 1: outputs[5]}
	closure := func(roots []int32, blocked map[int32]bool, sequential bool) map[int32]bool {
		set := map[int32]bool{}
		for _, id := range roots {
			if !blocked[id] {
				set[id] = true
			}
		}
		for changed := true; changed; {
			changed = false
			for id := range set {
				for _, in := range deps[id] {
					if !blocked[in] && !set[in] {
						set[in], changed = true, true
					}
				}
				if ri, ok := registers[id]; sequential && ok && !set[ri] {
					set[ri], changed = true, true
				}
			}
		}
		return set
	}
	filter := func(set, allowed map[int32]bool) map[int32]bool {
		out := map[int32]bool{}
		for id := range set {
			if allowed[id] {
				out[id] = true
			}
		}
		return out
	}
	check := func(name string, got *Vec_Int_t, want map[int32]bool, postorder bool) {
		t.Helper()
		seen := map[int32]bool{}
		for _, id := range Vec_IntArray(got) {
			if seen[id] || !want[id] {
				t.Fatalf("%s: unexpected/repeated ID %d in %v; want %v", name, id, Vec_IntArray(got), want)
			}
			if postorder {
				for _, in := range deps[id] {
					if want[in] && !seen[in] {
						t.Fatalf("%s: node %d precedes dependency %d", name, id, in)
					}
				}
			}
			seen[id] = true
		}
		if len(seen) != len(want) {
			t.Fatalf("%s: got %v; want %v", name, seen, want)
		}
	}
	work := Vec_IntAlloc(32)
	for mask := 0; mask < 64; mask++ {
		var roots []int32
		for i, id := range outputs {
			if mask&(1<<i) != 0 {
				roots = append(roots, id)
			}
		}
		set := closure(roots, nil, false)
		wantCis, wantAnds := filter(set, cis), filter(set, internal)
		n := int32(len(roots))
		Gia_ManCollectCis(p, roots, n, work)
		check(fmt.Sprintf("CIs mask%d", mask), work, wantCis, false)
		if got := Gia_ManSuppSize(p, roots, n); got != int32(len(wantCis)) {
			t.Fatalf("mask %d: support %d, want %d", mask, got, len(wantCis))
		}
		if got := Gia_ManConeSize(p, roots, n); got != int32(len(wantAnds)) {
			t.Fatalf("mask %d: cone %d, want %d", mask, got, len(wantAnds))
		}
		wantNodes := filter(set, internal)
		for id := range wantCis {
			wantNodes[id] = true
		}
		check("nodes/CIs", Gia_ManCollectNodesCis(p, roots, n), wantNodes, true)
		for _, cut := range []bool{false, true} {
			var leaves *Vec_Int_t
			blocked := map[int32]bool{}
			if cut {
				leaves = Vec_IntAlloc(2)
				for _, id := range []int32{x >> 1, buf >> 1} {
					Vec_IntPush(leaves, id)
					blocked[id] = true
				}
			}
			Gia_ManIncrementTravId(p)
			Gia_ManCollectAnds(p, roots, n, work, leaves)
			check("AND cut", work, filter(closure(roots, blocked, false), internal), true)
		}
	}
	Gia_ManStaticFanoutStart(p)
	for _, root := range []int32{x >> 1, z >> 1, buf >> 1, m >> 1, y >> 1, 0} {
		roots := Vec_IntAlloc(1)
		Vec_IntPush(roots, root)
		Gia_ManCollectTfi(p, roots, work)
		check("TFI", work, filter(closure([]int32{root}, nil, false), internal), true)
		if root == 0 {
			continue // TFO requires internal roots.
		}
		forward := map[int32]bool{root: true}
		for changed := true; changed; {
			changed = false
			for id, inputs := range deps {
				for _, in := range inputs {
					if forward[in] && !forward[id] {
						forward[id], changed = true, true
					}
				}
			}
		}
		Gia_ManCollectTfo(p, roots, work)
		check("TFO", work, filter(forward, internal), false)
	}
	for mask := 0; mask < 16; mask++ {
		roots, pos := []int32{0}, []int32{}
		for i := 0; i < 4; i++ {
			if mask&(1<<i) != 0 {
				pos, roots = append(pos, int32(i)), append(roots, outputs[i])
			}
		}
		got := Gia_ManCollectSeq(p, pos, int32(len(pos)))
		check("sequential", got, closure(roots, nil, true), true)
		if Vec_IntEntry(got, 0) != 0 {
			t.Fatal("sequential order must start with the constant")
		}
	}
}

func BenchmarkDfsSupportWalks(b *testing.B) {
	p := Gia_ManStart(2048)
	defer Gia_ManStop(p)
	a, other := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for i := 0; i < 1024; i++ {
		a = Gia_ManAppendAnd(p, a, other)
	}
	Gia_ManAppendCo(p, a)
	root := []int32{p.nObjs - 1}
	work := Vec_IntAlloc(1024)
	Gia_ManIncrementTravId(p)
	for _, operation := range []struct {
		name string
		run  func()
	}{
		{"support", func() { Gia_ManSuppSize(p, root, 1) }},
		{"CIs", func() { Gia_ManCollectCis(p, root, 1, work) }},
		{"cone", func() { Gia_ManConeSize(p, root, 1) }},
		{"sequential", func() { Gia_ManCollectSeq(p, []int32{0}, 1) }},
	} {
		b.Run(operation.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				operation.run()
			}
		})
	}
}
