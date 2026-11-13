package aig

import "testing"

func TestPartitionSupportsAndCoverage(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	p.AddOutput(p.AND(a, c))
	p.AddOutput(p.Const1())
	supports := Aig_ManSupports(p)
	if len(supports) != 3 || len(supports[0]) != 3 || len(supports[2]) != 1 {
		t.Fatal(supports)
	}
	inverse := Aig_ManSupportsInverse(p)
	if len(inverse[0]) != 2 || len(inverse[1]) != 1 || len(inverse[2]) != 1 {
		t.Fatal(inverse)
	}
	var partitionSupports [][]int32
	parts := Aig_ManPartitionSmart(p, 2, 0, &partitionSupports)
	seen := map[int32]bool{}
	Aig_ManSetCioIds(p)
	for i, part := range parts {
		var expected []int32
		for _, id := range part {
			if seen[id] {
				t.Fatal("duplicated output")
			}
			seen[id] = true
			expected = partUnion(expected, partOutputSupport(p, p.outputs[id]))
		}
		if len(expected) != len(partitionSupports[i]) {
			t.Fatal(parts, partitionSupports)
		}
		for j, v := range expected {
			if v != partitionSupports[i][j] {
				t.Fatal(expected, partitionSupports[i])
			}
		}
	}
	if len(seen) != len(p.outputs) {
		t.Fatal("output omitted")
	}
}
func TestPartitionCompactionFlushesPendingBeforeLarge(t *testing.T) {
	parts := [][]int32{{0}, {1}, {2}}
	supports := [][]int32{{0}, {1, 2, 3, 4}, {5}}
	Aig_ManPartitionCompact(&parts, &supports, 3)
	seen := map[int32]bool{}
	for _, part := range parts {
		for _, id := range part {
			seen[id] = true
		}
	}
	if len(seen) != 3 || len(parts) != len(supports) {
		t.Fatal("pending partition lost", parts, supports)
	}
}
func TestPartitionPoolOwnershipAndSortedMerge(t *testing.T) {
	p := Part_ManStart(1024, 64)
	block := Part_ManFetch(p, 19)
	if len(block.Data) != 64 {
		t.Fatal(len(block.Data))
	}
	Part_ManRecycle(p, block, 19)
	reused := Part_ManFetch(p, 19)
	if reused != block {
		t.Fatal("not reused")
	}
	a, b := Part_ManFetchEntry(p, 3, 1), Part_ManFetchEntry(p, 3, 1)
	a.pOuts = append(a.pOuts, 0, 2, 4)
	b.pOuts = append(b.pOuts, 1, 2, 5)
	merged := Part_ManMergeEntry(p, a, b, 2)
	want := []int32{0, 1, 2, 4, 5}
	got := Part_ManTransferEntry(merged)
	if len(got) != len(want) {
		t.Fatal(got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Fatal(got)
		}
	}
	got[0] = 99
	if merged.pOuts[0] != 0 {
		t.Fatal("transfer exposes pool storage")
	}
	Part_ManRecycleEntry(p, merged)
	defer func() {
		if recover() == nil {
			t.Fatal("double recycle accepted")
		}
	}()
	Part_ManRecycleEntry(p, merged)
}
func TestPartitionedMiterAndInputMaps(t *testing.T) {
	a, b := NewGraph(), NewGraph()
	x, y, z := a.NewInput(), a.NewInput(), a.NewInput()
	u, v, w := b.NewInput(), b.NewInput(), b.NewInput()
	a.AddOutput(a.NewXor(x, y))
	a.AddOutput(z)
	b.AddOutput(b.NewXor(u, v))
	b.AddOutput(w.Not())
	for _, smart := range []int32{0, 1} {
		miters := Aig_ManMiterPartitioned(a, b, 1, smart)
		if len(miters) == 0 {
			t.Fatal("missing miters")
		}
		for _, q := range miters {
			if err := q.Check(); err != nil {
				t.Fatal(err)
			}
			for mask := 0; mask < 8; mask++ {
				global := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				local := make([]bool, len(q.PartitionInputs))
				for i, index := range q.PartitionInputs {
					local[i] = global[index]
				}
				out, err := q.Evaluate(local)
				if err != nil {
					t.Fatal(err)
				}
				want := false
				for _, index := range q.PartitionOutputs {
					if index == 1 {
						want = true
					}
				}
				if out[0] != want {
					t.Fatal(q.PartitionOutputs, out, want)
				}
			}
		}
	}
}
func TestPartitionExactFraigingAndChoices(t *testing.T) {
	// Test-only oracle for these three-input fixtures. Production callers
	// must supply the FRA backend; this checks partition transfer itself.
	backend := func(g *Graph, _, _ int32) *Graph {
		Aig_ManReprStart(g, int32(len(g.nodes)))
		classes := map[uint8]Edge{}
		inputPos := map[int]int{}
		for i, input := range g.inputs {
			inputPos[input.NodeID()] = i
		}
		for _, edge := range g.LiveEdges() {
			n := g.Node(edge)
			if n.Kind == Output {
				continue
			}
			var truth uint8
			for bits := 0; bits < 8; bits++ {
				var value func(Edge) bool
				value = func(e Edge) bool {
					x := g.Node(e)
					var v bool
					switch x.Kind {
					case Constant:
						v = true
					case Input:
						v = bits&(1<<inputPos[e.NodeID()]) != 0
					case And:
						v = value(x.Fanins[0]) && value(x.Fanins[1])
					case ExclusiveOr:
						v = value(x.Fanins[0]) != value(x.Fanins[1])
					case Buffer:
						v = value(x.Fanins[0])
					}
					return v != e.Complemented()
				}
				if value(edge) != n.Phase {
					truth |= 1 << uint(bits)
				}
			}
			if repr, ok := classes[truth]; ok {
				Aig_ObjSetRepr_(g, repr, edge)
			} else {
				classes[truth] = edge
			}
		}
		return g
	}
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	simple := p.AND(a, b)
	redundant := p.AND(simple, p.OR(a, c))
	p.AddOutput(redundant)
	reduced := Aig_ManFraigPartitioned(p, 1, 0, 0, 0, backend)
	if err := reduced.Check(); err != nil {
		t.Fatal(err)
	}
	if Aig_ManNodeNum(reduced) >= Aig_ManNodeNum(p) {
		t.Fatal("equivalent redundancy not reduced")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		before, _ := p.Evaluate(inputs)
		after, err := reduced.Evaluate(inputs)
		if err != nil || before[0] != after[0] {
			t.Fatal(mask, before, after, err)
		}
	}
	versions := make([]*Graph, 2)
	for i := range versions {
		q := NewGraph()
		x, y, z := q.NewInput(), q.NewInput(), q.NewInput()
		var root Edge
		if i == 0 {
			root = q.AND(q.AND(x, y), z)
		} else {
			root = q.AND(x, q.AND(y, z))
		}
		q.AddOutput(root)
		versions[i] = q
	}
	for _, result := range []*Graph{Aig_ManChoiceConstructive(versions, 0), Aig_ManChoicePartitioned(versions, 2, 0, 0, 0, backend)} {
		if err := result.Check(); err != nil {
			t.Fatal(err)
		}
		if len(result.equivs) == 0 {
			t.Fatal("alternative implementations not retained")
		}
		for mask := 0; mask < 8; mask++ {
			out, err := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil || out[0] != (mask == 7) {
				t.Fatal(out, err)
			}
		}
	}
}

