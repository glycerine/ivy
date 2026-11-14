package aig

import "testing"

func TestGreedyJustificationWitnesses(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	roots := []Edge{p.AND(a, b), p.OR(a, b), p.NewXor(a, b), p.Const0(), p.Const1()}
	roots = append(roots, p.AddOutput(roots[2].Not()))
	Aig_ManSetCioIds(p)
	for _, root := range roots {
		for target := int32(0); target < 2; target++ {
			var support []int32
			found := Aig_ObjFindSatAssign(p, root, target, &support)
			if root == p.Const0() && target == 1 || root == p.Const1() && target == 0 {
				if found {
					t.Fatal("impossible constant witness")
				}
				continue
			}
			if !found {
				t.Fatalf("failed simple satisfiable root %d target %d", root, target)
			}
			value := Aig_ObjTerSimulate(p, root, support)
			want := AIG_VAL0
			if target == 1 {
				want = AIG_VAL1
			}
			if value != want {
				t.Fatalf("invalid witness root %d target %d support %v value %d", root, target, support, value)
			}
		}
	}
}
func TestTernaryPartialAssignments(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	if Aig_ObjTerSimulate(p, root, nil) != AIG_VALX {
		t.Fatal("unknown cone not X")
	}
	if Aig_ObjTerSimulate(p, root, []int32{1}) != AIG_VAL0 {
		t.Fatal("zero input did not control AND")
	}
	if Aig_ObjTerSimulate(p, root, []int32{0}) != AIG_VALX {
		t.Fatal("partial AND incorrectly resolved")
	}
	if Aig_ObjTerSimulate(p, root.Not(), []int32{0, 2}) != AIG_VAL0 {
		t.Fatal("complemented root incorrect")
	}
}
func TestJustificationTraversalWrap(t *testing.T) {
	p := NewGraph()
	a := p.NewInput()
	p.travID = 1<<31 - 1
	p.node(a).TravID = 1
	var support []int32
	if !Aig_ObjFindSatAssign(p, a, 1, &support) || p.travID != 2 || len(support) != 1 {
		t.Fatal(p.travID, support)
	}
}
