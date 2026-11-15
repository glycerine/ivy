package gia

import "testing"

func TestEdgeForwardAssignmentTieAndCapacityBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name     string
		times    [4]int32
		capacity [2]int
		two      int32
		want     int32
		fast     []int32
	}{
		{"first level", [4]int32{}, [2]int{}, 1, 10, nil},
		{"unique", [4]int32{10, 1}, [2]int{}, 0, 12, []int32{1}},
		{"runner dominates", [4]int32{10, 9}, [2]int{}, 0, 19, []int32{1}},
		{"two ties disabled", [4]int32{10, 10}, [2]int{}, 0, 20, nil},
		{"two ties enabled", [4]int32{10, 10}, [2]int{}, 1, 12, []int32{1, 2}},
		{"three ties", [4]int32{10, 10, 10}, [2]int{}, 1, 20, nil},
		{"one slot remains", [4]int32{10, 1}, [2]int{1, 0}, 0, 12, []int32{1}},
		{"saturated unique", [4]int32{10, 1}, [2]int{2, 0}, 0, 20, nil},
		{"saturated second tie", [4]int32{10, 10}, [2]int{0, 2}, 1, 20, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := Gia_ManStart(32)
			defer Gia_ManStop(p)
			var leaves [4]int32
			for i := range leaves {
				leaves[i] = Gia_ManAppendCi(p)
			}
			var buffers [2][2]int32
			for i := range buffers {
				for j := range buffers[i] {
					buffers[i][j] = Gia_ManAppendBuf(p, leaves[i]) >> 1
				}
			}
			root := leaves[0]
			for _, lit := range leaves[1:] {
				root = Gia_ManAppendAnd(p, root, lit)
			}
			Gia_ManAppendCo(p, root)
			p.vMapping2 = Vec_WecStart(int(p.nObjs))
			for _, lit := range leaves {
				Vec_WecPush(p.vMapping2, int(root>>1), lit>>1)
			}
			for i := range buffers {
				for _, id := range buffers[i] {
					Vec_WecPush(p.vMapping2, int(id), leaves[i]>>1)
				}
			}
			p.VEdge1, p.VEdge2 = Vec_IntStart(int(p.nObjs)), Vec_IntStart(int(p.nObjs))
			for i, occupied := range tc.capacity {
				for _, id := range buffers[i][:occupied] {
					Gia_ObjEdgeAdd(int32(i+1), id, p.VEdge1, p.VEdge2)
					Gia_ObjEdgeAdd(id, int32(i+1), p.VEdge1, p.VEdge2)
				}
			}
			delays := Vec_IntStart(int(p.nObjs))
			for i, value := range tc.times {
				Vec_IntWriteEntry(delays, i+1, value)
			}
			if got := Gia_ObjComputeEdgeDelay(p, root>>1, delays, p.VEdge1, p.VEdge2, tc.two); got != tc.want || Vec_IntEntry(delays, int(root>>1)) != tc.want {
				t.Fatalf("delay %d, want %d", got, tc.want)
			}
			var want [5]bool
			for _, id := range tc.fast {
				want[id] = true
			}
			for id := int32(1); id <= 4; id++ {
				if got := Gia_ObjHaveEdge(p, root>>1, id) != 0; got != want[id] {
					t.Fatalf("fast input %d: %t, want %t", id, got, want[id])
				}
			}
			arcs := map[int32][]int32{}
			for id := int32(1); id < p.nObjs; id++ {
				if Gia_ObjIsLut2(p, id) != 0 {
					arcs[id] = Vec_IntArray(Gia_ObjLutFanins2(p, id))
				}
			}
			edgeReviewAssignment(t, p, arcs, 2)
		})
	}
}

func edgeReviewPair(a, b int32) [2]int32 {
	return [2]int32{min(a, b), max(a, b)}
}

// Enumerate complete paths rather than reuse the port's timing recurrence.
func edgeReviewPathDelay(root int32, arcs map[int32][]int32, fast map[[2]int32]bool, slowCost, fastCost int32) int32 {
	type path struct{ id, cost int32 }
	stack := []path{{id: root}}
	var longest int32
	for len(stack) != 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(arcs[x.id]) == 0 {
			longest = max(longest, x.cost)
		}
		for _, next := range arcs[x.id] {
			cost := slowCost
			if fast[edgeReviewPair(x.id, next)] {
				cost = fastCost
			}
			stack = append(stack, path{next, x.cost + cost})
		}
	}
	return longest
}

func edgeReviewAssignment(t *testing.T, p *Gia_Man_t, arcs map[int32][]int32, limit int32) map[[2]int32]bool {
	t.Helper()
	allowed, edges := map[[2]int32]bool{}, map[[2]int32]bool{}
	for id, inputs := range arcs {
		for _, in := range inputs {
			allowed[edgeReviewPair(id, in)] = true
		}
	}
	for id := int32(0); id < p.nObjs; id++ {
		a, b := Vec_IntEntry(p.VEdge1, int(id)), Vec_IntEntry(p.VEdge2, int(id))
		if a != 0 && a == b {
			t.Fatalf("duplicate assigned edge at %d", id)
		}
		var degree int32
		for _, next := range []int32{a, b} {
			if next == 0 {
				continue
			}
			degree++
			key := edgeReviewPair(id, next)
			if !allowed[key] || next <= 0 || next >= p.nObjs {
				t.Fatalf("assigned edge %d-%d is not a mapped dependency", id, next)
			}
			if Vec_IntEntry(p.VEdge1, int(next)) != id && Vec_IntEntry(p.VEdge2, int(next)) != id {
				t.Fatalf("asymmetric assigned edge %d-%d", id, next)
			}
			edges[key] = true
		}
		if degree > limit {
			t.Fatalf("node %d has degree %d above limit %d", id, degree, limit)
		}
	}
	if got := Gia_ManEvalEdgeCount(p); got != int32(len(edges)) {
		t.Fatalf("edge count %d, want %d", got, len(edges))
	}
	return edges
}