func TestPartitionFraigRestoresCopiesAndForwardsLimits(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.copies = map[int]Edge{0: 0, a.NodeID(): 42}
	called := false
	runPartitionFraig(g, 17, 9, func(part *Graph, conflicts, level int32) *Graph {
		called = true
		if part != g || conflicts != 17 || level != 9 {
			t.Fatalf("backend arguments: %p %d %d", part, conflicts, level)
		}
		part.copies = map[int]Edge{a.NodeID(): 100}
		return part
	})
	if !called || g.copies[a.NodeID()] != 42 || g.copies[0] != 0 {
		t.Fatalf("partition copy map not restored: %v", g.copies)
	}
}

// ABC transfers source representatives in object-ID order. When multiple
// partition objects map to one destination, the later source entry wins.
func TestPartitionRepresentativeTransferCollisionOrder(t *testing.T) {
	dst := NewGraph()
	a, b, shared := dst.NewInput(), dst.NewInput(), dst.NewInput()
	Aig_ManReprStart(dst, int32(len(dst.nodes)))
	fragment := NewGraph()
	x, y, first, last := fragment.NewInput(), fragment.NewInput(), fragment.NewInput(), fragment.NewInput()
	Aig_ManReprStart(fragment, int32(len(fragment.nodes)))
	Aig_ObjSetRepr_(fragment, x, first)
	Aig_ObjSetRepr_(fragment, y, last)
	fragment.copies = map[int]Edge{x.NodeID(): a, y.NodeID(): b, first.NodeID(): shared, last.NodeID(): shared}
	Aig_ManTransferRepr(dst, fragment)
	if got := Aig_ObjRepr(dst, shared); got != b {
		t.Fatalf("collision representative %v, want later source mapping %v", got, b)
	}
}
