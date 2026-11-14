package aig

import "testing"

// aigPart.c checks every live previous object, not only the output cone,
// before accumulating another version into the destination graph.
func TestConstructiveChoiceRequiresPreviousCopyMap(t *testing.T) {
	for _, damage := range []string{"missing", "invalid", "unreachable"} {
		t.Run(damage, func(t *testing.T) {
			previous := NewGraph()
			a, b := previous.NewInput(), previous.NewInput()
			previous.AddOutput(previous.AND(a, b))
			unused := previous.OR(a, b)
			current := Aig_ManDupOrdered(previous)
			destination := Aig_ManDupOrdered(previous)
			Aig_ManReprStart(destination, int32(len(destination.nodes)))
			switch damage {
			case "missing":
				delete(previous.copies, a.NodeID())
			case "invalid":
				previous.copies[a.NodeID()] = InvalidEdge
			case "unreachable":
				delete(previous.copies, unused.NodeID())
			}
			defer func() {
				if recover() == nil {
					t.Fatal("accepted a previous object without a destination copy")
				}
			}()
			Aig_ManChoiceConstructiveOne(destination, previous, current)
		})
	}
}

func TestConstructiveChoiceChainsPreviousCopyMaps(t *testing.T) {
	versions := make([]*Graph, 3)
	for i := range versions {
		g := NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		var root Edge
		switch i {
		case 0:
			root = g.AND(g.AND(a, b), c)
		case 1:
			root = g.AND(a, g.AND(b, c))
		case 2:
			root = g.AND(g.AND(a, c), b)
		}
		g.AddOutput(root.Not())
		versions[i] = g
	}
	destination := Aig_ManDupOrdered(versions[0])
	Aig_ManReprStart(destination, int32(len(destination.nodes)))
	for i := 1; i < len(versions); i++ {
		Aig_ManChoiceConstructiveOne(destination, versions[i-1], versions[i])
		for id := range versions[i].nodes {
			copy, ok := versions[i].copies[id]
			if !ok || !destination.valid(copy) {
				t.Fatalf("version %d object %d lost destination copy", i, id)
			}
		}
		if versions[i].copies[versions[i].outputs[0].NodeID()] != destination.outputs[0] {
			t.Fatal("output copy must refer to the destination CO")
		}
		copied := versions[i].copies[versions[i].outputDriver(versions[i].outputs[0]).NodeID()]
		original := destination.outputDriver(destination.outputs[0])
		if Aig_ObjFindReprTransitive(destination, Aig_Regular(copied)) != Aig_Regular(original) {
			t.Fatal("missing output-driver representative")
		}
	}
	result := Aig_ManChoiceConstructive(versions, 0)
	if err := result.Check(); err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		out, err := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || out[0] != (mask != 7) {
			t.Fatalf("mask %d: output %v, error %v", mask, out, err)
		}
	}
}
