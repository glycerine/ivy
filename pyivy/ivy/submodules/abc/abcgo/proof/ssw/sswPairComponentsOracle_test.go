package ssw

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"reflect"
	"testing"
)

// Pair classes represent undirected connected components, with the least ID
// as representative. Check against a separate graph-search implementation.
func TestPairClassesAgainstConnectedComponents(t *testing.T) {
	random := rand.New(rand.NewSource(51823))
	for trial := 0; trial < 500; trial++ {
		const n = 12
		graph := make([][]int, n)
		used := make([]bool, n)
		pairs := vec.Vec_IntAlloc(64)
		for edge := 0; edge < random.Intn(35); edge++ {
			// Keep a fixed representative: C reuses the pair index on merges.
			a, b := 0, 1+random.Intn(n-1)
			vec.Vec_IntPushTwo(pairs, int32(a), int32(b))
			used[a], used[b] = true, true
			graph[a] = append(graph[a], b)
			graph[b] = append(graph[b], a)
		}
		classes := Ssw_TransformPairsIntoTempClasses(pairs, n)
		seen := make([]bool, n)
		for root := 0; root < n; root++ {
			if !used[root] || seen[root] {
				continue
			}
			component := make([]bool, n)
			component[root], seen[root] = true, true
			stack := []int{root}
			for len(stack) > 0 {
				v := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				for _, next := range graph[v] {
					if !seen[next] {
						seen[next], component[next] = true, true
						stack = append(stack, next)
					}
				}
			}
			want := []int32{}
			for i, member := range component {
				if member {
					want = append(want, int32(i))
				}
			}
			if len(want) < 2 {
				if classes[root] != nil {
					t.Fatalf("trial %d singleton class", trial)
				}
				continue
			}
			if classes[root] == nil || !reflect.DeepEqual(vec.Vec_IntArray(classes[root]), want) {
				t.Fatalf("trial %d root %d got %v want %v", trial, root, classes[root], want)
			}
			for _, member := range want[1:] {
				if classes[member] != nil {
					t.Fatalf("trial %d duplicate class %d", trial, member)
				}
			}
		}
	}
}