func TestEdgeAssignmentIndependentPathInvariants(t *testing.T) {
	for seed := uint32(33); seed <= 64; seed++ {
		p := edgeReviewMappedGraph(seed)
		Edg_ManToMapping(p)
		backward, forward := map[int32][]int32{}, map[int32][]int32{}
		for id := int32(1); id < p.nObjs; id++ {
			if Gia_ObjIsLut(p, id) == 0 {
				continue
			}
			backward[id] = append([]int32(nil), Gia_ObjLutFanins(p, id)...)
			for _, in := range backward[id] {
				forward[in] = append(forward[in], id)
			}
		}
		for two := int32(0); two < 2; two++ {
			result := Gia_ManComputeEdgeDelay(p, two)
			fast := edgeReviewAssignment(t, p, backward, 2)
			for id := range backward {
				want := edgeReviewPathDelay(id, backward, fast, 10, 2)
				if got := Vec_IntEntry(p.vEdgeDelay, int(id)); got != want {
					t.Fatalf("seed%d/two%d: LUT%d forward %d, want %d", seed, two, id, got, want)
				}
			}
			var want int32
			for i := int32(0); i < Gia_ManCoNum(p); i++ {
				co := Gia_ManCo(p, i)
				want = max(want, edgeReviewPathDelay(Gia_ObjFaninId0(co, co.id), backward, fast, 10, 2))
			}
			if result != want || Gia_ManEvalEdgeDelay(p) != want {
				t.Fatalf("seed%d/two%d: assignment/evaluation disagree with output paths %d", seed, two, want)
			}
		}
		result := Gia_ManComputeEdgeDelay2(p)
		fast := edgeReviewAssignment(t, p, backward, 2)
		var want int32
		for id := range backward {
			if got, expected := Vec_IntEntry(p.vEdgeDelay, int(id)), edgeReviewPathDelay(id, forward, fast, 1, 0); got != expected {
				t.Fatalf("seed%d: LUT%d reverse %d, want %d", seed, id, got, expected)
			}
		}
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			id := Gia_ManCi(p, i).id
			expected := edgeReviewPathDelay(id, forward, fast, 1, 0)
			if got := Vec_IntEntry(p.vEdgeDelay, int(id)); got != expected {
				t.Fatalf("seed%d: CI%d reverse %d, want %d", seed, id, got, expected)
			}
			want = max(want, expected)
		}
		if result != want {
			t.Fatalf("seed%d: reverse maximum %d, want %d", seed, result, want)
		}
		for limit := int32(0); limit < 3; limit++ {
			result = Edg_ManAssignEdgeNew(p, limit, 0)
			fast = edgeReviewAssignment(t, p, backward, limit)
			f, r := Edg_ManEvalEdgeDelay(p), Edg_ManEvalEdgeDelayR(p)
			var maxF, maxR int32
			for id := range backward {
				wantF := edgeReviewPathDelay(id, backward, fast, 1, 0)
				wantR := edgeReviewPathDelay(id, forward, fast, 1, 0)
				if Vec_IntEntry(p.vEdgeDelay, int(id)) != wantF || Vec_IntEntry(p.vEdgeDelayR, int(id)) != wantR {
					t.Fatalf("seed%d/limit%d: LUT%d unit path delays differ", seed, limit, id)
				}
				maxF, maxR = max(maxF, wantF), max(maxR, wantR)
			}
			if result != maxF || f != maxF || r != maxR || f != r+1 {
				t.Fatalf("seed%d/limit%d: optimized delay %d, F/R %d/%d, want %d/%d", seed, limit, result, f, r, maxF, maxR)
			}
			levels := Vec_WecStart(0)
			Edg_ManCollectCritEdges(p, levels, f)
			seen := map[int32]bool{}
			for level := 0; level < Vec_WecSize(levels); level++ {
				for _, id := range Vec_IntArray(Vec_WecEntry(levels, level)) {
					if seen[id] || Vec_IntEntry(p.vEdgeDelay, int(id)) != int32(level) || Vec_IntEntry(p.vEdgeDelayR, int(id))+int32(level) != f {
						t.Fatal("incorrect/repeated critical node", id)
					}
					seen[id] = true
				}
			}
			for id := range backward {
				if want := Vec_IntEntry(p.vEdgeDelay, int(id))+Vec_IntEntry(p.vEdgeDelayR, int(id)) == f; seen[id] != want {
					t.Fatalf("critical-node membership differs at %d", id)
				}
			}
		}
		Gia_ManStop(p)
	}
}
