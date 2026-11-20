package rwr

// rwrEdge is an index into the rewrite forest and its complemented attribute.
// C encodes the same information in the low bit of each node pointer.
type rwrEdge struct {
	node  int
	compl bool
}

type rwrForestNode struct {
	truth         uint16
	level, volume int
	xor           bool
	a, b          rwrEdge
}

// rwrLoadForest implements the pair decoding and node construction in
// Rwr_ManLoadFromArray and Rwr_ManAddNode. The first five entries are leaves.
func rwrLoadForest(data []uint16) []rwrForestNode {
	forest := []rwrForestNode{{truth: 0}, {truth: 0xaaaa}, {truth: 0xcccc}, {truth: 0xf0f0}, {truth: 0xff00}}
	for i := 0; ; i += 2 {
		if i+1 >= len(data) {
			panic("rwr: unterminated subgraph array")
		}
		x, y := data[i], data[i+1]
		if x == 0 && y == 0 {
			return forest
		}
		xor := x&1 != 0
		x >>= 1
		a, b := rwrEdge{int(x >> 1), x&1 != 0}, rwrEdge{int(y >> 1), y&1 != 0}
		if a.node >= len(forest) || b.node >= len(forest) {
			panic("rwr: invalid subgraph reference")
		}
		left, right := forest[a.node].truth, forest[b.node].truth
		var truth uint16
		if xor {
			// C's XOR entries use regular children.
			truth = left ^ right
		} else {
			if a.compl {
				left = ^left
			}
			if b.compl {
				right = ^right
			}
			truth = left & right
		}
		seen := make([]bool, len(forest))
		var volume func(int) int
		volume = func(id int) int {
			if id < 5 || seen[id] {
				return 0
			}
			seen[id] = true
			n := forest[id]
			v := 1 + volume(n.a.node) + volume(n.b.node)
			if n.xor {
				v++
			}
			return v
		}
		v := 1 + volume(a.node) + volume(b.node)
		if xor {
			v++
		}
		forest = append(forest, rwrForestNode{truth, 1 + max(forest[a.node].level, forest[b.node].level), v, xor, a, b})
	}
}
