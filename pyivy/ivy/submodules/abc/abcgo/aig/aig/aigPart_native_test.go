package aig

import "testing"

// The original C copies native EXOR as AND. Preserve its AND-only fragment
// representation while repairing that Boolean error in the Go port.
func TestDupPartNativeXorPreservesItsFunction(t *testing.T) {
	src := NewGraph()
	x := src.NewInput()
	y := src.NewInput()
	src.AddOutput(src.NewXor(x, y))
	dst := NewGraph()
	var support []int32
	outputs := Aig_ManDupPart(dst, src, []int32{0}, &support, 0)
	dst.AddOutput(outputs[0])
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, err := dst.Evaluate(in)
		if err != nil || got[0] != (in[0] != in[1]) {
			t.Fatalf("input=%v output=%v error=%v", in, got, err)
		}
	}
}

func TestPartitionedMiterDistinguishesNativeXorFromAnd(t *testing.T) {
	for _, smart := range []int32{0, 1} {
		a, b := NewGraph(), NewGraph()
		x, y := a.NewInput(), a.NewInput()
		u, v := b.NewInput(), b.NewInput()
		a.AddOutput(a.NewXor(x, y))
		b.AddOutput(b.AND(u, v))
		miters := Aig_ManMiterPartitioned(a, b, 1, smart)
		if len(miters) != 1 {
			t.Fatal("expected one comparison fragment")
		}
		q := miters[0]
		for mask := 0; mask < 4; mask++ {
			global := []bool{mask&1 != 0, mask&2 != 0}
			local := make([]bool, len(q.PartitionInputs))
			for i, index := range q.PartitionInputs {
				local[i] = global[index]
			}
			out, err := q.Evaluate(local)
			if err != nil || out[0] != (mask != 0) {
				t.Errorf("smart %d assignment %d: mismatch %v (%v), want %t", smart, mask, out, err, mask != 0)
			}
		}
	}
}

func TestOriginalCPartitionNativeXorLimitation(t *testing.T) {
	got := runAigSourceFixture(t, "partition_xor_upstream.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigPart.c", "aigDfs.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	want := "native-XOR 0 smart 0 mismatch truth 14\nnative-XOR 0 smart 1 mismatch truth 14\nnative-XOR 1 smart 0 mismatch truth 0\nnative-XOR 1 smart 1 mismatch truth 0\n"
	if got != want {
		t.Fatalf("original C limitation: %q, want %q", got, want)
	}
}

func TestDupPartNativeXorCopyMaps(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		src := NewGraph()
		x, y := src.NewInput(), src.NewInput()
		root := src.NewXor(dupPhase(x, signs&1 != 0), dupPhase(y, signs&2 != 0))
		src.AddOutput(dupPhase(root, signs&4 != 0))
		whole := Aig_ManDupPartAll(src, []int32{0})
		for mask := 0; mask < 4; mask++ {
			global := []bool{mask&1 != 0, mask&2 != 0}
			local := make([]bool, len(whole.inputs))
			for i, input := range whole.inputs {
				original := whole.copies[input.NodeID()]
				if original.NodeID() == x.NodeID() {
					local[i] = global[0]
				} else {
					local[i] = global[1]
				}
			}
			sv := aigSimulate(src, aigUtilOrder(src), global)
			dv := aigSimulate(whole, aigUtilOrder(whole), local)
			for id, target := range src.copies {
				if sv[id] != aigUtilEdgeValue(target, dv) {
					t.Fatalf("signs %d assignment %d forward copy of %d changed value", signs, mask, id)
				}
			}
			for id, original := range whole.copies {
				if dv[id] != aigUtilEdgeValue(original, sv) {
					t.Errorf("signs %d assignment %d reverse copy of %d changed value", signs, mask, id)
				}
			}
		}
		if len(whole.copies) != 5 {
			t.Fatalf("reverse map has %d entries, want the five original objects only", len(whole.copies))
		}
		for _, e := range whole.LiveEdges() {
			if whole.Node(e).Kind == ExclusiveOr {
				t.Fatal("partition contains a native XOR")
			}
		}
	}
}

func TestDupPartNativeXorInverseAndAliases(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		src := NewGraph()
		x, y := src.NewInput(), src.NewInput()
		root := src.NewXor(dupPhase(x, signs&1 != 0), dupPhase(y, signs&2 != 0))
		buffer := Aig_ObjCreate(src, &Node{Kind: Buffer, Fanins: [2]Edge{root.Not(), InvalidEdge}})
		for _, e := range []Edge{dupPhase(root, signs&4 != 0), root.Not(), src.Const0(), x, buffer} {
			src.AddOutput(e)
		}
		fragment := NewGraph()
		fragment.NewInput()
		fragment.NewInput()
		support := []int32{1, 0}
		for _, e := range Aig_ManDupPart(fragment, src, []int32{0, 1, 2, 3, 4}, &support, 0) {
			fragment.AddOutput(e)
		}
		restored := NewGraph()
		restored.NewInput()
		restored.NewInput()
		for _, e := range Aig_ManDupPart(restored, fragment, nil, &support, 1) {
			restored.AddOutput(e)
		}
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for _, q := range []*Graph{fragment, restored} {
				local := inputs
				if q == fragment {
					local = []bool{inputs[1], inputs[0]}
				}
				got, err := q.Evaluate(local)
				if err != nil {
					t.Fatal(err)
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("signs %d assignment %d output %d: got %t want %t", signs, mask, i, got[i], want[i])
					}
				}
			}
		}
	}
}

func TestPartitionNativeXorFraigAndChoiceTransfer(t *testing.T) {
	calls := 0
	backend := func(g *Graph, _, _ int32) *Graph {
		calls++
		Aig_ManReprStart(g, int32(len(g.nodes)))
		order := aigUtilOrder(g)
		values := make([][]bool, 4)
		for mask := range values {
			values[mask] = aigSimulate(g, order, []bool{mask&1 != 0, mask&2 != 0})
		}
		classes := map[uint8]Edge{}
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			if n.Kind == ExclusiveOr {
				t.Fatal("backend received native XOR")
			}
			if n.Kind == Output {
				continue
			}
			var truth uint8
			for mask, v := range values {
				if v[e.NodeID()] != n.Phase {
					truth |= 1 << uint(mask)
				}
			}
			if previous, ok := classes[truth]; ok {
				Aig_ObjSetRepr_(g, previous, e)
			} else {
				classes[truth] = e
			}
		}
		// Real FRA uses this scratch map; the driver must restore it.
		g.copies = nil
		return g
	}
	makeGraph := func(native bool) *Graph {
		g := NewGraph()
		x, y := g.NewInput(), g.NewInput()
		xor := g.XOR(x, y)
		if native {
			xor = g.NewXor(x, y)
		}
		g.AddOutput(xor)
		g.AddOutput(g.AND(x, y))
		return g
	}
	original := makeGraph(true)
	reduced := Aig_ManFraigPartitioned(original, 2, 0, 0, 0, backend)
	choice := Aig_ManChoicePartitioned([]*Graph{makeGraph(true), makeGraph(false)}, 2, 0, 0, 0, backend)
	for _, g := range []*Graph{reduced, choice} {
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 4; mask++ {
			got, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
			if err != nil || got[0] != (mask == 1 || mask == 2) || got[1] != (mask == 3) {
				t.Fatalf("assignment %d: outputs %v, error %v", mask, got, err)
			}
		}
	}
	if calls < 2 {
		t.Fatal("partition backend was not exercised")
	}
}
